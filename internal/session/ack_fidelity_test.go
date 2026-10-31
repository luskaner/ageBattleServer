package session

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"testing"
	"time"
)

// The per-tick ack against the protocol's own limits: how the drain takes the
// queued chunks, how much one 0x0A may carry, how the checksum reconciliation
// picks its reference, what refreshes a peer's sync-timeout stamp, and what the
// desync notice carries.

// ackSub is one sub-message of a parsed 0x0A: the peer id and the chunk.
type ackSub struct {
	slot    byte
	payload []byte
}

// parsedAck is a 0x0A body read back field by field.
type parsedAck struct {
	subs    []ackSub
	pending []byte
	// block is the checksum block after its u16 length: nil for [00 00].
	block []byte
}

// parseAck reads a 0x0A body the way a client does, and fails the test unless
// every byte is consumed. A sub count that disagrees with the subs that follow
// misaligns everything after it, so exact consumption is what proves the count.
func parseAck(t *testing.T, body []byte) parsedAck {
	t.Helper()
	var p parsedAck
	if len(body) < 5 || body[0] != 0xfe {
		t.Fatalf("not a 0x0A body: %x", body)
	}
	n := int(body[1])
	off := 2
	for i := range n {
		if off+3 > len(body) {
			t.Fatalf("sub %d of %d runs past the body (%d bytes)", i+1, n, len(body))
		}
		l := int(binary.LittleEndian.Uint16(body[off+1:]))
		if off+3+l > len(body) {
			t.Fatalf("sub %d of %d claims %d bytes past the body (%d bytes)", i+1, n, l, len(body))
		}
		p.subs = append(p.subs, ackSub{slot: body[off], payload: body[off+3 : off+3+l]})
		off += 3 + l
	}
	if off >= len(body) {
		t.Fatalf("no pending list after %d subs: %d bytes", n, len(body))
	}
	k := int(body[off])
	if off+1+k+2 > len(body) {
		t.Fatalf("pending list of %d runs past the body", k)
	}
	p.pending = body[off+1 : off+1+k]
	off += 1 + k
	blen := int(binary.LittleEndian.Uint16(body[off:]))
	off += 2
	if off+blen != len(body) {
		t.Fatalf("the checksum block claims %d bytes and %d are left: the body does not parse exactly", blen, len(body)-off)
	}
	if blen > 0 {
		p.block = body[off:]
	}
	return p
}

// chunk is a client upload as QueueGameData takes it: [len u16 LE][payload].
func chunk(payload []byte) []byte {
	out := binary.LittleEndian.AppendUint16(nil, uint16(len(payload)))
	return append(out, payload...)
}

// sizedChunk is a chunk whose payload is n bytes of fill.
func sizedChunk(n int, fill byte) []byte {
	return chunk(bytes.Repeat([]byte{fill}, n))
}

// checksumBlock is the 0x0A's checksum block for the given 8-byte records.
func checksumBlock(records ...[]byte) []byte {
	out := []byte{0x02}
	out = binary.LittleEndian.AppendUint16(out, uint16(len(records)))
	for _, r := range records {
		out = append(out, r...)
	}
	return out
}

// record is one 8-byte checksum record: two u32 LE.
func record(d0, d1 uint32) []byte {
	out := binary.LittleEndian.AppendUint32(nil, d0)
	return binary.LittleEndian.AppendUint32(out, d1)
}

// ---------------------------------------------------------------------------
// The drain: one chunk from every peer per pass, seventeen passes, two limits
// ---------------------------------------------------------------------------

