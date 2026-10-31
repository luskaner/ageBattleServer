package session

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
)

// Byte-exact regression tests for the EndOfGameData (0x0A) body format,
// derived from the original server's captures _6 (AoE2, 2 players) and _24
// (AoE1, 1 player): main_log_6.csv / main_log_24.csv ack bodies.

// twoPeers is the queue-record list of a two-player session: one record per
// peer, in the order the peers joined, both marked as being in the started game.
// The ids were handed out in join order, which is the common case, so the
// record order and the slot order agree here.
var twoPeers = []QueueEntry{
	{Slot: 1, Participating: true},
	{Slot: 2, Participating: true},
}

func TestAckNilBeforeStart(t *testing.T) {
	var a AckState
	if b := a.PopTick(twoPeers); b != nil {
		t.Fatalf("PopTick before any queue = %x, want nil", b)
	}
}

// Heartbeat observed 3124 times: fe 00 02 01 02 00 00 (7 bytes).
func TestAckHeartbeat(t *testing.T) {
	a := AckState{}
	a.QueueDataSub(1, []byte{0x00, 0x00})
	if b := a.PopTick(twoPeers); !bytes.Equal(b, []byte{0xfe, 0x01, 0x01, 0x00, 0x00, 0x01, 0x02, 0x00, 0x00}) {
		t.Fatalf("flush = %x", b)
	}
	if b := a.PopTick(twoPeers); !bytes.Equal(b, []byte{0xfe, 0x00, 0x02, 0x01, 0x02, 0x00, 0x00}) {
		t.Fatalf("heartbeat = %x", b)
	}
	if b := a.PopTick(twoPeers); !bytes.Equal(b, []byte{0xfe, 0x00, 0x02, 0x01, 0x02, 0x00, 0x00}) {
		t.Fatalf("second heartbeat = %x", b)
	}
}

// Standalone checksum ack, one entry (N=1), observed exactly as:
// fe 00 02 01 02 0b 00 02 01 00 3f 8b 1b 00 b0 00 00 00 (18 bytes).
func TestAckStandaloneOneEntry(t *testing.T) {
	a := AckState{}
	entries := []byte{0x3f, 0x8b, 0x1b, 0x00, 0xb0, 0x00, 0x00, 0x00}
	a.QueueEntries(1, entries)
	a.QueueEntries(2, entries) // both active peers must contribute

	// The tick drains, then reconciles, then builds the frame, so the
	// reconciled entries ride in the *same* tick that received them: there is
	// no one-tick deferral for this block.
	a.ReconcileChecksums(twoPeers)
	want := []byte{0xfe, 0x00, 0x02, 0x01, 0x02, 0x0b, 0x00, 0x02, 0x01, 0x00}
	want = append(want, entries...)
	if b := a.PopTick(twoPeers); !bytes.Equal(b, want) {
		t.Fatalf("standalone = %x, want %x", b, want)
	}

	// The next tick has nothing left to merge.
	a.ReconcileChecksums(twoPeers)
	if b := a.PopTick(twoPeers); !bytes.Equal(b, []byte{0xfe, 0x00, 0x02, 0x01, 0x02, 0x00, 0x00}) {
		t.Fatalf("after ack = %x, want a plain heartbeat", b)
	}
}

// Flush #1 observed: fe 02 [01][1f 02][543B zlib] [02][02 00][00 30] 00 00 00
// — both players' chunks merged into one flush, multi-sub tail 00 00 00.
func TestAckFlushTwoSubs(t *testing.T) {
	a := AckState{}
	chunk1 := make([]byte, 543)
	for i := range chunk1 {
		chunk1[i] = byte(i)
	}
	p1 := []byte{0x1f, 0x02}
	p1 = append(p1, chunk1...)
	p2 := []byte{0x02, 0x00, 0x00, 0x30}
	a.QueueDataSub(1, p1)
	a.QueueDataSub(2, p2)

	body := a.PopTick(twoPeers)
	want := append([]byte{0xfe, 0x02, 0x01, 0x1f, 0x02}, chunk1...)
	want = append(want, 0x02, 0x02, 0x00, 0x00, 0x30, 0x00, 0x00, 0x00)
	if !bytes.Equal(body, want) {
		t.Fatalf("merged flush = %d bytes, want %d bytes", len(body), len(want))
	}
}

