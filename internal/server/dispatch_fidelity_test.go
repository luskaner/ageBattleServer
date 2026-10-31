package server

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"strings"
	"testing"
	"time"

	"github.com/luskaner/ageBattleServer/internal/protocol"
	"github.com/luskaner/ageBattleServer/internal/session"
)

// The tests in this file pin three rules of the message dispatcher:
//
//   - A relay (0x13) is [toPeerID][fromPeerID][nested message]. The nested
//     message runs through the same dispatcher as a top-level one, as if the
//     sender had sent it. The relay is forwarded to toPeerID only when the
//     nested message is a peer message (0x0F) whose handling asks for it; any
//     other nested type is handled and goes nowhere.
//   - A peer message whose flagBits carry 0x04 has a zlib-compressed inner
//     message. The server inflates it to classify it and to run its effect,
//     and forwards the compressed bytes it received.
//   - A top-level message whose body is not consumed exactly by its layout is
//     dropped without an effect.

// testKey is the session key the RequestJoin tests use: title 24 and a session
// ID that fits in 48 bits, the shape a real client sends.
var testKey = session.SessionKey{TitleID: 24, SessionID: 0x5A69F82C}

// keyedTwoPeerServer is twoPeerServer with a session created under testKey, so a
// RequestJoin can name the sender's own session.
func keyedTwoPeerServer(t *testing.T) (*Server, *session.Session, *mockMsgClient, *mockMsgClient) {
	t.Helper()
	srv := createTestServer(t)
	c1 := &mockMsgClient{}
	c2 := &mockMsgClient{}
	srv.AddClient(c1)
	srv.AddClient(c2)
	sess := srv.sessionManager.GetOrCreateSession(testKey)
	if sess == nil {
		t.Fatal("GetOrCreateSession returned nil")
	}
	for id, c := range map[byte]*mockMsgClient{1: c1, 2: c2} {
		if err := srv.sessionManager.AddPeer(sess, sess.NewPeer(id, c)); err != nil {
			t.Fatalf("AddPeer %d: %v", id, err)
		}
	}
	return srv, sess, c1, c2
}

// peerMessage builds a whole 0x0F message, type byte included: the 15-byte
// header with the given flagBits, then the inner message. The header carries the
// values the game puts in a fresh one, so a test can check it travels unchanged.
func peerMessage(flagBits byte, inner []byte) []byte {
	h := protocol.PeerHeader{
		FlagBits:                      flagBits,
		StartACKReliableNonSequential: 0xFFFFFFFF,
		LastReliableSequentialID:      0xFFFFFFFF,
	}
	msg := append([]byte{protocol.MsgPeer}, h.Serialize()...)
	return append(msg, inner...)
}

// compressedPeerMessage builds a 0x0F message the way the game sends its
// RequestJoin and the first AddPeer of a connection: flagBits 0x04 and the
// inner message zlib-compressed.
func compressedPeerMessage(t *testing.T, inner []byte) []byte {
	t.Helper()
	return peerMessage(protocol.FlagCompressed, zlibBytes(t, inner))
}

func zlibBytes(t *testing.T, b []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	if _, err := zw.Write(b); err != nil {
		t.Fatalf("zlib write: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zlib close: %v", err)
	}
	return buf.Bytes()
}

// relayBody builds the body of a 0x13 (everything after its type byte).
func relayBody(to, from byte, nested []byte) []byte {
	return append([]byte{to, from}, nested...)
}

func requestJoinInner(key session.SessionKey, peerKey uint64) []byte {
	inner := []byte{protocol.PeerMsgRequestJoin}
	inner = append(inner, innerBlob("")...)
	inner = append(inner, innerBlob("custom")...)
	packed := uint64(key.SessionID)&0xFFFFFFFFFFFF | uint64(uint16(key.TitleID))<<48
	inner = binary.LittleEndian.AppendUint64(inner, packed)
	return binary.LittleEndian.AppendUint64(inner, peerKey)
}