// One pass takes ONE chunk from every peer that has one, in join order, and the
// next pass starts again at the first peer. So two chunks of peer 1 and one of
// peer 2 come out as 1, 2, 1, and not grouped by peer.
func TestAckDrainInterleavesOneChunkPerPeerPerPass(t *testing.T) {
	s := joinedIn(t, 1, 1, 2)
	s.QueueGameData(1, []byte{0x02, 0x00, 0x11, 0x11})
	s.QueueGameData(1, []byte{0x02, 0x00, 0x12, 0x12})
	s.QueueGameData(2, []byte{0x02, 0x00, 0x22, 0x22})

	body, _ := s.PopAckTick()
	want := []byte{0xfe, 0x03,
		0x01, 0x02, 0x00, 0x11, 0x11, // pass 1: peer 1's first chunk
		0x02, 0x02, 0x00, 0x22, 0x22, // pass 1: peer 2's only chunk
		0x01, 0x02, 0x00, 0x12, 0x12, // pass 2: peer 1's second chunk
		0x00,       // nobody pending: both peers were flushed on the first pass
		0x00, 0x00} // no checksum entries
	if !bytes.Equal(body, want) {
		t.Fatalf("flush = %x,\nwant     %x (sub peer ids 1,2,1)", body, want)
	}
}

// One tick makes at most seventeen passes, so a peer gets at most seventeen
// chunks into one 0x0A and the rest wait for the next tick. Two peers with forty
// chunks each come out as 34, 34 and 12, every chunk exactly once and in its
// peer's upload order.
//
// The merged checksum entries ride only the pass that ends the tick's drain with
// nothing more to take. A tick whose seventeenth pass still left chunks behind
// carries none, and the entries wait in the merged list.
func TestAckDrainTakesSeventeenChunksAPeerATick(t *testing.T) {
	const perPeer = 40
	s := joinedIn(t, 1, 1, 2)
	for i := range perPeer {
		s.QueueGameData(1, chunk([]byte{0x01, byte(i)}))
		s.QueueGameData(2, chunk([]byte{0x02, byte(i)}))
	}
	entry := record(0xAABBCCDD, 7)
	s.QueueChecksumEntries(1, entry)
	s.QueueChecksumEntries(2, entry)

	next := map[byte]int{1: 0, 2: 0}
	wantCounts := []int{34, 34, 12}
	for tick, wantCount := range wantCounts {
		body, notice := s.PopAckTick()
		if notice != nil {
			t.Fatalf("tick %d: agreeing peers raised a desync notice %x", tick+1, notice)
		}
		if len(body) >= ackMaxSize {
			t.Fatalf("tick %d: the body is %d bytes, at or over the %d limit", tick+1, len(body), ackMaxSize)
		}
		p := parseAck(t, body)
		if len(p.subs) != wantCount {
			t.Fatalf("tick %d carries %d chunks, want %d (seventeen passes, one chunk a peer a pass)",
				tick+1, len(p.subs), wantCount)
		}
		for i, sub := range p.subs {
			wantSlot := byte(1 + i%2)
			if sub.slot != wantSlot {
				t.Fatalf("tick %d, sub %d is from peer %d, want %d: the passes interleave the peers", tick+1, i, sub.slot, wantSlot)
			}
			if len(sub.payload) != 2 || sub.payload[0] != sub.slot || int(sub.payload[1]) != next[sub.slot] {
				t.Fatalf("tick %d, sub %d = %x, want peer %d's chunk %d", tick+1, i, sub.payload, sub.slot, next[sub.slot])
			}
			next[sub.slot]++
		}
		if len(p.pending) != 0 {
			t.Errorf("tick %d lists %x as pending, want nobody: both peers were flushed on the first pass", tick+1, p.pending)
		}
		switch tick {
		case 0, 1:
			// Both ticks end with chunks still queued, so the entry merged after
			// the first tick is not attached to either.
			if p.block != nil {
				t.Errorf("tick %d attached checksum block %x although its last pass left chunks behind", tick+1, p.block)
			}
		case 2:
			if want := checksumBlock(entry); !bytes.Equal(p.block, want) {
				t.Errorf("tick 3 checksum block = %x, want %x: the drain ended with nothing left, so the entry "+
					"merged two ticks ago rides this one", p.block, want)
			}
		}
	}
	if next[1] != perPeer || next[2] != perPeer {
		t.Fatalf("delivered %d and %d chunks, want %d each", next[1], next[2], perPeer)
	}
	if body, _ := s.PopAckTick(); len(parseAck(t, body).subs) != 0 {
		t.Fatalf("a fourth tick still carries chunks: %x", body)
	}
}

