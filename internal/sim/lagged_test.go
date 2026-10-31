package sim_test

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/luskaner/ageBattleServer/internal/sim"
)

// What the original does with an observer that does not present the session's
// credential, measured on the committed captures instead of assumed.
//
// The rule is one expression in the binary: the observer is lagged unless its
// StartObserveRequest's passwordHash is exactly the u64 the players published
// with 0x1D, so "lagged" and "the client sent a hash" are opposites. This server
// used to decide it the second way, and every observer in every committed
// capture sits on the wrong side of that difference, so this is the measurement
// that decides which is right for the client that actually plays:
//
//   - 0x1D (SetObservableParamatersMessage) carries [u32 observable][u64 hash]
//     [u32 lag in seconds], and the original keeps that u64 as the session's
//     observable credential.
//   - 0x1C (StartObserveRequest Vx) carries the passwordHash at offset 16 of
//     its body.
//
// It also pins the third field of that 0x1D, the lag, because it is why the
// harness never saw the difference: a lag of 0 makes the lagged target equal the
// session's frame count, so a wrong lagged flag and a wrong target produce the
// same bytes in every capture that exists. That is a fact about the evidence,
// not a licence to keep the bug.
func TestTheCommittedObserversAreLaggedUnderTheOriginalsRule(t *testing.T) {
	dirs, err := filepath.Glob("testdata/*")
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(dirs)

	type captureFacts struct {
		dir        string
		hashes     []uint64
		lags       []uint32
		hashesSeen int
		requests   []uint64
	}
	var facts []captureFacts
	for _, dir := range dirs {
		if _, err := os.Stat(filepath.Join(dir, "metadata.json")); err != nil {
			continue
		}
		cap, err := sim.Load(dir)
		if err != nil {
			t.Fatalf("%s: %v", dir, err)
		}
		f := captureFacts{dir: dir}
		for _, c := range cap.Conns {
			for _, m := range c.Up {
				switch m.Type {
				case 0x1D:
					if len(m.Body) < 16 {
						t.Fatalf("%s: 0x1D with %d bytes", dir, len(m.Body))
					}
					if binary.LittleEndian.Uint32(m.Body[0:4]) == 0 {
						continue // not announcing the session observable
					}
					f.hashesSeen++
					f.hashes = append(f.hashes, binary.LittleEndian.Uint64(m.Body[4:12]))
					f.lags = append(f.lags, binary.LittleEndian.Uint32(m.Body[12:16]))
				case 0x1C:
					if len(m.Body) < 24 {
						t.Fatalf("%s: 0x1C with %d bytes", dir, len(m.Body))
					}
					f.requests = append(f.requests, binary.LittleEndian.Uint64(m.Body[16:24]))
				}
			}
		}
		facts = append(facts, f)
	}

	// Both sides have to be there or the comparison says nothing: a capture with
	// no 0x1D and one with no 0x1C each prove nothing about the rule.
	withBoth := 0
	for _, f := range facts {
		if len(f.hashes) == 0 || len(f.requests) == 0 {
			continue
		}
		withBoth++
		published := f.hashes[0]
		if published == 0 {
			t.Errorf("%s: 0x1D published hash 0, want a non-zero credential", f.dir)
		}
		for _, h := range f.hashes[1:] {
			if h != published {
				t.Errorf("%s: 0x1D published two different hashes (%#x and %#x)", f.dir, published, h)
			}
		}
		for _, lag := range f.lags {
			if lag != 0 {
				t.Errorf("%s: 0x1D published a lag of %d seconds, want 0: that is why the lagged "+
					"target is invisible in this capture", f.dir, lag)
			}
		}
		for _, req := range f.requests {
			lagged := published == 0 || req != published
			if !lagged {
				t.Errorf("%s: an observer presented the published hash %#x, so it is NOT lagged; "+
					"no capture has shown that and this test was written to record that none has",
					f.dir, req)
			}
		}
	}
	if withBoth == 0 {
		t.Fatal("no capture carries both a 0x1D and a 0x1C: this test would be comparing nothing")
	}
	for _, f := range facts {
		if len(f.hashes) > 0 || len(f.requests) > 0 {
			fmt.Printf("lagged-fact %s: 0x1D x%d hash=%v lag=%v 0x1C x%d passwordHash=%v\n",
				f.dir, len(f.hashes), f.hashes, f.lags, len(f.requests), f.requests)
		}
	}
}