func requestJoinDeniedInner(peerKey uint64) []byte {
	inner := []byte{protocol.PeerMsgRequestJoinDenied}
	inner = append(inner, innerBlob("")...)
	inner = binary.LittleEndian.AppendUint32(inner, 3)
	inner = append(inner, innerBlob("lobby full")...)
	return binary.LittleEndian.AppendUint64(inner, peerKey)
}

func removePeerInner(peerID byte) []byte {
	inner := []byte{protocol.PeerMsgRemovePeer}
	inner = binary.LittleEndian.AppendUint32(inner, 7)
	inner = append(inner, innerBlob("left")...)
	return append(inner, peerID)
}

func addPeerInner(peerID byte, isHost bool, globalID uint64) []byte {
	inner := []byte{protocol.PeerMsgAddPeer}
	inner = append(inner, innerBlob("")...)
	inner = append(inner, innerBlob("custom")...)
	host := byte(0)
	if isHost {
		host = 1
	}
	inner = append(inner, peerID, host, 1)
	return binary.LittleEndian.AppendUint64(inner, globalID)
}

// assertOnlyFrame fails unless c received exactly one frame, equal to want.
func assertOnlyFrame(t *testing.T, name string, c *mockMsgClient, want []byte) {
	t.Helper()
	if len(c.sent) != 1 {
		t.Errorf("%s received %d frames, want exactly 1: %x", name, len(c.sent), c.sent)
		return
	}
	if !bytes.Equal(c.sent[0], want) {
		t.Errorf("%s received\n %x\nwant\n %x", name, c.sent[0], want)
	}
}

func assertNothingSent(t *testing.T, name string, c *mockMsgClient) {
	t.Helper()
	if len(c.sent) != 0 {
		t.Errorf("%s must receive nothing, got %d frames: %x", name, len(c.sent), c.sent)
	}
}

// assertGraceClose checks that the peer's connection is scheduled to close 30 s
// after the message that removed it, and is still open now. The deadline is the
// mechanism the session tick closes connections by, so the 30 s are not waited
// out: the tick right after the message must leave the connection open.
func assertGraceClose(t *testing.T, sess *session.Session, id byte, c *mockMsgClient, before, after time.Time) {
	t.Helper()
	at, _, ok := sess.GetPeer(id).CloseDeadline()
	if !ok {
		t.Fatalf("peer %d has no close deadline", id)
	}
	if at.Before(before.Add(30*time.Second)) || at.After(after.Add(30*time.Second)) {
		t.Errorf("peer %d closes at %v, want 30 s after the message (%v .. %v)",
			id, at, before.Add(30*time.Second), after.Add(30*time.Second))
	}
	sess.Tick()
	if c.closed {
		t.Errorf("peer %d was closed before its 30 s grace period ended", id)
	}
}

func assertNoCloseScheduled(t *testing.T, sess *session.Session, id byte) {
	t.Helper()
	if _, _, ok := sess.GetPeer(id).CloseDeadline(); ok {
		t.Errorf("peer %d must not be scheduled to close", id)
	}
}

// A relayed RequestJoinDenied is consumed: the relay is not forwarded, the peer
// message is sent as a 0x0F to every joiner and member of the session (the sender
// and the denied peer included), and the denied peer is closed after 30 s.
func TestRelayedRequestJoinDeniedIsConsumedAndRebroadcast(t *testing.T) {
	srv, sess, a, b := keyedTwoPeerServer(t)
	sess.SetPeerKey(2, 0xB0B)

	nested := peerMessage(0, requestJoinDeniedInner(0xB0B))
	before := time.Now()
	srv.dispatchMessage(a, protocol.MsgLobbyRelay, relayBody(2, 1, nested))
	after := time.Now()

	want := frameMessage(protocol.MsgPeer, nested[1:])
	assertOnlyFrame(t, "the sender", a, want)
	assertOnlyFrame(t, "the denied peer", b, want)
	assertGraceClose(t, sess, 2, b, before, after)
	assertNoCloseScheduled(t, sess, 1)
}

