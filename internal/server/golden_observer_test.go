package server

import (
	"encoding/binary"
	"testing"

	"github.com/luskaner/ageBattleServer/internal/protocol"
	"github.com/luskaner/ageBattleServer/internal/session"
)

// Captured observer handshakes, byte for byte, from two runs whose binaries were
// arranged in OPPOSITE positions (cross-checked with the generated start marker
// in the proxy log, which identifies our build).
//
//	_26  original as main:  0x1C -> 0x19=00 -> 0x1E=00 -> 7343 pages. WORKS.
//	_30  ours as main:       0x1C -> 0x19=03 -> no 0x1E at all       -> FAILS.
//
// The original answered 0 to the real client request and then served the replay
// normally. We answered 3, the client never sent ActivateObserver, and it then
// reported that the start-of-match data was gone. So 0 is the accept answer on
// the wire, and a 3 is what breaks the client. This test exists so that a
// regression here fails in one command instead of after a 20 minute game.
//
// Which arm of the cascade a request lands in does not depend on how much of the
// match is stored: the original registers the observer in the arm that answers 0,
// and that arm is the one that appends the entry, while the 3 is the arm that
// registers nobody. These vectors are the ones the original answered 0 to, and the store
// here is left in the state that capture's was in: a handful of frames and
// nothing dropped. That is not a detail: it is the same state a whole match is
// in, because the store is 8 blocks of 64 KiB and the match's acks fit.
var capturedStartObserve = []struct {
	name string
	body string
	// frames already in the store when the observer asked
	frames int
}{
	{"_26 observer asks at the start of the match",
		"c850a26b000018008bce7129c5c0428c0000000000000000", 2},
	{"_30 observer asks two minutes in",
		"d96a4b6300001800070171bfc60c37e30000000000000000", 2},
}

func TestStartObserveMatchesCapturedClientRequest(t *testing.T) {
	for _, v := range capturedStartObserve {
		t.Run(v.name, func(t *testing.T) {
			srv := observerTestServer(t)
			client := &mockMsgClient{}

			raw, err := hexDecode(v.body)
			if err != nil {
				t.Fatalf("bad vector: %v", err)
			}
			// The handler keys the session on (titleID, sessionID) packed in the
			// first u64, so the vector is what decides which session is found. The
			// session is created the way the real client creates it, with a
			// LobbyConnect carrying the same packed key.
			packed := binary.LittleEndian.Uint64(raw[:8])
			key := session.SessionKey{
				TitleID:   int16(packed >> 48),
				SessionID: int64(packed & 0xFFFFFFFFFFFF),
			}
			// A first connect, reconnect state 0: a reconnect to a session that
			// does not exist yet is closed and creates nothing.
			lobby := make([]byte, 17)
			binary.LittleEndian.PutUint64(lobby[:8], packed)
			srv.dispatchMessage(client, protocol.MsgLobbyConnect, lobby)
			sess := srv.sessionManager.GetSession(key)
			if sess == nil {
				t.Fatalf("LobbyConnect did not create session %+v", key)
			}
			// What the store is fed by: the 0x0A the session tick builds.
			for range v.frames {
				sess.RecordFrame(protocol.MsgEndOfGameData, []byte{0xFE, 0x00, 0x00})
			}

			srv.dispatchMessage(client, 0x1C, raw)

			last := client.lastSent()
			if len(last) != 4 || last[2] != protocol.MsgStartObserveResp {
				t.Fatalf("no StartObserveResponse, got %x", last)
			}
			// THE assertion. The original puts 0 here, for this exact input.
			if last[3] != 0 {
				t.Fatalf("StartObserveResponse = %d, want 0 (capture _26: a 3 makes the client skip ActivateObserver)", last[3])
			}
			// And acceptance is proven by state, not by the byte, because 0 is also
			// the "no replay" answer.
			if !client.observer {
				t.Error("the client should be registered as an observer")
			}
			if !sess.HasObserver(client) {
				t.Error("the client should be in the session's observer list")
			}

			// The client sends ActivateObserver next and the server must answer and
			// start serving; in _30 this is exactly where the run died.
			srv.dispatchMessage(client, protocol.MsgActivateObserver, nil)
			found := false
			for _, f := range client.sent {
				if len(f) >= 3 && f[2] == protocol.MsgActivateObserverResp {
					found = true
				}
			}
			if !found {
				t.Fatal("no ActivateObserverResponse: the observer never starts (this is the _30 failure)")
			}
		})
	}
}

func hexDecode(s string) ([]byte, error) {
	out := make([]byte, len(s)/2)
	for i := range out {
		hi, err := hexVal(s[2*i])
		if err != nil {
			return nil, err
		}
		lo, err := hexVal(s[2*i+1])
		if err != nil {
			return nil, err
		}
		out[i] = hi<<4 | lo
	}
	return out, nil
}

func hexVal(c byte) (byte, error) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', nil
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, nil
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, nil
	}
	return 0, errBadHex
}

var errBadHex = errorString("bad hex digit")

type errorString string

func (e errorString) Error() string { return string(e) }