// Single-player flush with tail advertising the other player's slot:
// sub 01 -> ... 01 02 00 00 (observed 3124-3131-ish), sub 02 -> ... 01 01 00 00.
func TestAckFlushSingleSubTails(t *testing.T) {
	a := AckState{}
	a.QueueDataSub(1, []byte{0x00, 0x00})
	if b := a.PopTick(twoPeers); !bytes.Equal(b, []byte{0xfe, 0x01, 0x01, 0x00, 0x00, 0x01, 0x02, 0x00, 0x00}) {
		t.Fatalf("sub-1 flush = %x", b)
	}
	a.QueueDataSub(2, []byte{0x00, 0x00})
	if b := a.PopTick(twoPeers); !bytes.Equal(b, []byte{0xfe, 0x01, 0x02, 0x00, 0x00, 0x01, 0x01, 0x00, 0x00}) {
		t.Fatalf("sub-2 flush = %x", b)
	}
}

// Multi-sub flush whose subs all belong to the same player still advertises
// the other player's slot (capture _9: (1,1) -> 01 02 00 00, (2,2) -> 01 01 00 00).
func TestAckFlushSameSlotTails(t *testing.T) {
	a := AckState{}
	a.QueueDataSub(1, []byte{0x00, 0x00})
	a.QueueDataSub(1, []byte{0x00, 0x00})
	if b := a.PopTick(twoPeers); !bytes.Equal(b, []byte{0xfe, 0x02, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x01, 0x02, 0x00, 0x00}) {
		t.Fatalf("(1,1) flush = %x", b)
	}
	a.QueueDataSub(2, []byte{0x00, 0x00})
	a.QueueDataSub(2, []byte{0x00, 0x00})
	if b := a.PopTick(twoPeers); !bytes.Equal(b, []byte{0xfe, 0x02, 0x02, 0x00, 0x00, 0x02, 0x00, 0x00, 0x01, 0x01, 0x00, 0x00}) {
		t.Fatalf("(2,2) flush = %x", b)
	}
}

// Flush with pending checksum entries, tail as observed:
// 01 02 33 00 02 06 00 + 48 bytes (6 entries). Both active peers must have
// contributed.
//
// The merged block does NOT ride the flush that triggered it. The original sends
// the 0x0A and only then reconciles, so the first tick goes out with an empty
// tail and the merged entries reach the peers on
// the following tick. An earlier revision reconciled first and shipped them in the
// same tick; that was wrong.
func TestAckFlushWithEntries(t *testing.T) {
	a := AckState{}
	entries := make([]byte, 48)
	for i := range entries {
		entries[i] = byte(i)
	}
	a.QueueEntries(1, entries)
	a.QueueEntries(2, entries) // both active peers must contribute
	a.QueueDataSub(1, []byte{0x00, 0x00})

	// The flush goes out first, before any reconciliation: peer 2 has no sub, so
	// it is the only entry in the pending list, and the tail is empty.
	if body := a.PopTick(twoPeers); !bytes.Equal(body, []byte{0xfe, 0x01, 0x01, 0x00, 0x00, 0x01, 0x02, 0x00, 0x00}) {
		t.Fatalf("flush before reconciliation = %x", body)
	}

	// Six entries need six rounds, and the original runs at most ten. The loop
	// stops on reconcileStop, i.e. when there is nothing left to consume.
	for range reconcileMaxRounds {
		if a.ReconcileChecksums(twoPeers) == reconcileStop {
			break
		}
	}

	// The next tick carries the merged block. No subs this time, so both peers
	// are pending.
	body := a.PopTick(twoPeers)
	want := []byte{0xfe, 0x00, 0x02, 0x01, 0x02, 0x33, 0x00, 0x02, 0x06, 0x00}
	want = append(want, entries...)
	if !bytes.Equal(body, want) {
		t.Fatalf("flush with entries = %x, want %x", body, want)
	}
}