// A relayed RemovePeer is forwarded unchanged to toPeerID, and its target is
// closed after 30 s.
func TestRelayedRemovePeerClosesTargetAfterGrace(t *testing.T) {
	srv, sess, a, b := keyedTwoPeerServer(t)

	relay := relayBody(2, 1, peerMessage(0, removePeerInner(2)))
	before := time.Now()
	srv.dispatchMessage(a, protocol.MsgLobbyRelay, relay)
	after := time.Now()

	assertNothingSent(t, "the sender", a)
	assertOnlyFrame(t, "the addressed peer", b, frameMessage(protocol.MsgLobbyRelay, relay))
	assertGraceClose(t, sess, 2, b, before, after)
	assertNoCloseScheduled(t, sess, 1)
}

// A relay whose nested message is not a peer message is handled as that message
// and forwarded to nobody. Here the nested message is a top-level GameData with
// a tag other than 0xFE, which the server ignores.
func TestRelayedNonPeerMessageReachesNobody(t *testing.T) {
	srv, _, a, b := keyedTwoPeerServer(t)

	srv.dispatchMessage(a, protocol.MsgLobbyRelay, relayBody(2, 1, []byte{0x09, 0xFD, 0x03, 0x00, 0xAA, 0xBB, 0xCC}))

	assertNothingSent(t, "the sender", a)
	assertNothingSent(t, "the addressed peer", b)
}

// A relay with no nested message, or with a peer message too short for its
// header, is forwarded to nobody.
func TestRelayWithoutAValidNestedMessageReachesNobody(t *testing.T) {
	for name, body := range map[string][]byte{
		"only the peer IDs":    {0x02, 0x01},
		"one peer ID":          {0x02},
		"short peer message":   {0x02, 0x01, protocol.MsgPeer, 0x00, 0x00, 0x00},
		"empty peer message":   {0x02, 0x01, protocol.MsgPeer},
		"empty relay":          {},
		"nested unknown type":  {0x02, 0x01, 0x0E, 0x01},
		"nested ping (0x11)":   {0x02, 0x01, protocol.MsgPing, 0x00},
		"nested 0x0F, 15 byte": append([]byte{0x02, 0x01}, peerMessage(0, nil)[:15]...),
	} {
		t.Run(name, func(t *testing.T) {
			srv, _, a, b := keyedTwoPeerServer(t)
			srv.dispatchMessage(a, protocol.MsgLobbyRelay, body)
			assertNothingSent(t, "the sender", a)
			assertNothingSent(t, "the addressed peer", b)
		})
	}
}

// A peer message with a bare header and no inner message is forwarded: there is
// nothing for the server to take.
func TestRelayedBarePeerHeaderIsForwarded(t *testing.T) {
	srv, _, a, b := keyedTwoPeerServer(t)

	relay := relayBody(2, 1, peerMessage(0, nil))
	srv.dispatchMessage(a, protocol.MsgLobbyRelay, relay)

	assertNothingSent(t, "the sender", a)
	assertOnlyFrame(t, "the addressed peer", b, frameMessage(protocol.MsgLobbyRelay, relay))
}

// The forwarding decision belongs to the message, not to the connection: a
// client that reconnected sends its game traffic on a fresh connection, and its
// first self-addressed game relay is still a relay, forwarded unchanged.
func TestReconnectedClientGameRelayStaysARelay(t *testing.T) {
	srv, sess, a, b := keyedTwoPeerServer(t)

	srv.dispatchMessage(a, protocol.MsgLobbyRelay, relayBody(1, 1, peerMessage(0, []byte{protocol.PeerMsgGameData, 0xFD, 0x01, 0x00, 0x11})))
	a.sent, b.sent = nil, nil

	// The player comes back on a new connection that takes over its slot.
	a2 := &mockMsgClient{}
	srv.AddClient(a2)
	if _, ok := sess.ReconnectPeer(1, a2, 0); !ok {
		t.Fatal("ReconnectPeer found no peer 1")
	}
	if err := srv.sessionManager.AddPeer(sess, sess.GetPeer(1)); err != nil {
		t.Fatalf("AddPeer after reconnect: %v", err)
	}

	game := relayBody(1, 1, peerMessage(0, []byte{protocol.PeerMsgGameData, 0xFD, 0x03, 0x00, 0xAA, 0xBB, 0xCC}))
	srv.dispatchMessage(a2, protocol.MsgLobbyRelay, game)

	assertOnlyFrame(t, "the reconnected sender", a2, frameMessage(protocol.MsgLobbyRelay, game))
	assertNothingSent(t, "the other peer", b)
	assertNothingSent(t, "the replaced connection", a)
}

