package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"slices"
	"testing"

	"github.com/luskaner/ageBattleServer/internal/protocol"
	"github.com/luskaner/ageBattleServer/internal/session"
)

// WHAT THIS FILE IS
//
// The session lifecycle seen from the wire: what a reconnect answers and in which
// order, what a closed connection takes out of its session, and how long a
// session lives once its players are gone. Every connect here is the one a 2026
// client sends (message version 5, no timeout override), built by
// protocol.BuildLobbyConnect, so each body has the exact length the parser
// accepts.

const (
	lcSecretA = 0x0A0A_0A0A_1111_1111
	lcSecretB = 0x0B0B_0B0B_2222_2222
)

// lcConnect is a version-5 LobbyConnect for session sessionID. The count is only
// written for state 2, the one state whose message carries it.
func lcConnect(sessionID int64, secret uint64, state byte, count int32) []byte {
	return protocol.BuildLobbyConnect(0, sessionID, secret, state, count, 5, 0, 0)
}

// lcObserve is the 24-byte StartObserveRequest (0x1C) for session sessionID.
func lcObserve(sessionID, peerIdentifier, passwordHash uint64) []byte {
	body := make([]byte, 24)
	binary.LittleEndian.PutUint64(body[:8], sessionID) // titleID 0 in the top 16 bits
	binary.LittleEndian.PutUint64(body[8:16], peerIdentifier)
	binary.LittleEndian.PutUint64(body[16:24], passwordHash)
	return body
}

// lcPing is a LobbyPing body as the client sends it: its clock in ms, a u64.
var lcPing = []byte{0x10, 0x20, 0x30, 0x40, 0x00, 0x00, 0x00, 0x00}

// lcBodies returns the bodies of the messages of type typ sent to c, in order.
// The mock frames every message as [u16 BE length][type][body].
func lcBodies(c *mockMsgClient, typ byte) [][]byte {
	var out [][]byte
	for _, f := range c.sent {
		if len(f) >= 3 && f[2] == typ {
			out = append(out, f[3:])
		}
	}
	return out
}

// lcTypes returns the type of every message sent to c, in order.
func lcTypes(c *mockMsgClient) []byte {
	out := make([]byte, 0, len(c.sent))
	for _, f := range c.sent {
		if len(f) >= 3 {
			out = append(out, f[2])
		}
	}
	return out
}

// lcAckTail splits a 0x0A body, [FE][subCount] subs [K] pending [block], into
// its pending list and its checksum block.
func lcAckTail(t *testing.T, body []byte) (pending, block []byte) {
	t.Helper()
	if len(body) < 2 || body[0] != 0xFE {
		t.Fatalf("not a 0x0A body: %x", body)
	}
	off := 2
	for range int(body[1]) {
		if off+3 > len(body) {
			t.Fatalf("the 0x0A body ends inside a sub-message: %x", body)
		}
		off += 3 + int(binary.LittleEndian.Uint16(body[off+1:off+3]))
	}
	if off >= len(body) {
		t.Fatalf("the 0x0A body has no pending list: %x", body)
	}
	k := int(body[off])
	off++
	if off+k > len(body) {
		t.Fatalf("the 0x0A pending list overruns the body: %x", body)
	}
	return body[off : off+k], body[off+k:]
}

// lcTick runs one simulation tick in the main loop's order: the session ack
// first, then the session update that serves the observers.
func lcTick(srv *Server) {
	srv.sendEndOfGameDataTicks()
	srv.tickSessions()
}

// lcGameRelay is a 0x13 from peer `from` to peer `to` carrying a 0x0F game
// message, the shape the committed captures relay all game traffic in: a 15-byte
// reliability header and an inner message the server does not act on, so the
// relay forwards it to its target.
func lcGameRelay(to, from byte) []byte {
	b := []byte{to, from, protocol.MsgPeer}
	b = append(b, protocol.PeerHeader{}.Serialize()...)
	return append(b, 0x0D, 0xC1, 0xC2)
}