// The sub count is a u8, and a pass takes a chunk only while the 0x0A holds
// fewer than 255. Fifteen peers with seventeen chunks each fill exactly 255;
// sixteen peers leave seventeen chunks for the next tick.
//
// A 0x0A that holds 255 sub-messages carries no checksum block either, even when
// its drain ended with nothing left: the block needs the same room.
func TestAckDrainStopsAt255SubMessages(t *testing.T) {
	t.Run("15 peers fill 255 exactly", func(t *testing.T) {
		const peers = 15
		ids := make([]byte, peers)
		for i := range ids {
			ids[i] = byte(i + 1)
		}
		s := joinedIn(t, 1, ids...)
		entry := record(1, 2)
		for _, id := range ids {
			s.QueueGameData(id, chunk([]byte{id}))
			s.QueueChecksumEntries(id, entry)
		}
		if body, _ := s.PopAckTick(); len(parseAck(t, body).subs) != peers {
			t.Fatalf("the first tick should flush one chunk a peer: %x", body)
		}
		for _, id := range ids {
			for range ackDrainPasses {
				s.QueueGameData(id, chunk([]byte{id}))
			}
		}
		body, _ := s.PopAckTick()
		if body[1] != 0xff {
			t.Fatalf("count byte = %d, want 255", body[1])
		}
		p := parseAck(t, body)
		if len(p.subs) != 255 {
			t.Fatalf("the 0x0A carries %d chunks, want 255", len(p.subs))
		}
		if p.block != nil {
			t.Errorf("a 0x0A with 255 sub-messages attached checksum block %x", p.block)
		}
		for _, id := range ids {
			if !s.Ack.QueueEmpty(id) {
				t.Errorf("peer %d still has chunks queued after a 255-chunk tick", id)
			}
		}
		next, _ := s.PopAckTick()
		if p := parseAck(t, next); len(p.subs) != 0 || !bytes.Equal(p.block, checksumBlock(entry)) {
			t.Errorf("the tick after the full one = %x, want no chunks and the entry the full one could not carry", next)
		}
	})

	t.Run("16 peers leave 17 for the next tick", func(t *testing.T) {
		const peers = 16
		ids := make([]byte, peers)
		for i := range ids {
			ids[i] = byte(i + 1)
		}
		s := joinedIn(t, 1, ids...)
		for _, id := range ids {
			for range ackDrainPasses {
				s.QueueGameData(id, chunk([]byte{id}))
			}
		}
		body, _ := s.PopAckTick()
		if body[1] != 0xff {
			t.Fatalf("count byte = %d, want 255: 16 x 17 chunks must not wrap the u8 count", body[1])
		}
		if p := parseAck(t, body); len(p.subs) != 255 {
			t.Fatalf("the first 0x0A carries %d chunks, want 255", len(p.subs))
		}
		body, _ = s.PopAckTick()
		p := parseAck(t, body)
		if len(p.subs) != peers*ackDrainPasses-255 {
			t.Fatalf("the second 0x0A carries %d chunks, want the %d left over", len(p.subs), peers*ackDrainPasses-255)
		}
		// Peers 1..15 have one chunk left, peer 16 has two: the first pass takes
		// one from each in join order, the second pass peer 16's last.
		for i, sub := range p.subs {
			want := byte(i + 1)
			if i == peers {
				want = peers
			}
			if sub.slot != want {
				t.Fatalf("second tick sub %d is from peer %d, want %d", i, sub.slot, want)
			}
		}
	})
}