// A relayed RequestJoin naming the sender's own session is consumed and sent as
// a 0x0F to every peer of the session, the sender included. The game sends it
// compressed and self-addressed before it knows its peer ID (to = from = 0), so
// the inner type is only visible after inflating it. The bytes sent out are the
// received ones, still compressed.
func TestRelayedRequestJoinIsConsumedAndRebroadcast(t *testing.T) {
	srv, sess, host, joiner := keyedTwoPeerServer(t)

	nested := compressedPeerMessage(t, requestJoinInner(testKey, 0x10E))
	srv.dispatchMessage(joiner, protocol.MsgLobbyRelay, relayBody(0, 0, nested))

	want := frameMessage(protocol.MsgPeer, nested[1:])
	assertOnlyFrame(t, "the host", host, want)
	assertOnlyFrame(t, "the joiner", joiner, want)
	if got := sess.GetPeer(2).Key; got != 0x10E {
		t.Errorf("joiner key = %#x, want the RequestJoin's peer identifier 0x10e", got)
	}
}

// A connection's first RequestJoin is what decides its session, so it is not
// compared with the session the connection's LobbyConnect named: one naming
// another (titleID, sessionID) is still consumed and re-broadcast to the
// sender's session.
func TestRequestJoinIsNotComparedWithTheConnectSession(t *testing.T) {
	srv, sess, a, b := keyedTwoPeerServer(t)

	other := session.SessionKey{TitleID: -1, SessionID: testKey.SessionID + 1}
	nested := peerMessage(0, requestJoinInner(other, 0x10E))
	srv.dispatchMessage(a, protocol.MsgLobbyRelay, relayBody(2, 1, nested))

	want := frameMessage(protocol.MsgPeer, nested[1:])
	assertOnlyFrame(t, "the sender", a, want)
	assertOnlyFrame(t, "the other peer", b, want)
	if got := sess.GetPeer(1).Key; got != 0x10E {
		t.Errorf("sender key = %#x, want the RequestJoin's peer identifier 0x10e", got)
	}
}

// A relayed GameData tagged 0xFE is queued for the next 0x0A and not forwarded.
func TestRelayedGameDataTaggedFEIsQueuedNotForwarded(t *testing.T) {
	srv, sess, a, b := keyedTwoPeerServer(t)

	srv.dispatchMessage(a, protocol.MsgLobbyRelay, relayBody(2, 1, peerMessage(0, []byte{protocol.PeerMsgGameData, 0xFE, 0x02, 0x00, 0xAA, 0xBB})))

	assertNothingSent(t, "the sender", a)
	assertNothingSent(t, "the addressed peer", b)
	body, _ := sess.PopAckTick()
	if !bytes.Contains(body, []byte{0x01, 0x02, 0x00, 0xAA, 0xBB}) {
		t.Errorf("next 0x0A = %x, want it to carry peer 1's chunk 01 0200 aabb", body)
	}
}

// A compressed AddPeer (the game compresses the first one of each connection) is
// forwarded unchanged, and its effect is applied from the inflated bytes.
func TestRelayedCompressedAddPeerTakesEffect(t *testing.T) {
	srv, sess, a, b := keyedTwoPeerServer(t)

	relay := relayBody(2, 1, compressedPeerMessage(t, addPeerInner(2, true, 0xC0FFEE)))
	srv.dispatchMessage(a, protocol.MsgLobbyRelay, relay)

	assertNothingSent(t, "the sender", a)
	assertOnlyFrame(t, "the addressed peer", b, frameMessage(protocol.MsgLobbyRelay, relay))
	if got := sess.HostID(); got != 2 {
		t.Errorf("HostID = %d, want 2 from the inflated AddPeer", got)
	}
	if got := sess.GetPeer(2).Key; got != 0xC0FFEE {
		t.Errorf("peer 2 key = %#x, want the AddPeer's globalID", got)
	}
}

