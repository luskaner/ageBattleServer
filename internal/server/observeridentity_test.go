package server

import (
	"encoding/binary"
	"testing"

	"github.com/luskaner/ageBattleServer/internal/protocol"
	"github.com/luskaner/ageBattleServer/internal/session"
)

// THE peerIdentifier ON THE WIRE.
//
// The field is the third u64 of the StartObserveRequest Vx, offset 9 of the body,
// and it is what orders the observer list. These tests pin the OFFSET and not only
// the semantics: if the identifier were read from somewhere else,
// the passwordHash's 0 or the 0x19's own would answer the same way and the whole
// suite would pass without ever having read the field.
//
// The body dispatched here starts AFTER the version byte (what parseWsMessage
// leaves as body[0]), so the wire's offset 9 is the 8 of this body. The 24 bytes
// are capture _26's, with the identifier changed so it does not depend on the
// store.

// observeRequest builds the 24-byte Vx body of a session-1 request with the
// identifier and the hash the caller wants.
func observeRequest(peerIdentifier, passwordHash uint64) []byte {
	body := make([]byte, 24)
	binary.LittleEndian.PutUint64(body[:8], 1) // titleID 0 + sessionID 1 in 48 bits
	binary.LittleEndian.PutUint64(body[8:16], peerIdentifier)
	binary.LittleEndian.PutUint64(body[16:24], passwordHash)
	return body
}

// TestThePeerIdentifierIsReadFromBodyOffsetEight pins the wire offset. The
// capture's own bytes are the vector: _26 answered 0 and served 7343 pages, and
// the identifier it carries is the third u64.
func TestThePeerIdentifierIsReadFromBodyOffsetEight(t *testing.T) {
	srv := observerTestServer(t)
	client := &mockMsgClient{}
	testSessionWithPeer(t, srv, client)
	sess := srv.sessionManager.GetSession(session.SessionKey{SessionID: 1})
	if sess == nil {
		t.Fatal("the helper did not create session 1")
	}

	srv.dispatchMessage(client, 0x1C, observeRequest(0xFEEDFACECAFEBEEF, 0))
	if got := client.PeerIdentifier(); got != 0xFEEDFACECAFEBEEF {
		t.Fatalf("connection identifier = %#x, want %#x: the field is at body offset 8 (wire 9)",
			got, uint64(0xFEEDFACECAFEBEEF))
	}
	if !sess.HasObserver(client) {
		t.Fatal("the observer should be in the list")
	}
}

// The two observers of 2026-09-29_012528_03cb67c7cf19 over the wire, with the
// two identifiers its two 0x1C bodies really carry at offset 9. Both must be
// served, from frame 0, and neither may be armed by the other's request.
func TestTwoObserversOverTheWireWithDistinctIdentifiers(t *testing.T) {
	srv := observerTestServer(t)
	// A player creates the session the way the game does.
	player := &mockMsgClient{}
	testSessionWithPeer(t, srv, player)
	sess := srv.sessionManager.GetSession(session.SessionKey{SessionID: 1})
	if sess == nil {
		t.Fatal("the request join did not create session 1")
	}
	// The store the session tick fills: 600 frames of 2 bytes, well inside the
	// default 64 KiB block, so Base() stays at 0 and the 0 arm is the one that
	// registers.
	for range 600 {
		sess.RecordFrame(protocol.MsgEndOfGameData, []byte{0xFE, 0x01})
	}

	first, second := &mockMsgClient{}, &mockMsgClient{}
	srv.dispatchMessage(first, 0x1C, observeRequest(0xC5C0428C2971CE8B, 0))
	srv.dispatchMessage(second, 0x1C, observeRequest(0xA9AE099C85ED50C4, 0))
	if got := sess.ObserverCount(); got != 2 {
		t.Fatalf("ObserverCount = %d, want 2", got)
	}
	if first.PeerIdentifier() == second.PeerIdentifier() {
		t.Fatal("the two requests carried distinct identifiers and they landed equal")
	}

	// The second one activating must arm the second entry, not the first.
	second.sent, first.sent = nil, nil
	srv.dispatchMessage(second, protocol.MsgActivateObserver, nil)
	if len(second.sent) == 0 {
		t.Fatal("the activating connection got no answer")
	}
	if len(first.sent) != 0 {
		t.Errorf("the other observer got %d messages: the lookup is by identifier", len(first.sent))
	}
	// Only the activated one is served, and from the first frame. The update walks
	// the list with a stride of one whole entry, so two passes are what it takes to
	// reach the second entry.
	first.sent, second.sent = nil, nil
	srv.tickSessions()
	srv.tickSessions()
	if n := countType(second.sent, protocol.MsgEndOfGameData); n != 200 {
		t.Errorf("the activated observer got %d frames in one visit, want the 200 cap", n)
	}
	if n := countType(first.sent, protocol.MsgEndOfGameData); n != 0 {
		t.Errorf("the observer that never activated got %d frames, want 0", n)
	}
}