// The running size of the 0x0A, counted as the message type byte plus the body,
// has to stay below 0xF7FF. A chunk that would take it to 0xF7FF or past stays
// queued for the next tick.
func TestAckDrainByteLimitAtTheBoundary(t *testing.T) {
	onePeer := []QueueEntry{{Slot: 1, Participating: true}}
	// Base size 6, then 3 header bytes per chunk: 6 + (3+60000) + (3+b).
	const first = 60000
	const fits = ackMaxSize - 1 - 6 - (3 + first) - 3 // the size lands on 0xF7FE
	if fits != 3474 {
		t.Fatalf("fits = %d: the arithmetic of this test changed", fits)
	}

	var a AckState
	a.QueueDataSub(1, sizedChunk(first, 0xA1))
	a.QueueDataSub(1, sizedChunk(fits, 0xB1))
	body := a.PopTick(onePeer)
	if p := parseAck(t, body); len(p.subs) != 2 {
		t.Fatalf("a second chunk that leaves the size at 0xF7FE must ride the same 0x0A: got %d chunks", len(p.subs))
	}
	if got := 1 + len(body); got != ackMaxSize-1 {
		t.Fatalf("type byte + body = %d, want %d", got, ackMaxSize-1)
	}

	var b AckState
	b.QueueDataSub(1, sizedChunk(first, 0xA2))
	b.QueueDataSub(1, sizedChunk(fits+1, 0xB2))
	body = b.PopTick(onePeer)
	p := parseAck(t, body)
	if len(p.subs) != 1 || len(p.subs[0].payload) != first {
		t.Fatalf("a chunk that would take the size to 0xF7FF must wait: got %d chunks", len(p.subs))
	}
	if len(p.pending) != 0 {
		t.Errorf("pending = %x: peer 1 was flushed on the first pass, a refusal on a later pass lists nobody", p.pending)
	}
	if b.QueueEmpty(1) {
		t.Fatal("the refused chunk must stay queued")
	}
	body = b.PopTick(onePeer)
	if p := parseAck(t, body); len(p.subs) != 1 || len(p.subs[0].payload) != fits+1 {
		t.Fatalf("the next tick must carry the chunk the first one could not: %d chunks", len(p.subs))
	}
}

// A chunk too big for any 0x0A is never taken. On the pass that builds the
// pending list its peer is listed instead, exactly as a peer with nothing
// queued is.
func TestAckDrainListsAPeerWhoseChunkDoesNotFit(t *testing.T) {
	onePeer := []QueueEntry{{Slot: 1, Participating: true}}
	var a AckState
	a.QueueDataSub(1, sizedChunk(ackMaxSize, 0xCC))
	for tick := range 2 {
		body := a.PopTick(onePeer)
		if want := []byte{0xfe, 0x00, 0x01, 0x01, 0x00, 0x00}; !bytes.Equal(body, want) {
			t.Fatalf("tick %d = %d bytes starting %x, want %x", tick+1, len(body), body[:min(len(body), 8)], want)
		}
	}
	if a.QueueEmpty(1) {
		t.Fatal("the oversized chunk must stay queued, not be dropped")
	}
}

// The checksum block goes in only if the size stays below 0xF7FF with it. If it
// does not fit, the merged entries wait for a 0x0A with room.
func TestAckChecksumBlockByteLimitAtTheBoundary(t *testing.T) {
	onePeer := []QueueEntry{{Slot: 1, Participating: true}}
	entry := record(0x01020304, 0x05060708)
	// 6 + (3+payload) + 0 pending + (3+8) for one entry.
	const fits = ackMaxSize - 1 - 6 - 3 - 11
	if fits != 63466 {
		t.Fatalf("fits = %d: the arithmetic of this test changed", fits)
	}

	var a AckState
	a.QueueEntries(1, entry)
	a.ReconcileChecksums(onePeer)
	a.QueueDataSub(1, sizedChunk(fits, 0xD1))
	body := a.PopTick(onePeer)
	if p := parseAck(t, body); !bytes.Equal(p.block, checksumBlock(entry)) {
		t.Fatalf("a block that leaves the size at 0xF7FE must ride: block %x", p.block)
	}
	if got := 1 + len(body); got != ackMaxSize-1 {
		t.Fatalf("type byte + body = %d, want %d", got, ackMaxSize-1)
	}

	var b AckState
	b.QueueEntries(1, entry)
	b.ReconcileChecksums(onePeer)
	b.QueueDataSub(1, sizedChunk(fits+1, 0xD2))
	if p := parseAck(t, b.PopTick(onePeer)); p.block != nil {
		t.Fatalf("a block that would take the size to 0xF7FF must wait: block %x", p.block)
	}
	want := append([]byte{0xfe, 0x00, 0x01, 0x01, 0x0b, 0x00}, checksumBlock(entry)...)
	if body := b.PopTick(onePeer); !bytes.Equal(body, want) {
		t.Fatalf("the next 0x0A = %x, want %x: the entry that did not fit rides the next tick", body, want)
	}
}