// Solo session (capture _24, AoE1): the pending list holds the single peer,
// so heartbeats shrink to 6 bytes and a flush that carries the only player's
// data ends in an empty pending list (`00 00 00` / `00 <block>`).
func TestAckSoloSessionTails(t *testing.T) {
	a := AckState{}
	onePeer := []QueueEntry{{Slot: 1, Participating: true}}
	a.QueueDataSub(1, []byte{0x00, 0x00})
	if b := a.PopTick(onePeer); !bytes.Equal(b, []byte{0xfe, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00}) {
		t.Fatalf("solo flush = %x", b)
	}
	if b := a.PopTick(onePeer); !bytes.Equal(b, []byte{0xfe, 0x00, 0x01, 0x01, 0x00, 0x00}) {
		t.Fatalf("solo heartbeat = %x", b)
	}

	entries := []byte{0x3f, 0x8b, 0x1b, 0x00, 0xb0, 0x00, 0x00, 0x00}
	a.QueueEntries(1, entries)
	a.ReconcileChecksums(onePeer)
	want := []byte{0xfe, 0x00, 0x01, 0x01, 0x0b, 0x00, 0x02, 0x01, 0x00}
	want = append(want, entries...)
	if b := a.PopTick(onePeer); !bytes.Equal(b, want) {
		t.Fatalf("solo standalone = %x, want %x", b, want)
	}

	// Entries and data arriving in the same tick ride the same flush: the tick
	// drains, reconciles, then builds the frame.
	a.QueueEntries(1, entries)
	a.QueueDataSub(1, []byte{0x00, 0x01})
	a.ReconcileChecksums(onePeer)
	wantFlush := append([]byte{0xfe, 0x01, 0x01, 0x00, 0x01, 0x00, 0x0b, 0x00, 0x02, 0x01, 0x00}, entries...)
	if b := a.PopTick(onePeer); !bytes.Equal(b, wantFlush) {
		t.Fatalf("solo flush with entries = %x, want %x", b, wantFlush)
	}
}

// Acks are session-wide: both players' chunks from one tick merge into a
// single flush body; the player count byte reflects the session membership.
func TestAckSessionWideMerge(t *testing.T) {
	a := AckState{}
	a.QueueDataSub(1, []byte{0x02, 0x00, 0xaa, 0xbb})
	a.QueueDataSub(2, []byte{0x02, 0x00, 0xcc, 0xdd})
	body := a.PopTick(twoPeers)
	want := []byte{0xfe, 0x02,
		0x01, 0x02, 0x00, 0xaa, 0xbb,
		0x02, 0x02, 0x00, 0xcc, 0xdd,
		0x00, 0x00, 0x00}
	if !bytes.Equal(body, want) {
		t.Fatalf("merged body = %x, want %x", body, want)
	}
}

// Reconciliation needs every active peer to have an entry before it merges
// anything: with one peer silent it returns early and consumes nothing, unless
// every active peer has at least one stored entry.
func TestAckReconcileWaitsForEveryPeer(t *testing.T) {
	a := AckState{}
	a.QueueEntries(1, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08})

	if got := a.ReconcileChecksums(twoPeers); got != reconcileStop {
		t.Fatalf("one peer missing must do nothing, got %d", got)
	}
	if len(a.merged) != 0 {
		t.Fatalf("merged = %x, want nothing while a peer is silent", a.merged)
	}
	if a.pendingEntries(1) != 1 {
		t.Error("nothing may be consumed while a peer is silent")
	}
}

// When every active peer agrees on the front entry, the first active peer's
// entry is merged and one entry is popped from each, so the vectors stay aligned
// position by position.
func TestAckReconcileMergesAlignedEntries(t *testing.T) {
	a := AckState{}
	e1 := []byte{0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11}
	e2 := []byte{0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11} // identical
	e3 := []byte{0x99, 0x9a, 0x9b, 0x9c, 0x9d, 0x9e, 0x9f, 0xa0} // next position
	a.QueueEntries(1, append(append([]byte{}, e1...), e3...))
	a.QueueEntries(2, append(append([]byte{}, e2...), e3...))

	if got := a.ReconcileChecksums(twoPeers); got != reconcileAgreed {
		t.Fatalf("identical front entries must agree, got %d", got)
	}
	if !bytes.Equal(a.merged, e1) {
		t.Fatalf("merged = %x, want the first peer's entry %x", a.merged, e1)
	}
	if a.pendingEntries(1) != 1 || a.pendingEntries(2) != 1 {
		t.Errorf("one entry should have been popped from each, got %d/%d", a.pendingEntries(1), a.pendingEntries(2))
	}

	// A second round merges the next aligned position.
	if got := a.ReconcileChecksums(twoPeers); got != reconcileAgreed {
		t.Fatalf("round two should also agree, got %d", got)
	}
	want := append(append([]byte{}, e1...), e3...)
	if !bytes.Equal(a.merged, want) {
		t.Fatalf("merged after two rounds = %x, want %x", a.merged, want)
	}
}