// The re-broadcast is a property of the peer message, not of the relay: a
// top-level 0x0F RequestJoinDenied is sent to the session the same way.
func TestPeerRequestJoinDeniedIsRebroadcast(t *testing.T) {
	srv, sess, a, b := keyedTwoPeerServer(t)
	sess.SetPeerKey(2, 0xB0B)

	msg := peerMessage(0, requestJoinDeniedInner(0xB0B))
	before := time.Now()
	srv.dispatchMessage(a, protocol.MsgPeer, msg[1:])
	after := time.Now()

	want := frameMessage(protocol.MsgPeer, msg[1:])
	assertOnlyFrame(t, "the sender", a, want)
	assertOnlyFrame(t, "the denied peer", b, want)
	assertGraceClose(t, sess, 2, b, before, after)
}

// A compressed RemovePeer has the same effect as the uncompressed one, and the
// relay forwards the bytes it received.
func TestCompressedRemovePeerMatchesUncompressedControl(t *testing.T) {
	for name, compressed := range map[string]bool{"uncompressed control": false, "compressed": true} {
		t.Run(name, func(t *testing.T) {
			srv, sess, a, b := keyedTwoPeerServer(t)
			nested := peerMessage(0, removePeerInner(2))
			if compressed {
				nested = compressedPeerMessage(t, removePeerInner(2))
			}
			relay := relayBody(2, 1, nested)

			before := time.Now()
			srv.dispatchMessage(a, protocol.MsgLobbyRelay, relay)
			after := time.Now()

			assertOnlyFrame(t, "the addressed peer", b, frameMessage(protocol.MsgLobbyRelay, relay))
			assertGraceClose(t, sess, 2, b, before, after)
		})
	}
}

// The same holds on the top-level 0x0F path, which forwards nothing.
func TestCompressedPeerRemovePeerSchedulesClose(t *testing.T) {
	srv, sess, a, b := keyedTwoPeerServer(t)

	msg := compressedPeerMessage(t, removePeerInner(2))
	before := time.Now()
	srv.dispatchMessage(a, protocol.MsgPeer, msg[1:])
	after := time.Now()

	assertNothingSent(t, "the sender", a)
	assertNothingSent(t, "the removed peer", b)
	assertGraceClose(t, sess, 2, b, before, after)
}

// A compressed peer message that does not inflate has no type the server can
// read. It must not panic; it has no effect and is forwarded unchanged, as any
// unhandled inner type is.
func TestUndecodableCompressedPeerMessageIsForwardedUnchanged(t *testing.T) {
	for name, payload := range map[string][]byte{
		"not zlib":          {0x01, 0x02, 0x03, 0x04},
		"empty":             nil,
		"zlib header only":  zlibBytes(t, removePeerInner(2))[:2],
		"garbage after hdr": {0x78, 0x9C, 0xFF, 0xFF, 0xFF, 0xFF},
	} {
		t.Run(name, func(t *testing.T) {
			srv, sess, a, b := keyedTwoPeerServer(t)
			relay := relayBody(2, 1, peerMessage(protocol.FlagCompressed, payload))
			srv.dispatchMessage(a, protocol.MsgLobbyRelay, relay)

			assertNothingSent(t, "the sender", a)
			assertOnlyFrame(t, "the addressed peer", b, frameMessage(protocol.MsgLobbyRelay, relay))
			assertNoCloseScheduled(t, sess, 2)
		})
	}
}