// C-1. A reconnect moves only the socket: the peer keeps its slot and its place
// in the session, so everything the new connection sends is that peer's traffic.
// The session lookup by connection has to follow the move. Otherwise the new
// connection is a stranger to the session: its relays and uploads are dropped,
// and its pings fall back to the budget of a connection that has no session.
func TestAReconnectedPlayerKeepsItsSessionTraffic(t *testing.T) {
	srv := createTestServer(t)
	a, b, aAgain := &mockMsgClient{}, &mockMsgClient{}, &mockMsgClient{}
	for _, c := range []*mockMsgClient{a, b, aAgain} {
		srv.AddClient(c)
	}
	srv.dispatchMessage(a, protocol.MsgLobbyConnect, lcConnect(81, lcSecretA, 0, 0))
	srv.dispatchMessage(b, protocol.MsgLobbyConnect, lcConnect(81, lcSecretB, 0, 0))
	sess := srv.sessionManager.GetSession(session.SessionKey{SessionID: 81})
	if sess == nil {
		t.Fatal("the first connect should create session 81")
	}
	slotA, slotB := sess.PeerIDForConn(a), sess.PeerIDForConn(b)

	// The client comes back on a new connection with its secret and no frame yet.
	srv.dispatchMessage(aAgain, protocol.MsgLobbyConnect, lcConnect(81, lcSecretA, 2, 0))
	if got := srv.sessionManager.FindSessionByPeer(aAgain); got != sess {
		t.Fatalf("the reconnected connection resolves to session %p, want the peer's session %p", got, sess)
	}
	if got := srv.sessionManager.FindSessionByPeer(a); got != nil {
		t.Errorf("the replaced connection still resolves to session %p, want none", got)
	}

	relay := lcGameRelay(slotB, slotA)
	srv.dispatchMessage(aAgain, protocol.MsgLobbyRelay, relay)
	if got := lcBodies(b, protocol.MsgLobbyRelay); len(got) != 1 || !bytes.Equal(got[0], relay) {
		t.Errorf("B received the relays %x, want the reconnected player's %x", got, relay)
	}

	srv.dispatchMessage(aAgain, protocol.MsgGameData, []byte{0xFE, 0x02, 0x00, 0x5A, 0x5B})
	srv.sendEndOfGameDataTicks()
	acks := lcBodies(b, protocol.MsgEndOfGameData)
	if len(acks) == 0 {
		t.Fatal("B received no 0x0A: the reconnected player's upload never reached the session")
	}
	if sub := []byte{slotA, 0x02, 0x00, 0x5A, 0x5B}; !bytes.Contains(acks[len(acks)-1], sub) {
		t.Errorf("B's 0x0A %x does not carry the reconnected player's chunk %x", acks[len(acks)-1], sub)
	}

	// A connection in a session has every ping echoed; one without stops at the
	// budget.
	aAgain.sent = nil
	for range lobbyPingEchoBudget + 1 {
		srv.dispatchMessage(aAgain, protocol.MsgLobbyPing, lcPing)
	}
	if got := len(lcBodies(aAgain, protocol.MsgLobbyPing)); got != lobbyPingEchoBudget+1 {
		t.Errorf("the reconnected connection got %d ping echoes, want %d: it is in a session",
			got, lobbyPingEchoBudget+1)
	}
}