// ---------------------------------------------------------------------------
// The reconciliation's reference entry
// ---------------------------------------------------------------------------

// The reference is the first active peer whose front entry is not (0,0). A
// leading (0,0) is skipped, so (0,0) against X agrees and X is merged.
func TestReconcileLeadingZeroEntryIsAWildcard(t *testing.T) {
	s := joinedIn(t, 1, 1, 2)
	x := record(0x11223344, 0x55667788)
	s.QueueGameData(1, chunk([]byte{0xaa}))
	s.QueueChecksumEntries(1, record(0, 0))
	s.QueueChecksumEntries(2, x)

	if _, notice := s.PopAckTick(); notice != nil {
		t.Fatalf("(0,0) against X raised desync notice %x", notice)
	}
	if s.DesyncCount != 0 {
		t.Fatalf("DesyncCount = %d, want 0", s.DesyncCount)
	}
	body, _ := s.PopAckTick()
	if p := parseAck(t, body); !bytes.Equal(p.block, checksumBlock(x)) {
		t.Fatalf("checksum block = %x, want X merged: %x", p.block, checksumBlock(x))
	}
}

// Only a LEADING (0,0) is skipped. Once the reference is set, every later
// active peer is compared with it, and (0,0) is a value like any other.
func TestReconcileZeroEntryAfterTheReferenceDisagrees(t *testing.T) {
	var a AckState
	a.QueueEntries(1, record(9, 9))
	a.QueueEntries(2, record(0, 0))
	if got := a.ReconcileChecksums(twoPeers); got != reconcileDesynced {
		t.Fatalf("X against a later (0,0) = %d, want reconcileDesynced", got)
	}
}

// With every front at (0,0) there is no reference to find, and the merged entry
// is (0,0).
func TestReconcileAllZeroEntriesMergeZero(t *testing.T) {
	var a AckState
	a.QueueEntries(1, record(0, 0))
	a.QueueEntries(2, record(0, 0))
	if got := a.ReconcileChecksums(twoPeers); got != reconcileAgreed {
		t.Fatalf("all (0,0) = %d, want reconcileAgreed", got)
	}
	if !bytes.Equal(a.merged, record(0, 0)) {
		t.Fatalf("merged = %x, want one (0,0) entry", a.merged)
	}
}

// A round consumes the front entry of EVERY peer's queue, active or not, so a
// peer outside the started game cannot pile up checksum entries that no round
// ever takes.
func TestReconcilePopsEveryRecordEvenANonParticipatingOne(t *testing.T) {
	entries := []QueueEntry{
		{Slot: 1, Participating: true},
		{Slot: 2, Participating: true},
		{Slot: 3, Participating: false},
	}
	var a AckState
	a.QueueEntries(1, record(5, 6))
	a.QueueEntries(2, record(5, 6))
	a.QueueEntries(3, append(record(1, 1), record(2, 2)...))
	if got := a.ReconcileChecksums(entries); got != reconcileAgreed {
		t.Fatalf("round = %d, want reconcileAgreed", got)
	}
	if got := a.pendingEntries(3); got != 1 {
		t.Fatalf("the non-participating peer has %d entries after a round, want 1: the round pops every queue", got)
	}
}

// ---------------------------------------------------------------------------
// The sync-timeout stamp
// ---------------------------------------------------------------------------