// A disagreement is reported as reconcileDesynced and nothing is merged: that is
// where the original broadcasts its 0xFB desync notice and counts it. Note that
// a desync is not a stop
// condition; the caller's loop only breaks when reconciliation reports there is
// nothing left to do.
//
// A disagreement still consumes one entry from every peer. Both outcomes, the
// agreeing one and the disagreeing one, reach the same step that pops one entry
// from each peer. Consuming on disagreement is what keeps a stuck peer from
// wedging the queue.
func TestAckReconcileReportsDisagreement(t *testing.T) {
	a := AckState{}
	a.QueueEntries(1, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08})
	a.QueueEntries(2, []byte{0x01, 0x02, 0x03, 0x04, 0xff, 0xfe, 0xfd, 0xfc})

	if got := a.ReconcileChecksums(twoPeers); got != reconcileDesynced {
		t.Fatalf("differing front entries must report a desync, got %d", got)
	}
	if len(a.merged) != 0 {
		t.Errorf("a disagreement must not merge anything, got %x", a.merged)
	}
	if a.pendingEntries(1) != 0 || a.pendingEntries(2) != 0 {
		t.Errorf("a disagreement must consume one entry from each peer, got %d/%d",
			a.pendingEntries(1), a.pendingEntries(2))
	}
	// With both vectors drained the next round stops instead of desyncing again.
	if got := a.ReconcileChecksums(twoPeers); got != reconcileStop {
		t.Errorf("after a disagreement drains the queues the next round must stop, got %d", got)
	}
}

// The entries block of the 0x0A tail is the checksum message: [0x02][u16 count
// LE][count * 8 bytes] wrapped in a [u16 len] prefix, with the length written as
// the byte length of the block and each 8-byte element written as two u32 LE. An
// earlier revision wrote the count as a u8 plus a reserved 0x00, which emits the
// same bytes below 256 but truncates above it; this pins the real u16.
func TestAckTailEntriesBlockLayout(t *testing.T) {
	a := AckState{}
	const n = 6
	entries := make([]byte, 0, n*8)
	for i := range n {
		entries = binary.LittleEndian.AppendUint32(entries, uint32(0x1000+i)) // crc
		entries = binary.LittleEndian.AppendUint32(entries, uint32(i+1))       // page
	}
	a.QueueEntries(1, entries)
	a.QueueEntries(2, entries)
	a.QueueDataSub(1, []byte{0x00, 0x00})

	// The tick sends the 0x0A first, so this flush carries an empty tail, and
	// only then does the reconciliation merge the six agreed positions.
	a.PopTick(twoPeers)
	for range reconcileMaxRounds {
		if a.ReconcileChecksums(twoPeers) == reconcileStop {
			break
		}
	}
	body := a.PopTick(twoPeers)

	// fe 00 | pending count 02, peers 01 02 | u16 len | 02 | count | entries
	want := []byte{0xfe, 0x00, 0x02, 0x01, 0x02}
	want = binary.LittleEndian.AppendUint16(want, uint16(3+len(entries)))
	want = append(want, 0x02) // marker, then the u16 count as two bytes
	want = binary.LittleEndian.AppendUint16(want, n)
	want = append(want, entries...)
	if !bytes.Equal(body, want) {
		t.Fatalf("tail layout = %x, want %x", body, want)
	}

	// Re-read it the way the original parses it.
	const off = 5
	if got := int(binary.LittleEndian.Uint16(body[off : off+2])); got != 3+len(entries) {
		t.Errorf("blob length = %d, want %d", got, 3+len(entries))
	}
	if body[off+2] != 0x02 {
		t.Errorf("marker = %#x, want 0x02", body[off+2])
	}
	if got := int(binary.LittleEndian.Uint16(body[off+3 : off+5])); got != n {
		t.Errorf("u16 count = %d, want %d", got, n)
	}
	if len(body)-off-5 != n*8 {
		t.Errorf("payload = %d bytes, want %d", len(body)-off-5, n*8)
	}
}

