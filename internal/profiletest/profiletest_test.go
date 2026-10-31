package profiletest

import (
	"bytes"
	"compress/gzip"
	"os"
	"runtime/pprof"
	"testing"
	"runtime"
	"time"
)

// A hand-built profile is the only way to pin the field numbers and the
// "records are not samples" rule without trusting a real one, because a real one
// only exists as something pprof has to be asked about.
//
// The three cases are the three ways this went wrong: counting field 1 (which is
// sample_type) gives 2 instead of 300; counting Sample RECORDS instead of summing
// Sample.value[0] gives 3 instead of 300; and not decompressing gives 0.
func TestSampleCountReadsValuesNotRecords(t *testing.T) {
	var prof bytes.Buffer
	// Two sample_type values (field 1), which is what Go writes and what must NOT
	// be counted: {"samples","count"} and {"cpu","nanoseconds"}.
	writeLengthDelimited(&prof, 1, append(append([]byte{0x08}, 0x01), 0x02))
	// Three Sample records (field 2), each carrying value = [100, 1000].
	for range 3 {
		var sample bytes.Buffer
		writeVarintField(&sample, 1, 7)  // location_id
		writeVarintField(&sample, 2, 100) // value[0] = samples
		writeVarintField(&sample, 2, 1000)
		writeLengthDelimited(&prof, 2, sample.Bytes())
	}
	// A location (field 4) and a varint field 9, which is time_nanos: the field a
	// walker that skips by moving one byte desynchronises on.
	writeLengthDelimited(&prof, 4, []byte{0x08, 0x01})
	writeVarintField(&prof, 9, 1234567890123)

	if got := SampleCount(prof.Bytes()); got != 300 {
		t.Errorf("SampleCount = %d, want 300: three Sample records carrying value[0] = 100 each. Counting the "+
			"records would give 3 and counting field 1 would give 2, so this pins which of the two it is", got)
	}
}

// SampleCount on an empty profile is 0 and not a panic: that is the case the drivers
// above it are asserting against, so it has to be a value they can compare.
func TestSampleCountOnEmptyAndGarbage(t *testing.T) {
	if got := SampleCount(nil); got != 0 {
		t.Errorf("SampleCount(nil) = %d, want 0", got)
	}
	if got := SampleCount([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}); got != 0 {
		t.Errorf("SampleCount on a varint that never terminates = %d, want 0", got)
	}
	if got := SampleCount([]byte{0x0a, 0x7f, 0x01}); got != 0 {
		t.Errorf("SampleCount on a length-delimited field that runs off the end = %d, want 0", got)
	}
	if got := SampleCount([]byte{0x1f, 0x8b, 0x00}); got != 0 {
		t.Errorf("SampleCount on a truncated gzip stream = %d, want 0", got)
	}
}

