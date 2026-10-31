package server

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/luskaner/ageBattleServer/internal/protocol"
	"github.com/luskaner/ageBattleServer/internal/session"
)

// WHAT THIS FILE IS
//
// The per-peer count of sync messages the session has taken off that peer. The
// sync-message handler counts there once per message, and the only reader in the
// whole original is the reconnect: the LobbyConnect handler clones the client's
// message and overwrites its dataReceivedCount with this number before echoing it
// back, so the client is told how many of its uploads the server took.
//
// It is modelled here rather than left out because both halves of it are in the
// decompilation with nothing left to interpret, and because it was the one field
// of the §32.5 list whose writer AND reader were both identified. The limit is
// written down instead of papered over: NO committed capture contains a
// reconnect, so none of this can be compared against the original's bytes. What
// the tests below pin is that the counter moves where the original says it moves
// and reaches the echo where the original says it reaches.

// syncUploadCount is how many 0xFE uploads the tests put through before the
// reconnect. Three, so a counter that starts at one, that counts the connect
// itself, or that counts a different message type is a different number and not
// a coincidence.
const syncUploadCount = 3

// syncTestSecret keys the session both ends of the reconnect name.
const syncTestSecret = 0x0BAD_C0DE_0BAD_C0DE

// syncTestSessionID is the session both ends name.
const syncTestSessionID = 0x2B

// syncConnect builds a LobbyConnect body of the exact length ParseLobbyConnect
// accepts for that state: 17 + 6 for state 0 and 21 + 6 once the
// dataReceivedCount is present.
func syncConnect(state byte, dataReceived uint32) []byte {
	n := 23
	if state == 2 {
		n = 27
	}
	b := make([]byte, n)
	binary.LittleEndian.PutUint64(b[:8], syncTestSessionID)
	binary.LittleEndian.PutUint64(b[8:16], syncTestSecret)
	b[16] = state
	if state == 2 {
		binary.LittleEndian.PutUint32(b[17:21], dataReceived)
	}
	// The version fields the parser reads after the count, left at zero: this is
	// a version-0 client, which is what makes the echo comparison below exact
	// instead of approximate.
	return b
}

// uploadSync pushes one well-formed 0xFE game-data message:
// [0xfe][u16 LE length = 1][payload]. The slot is not part of the message: the
// server takes the sender's slot from its connection.
func uploadSync(t *testing.T, srv *Server, c *mockMsgClient, _ byte, payload byte) {
	t.Helper()
	srv.dispatchMessage(c, protocol.MsgGameData, []byte{0xfe, 0x01, 0x00, payload})
}

// echoedLobbyConnect returns the body of the one echoed connect on c, or nil.
func echoedLobbyConnect(t *testing.T, c *mockMsgClient) []byte {
	t.Helper()
	for _, f := range c.sent {
		if len(f) >= 3 && f[2] == protocol.MsgLobbyConnect {
			return f[3:]
		}
	}
	t.Fatal("no LobbyConnect was echoed to the reconnecting connection")
	return nil
}