// An inflate error does not discard what was inflated before it. A stream cut
// short yields the start of its message, which is classified by its first byte:
// here the start of a RemovePeer, whose layout is then incomplete, so it keeps
// the RemovePeer's verdict (forwarded) and has no effect. A stream whose
// checksum is wrong yields the whole message, which takes effect.
func TestPartiallyInflatedPeerMessageIsClassifiedByWhatInflated(t *testing.T) {
	removal := []byte{protocol.PeerMsgRemovePeer, 7, 0, 0, 0}
	removal = append(removal, innerBlob(strings.Repeat("connection lost; ", 40))...)
	removal = append(removal, 2)
	full := zlibBytes(t, removal)
	cut := full[:len(full)/2]
	if got := protocol.InflatePeerPayload(cut); len(got) == 0 || len(got) >= len(removal) || got[0] != protocol.PeerMsgRemovePeer {
		t.Fatalf("precondition: the cut stream inflates to %d bytes, want a strict prefix of the %d-byte RemovePeer", len(got), len(removal))
	}
	badSum := append([]byte(nil), full...)
	badSum[len(badSum)-1] ^= 0xFF

	for _, tc := range []struct {
		name      string
		payload   []byte
		scheduled bool
	}{{"truncated stream", cut, false}, {"corrupt checksum", badSum, true}} {
		t.Run(tc.name, func(t *testing.T) {
			srv, sess, a, b := keyedTwoPeerServer(t)
			relay := relayBody(2, 1, peerMessage(protocol.FlagCompressed, tc.payload))
			srv.dispatchMessage(a, protocol.MsgLobbyRelay, relay)

			assertNothingSent(t, "the sender", a)
			assertOnlyFrame(t, "the addressed peer", b, frameMessage(protocol.MsgLobbyRelay, relay))
			if _, _, ok := sess.GetPeer(2).CloseDeadline(); ok != tc.scheduled {
				t.Errorf("peer 2 scheduled to close = %v, want %v", ok, tc.scheduled)
			}
		})
	}
}

// uploadAck runs one 0x0A tick of sess and returns its body.
func uploadAck(t *testing.T, sess *session.Session) []byte {
	t.Helper()
	body, _ := sess.PopAckTick()
	if body == nil {
		t.Fatal("no 0x0A was built")
	}
	return body
}

// A 0x09 upload whose declared length is not its actual length is dropped: the
// next 0x0A carries only the valid upload of peer 1, and peer 2 stays pending.
func TestMalformedGameDataUploadIsDropped(t *testing.T) {
	for name, upload := range map[string][]byte{
		"declared longer":  {0xFE, 0x05, 0x00, 0xAA, 0xBB, 0xCC, 0xDD},
		"declared shorter": {0xFE, 0x03, 0x00, 0xAA, 0xBB, 0xCC, 0xDD},
		"no length":        {0xFE, 0x05},
	} {
		t.Run(name, func(t *testing.T) {
			srv, sess, a, b := keyedTwoPeerServer(t)
			srv.dispatchMessage(a, protocol.MsgGameData, []byte{0xFE, 0x01, 0x00, 0x11})
			srv.dispatchMessage(b, protocol.MsgGameData, upload)

			want := []byte{0xFE, 0x01, 0x01, 0x01, 0x00, 0x11, 0x01, 0x02, 0x00, 0x00}
			if got := uploadAck(t, sess); !bytes.Equal(got, want) {
				t.Errorf("0x0A = %x, want %x (peer 1's chunk, peer 2 pending)", got, want)
			}
			if got := sess.SyncMessagesAccepted(2); got != 0 {
				t.Errorf("peer 2 accepted uploads = %d, want 0", got)
			}
		})
	}
}

// The same rule holds for the upload a 0x17 carries after its checksum entries:
// it is a nested 0x09 and must be consumed exactly as well.
func TestMalformedChecksumUploadIsDropped(t *testing.T) {
	srv, sess, a, b := keyedTwoPeerServer(t)
	srv.dispatchMessage(a, protocol.MsgGameData, []byte{0xFE, 0x01, 0x00, 0x11})
	// No checksum entries, then a 0x09 that declares 5 bytes and carries 4.
	srv.dispatchMessage(b, protocol.MsgGameDataChecksum, []byte{0x00, 0x09, 0xFE, 0x05, 0x00, 0xAA, 0xBB, 0xCC, 0xDD})

	want := []byte{0xFE, 0x01, 0x01, 0x01, 0x00, 0x11, 0x01, 0x02, 0x00, 0x00}
	if got := uploadAck(t, sess); !bytes.Equal(got, want) {
		t.Errorf("0x0A = %x, want %x (peer 1's chunk, peer 2 pending)", got, want)
	}
}