// C-2/D-4. A closed connection takes its peer out of the session: the slot, the
// game-data queue record and the checksum vector go with it. The reconciliation
// waits for an entry from every active peer, so a departed peer that stayed a
// member with an empty vector stalled it for the rest of the match. It also
// stayed on the pending list until its sync timeout.
func TestAClosedConnectionLeavesTheChecksumReconciliation(t *testing.T) {
	srv := createTestServer(t)
	players := []*mockMsgClient{{}, {}, {}}
	for i, c := range players {
		srv.AddClient(c)
		srv.dispatchMessage(c, protocol.MsgLobbyConnect, lcConnect(82, lcSecretA+uint64(i), 0, 0))
	}
	sess := srv.sessionManager.GetSession(session.SessionKey{SessionID: 82})
	if sess == nil || sess.PeerCount() != 3 {
		t.Fatal("three connects should give session 82 three peers")
	}

	// The game starts: the first upload creates the frame store, and every peer
	// present is in the started game from then on.
	srv.dispatchMessage(players[0], protocol.MsgGameData, []byte{0xFE, 0x01, 0x00, 0x77})
	srv.sendEndOfGameDataTicks()
	if !sess.Recording {
		t.Fatal("the first ack should have created the frame store")
	}

	// Player 3's connection closes; the transport reports that through RemoveClient.
	srv.RemoveClient(players[2])
	if got := sess.PeerCount(); got != 2 {
		t.Errorf("PeerCount() = %d after a close, want 2: the closed connection's peer leaves", got)
	}

	entry := []byte{0x44, 0x33, 0x22, 0x11, 0x07, 0x00, 0x00, 0x00}
	for _, c := range players[:2] {
		srv.dispatchMessage(c, protocol.MsgGameDataChecksum, append([]byte{0x01}, entry...))
	}
	players[0].sent = nil
	// The merged entry rides on the tick after the one that merges it: the
	// reconciliation runs once the 0x0A is built.
	srv.sendEndOfGameDataTicks()
	srv.sendEndOfGameDataTicks()
	acks := lcBodies(players[0], protocol.MsgEndOfGameData)
	if len(acks) != 2 {
		t.Fatalf("player 1 received %d 0x0A over two ticks, want 2", len(acks))
	}
	pending, block := lcAckTail(t, acks[1])
	if got := slices.Sorted(slices.Values(pending)); !bytes.Equal(got, []byte{1, 2}) {
		t.Errorf("pending peers = %v, want {1, 2}: the closed connection's peer is gone", pending)
	}
	if !bytes.Contains(block, entry) {
		t.Errorf("the checksum block %x does not carry the entry both remaining players sent (%x)", block, entry)
	}
}

// C-3. The session outlives its players while anybody still watches it. With the
// last player gone the observers are served the live end of the store, the lag
// included, and then the end-of-game marker. The session itself goes when its
// last observer does.
func TestObserversOutliveThePlayers(t *testing.T) {
	srv := observerTestServer(t)
	const (
		lagSeconds = 1
		lagPages   = lagSeconds * 8 // the default 125 ms period is 8 ticks a second
		frames     = lagPages + 12
		hash       = 0x21616C6661AA7A42
		sessionID  = 83
	)
	player, watcher := &mockMsgClient{}, &mockMsgClient{}
	srv.AddClient(player)
	srv.AddClient(watcher)
	srv.dispatchMessage(player, protocol.MsgLobbyConnect, lcConnect(sessionID, lcSecretA, 0, 0))
	key := session.SessionKey{SessionID: sessionID}
	sess := srv.sessionManager.GetSession(key)
	if sess == nil {
		t.Fatal("the player's connect should create the session")
	}
	// Observable, a published credential and a lag of one second.
	params := make([]byte, 16)
	binary.LittleEndian.PutUint32(params[0:4], 1)
	binary.LittleEndian.PutUint64(params[4:12], hash)
	binary.LittleEndian.PutUint32(params[12:16], lagSeconds)
	srv.dispatchMessage(player, protocol.MsgSetObservableParameters, params)

	// The observer announces version 5 and presents no credential, so it is lagged.
	srv.dispatchMessage(watcher, protocol.MsgLobbyConnect, lcConnect(sessionID, lcSecretB, 0, 0))
	srv.dispatchMessage(watcher, 0x1C, lcObserve(sessionID, 0x5157, 0))
	srv.dispatchMessage(watcher, protocol.MsgActivateObserver, nil)
	if !sess.HasObserver(watcher) {
		t.Fatal("the observer should be registered")
	}

	for i := range frames {
		sess.RecordFrame(protocol.MsgEndOfGameData, []byte{0xFE, byte(i)})
	}
	watcher.sent = nil
	lcTick(srv)
	if got := len(lcBodies(watcher, protocol.MsgEndOfGameData)); got != frames-lagPages {
		t.Fatalf("the lagged observer got %d frames while the player plays, want %d", got, frames-lagPages)
	}

	srv.RemoveClient(player)
	if srv.sessionManager.GetSession(key) != sess {
		t.Fatal("the session was erased while an observer still watches it")
	}

	// Two whole strides of the observer step: the first serves the rest, the
	// second would show a repeated marker.
	for range 2 * (srv.observerUpdateIntervalTicks() + 1) {
		lcTick(srv)
	}
	got := lcBodies(watcher, protocol.MsgEndOfGameData)
	if len(got) != frames {
		t.Fatalf("the observer got %d frames in all, want the whole store's %d", len(got), frames)
	}
	if last := got[len(got)-1]; !bytes.Equal(last, []byte{0xFE, frames - 1}) {
		t.Errorf("the observer's last frame is %x, want the store's last %x", last, []byte{0xFE, frames - 1})
	}
	if n := len(lcBodies(watcher, protocol.MsgEndOfGameDataObserver)); n != 1 {
		t.Fatalf("the observer got %d end-of-game markers, want exactly 1", n)
	}
	if types := lcTypes(watcher); types[len(types)-1] != protocol.MsgEndOfGameDataObserver {
		t.Errorf("the observer's messages end with type 0x%02x, want the marker after the last frame",
			types[len(types)-1])
	}

	srv.RemoveClient(watcher)
	if n := srv.sessionManager.SessionCount(); n != 0 {
		t.Errorf("SessionCount() = %d after the last observer left, want 0", n)
	}
}