// TestSyncMessageCountReachesTheReconnectEcho is the whole chain in one test: N
// uploads, a reconnect on state 2, and the echoed connect carrying N in the
// field the original overwrites.
func TestSyncMessageCountReachesTheReconnectEcho(t *testing.T) {
	srv := createTestServer(t)
	first := &mockMsgClient{}
	srv.dispatchMessage(first, protocol.MsgLobbyConnect, syncConnect(0, 0))
	sess := srv.sessionManager.GetSession(session.SessionKey{SessionID: syncTestSessionID})
	if sess == nil {
		t.Fatal("LobbyConnect should create the session")
	}
	slot := sess.PeerIDForConn(first)
	if slot == 0 {
		t.Fatal("the first connection should hold slot 1")
	}

	for i := range syncUploadCount {
		uploadSync(t, srv, first, slot, byte(i))
	}
	// The store needs a frame for the reconnect to have something to resend, and
	// PageCount has to be past the client's declared resume point so the resend
	// arm really runs and really is separate from the echo.
	for i := 1; i <= syncUploadCount; i++ {
		sess.RecordFrame(protocol.MsgEndOfGameData, []byte{0xFE, byte(i)})
	}

	if got := sess.SyncMessagesAccepted(slot); got != syncUploadCount {
		t.Fatalf("accepted sync messages = %d, want %d: the counter moves one per 0xFE upload",
			got, syncUploadCount)
	}

	second := &mockMsgClient{}
	request := syncConnect(2, 1)
	srv.dispatchMessage(second, protocol.MsgLobbyConnect, request)

	echo := echoedLobbyConnect(t, second)
	if len(echo) != len(request) {
		t.Fatalf("the echo is %d bytes, want the request's %d: the reconnect rewrites one field of a clone, "+
			"it does not resize the message", len(echo), len(request))
	}
	if got := binary.LittleEndian.Uint32(echo[17:21]); got != syncUploadCount {
		t.Errorf("the echoed connect carries dataReceivedCount = %d, want %d (the reconnect copies "+
			"the peer's own count into that field)", got, syncUploadCount)
	}
	// Everything else in the echo is the client's own message: the original's
	// clone differs in that one field and nothing else.
	if !bytes.Equal(echo[:17], request[:17]) || !bytes.Equal(echo[21:], request[21:]) {
		t.Errorf("the rewrite touched more than the count field:\n got %x\nwant %x", echo, request)
	}
	// And the resend still started from the CLIENT's count, not from the counter:
	// three frames stored, the client says it has up to index 0, so the original's
	// range [1, 2] puts frames 2 and 3 on the wire.
	var frames int
	for _, f := range second.sent {
		if len(f) >= 3 && f[2] == protocol.MsgEndOfGameData {
			frames++
		}
	}
	if frames != syncUploadCount-1 {
		t.Errorf("the reconnect resent %d frames, want %d: the resend range is the client's own "+
			"dataReceivedCount, read before the copy overwrites it",
			frames, syncUploadCount-1)
	}
}

// TestSyncMessageCountIgnoresEverythingElse pins the other half of the counting
// rule, which is what makes the number mean something: only a 0xFE message from a
// peer that is in the session counts. The original's guard is a pair: the peer
// has to exist, and the session must not be recording. The membership half is
// modelled here and the recording half is a recorded divergence
// (SESSION-JOURNAL.md 32.4, item 3), so this test does not claim it.
func TestSyncMessageCountIgnoresEverythingElse(t *testing.T) {
	srv := createTestServer(t)
	c := &mockMsgClient{}
	srv.dispatchMessage(c, protocol.MsgLobbyConnect, syncConnect(0, 0))
	sess := srv.sessionManager.GetSession(session.SessionKey{SessionID: syncTestSessionID})
	if sess == nil {
		t.Fatal("LobbyConnect should create the session")
	}
	slot := sess.PeerIDForConn(c)
	if slot != 1 {
		t.Fatalf("slot = %d, want 1", slot)
	}
	start := sess.SyncMessagesAccepted(slot)
	if start != 0 {
		t.Fatalf("a fresh peer already counts %d accepted sync messages, want 0", start)
	}

	// Not a 0xFE message.
	srv.dispatchMessage(c, protocol.MsgGameData, []byte{0x09, slot, 0xAA})
	// A 0xFE message from a connection that is not in any session.
	outsider := &mockMsgClient{}
	uploadSync(t, srv, outsider, slot, 0xAA)
	// A checksum message, which the original reaches by a different function.
	srv.dispatchMessage(c, protocol.MsgGameDataChecksum, []byte{0x00})
	// A second peer that DOES upload: the counter is per peer, not per session.
	other := &mockMsgClient{}
	srv.dispatchMessage(other, protocol.MsgLobbyConnect, syncConnect(0, 0))
	otherSlot := sess.PeerIDForConn(other)
	if otherSlot == 0 || otherSlot == slot {
		t.Fatalf("the second connection got slot %d, want a slot of its own", otherSlot)
	}
	uploadSync(t, srv, other, otherSlot, 0xBB)

	if got := sess.SyncMessagesAccepted(slot); got != start {
		t.Errorf("peer %d counted %d accepted sync messages, want %d: only a 0xFE upload from THAT peer "+
			"counts (the count walks the peer's own record and the guard needs the peer to exist)",
			slot, got, start)
	}
	if got := sess.SyncMessagesAccepted(otherSlot); got != 1 {
		t.Errorf("peer %d counted %d accepted sync messages, want 1: its own upload", otherSlot, got)
	}
	if got := sess.SyncMessagesAccepted(200); got != 0 {
		t.Errorf("a slot with no peer reported %d accepted sync messages, want 0", got)
	}
}