// TestSampleCountOnARealProfile is the one that matters: a profile the RUNTIME
// wrote, counted by this code and by `go tool pprof -sample_index=samples`, which
// have to agree.
//
// The comparison with pprof itself cannot run here (it is a subprocess and its text
// output is not an API), so what this asserts is the part that can: the count is
// positive, it is at most the number of ticks' worth of profiling the window could
// have produced, and it is stable across a re-read. The agreement with pprof is
// recorded in SESSION-JOURNAL.md 39, measured on six files.
func TestSampleCountOnARealProfile(t *testing.T) {
	var buf bytes.Buffer
	// The rate is raised so the count is a three-figure number rather than the ~25 a
	// 250 ms window gives at the default 100 Hz. StartCPUProfile's own call to set
	// 100 is refused, which is expected and harmless; what matters is the count.
	runtime.SetCPUProfileRate(1000)
	if err := pprof.StartCPUProfile(&buf); err != nil {
		t.Fatalf("StartCPUProfile: %v", err)
	}
	// A quarter of a second of real work at a raised rate. The rate is restored by
	// StopCPUProfile.
	deadline := time.Now().Add(250 * time.Millisecond)
	x := 0
	for time.Now().Before(deadline) {
		for i := range 4000 {
			x += i
		}
	}
	pprof.StopCPUProfile()
	if x == 0 {
		t.Fatal("the loop optimised itself away, so the profile is of nothing")
	}

	raw := buf.Bytes()
	if len(raw) < 2 || raw[0] != 0x1f || raw[1] != 0x8b {
		t.Fatalf("the runtime wrote %d bytes that do not start with the gzip magic %x %x, so the decompression "+
			"this relies on is not happening and SampleCount would read 0", len(raw), raw[0], raw[1])
	}
	got := SampleCount(raw)
	// The MAGNITUDE of the count is deliberately not asserted, and that is a repair
	// rather than a style choice. This test used to require 100 samples from a 250 ms
	// window and it failed once in one full-suite run (1 failure, at 0,57 s against
	// 0,28 s standalone) and then passed 8 times standalone and 5 more times beside
	// the other packages. The cause is NOT established and is not guessed at here.
	//
	// What it can be is the runtime's achieved sampling rate: a thread that gets
	// descheduled delivers fewer profiler ticks in the same wall time, and the full
	// suite is the one place in this repo where four packages run flat out for
	// thirteen minutes (internal/sim replays captures in real time). Asserting how
	// many samples the machine managed to take is asserting the machine, which is
	// the shape of flake SESSION-JOURNAL.md 38.4 is about, and this repo treats a test
	// that fails once in a full-suite run as a broken test however innocent it looks.
	//
	// So the decoder is proved by the things that do not depend on throughput:
	//
	//   - TestSampleCountReadsValuesNotRecords pins the field numbers and the
	//     values-not-records rule on a hand-built profile, exactly;
	//   - Contains below finds THIS TEST's own function in the profile, which is a
	//     positive decode of a profile the runtime wrote, and is the assertion that
	//     actually catches a broken field walk;
	//   - the count is stable across two reads and across the bytes on disk;
	//   - and the ceiling below catches the walk summing the wrong field.
	//
	// The floor that remains is 1, because a decode that fails reads exactly 0 and
	// that is the only zero worth stopping for.
	if got < 1 {
		t.Fatalf("SampleCount = %d on a profile of a quarter second of real work, so the decode returned nothing. "+
			"The file starts %x %x, so this is the decompression or the field walk, not the profiler", got,
			raw[0], raw[1])
	}
	// One second of samples is the hard ceiling whatever the rate, and this window is
	// a quarter of one. This one IS load-independent: only a walk summing the wrong
	// field can exceed it.
	if got > 5000 {
		t.Errorf("SampleCount = %d on a 250 ms profile, over the 5000 that is more than 5000 samples a second. A "+
			"count that high means the walk is summing the wrong field", got)
	}
	if again := SampleCount(raw); again != got {
		t.Errorf("SampleCount is not stable across two reads of the same bytes: %d then %d", got, again)
	}

	// And the same bytes on disk, which is where the drivers actually read them
	// from: gzip's reader is not consumable twice, so a driver that kept the
	// decompressed buffer would read 0 the second time. Writing and re-reading is
	// what catches that.
	path := t.TempDir() + "/cpu.prof"
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	fromDisk, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if disk := SampleCount(fromDisk); disk != got {
		t.Errorf("SampleCount on the bytes read back from disk = %d, want the %d of the same bytes in memory", disk, got)
	}
	// Contains has to see this test's own function, which is the only non-runtime
	// code in the profile, and it has to see it through the gzip the same way.
	if !Contains(raw, "profiletest.TestSampleCountOnARealProfile") {
		t.Error(`Contains did not find this test's own function in a profile of this test, so it is not looking ` +
			"where the function names are")
	}
	if Contains(raw, "no.such.function.anywhere") {
		t.Error("Contains found a name that is not in the profile")
	}
	if Contains(nil, "anything") {
		t.Error("Contains on nil = true")
	}
}

// writeLengthDelimited writes one length-delimited protobuf field.
func writeLengthDelimited(buf *bytes.Buffer, field uint64, payload []byte) {
	buf.WriteByte(byte(field<<3 | 2))
	writeVarint(buf, uint64(len(payload)))
	buf.Write(payload)
}

// writeVarintField writes one varint protobuf field.
func writeVarintField(buf *bytes.Buffer, field uint64, v uint64) {
	buf.WriteByte(byte(field << 3))
	writeVarint(buf, v)
}

func writeVarint(buf *bytes.Buffer, v uint64) {
	for v >= 0x80 {
		buf.WriteByte(byte(v) | 0x80)
		v >>= 7
	}
	buf.WriteByte(byte(v))
}

// A gzip round trip is what Go 1.27 does to a profile, and it is why SampleCount
// decompresses: this is the case where a profile exists on disk and reads as 0.
//
// The sample is built as a LENGTH-DELIMITED field 2 holding value[0], because that
// is the shape a Sample has. Writing field 2 as a bare varint reads 0 here, and that
// is the fourth way to get this wrong: field 2 with wire type 0 is not a Sample.
func TestSampleCountDecompresses(t *testing.T) {
	var sample bytes.Buffer
	writeVarintField(&sample, 2, 42)
	var plain bytes.Buffer
	writeLengthDelimited(&plain, 2, sample.Bytes())
	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	if _, err := zw.Write(plain.Bytes()); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}
	if got := SampleCount(gz.Bytes()); got != 42 {
		t.Errorf("SampleCount on a gzipped profile = %d, want 42", got)
	}
	// And the same bytes uncompressed, to pin that the decompression is the only
	// thing standing between this and a zero.
	if got := SampleCount(plain.Bytes()); got != 42 {
		t.Errorf("SampleCount on the same profile uncompressed = %d, want 42", got)
	}
}