// A checksum upload from a peer of the session refreshes the peer's sync-timeout
// stamp, as a game-data upload does. A peer that keeps sending checksums every
// ten seconds is not reaped at t0+61 s, although its last game data is from t0.
func TestSessionChecksumUploadRefreshesTheSyncTimeoutStamp(t *testing.T) {
	t0 := time.Now()
	for _, sendChecksums := range []bool{false, true} {
		t.Run(fmt.Sprintf("checksums=%t", sendChecksums), func(t *testing.T) {
			s := NewSession(1)
			s.AddPeer(NewPeer(1, &mockConn{}))
			now := t0
			s.Ack.now = func() time.Time { return now }
			s.QueueGameData(1, chunk([]byte{0xaa, 0xbb}))
			s.PopAckTick() // drains the queue, so the timeout applies
			if sendChecksums {
				for i := 1; i <= 6; i++ {
					now = t0.Add(time.Duration(i) * 10 * time.Second)
					s.QueueChecksumEntries(1, record(uint32(i), uint32(i)))
				}
			}
			got := s.SyncTimedOutPeers(t0.Add(61 * time.Second))
			if sendChecksums && len(got) != 0 {
				t.Fatalf("SyncTimedOutPeers = %v at t0+61 s, want none: the last checksum upload was at t0+60 s", got)
			}
			if !sendChecksums && (len(got) != 1 || got[0] != 1) {
				t.Fatalf("SyncTimedOutPeers = %v at t0+61 s with no upload since t0, want [1]", got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// The desync notice
// ---------------------------------------------------------------------------

// rawNotice is the uncompressed notice payload: [01][ref u32 LE][count u16 LE]
// and then {peer u8, d0 u32 LE, d1 u32 LE} for every active peer, where ref is
// the reference entry's second dword.
func rawNotice(ref uint32, members ...[]byte) []byte {
	out := []byte{0x01}
	out = binary.LittleEndian.AppendUint32(out, ref)
	out = binary.LittleEndian.AppendUint16(out, uint16(len(members)))
	for _, m := range members {
		out = append(out, m...)
	}
	return out
}

// member is one {peer, d0, d1} element of the notice.
func member(slot byte, d0, d1 uint32) []byte {
	return append([]byte{slot}, record(d0, d1)...)
}

// readNotice checks the notice framing, [FB][len u16 LE][flag][payload], and
// returns the flag and the payload with any compression undone.
func readNotice(t *testing.T, notice []byte) (flag byte, raw []byte) {
	t.Helper()
	if len(notice) < 4 || notice[0] != 0xFB {
		t.Fatalf("not a desync notice: %x", notice)
	}
	if got := int(binary.LittleEndian.Uint16(notice[1:])); got != len(notice)-3 {
		t.Fatalf("notice length field = %d, and %d bytes follow it", got, len(notice)-3)
	}
	flag = notice[3]
	switch flag {
	case 0:
		return flag, notice[4:]
	case 1:
		r, err := zlib.NewReader(bytes.NewReader(notice[4:]))
		if err != nil {
			t.Fatalf("compressed notice payload is not a zlib stream: %v", err)
		}
		raw, err = io.ReadAll(r)
		if err != nil {
			t.Fatalf("inflate the notice payload: %v", err)
		}
		return flag, raw
	}
	t.Fatalf("notice flag = %d, want 0 or 1", flag)
	return 0, nil
}

// A disagreeing round builds a GameData notice whose body is
// [FB][len u16 LE][flag][payload]. The payload lists every active peer's front
// entry and the reference's second dword. Two peers make 25 payload bytes, which
// zlib does not make smaller, so they go raw with flag 0.
func TestAckDesyncNoticeCarriesTheRoundsChecksumList(t *testing.T) {
	var a AckState
	a.QueueEntries(1, record(0x0A0B0C0D, 0x01020304))
	a.QueueEntries(2, record(0xF0E0D0C0, 0x99887766))
	if got := a.ReconcileChecksums(twoPeers); got != reconcileDesynced {
		t.Fatalf("round = %d, want reconcileDesynced", got)
	}
	payload := rawNotice(0x01020304,
		member(1, 0x0A0B0C0D, 0x01020304),
		member(2, 0xF0E0D0C0, 0x99887766))
	want := []byte{0xFB}
	want = binary.LittleEndian.AppendUint16(want, uint16(1+len(payload)))
	want = append(want, 0x00)
	want = append(want, payload...)
	if got := a.DesyncNotice(); !bytes.Equal(got, want) {
		t.Fatalf("notice = %x,\nwant     %x", got, want)
	}
}

// Every disagreeing round builds its own notice, from that round's front
// entries. A round that stops builds none.
func TestAckDesyncNoticeOnePerDisagreeingRound(t *testing.T) {
	var a AckState
	a.QueueEntries(1, append(record(1, 10), record(2, 20)...))
	a.QueueEntries(2, append(record(3, 30), record(4, 40)...))
	wants := [][]byte{
		rawNotice(10, member(1, 1, 10), member(2, 3, 30)),
		rawNotice(20, member(1, 2, 20), member(2, 4, 40)),
	}
	for round, want := range wants {
		if got := a.ReconcileChecksums(twoPeers); got != reconcileDesynced {
			t.Fatalf("round %d = %d, want reconcileDesynced", round+1, got)
		}
		if _, raw := readNotice(t, a.DesyncNotice()); !bytes.Equal(raw, want) {
			t.Fatalf("round %d notice payload = %x, want %x", round+1, raw, want)
		}
	}
	if got := a.ReconcileChecksums(twoPeers); got != reconcileStop {
		t.Fatalf("third round = %d, want reconcileStop", got)
	}
	if got := a.DesyncNotice(); got != nil {
		t.Fatalf("a round that stopped left notice %x", got)
	}
}

// The list holds every ACTIVE peer in join order, the skipped (0,0) one
// included, and nobody outside the started game. The reference is the first
// non-(0,0) entry.
func TestAckDesyncNoticeListsEveryActivePeer(t *testing.T) {
	entries := []QueueEntry{
		{Slot: 4, Participating: true},
		{Slot: 2, Participating: true},
		{Slot: 7, Participating: false},
		{Slot: 3, Participating: true},
	}
	var a AckState
	a.QueueEntries(4, record(0, 0))
	a.QueueEntries(2, record(0x100, 0x200))
	a.QueueEntries(7, record(0x5, 0x6))
	a.QueueEntries(3, record(0x100, 0x201))
	if got := a.ReconcileChecksums(entries); got != reconcileDesynced {
		t.Fatalf("round = %d, want reconcileDesynced", got)
	}
	want := rawNotice(0x200, member(4, 0, 0), member(2, 0x100, 0x200), member(3, 0x100, 0x201))
	if _, raw := readNotice(t, a.DesyncNotice()); !bytes.Equal(raw, want) {
		t.Fatalf("notice payload = %x,\nwant           %x", raw, want)
	}
}

// When zlib at level 1 makes the payload strictly smaller, the notice carries
// the compressed form with flag 1. Ten peers whose entries repeat compress well.
func TestAckDesyncNoticeCompressesWhenSmaller(t *testing.T) {
	var entries []QueueEntry
	var a AckState
	var members [][]byte
	for i := range 10 {
		slot := byte(i + 1)
		entries = append(entries, QueueEntry{Slot: slot, Participating: true})
		d1 := uint32(0x7000)
		if i == 9 {
			d1 = 0x7001
		}
		a.QueueEntries(slot, record(0x12345678, d1))
		members = append(members, member(slot, 0x12345678, d1))
	}
	if got := a.ReconcileChecksums(entries); got != reconcileDesynced {
		t.Fatalf("round = %d, want reconcileDesynced", got)
	}
	notice := a.DesyncNotice()
	flag, raw := readNotice(t, notice)
	want := rawNotice(0x7000, members...)
	if flag != 1 {
		t.Fatalf("flag = %d, want 1: %d raw bytes compress", flag, len(want))
	}
	if len(notice)-4 >= len(want) {
		t.Fatalf("compressed payload is %d bytes, not smaller than the %d raw ones", len(notice)-4, len(want))
	}
	if !bytes.Equal(raw, want) {
		t.Fatalf("inflated payload = %x,\nwant             %x", raw, want)
	}
}