// The upload a 0x17 carries is dispatched like a top-level 0x09: a valid one is
// queued as before.
func TestChecksumUploadIsQueued(t *testing.T) {
	srv, sess, a, _ := keyedTwoPeerServer(t)
	srv.dispatchMessage(a, protocol.MsgGameDataChecksum, []byte{0x00, 0x09, 0xFE, 0x02, 0x00, 0xAA, 0xBB})

	if got := uploadAck(t, sess); !bytes.Contains(got, []byte{0x01, 0x02, 0x00, 0xAA, 0xBB}) {
		t.Errorf("0x0A = %x, want it to carry peer 1's chunk 01 0200 aabb", got)
	}
}

// The StartObserveRequest layout is chosen by the type byte: 0x18 is 8 bytes and
// 0x1C is 24. A body of the other length is malformed and gets no reply.
func TestStartObserveLayoutFollowsTheTypeByte(t *testing.T) {
	newObserved := func(t *testing.T) (*Server, *mockMsgClient) {
		srv := observerTestServer(t)
		client := &mockMsgClient{}
		// The session is keyed under title -1, which is the title a 0x18
		// request carries, so both layouts would find it.
		sess := srv.sessionManager.GetOrCreateSession(session.SessionKey{TitleID: -1, SessionID: 1})
		sess.SetObserverConfig(int(srv.config.ObserverMax), srv.observerUpdateIntervalTicks())
		if err := srv.sessionManager.AddPeer(sess, sess.NewPeer(1, client)); err != nil {
			t.Fatalf("AddPeer: %v", err)
		}
		sess.RecordFrame(protocol.MsgEndOfGameData, []byte{0xFE, 0x00, 0x00})
		return srv, client
	}
	legacy := make([]byte, 8)
	binary.LittleEndian.PutUint64(legacy, 1)
	versioned := make([]byte, 24)
	binary.LittleEndian.PutUint64(versioned, 1|uint64(0xFFFF)<<48)

	for _, tc := range []struct {
		name    string
		typ     byte
		body    []byte
		replied bool
	}{
		{"0x18 with 8 bytes", protocol.MsgStartObserve, legacy, true},
		{"0x1C with 24 bytes", 0x1C, versioned, true},
		{"0x18 with 24 bytes", protocol.MsgStartObserve, versioned, false},
		{"0x1C with 8 bytes", 0x1C, legacy, false},
		{"0x18 with a trailing byte", protocol.MsgStartObserve, append(append([]byte(nil), legacy...), 0), false},
		{"0x1C with a trailing byte", 0x1C, append(append([]byte(nil), versioned...), 0), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, client := newObserved(t)
			srv.dispatchMessage(client, tc.typ, tc.body)
			if got := len(client.sent) != 0; got != tc.replied {
				t.Errorf("replied = %v, want %v: %x", got, tc.replied, client.sent)
			}
		})
	}
}