// The legacy 0x18 form has no peerIdentifier and the original writes 0 into the
// connection, so a legacy observer is known by 0. That is not
// "no identity": any other connection of the session whose identity is still the
// zero it starts with resolves it. The test says so out loud so nobody "fixes"
// it later by making 0 mean absent.
func TestTheLegacyObserveRequestIsKnownByZero(t *testing.T) {
	srv := observerTestServer(t)
	// A client that never set RL_TITLEID: its session is keyed (-1, sessionID),
	// and the legacy request leaves -1 in the same field, so it resolves.
	player := &mockMsgClient{}
	connect := make([]byte, 17) // a first connect: reconnect state 0
	binary.LittleEndian.PutUint64(connect[:8], 1)
	binary.LittleEndian.PutUint16(connect[6:8], 0xFFFF)
	srv.dispatchMessage(player, protocol.MsgLobbyConnect, connect)
	sess := srv.sessionManager.GetSession(session.SessionKey{TitleID: -1, SessionID: 1})
	if sess == nil {
		t.Fatal("the LobbyConnect did not create the (-1, 1) session")
	}
	for range 10 {
		sess.RecordFrame(protocol.MsgEndOfGameData, []byte{0xFE, 0x01})
	}

	legacy := make([]byte, 8)
	binary.LittleEndian.PutUint64(legacy, 1)
	client := &mockMsgClient{}
	srv.dispatchMessage(client, protocol.MsgStartObserve, legacy)
	if got := client.PeerIdentifier(); got != 0 {
		t.Fatalf("legacy identifier = %#x, want 0", got)
	}
	// A second connection of the same session that never sent a request.
	other := &mockMsgClient{}
	if !sess.HasObserver(other) {
		t.Fatal("a connection with identifier 0 must resolve the zero-identified entry: " +
			"that is the comparator, not an oversight")
	}
	if sess.HasObserver(&mockMsgClient{peerID: 1}) {
		t.Fatal("a connection with a non-zero identifier must not match a zero entry")
	}
}