// C-5. A state-2 reconnect answers with its connect first and only then resends
// the frames from the client's count.
func TestAStateTwoReconnectRepliesBeforeItResends(t *testing.T) {
	srv := createTestServer(t)
	first, second := &mockMsgClient{}, &mockMsgClient{}
	srv.dispatchMessage(first, protocol.MsgLobbyConnect, lcConnect(84, lcSecretA, 0, 0))
	sess := srv.sessionManager.GetSession(session.SessionKey{SessionID: 84})
	for i := 1; i <= 3; i++ {
		sess.RecordFrame(protocol.MsgEndOfGameData, []byte{0xFE, byte(i)})
	}

	srv.dispatchMessage(second, protocol.MsgLobbyConnect, lcConnect(84, lcSecretA, 2, 1))

	want := []byte{protocol.MsgLobbyConnect, protocol.MsgEndOfGameData, protocol.MsgEndOfGameData}
	if got := lcTypes(second); !bytes.Equal(got, want) {
		t.Fatalf("the reconnect sent the types %x, want %x: the reply, then frames 1 and 2", got, want)
	}
	frames := lcBodies(second, protocol.MsgEndOfGameData)
	if !bytes.Equal(frames[0], []byte{0xFE, 2}) || !bytes.Equal(frames[1], []byte{0xFE, 3}) {
		t.Errorf("the resent frames are %x, want frames 1 and 2 (fe02, fe03)", frames)
	}
	if second.closed {
		t.Error("a served reconnect must keep its connection")
	}
}

// C-6. A resend range the store cannot serve gets no reply and the connection is
// closed. Every state above 2 ends the same way: its message carries no count,
// the parser leaves -1 there, and a range that starts below zero is invalid.
// The socket has already moved to the peer by then, so the connection it
// replaced is closed as well.
func TestAnUnservableReconnectIsClosedWithoutAReply(t *testing.T) {
	for _, tc := range []struct {
		name     string
		state    byte
		count    int32
		frames   int
		pageSize uint
	}{
		{name: "the count is past the last frame", state: 2, count: 5, frames: 3},
		{name: "the client already holds every frame", state: 2, count: 3, frames: 3},
		{name: "the first frame asked for is no longer stored", state: 2, count: 0, frames: 10, pageSize: 64},
		{name: "state 3 carries no count", state: 3, frames: 3},
		{name: "state 3 against an empty store", state: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, err := New(&Config{LogLevel: "none", PageSize: tc.pageSize})
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			first, second := &mockMsgClient{}, &mockMsgClient{}
			srv.dispatchMessage(first, protocol.MsgLobbyConnect, lcConnect(86, lcSecretA, 0, 0))
			sess := srv.sessionManager.GetSession(session.SessionKey{SessionID: 86})
			for i := range tc.frames {
				sess.RecordFrame(protocol.MsgEndOfGameData, []byte{0xFE, byte(i)})
			}
			if tc.pageSize != 0 && sess.ReplayPageCount() >= tc.frames {
				t.Fatalf("the store still holds all %d frames: the case needs one that dropped its oldest", tc.frames)
			}

			srv.dispatchMessage(second, protocol.MsgLobbyConnect, lcConnect(86, lcSecretA, tc.state, tc.count))

			if len(second.sent) != 0 {
				t.Errorf("the reconnect was answered with the types %x, want nothing", lcTypes(second))
			}
			if !second.closed {
				t.Error("the reconnecting connection must be closed")
			}
			if !first.closed {
				t.Error("the connection the reconnect replaced must be closed")
			}
		})
	}
}