// BenchmarkAckQueueAndPop measures the per-tick game-data flush: two players
// queue a payload and PopTick builds the 0x0A body. Before P2 each sub
// allocated a new buffer; after, subs come from the pool.
func BenchmarkAckQueueAndPop(b *testing.B) {
	var a AckState
	payload := make([]byte, 128)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.QueueDataSub(1, payload)
		a.QueueDataSub(2, payload)
		_ = a.PopTick(twoPeers)
	}
}

// BenchmarkAckQueueAndPopSameSlot measures the common one-player-per-tick
// flush, where every sub belongs to the same slot. The largest case is the most
// one tick takes from a peer, ackDrainPasses chunks: more would pile up in the
// queue from one iteration to the next until the cap drops them.
func BenchmarkAckQueueAndPopSameSlot(b *testing.B) {
	for _, subs := range []int{2, 8, ackDrainPasses} {
		b.Run(fmt.Sprintf("subs=%d", subs), func(b *testing.B) {
			var a AckState
			payload := make([]byte, 128)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for range subs {
					a.QueueDataSub(1, payload)
				}
				_ = a.PopTick(twoPeers)
			}
		})
	}
}

// The original bounds each peer's pending queue and drops the incoming packet
// when it is full: the caps are 0x12C (300) entries and 0x80000
// (512 KiB) of payload, and it logs every dropped packet.

// queuedSubs is how many sub-messages the state holds across every slot.
func queuedSubs(a *AckState) int {
	n := 0
	for _, q := range a.queued {
		if q != nil {
			n += q.len()
		}
	}
	return n
}

func TestAckQueueCapEntries(t *testing.T) {
	var a AckState
	chunk := []byte{0x00, 0x00} // an empty payload: the entry cap binds, the byte cap never does

	for range maxAckQueueEntries {
		a.QueueDataSub(1, chunk)
	}
	if got := queuedSubs(&a); got != maxAckQueueEntries {
		t.Fatalf("queued %d subs, want %d", got, maxAckQueueEntries)
	}

	// One past the cap: dropped, and the queue does not grow.
	a.QueueDataSub(1, chunk)
	if got := queuedSubs(&a); got != maxAckQueueEntries {
		t.Errorf("queued %d subs after the cap, want %d (packet must be dropped)", got, maxAckQueueEntries)
	}
	// Dropping must not disable the tick: the original marks the peer as having
	// sent data on the drop path too.
	if b := a.PopTick(twoPeers); b == nil {
		t.Error("PopTick must still emit after a dropped packet")
	}
}

func TestAckQueueCapBytes(t *testing.T) {
	var a AckState
	chunk := make([]byte, 4096)
	// The byte cap charges the payload alone, the chunk minus its u16 length.
	for a.queued[1] == nil || a.queued[1].bytes+len(chunk)-2 < maxAckQueueBytes {
		a.QueueDataSub(1, chunk)
		if queuedSubs(&a) > maxAckQueueEntries {
			t.Fatal("entry cap hit before the byte cap; test payload is too large")
		}
	}
	before := queuedSubs(&a)
	a.QueueDataSub(1, chunk)
	if queuedSubs(&a) != before {
		t.Errorf("queued %d subs, want %d (packet over %d bytes must be dropped)", queuedSubs(&a), before, maxAckQueueBytes)
	}
}

func TestAckQueueCapIsPerSlot(t *testing.T) {
	var a AckState
	chunk := []byte{0x00, 0x00}
	for range maxAckQueueEntries {
		a.QueueDataSub(1, chunk)
	}
	// Slot 1 is full, but the other player is unaffected.
	a.QueueDataSub(2, chunk)
	if queuedSubs(&a) != maxAckQueueEntries+1 {
		t.Errorf("queued %d subs, want %d (slot 2 must not be capped by slot 1)", queuedSubs(&a), maxAckQueueEntries+1)
	}
}

func TestAckQueueCapResetsAfterTick(t *testing.T) {
	var a AckState
	chunk := []byte{0x00, 0x00}
	for range maxAckQueueEntries {
		a.QueueDataSub(1, chunk)
	}
	a.QueueDataSub(1, chunk) // dropped
	_ = a.PopTick(twoPeers)

	// The tick took chunks off the slot, so the peer can queue again (slots get
	// reused by later peers, so the accounting must not stay pinned at the cap).
	before := queuedSubs(&a)
	a.QueueDataSub(1, chunk)
	if queuedSubs(&a) != before+1 {
		t.Errorf("queued %d subs after a draining tick, want %d", queuedSubs(&a), before+1)
	}
}
