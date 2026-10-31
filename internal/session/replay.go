package session

import "sync"

// Frame is one stored message of the frame store. The type travels with the
// body because a stored packet holds the whole serialized message, and both
// readers put it back on the wire unchanged: the observer update and the
// reconnect resend send the stored object, not a body re-wrapped in a type of
// this server's choosing.
type Frame struct {
	Type byte
	Body []byte
}

// frameOverhead is the 57 bytes the store charges to every frame it keeps, on
// top of the message size. It is what a frame costs against its block, and it is
// why a block holds fewer frames than its byte count divided by the frame size.
const frameOverhead = 57

// Replay is the per-session frame store.
//
// It is a list of blocks, and that is not a detail of the implementation: what an
// observer joining late can be given is exactly the window the store still
// holds, so the shape of the store is the shape of the answer the client gets.
//
//   - A block holds blockSize bytes: the -pageSize setting, default 65536.
//   - The store keeps at most maxBlocks blocks, ceil(peers * f) with peers
//     floored at 8, where f is the -maxPagesToStorePerSession scale, 1.0 by
//     default. 8 blocks in any real match.
//   - The oldest block is dropped when a new block is needed and the count is
//     already at the cap, so the retained window is the most recent
//     maxBlocks*blockSize bytes and never more.
//   - maxFrames is an optional cap in frames: 60000/ticksPerSec+1 until the game
//     announces itself observable with maxObservers > 0, which is what AoE2DE
//     does, so in practice it does not bound anything.
//
// Base() is the absolute number of the first frame still held, which both the
// StartObserveRequest handler and the observer update read.
type Replay struct {
	mu        sync.RWMutex
	frames    []Frame
	total     int   // frames ever appended, the store's frame numbering
	base      int   // the absolute number of frames[0]
	blockSize int
	maxBlocks int
	maxFrames int // <= 0 disables the frame cap
	blockAt   []int // absolute number of the first frame of each live block
	blockUsed int   // bytes charged to the last (open) block
}

// NewReplay creates a frame store of at most maxBlocks blocks of blockSize
// bytes, with an optional cap of maxFrames retained frames (<= 0 disables it).
func NewReplay(maxBlocks, blockSize, maxFrames int) *Replay {
	if maxBlocks < 1 {
		maxBlocks = 1
	}
	if blockSize < 1 {
		blockSize = 1
	}
	return &Replay{maxBlocks: maxBlocks, blockSize: blockSize, maxFrames: maxFrames}
}

// MaxPages reports the store's byte window, which is what the cap means now that
// the store is bounded in bytes rather than in frames.
func (r *Replay) MaxPages() int { return r.maxBlocks * r.blockSize }

// BlockSize is the -pageSize value the store was built with.
func (r *Replay) BlockSize() int { return r.blockSize }

// Total returns the number of frames ever appended. It is the observer range's
// exclusive upper bound: the observer update serves up to it and
// ActivateObserver measures the match length against it.
func (r *Replay) Total() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.total
}

// Retained returns how many frames the store still holds.
func (r *Replay) Retained() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.frames)
}

// Base returns the absolute number of the oldest retained frame. It is 0 while
// the store still holds everything and grows once the first block is dropped.
func (r *Replay) Base() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.base
}

// AddFrame stores one message. The caller keeps ownership of nothing: the body
// is copied, because both callers hand over a buffer they are about to send or
// return to a pool.
func (r *Replay) AddFrame(msgType byte, body []byte) {
	f := Frame{Type: msgType, Body: append([]byte(nil), body...)}
	need := len(f.Body) + frameOverhead

	r.mu.Lock()
	defer r.mu.Unlock()
	switch {
	case len(r.blockAt) == 0:
		// First frame ever: it opens the first block.
		r.blockAt = append(r.blockAt, r.total)
	case need+r.blockUsed > r.blockSize:
		// The last block is full. A new block is opened, and the oldest one
		// goes when the count is already at the cap, which is the only place the
		// store ever forgets anything.
		if len(r.blockAt) >= r.maxBlocks {
			r.dropOldestBlockLocked()
		}
		r.blockAt = append(r.blockAt, r.total)
		r.blockUsed = 0
	}
	if r.maxFrames > 0 && r.total-r.base+1 > r.maxFrames {
		r.trimToMaxFramesLocked()
	}
	r.frames = append(r.frames, f)
	r.blockUsed += need
	r.total++
}

// dropOldestBlockLocked forgets the oldest live block. Called with the lock held
// and only when a new block is about to be opened, which is the only place the
// store ever forgets anything.
//
// The block is recycled rather than freed: its entry list is cleared, it is
// popped from the array and the emptied block is pushed back at the end, so the
// blocks are a ring. With a cap of one the window therefore starts at the frame
// that is about to be added.
func (r *Replay) dropOldestBlockLocked() {
	if len(r.blockAt) == 0 {
		return
	}
	r.blockAt = r.blockAt[1:]
	if len(r.blockAt) == 0 {
		r.base = r.total
		r.frames = r.frames[:0]
		return
	}
	drop := r.blockAt[0] - r.base
	r.base = r.blockAt[0]
	r.frames = append(r.frames[:0], r.frames[drop:]...)
}

// trimToMaxFramesLocked drops whole blocks from the front until the retained
// frame count is back inside the cap. The cap is applied only when a block is
// opened, and whole blocks are dropped, so the window it leaves is a
// block-quantum shorter than the cap.
func (r *Replay) trimToMaxFramesLocked() {
	for len(r.blockAt) > 1 && r.total-r.base+1 > r.maxFrames {
		r.dropOldestBlockLocked()
	}
}

// GetFrames returns the retained frames whose absolute numbers fall in
// [first, last] inclusive, inclusive on both ends.
//
// It returns nil when the range is not wholly inside what the store holds, and
// it does NOT clamp: a last at or past the frame count and a first below the
// oldest retained frame are both refused, and both callers treat the refusal as
// an observer that cannot be updated or as a failed resend, leaving their own
// counter where it was. Clamping here would invent frames that are never sent.
//
// The returned slice shares the store's frames; callers must not modify them and
// must not retain them across an AddFrame call.
func (r *Replay) GetFrames(first, last int) []Frame {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.frames) == 0 || first < 0 || last < first {
		return nil
	}
	if last >= r.total || first < r.base {
		return nil
	}
	return r.frames[first-r.base : last-r.base+1]
}