// C-6, the one exception: a count of 0 against an empty store is answered with the
// reply alone, and the connection stays.
func TestAZeroCountReconnectOnAnEmptyStoreOnlyReplies(t *testing.T) {
	srv := createTestServer(t)
	first, second := &mockMsgClient{}, &mockMsgClient{}
	srv.dispatchMessage(first, protocol.MsgLobbyConnect, lcConnect(87, lcSecretA, 0, 0))

	srv.dispatchMessage(second, protocol.MsgLobbyConnect, lcConnect(87, lcSecretA, 2, 0))

	if got := lcTypes(second); !bytes.Equal(got, []byte{protocol.MsgLobbyConnect}) {
		t.Errorf("the reconnect sent the types %x, want the reply alone", got)
	}
	if second.closed {
		t.Error("the reconnecting connection must stay open")
	}
	if !first.closed {
		t.Error("the connection the reconnect replaced must be closed")
	}
}

// C-7. A reconnect that finds no session is not a connect: nothing is stored on
// the connection, nothing is echoed, no session is created, and the connection
// is closed.
func TestAReconnectToAnUnknownSessionIsClosed(t *testing.T) {
	for _, state := range []byte{1, 2} {
		t.Run(fmt.Sprintf("state %d", state), func(t *testing.T) {
			srv := createTestServer(t)
			c := &mockMsgClient{}
			srv.dispatchMessage(c, protocol.MsgLobbyConnect, lcConnect(88, lcSecretA, state, 0))

			if len(c.sent) != 0 {
				t.Errorf("the reconnect was answered with the types %x, want nothing", lcTypes(c))
			}
			if !c.closed {
				t.Error("the connection must be closed")
			}
			if n := srv.sessionManager.SessionCount(); n != 0 {
				t.Errorf("SessionCount() = %d, want 0: a reconnect does not create a session", n)
			}
			if c.msgVersion != 0 {
				t.Errorf("the connection carries version %d, want 0: nothing is stored", c.msgVersion)
			}
		})
	}
}

// C-7, the other half: the session exists but no peer in it holds the secret.
func TestAReconnectWithAnUnknownSecretIsClosed(t *testing.T) {
	srv := createTestServer(t)
	first, stranger := &mockMsgClient{}, &mockMsgClient{}
	srv.dispatchMessage(first, protocol.MsgLobbyConnect, lcConnect(89, lcSecretA, 0, 0))
	sess := srv.sessionManager.GetSession(session.SessionKey{SessionID: 89})

	srv.dispatchMessage(stranger, protocol.MsgLobbyConnect, lcConnect(89, lcSecretB, 2, 0))

	if len(stranger.sent) != 0 {
		t.Errorf("the reconnect was answered with the types %x, want nothing", lcTypes(stranger))
	}
	if !stranger.closed {
		t.Error("the connection must be closed")
	}
	if got := sess.PeerCount(); got != 1 {
		t.Errorf("PeerCount() = %d, want 1: an unknown secret takes no slot", got)
	}
	if first.closed || sess.PeerIDForConn(first) == 0 {
		t.Error("the existing peer must keep its connection")
	}
}
