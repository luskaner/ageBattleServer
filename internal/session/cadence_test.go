package session

import (
	"testing"
)

// One ack per tick, and never more.
//
// The open question this pins down is "we tick differently", which a per-second
// count raised against capture 2026-09-28_205830_314309565d6e: four more
// EndOfGameData than the original in the one second where the ack stream starts.
// The capture settled it (internal/sim/cadence_test.go): in the run the capture
// itself recorded of our own binary, that second holds 2 and the original's 3,
// and both hold 8 in every second after it. So the surplus belongs to the replay
// and the cadence is the original's own: one ack per simulation tick.
//
// What is left to guard is the property the whole conclusion rests on, which is
// cheap and was nowhere written down: the tick produces EXACTLY one ack, once per
// tick, from the first tick after game data arrives. A server that emitted two in
// a tick would be the cadence bug the report would then be right about, and a
// server that deferred a tick's data to the next one would shift the whole stream
// by a tick, which is the same shape from the outside and a different defect.
//
// The number of ticks is a constant in this file and not a loop over a captured
// length: this test generates nothing, and the cap is here so that a change to it
// is a visible diff rather than a test that quietly grew.

// newCadenceSession is a session with ONE PEER IN SLOT 1, which is the
// precondition every upload below needs.
//
// It was added on 2026-10-06 and it is worth saying why, because the fixture was
// wrong in a way that hid a real rule: it used to be `NewSession(1)` alone, and
// `QueueGameData(1, ...)` on a slot that holds no peer. **An upload from a slot
// that is not a peer is refused**, so on that fixture every upload was dropped
// and the test was asserting an ack cadence the protocol cannot produce — the
// server's tick can only start once a real peer has sent something. The session
// has to have the peer for the question to mean anything.
func newCadenceSession() *Session {
	s := NewSession(1)
	s.AddPeer(NewPeer(1, &mockConn{}))
	return s
}

// ackCadenceTicks is how many ticks the invariant is checked over. 64 is eight
// seconds at the original's 125 ms period, which is longer than the 62 consecutive
// seconds both recorded sides of 2026-09-28_205830 agree on at 8 a second and is
// long enough that a drift of one tick in 64 would show up.
const ackCadenceTicks = 64

func TestTickIsOneAckOnceAndExactlyOneUntilItStops(t *testing.T) {
	s := newCadenceSession()
	s.SetObserverConfig(4, 8)

	// Nothing has been queued: the original emits no 0x0A at all before game data
	// (capture _4), and a tick that answered here would be a stream of acks where
	// the capture has none.
	for tick := range ackCadenceTicks {
		if ack, _ := s.PopAckTick(); ack != nil {
			t.Fatalf("tick %d produced an ack before any game data: %x", tick, ack)
		}
	}
	if s.PageCount != 0 {
		t.Fatalf("the frame store holds %d frames with nothing to store, want 0", s.PageCount)
	}

	// The first tick after data arrives carries it. Deferring it by a tick would
	// move the whole stream, and the capture's first ack is the one the report
	// localised.
	s.QueueGameData(1, []byte{0x02, 0x00, 0xAA})
	ack, _ := s.PopAckTick()
	if ack == nil {
		t.Fatal("the first tick after game data produced no ack")
	}
	if got, want := ack[1], byte(1); got != want {
		t.Fatalf("the first ack carries %d sub-message(s), want %d: the data queued before the tick has to be in it", got, want)
	}

	// One ack per tick from here, with or without new data: this is the 8 a second
	// both recorded sides of the capture hold.
	for tick := 1; tick < ackCadenceTicks; tick++ {
		if tick%2 == 0 {
			s.QueueGameData(1, []byte{0x02, 0x00, byte(tick)})
		}
		ack, notice := s.PopAckTick()
		if ack == nil {
			t.Fatalf("tick %d produced no ack: the original emits one per tick once game data has started", tick)
		}
		if notice != nil {
			t.Errorf("tick %d broadcast a desync notice with one peer: %x", tick, notice)
		}
		if tick%2 == 0 {
			// A flush: [0xfe][sub count][slot][u16 len][payload]. The one sub it
			// carries has to be the one queued for THIS tick.
			if ack[1] != 1 || ack[2] != 1 || ack[5] != byte(tick) {
				t.Fatalf("tick %d: the ack does not carry the sub-message queued for that tick: %x", tick, ack)
			}
		} else {
			// A heartbeat: the flush marker and no subs at all.
			if ack[1] != 0 {
				t.Fatalf("tick %d carried %d sub-message(s) with nothing queued for it: %x", tick, ack[1], ack)
			}
		}
	}

	// The store is fed by the tick, one frame per tick, so
	// the count of frames is the count of acks and nothing else. This is what an
	// observer is served out of.
	if got, want := s.PageCount, ackCadenceTicks; got != want {
		t.Errorf("the frame store holds %d frames after %d acks: one tick must append exactly one frame", got, want)
	}
	if got, want := s.ReplayPageCount(), ackCadenceTicks; got != want {
		t.Errorf("the store retains %d of %d frames, want all of them: a store this size does not roll over",
			got, want)
	}
	if s.replay.Base() != 0 {
		t.Errorf("Base = %d, want 0: nothing was dropped, so an observer joining now is served from frame 0", s.replay.Base())
	}
}