// Closing an observer's connection frees its entry, and THAT is a deliberate
// difference: nothing in the original removes an entry when its connection dies
// (the observer list is only appended, read, removed by StopObserve, and freed
// with the session), so the original leaves a dead entry that the observer update
// keeps writing to. Here the entry goes, because the entry holds a connection this
// server would otherwise keep writing to forever, and
// because a dead entry would eat an observer cap slot the game needs.
//
// The test exists to keep the difference visible and falsable: with the cap at 1,
// the second observer only gets in because the first entry was released.
func TestClosingAnObserverFreesItsCapSlot(t *testing.T) {
	srv := observerTestServer(t)
	player := &mockMsgClient{}
	testSessionWithPeer(t, srv, player)
	sess := srv.sessionManager.GetSession(session.SessionKey{SessionID: 1})
	if sess == nil {
		t.Fatal("the request join did not create session 1")
	}
	sess.MaxObservers = 1
	for range 10 {
		sess.RecordFrame(protocol.MsgEndOfGameData, []byte{0xFE, 0x01})
	}

	first := &mockMsgClient{}
	srv.AddClient(first)
	srv.dispatchMessage(first, 0x1C, observeRequest(0x1111, 0))
	if got := sess.ObserverCount(); got != 1 {
		t.Fatalf("ObserverCount = %d, want 1", got)
	}
	// Over the cap while it holds the slot.
	over := &mockMsgClient{}
	srv.dispatchMessage(over, 0x1C, observeRequest(0x2222, 0))
	if last := over.lastSent(); len(last) < 1 || last[len(last)-1] != 2 {
		t.Fatalf("the second observer over the cap got %x, want status 2", last)
	}

	// The first one goes away.
	srv.RemoveClient(first)
	if got := sess.ObserverCount(); got != 0 {
		t.Fatalf("ObserverCount = %d after the close, want 0", got)
	}
	// And now the slot is free.
	second := &mockMsgClient{}
	srv.dispatchMessage(second, 0x1C, observeRequest(0x3333, 0))
	if last := second.lastSent(); len(last) < 1 || last[len(last)-1] != 0 {
		t.Fatalf("the observer after the close got %x, want status 0", last)
	}
	if got := sess.ObserverCount(); got != 1 {
		t.Errorf("ObserverCount = %d, want 1", got)
	}
}

// The acceptance byte is a property of the ARM that ran, not of the list's
// contents afterwards. A connection that already holds an entry and sends a
// second request that lands in a rejecting arm must be answered with THAT arm's
// byte: the original writes the constant of the arm it took (3 for a store that
// has cycled, 2 for a full list, 4 for a session whose observers are off) and the
// arm is decided before anything is appended.
//
// This is the case the old `accepted := sess.HasObserver(client)` got wrong: the
// connection is in the list either way, so the query says "accepted" and the
// client is told 0 for a request the original answered 2. Reintroduce the query
// and this test fails on the byte.
func TestARepeatedRequestIsAnsweredWithItsOwnArmsByte(t *testing.T) {
	srv := observerTestServer(t)
	player := &mockMsgClient{}
	testSessionWithPeer(t, srv, player)
	sess := srv.sessionManager.GetSession(session.SessionKey{SessionID: 1})
	if sess == nil {
		t.Fatal("the request join did not create session 1")
	}
	sess.MaxObservers = 1
	for range 10 {
		sess.RecordFrame(protocol.MsgEndOfGameData, []byte{0xFE, 0x01})
	}

	// A second observer takes the single slot.
	holder := &mockMsgClient{}
	srv.dispatchMessage(holder, 0x1C, observeRequest(0xAAAA, 0))
	if last := holder.lastSent(); len(last) < 1 || last[len(last)-1] != 0 {
		t.Fatalf("the first observer got %x, want status 0", last)
	}

	// The holder asks again. It is over the cap now, so the original's cascade
	// takes the 2 arm and answers 2.
	holder.sent = nil
	srv.dispatchMessage(holder, 0x1C, observeRequest(0xBBBB, 0))
	if last := holder.lastSent(); len(last) < 1 || last[len(last)-1] != 2 {
		t.Fatalf("the repeated request got %x, want status 2: the byte is the arm's, "+
			"not a post-hoc question about the list", last)
	}
	// And it is still exactly one entry: the rejected arm appends nothing.
	if got := sess.ObserverCount(); got != 1 {
		t.Errorf("ObserverCount = %d, want 1: the 2 arm does not append", got)
	}
	// Its original entry still works, because the first entry is the first match.
	holder.sent = nil
	srv.dispatchMessage(holder, protocol.MsgActivateObserver, nil)
	if n := countType(holder.sent, protocol.MsgActivateObserverResp); n != 1 {
		t.Errorf("the holder got %d 0x1E after the rejected request, want 1", n)
	}
}

// countType counts the sent frames of one message type. lastSent's frames carry
// the [u16 BE length][type] prefix the bsPort transport writes.
func countType(sent [][]byte, msgType byte) int {
	n := 0
	for _, f := range sent {
		if len(f) >= 3 && f[2] == msgType {
			n++
		}
	}
	return n
}