// The activate, stop and leave messages are a bare type byte. One with a trailing
// byte does not parse and has no effect.
func TestEmptyMessagesWithTrailingBytesAreIgnored(t *testing.T) {
	t.Run("ActivateObserver", func(t *testing.T) {
		srv := observerTestServer(t)
		client := &mockMsgClient{}
		testSessionWithPeer(t, srv, client)
		sess := srv.sessionManager.GetSession(session.SessionKey{SessionID: 1})
		sess.RecordFrame(protocol.MsgEndOfGameData, []byte{0xFE, 0x00, 0x00})
		sess.RecordFrame(protocol.MsgEndOfGameData, []byte{0xFE, 0x00, 0x00})
		start := make([]byte, 24)
		binary.LittleEndian.PutUint64(start[:8], 1)
		srv.dispatchMessage(client, 0x1C, start)
		client.sent = nil

		srv.dispatchMessage(client, protocol.MsgActivateObserver, []byte{0x00})
		assertNothingSent(t, "the observer", client)

		srv.dispatchMessage(client, protocol.MsgActivateObserver, nil)
		if len(client.sent) != 1 || client.sent[0][2] != protocol.MsgActivateObserverResp {
			t.Errorf("the empty ActivateObserver must be answered, got %x", client.sent)
		}
	})
	t.Run("StopObserve", func(t *testing.T) {
		srv := observerTestServer(t)
		client := &mockMsgClient{observer: true}
		srv.dispatchMessage(client, protocol.MsgStopObserve, []byte{0x00})
		if !client.observer {
			t.Error("a StopObserve with a trailing byte must leave the observer flag set")
		}
		srv.dispatchMessage(client, protocol.MsgStopObserve, nil)
		if client.observer {
			t.Error("the empty StopObserve must clear the observer flag")
		}
	})
	t.Run("LeaveSimulation", func(t *testing.T) {
		srv := createTestServer(t)
		client := &mockMsgClient{}
		sess := testSessionWithPeer(t, srv, client)
		sess.GetPeer(1).Participating = true

		srv.dispatchMessage(client, protocol.MsgLeaveSimulation, []byte{0x00})
		if !sess.GetPeer(1).Participating {
			t.Error("a LeaveSimulation with a trailing byte must have no effect")
		}
		srv.dispatchMessage(client, protocol.MsgLeaveSimulation, nil)
		if sess.GetPeer(1).Participating {
			t.Error("the empty LeaveSimulation must clear the participating mark")
		}
	})
}

// SetObservableParameters is exactly u32, u64, u32: 16 bytes, no more, no less.
func TestSetObservableParametersNeedsExactlySixteenBytes(t *testing.T) {
	body := func(n int) []byte {
		b := make([]byte, n)
		binary.LittleEndian.PutUint32(b[0:4], 1)
		binary.LittleEndian.PutUint64(b[4:12], 0xFACE)
		return b
	}
	for _, tc := range []struct {
		n       int
		applied bool
	}{{15, false}, {17, false}, {16, true}} {
		srv := createTestServer(t)
		client := &mockMsgClient{}
		sess := testSessionWithPeer(t, srv, client)
		srv.dispatchMessage(client, protocol.MsgSetObservableParameters, body(tc.n))
		if got := sess.ObservableHash() == 0xFACE; got != tc.applied {
			t.Errorf("%d bytes: applied = %v, want %v", tc.n, got, tc.applied)
		}
	}
}

// Nesting is bounded by what a message the original accepts can hold: at most
// 65535 bytes, and every nesting level costs at least two bytes, so no such
// message nests deeper than 65535/2 levels. A message nested deeper is dropped
// instead of growing the stack without bound; one nested to the limit is still
// handled. The innermost message here is a LobbyPing, which a session member
// gets echoed.
func TestDeeplyNestedRelayIsDropped(t *testing.T) {
	const limit = 65535 / 2
	ping := []byte{protocol.MsgLobbyPing, 1, 2, 3, 4, 5, 6, 7, 8}
	nest := func(levels int) []byte {
		msg := make([]byte, 0, 3*levels+len(ping))
		for range levels {
			msg = append(msg, protocol.MsgLobbyRelay, 0x02, 0x01)
		}
		return append(msg, ping...)
	}
	srv, _, a, b := keyedTwoPeerServer(t)

	atLimit := nest(limit)
	srv.dispatchMessage(a, atLimit[0], atLimit[1:])
	assertOnlyFrame(t, "the sender", a, frameMessage(protocol.MsgLobbyPing, ping[1:]))
	assertNothingSent(t, "the other peer", b)
	a.sent = nil

	tooDeep := nest(limit + 1)
	srv.dispatchMessage(a, tooDeep[0], tooDeep[1:])
	assertNothingSent(t, "the sender", a)
	assertNothingSent(t, "the other peer", b)
}