// TestTwoPeersDoNotDoubleTheAckRate is the one way a session could tick "twice as
// fast" without looking like it in a single-player capture: a per-peer ack instead
// of one ack per tick.
//
// The original builds one body and sends it to every peer, so two players see the
// same acks, not twice as many, and the pending-peer tail
// is what tells them apart (capture _6: a flush for slot 1 ends `01 02`, a flush for
// both ends `00`). This is the invariant that keeps that from drifting.
func TestTwoPeersDoNotDoubleTheAckRate(t *testing.T) {
	one, two := NewSession(2), NewSession(3)
	one.SetObserverConfig(4, 8)
	two.SetObserverConfig(4, 8)
	// The peers exist only to be counted: PopAckTick builds the pending-peer tail
	// from the session's slots, and no connection is touched on this path.
	one.AddPeer(NewPeer(1, nil))
	two.AddPeer(NewPeer(1, nil))
	two.AddPeer(NewPeer(2, nil))

	one.QueueGameData(1, []byte{0x02, 0x00, 0xAA})
	two.QueueGameData(1, []byte{0x02, 0x00, 0xAA})
	two.QueueGameData(2, []byte{0x02, 0x00, 0xBB})

	first, _ := one.PopAckTick()
	both, _ := two.PopAckTick()
	if len(first) == 0 || len(both) == 0 {
		t.Fatal("both sessions were given game data and neither produced an ack")
	}
	// Same number of acks, so the same number of frames, whatever the peer count.
	if one.PageCount != 1 || two.PageCount != 1 {
		t.Fatalf("one peer gave %d frames and two peers gave %d: the tick is per session, not per peer",
			one.PageCount, two.PageCount)
	}
	// The two-peer ack is the one-peer one plus the second sub-message, which is
	// [slot][u16 len][payload] = 4 bytes. Nothing else may move: not a second
	// header, not a second pending list.
	if got, want := len(both), len(first)+4; got != want {
		t.Errorf("the two-peer ack is %d bytes and the one-peer ack is %d, a difference of %d; want exactly the "+
			"4 bytes of the extra sub-message", got, len(first), got-len(first))
	}
	// Both slots are represented in the flush, so neither is pending: the tail is a
	// zero count.
	if both[len(both)-3] != 0 {
		t.Errorf("both peers are in the flush, so the pending list must be empty: %x", both)
	}
	// And the next tick, with nothing queued, lists both of them, which is the
	// shape capture _6 measured on the original (a heartbeat from a two-player
	// session is `02 01 02`, a solo one is `01 01`).
	heartbeat, _ := two.PopAckTick()
	want := []byte{0xfe, 0x00, 0x02, 0x01, 0x02, 0x00, 0x00}
	if len(heartbeat) != len(want) {
		t.Fatalf("the two-player heartbeat is %x, want the %x capture _6 measured", heartbeat, want)
	}
	for i := range want {
		if heartbeat[i] != want[i] {
			t.Fatalf("the two-player heartbeat is %x, want the %x capture _6 measured", heartbeat, want)
		}
	}
}
