package main

import (
	"encoding/json"
	"os"
	"runtime"
	"strings"
	"testing"
)

// The classification is what turns a walk into an attribution, so it is pinned rather
// than trusted: these are the three numbers that decide the whole report, and a change in
// any of them changes what the tool says about a server.
func TestRegionTypeNamesSeparateImageFromPrivate(t *testing.T) {
	cases := []struct {
		in   uint32
		want string
	}{
		{memImage, "MEM_IMAGE"},
		{memPrivate, "MEM_PRIVATE"},
		{memMapped, "MEM_MAPPED"},
		{0, "MEM_UNKNOWN_0x0"},
	}
	for _, c := range cases {
		if got := regionTypeName(c.in); got != c.want {
			t.Errorf("regionTypeName(0x%x) = %q, want %q", c.in, got, c.want)
		}
	}
}

// Only the low base-protection byte is classified. The guard and modifier bits above it
// are dropped, so a region is not split into several rows over bits that do not change what
// the region is for. WRITECOPY is kept as its own name because a dirty image page is the
// case the tool exists to attribute.
func TestProtectionNameDropsGuardAndModifierBits(t *testing.T) {
	cases := []struct {
		in   uint32
		want string
	}{
		{0x02, "READONLY"},         // PAGE_READONLY
		{0x04, "READWRITE"},        // PAGE_READWRITE
		{0x08, "WRITECOPY"},        // PAGE_WRITECOPY
		{0x20, "EXECUTE_READ"},     // PAGE_EXECUTE_READ
		{0x40, "EXECUTE_READWRITE"},
		{0x40000004, "READWRITE"},  // + PAGE_GUARD
		{0x10000004, "READWRITE"},  // + PAGE_NOCACHE
		{0x01, "NOACCESS"},
		{0x00, "PROT_UNKNOWN_0x0"}, // a region nobody has written to yet
		{0xff, "PROT_UNKNOWN_0xff"},
	}
	for _, c := range cases {
		if got := protectionName(c.in); got != c.want {
			t.Errorf("protectionName(0x%x) = %q, want %q", c.in, got, c.want)
		}
	}
}

// MEM_FREE is not a claim about the process and must not be counted as its memory, and
// neither is a value that is a region type rather than a state.
func TestAddIgnoresStatesOtherThanCommitAndReserve(t *testing.T) {
	r := &Report{PID: 1}
	for _, state := range []uint32{0x10000, memImage, memPrivate, memMapped, 0x9999} {
		r.add("MEM_PRIVATE", "READWRITE", state, 4096)
	}
	if r.CommittedTotal != 0 {
		t.Fatalf("a non-state was counted as committed: %d bytes", r.CommittedTotal)
	}
	if len(r.Buckets) != 0 {
		t.Fatalf("got %d rows from ignored states: %+v", len(r.Buckets), r.Buckets)
	}

	r.add("MEM_PRIVATE", "READWRITE", memReserve, 8192)
	if r.CommittedTotal != 0 {
		t.Fatalf("a reservation was counted as committed: %d bytes", r.CommittedTotal)
	}
	if r.Buckets[0].Reserved != 8192 {
		t.Errorf("reserved = %d, want 8192", r.Buckets[0].Reserved)
	}
}

// A committed and a reserved region of the same type and protection are one row with two
// numbers, not two rows: the committed figure is the cost and the reserved figure is the
// address space, and reading them apart is the whole distinction this tool draws.
func TestAddMergesCommitAndReserveIntoOneRow(t *testing.T) {
	r := &Report{PID: 1}
	r.add("MEM_PRIVATE", "READWRITE", memReserve, 5*1024*1024*1024)
	r.add("MEM_PRIVATE", "READWRITE", memCommit, 4096)
	r.add("MEM_PRIVATE", "READWRITE", memCommit, 8192)
	if len(r.Buckets) != 1 {
		t.Fatalf("got %d rows, want 1: %+v", len(r.Buckets), r.Buckets)
	}
	if r.Buckets[0].Committed != 4096+8192 {
		t.Errorf("committed = %d, want 12288", r.Buckets[0].Committed)
	}
	if r.Buckets[0].Reserved != 5*1024*1024*1024 {
		t.Errorf("reserved = %d, want %d", r.Buckets[0].Reserved, 5*1024*1024*1024)
	}
	if r.Buckets[0].Regions != 2 {
		t.Errorf("regions = %d, want 2", r.Buckets[0].Regions)
	}
}

// The Go runtime reserves a large part of the address space and commits almost none of
// it, so a report that cannot print the two separately will read a reservation as a cost.
// The ordering invariant is that committed sorts first within a tie on type, because the
// cost is the number an operator pays.
func TestFinishSortsByCommittedAndTotalsBothFiguresSeparately(t *testing.T) {
	r := &Report{PID: 1}
	r.add("MEM_IMAGE", "EXEC", memCommit, 4*1024*1024)
	r.add("MEM_PRIVATE", "READWRITE", memCommit, 2*1024*1024)
	r.add("MEM_MAPPED", "READONLY", memCommit, 1024*1024)
	r.add("MEM_PRIVATE", "NOACCESS", memReserve, 5*1024*1024*1024)
	r.finish()

	if got, want := r.CommittedTotal, uint64(7*1024*1024); got != want {
		t.Errorf("committed total = %d, want %d", got, want)
	}
	if got, want := r.ReservedTotal, uint64(5*1024*1024*1024); got != want {
		t.Errorf("reserved total = %d, want %d", got, want)
	}
	if r.Buckets[0].Type != "MEM_IMAGE" || r.Buckets[3].Type != "MEM_PRIVATE" {
		t.Errorf("rows not ordered by committed size: %+v", r.Buckets)
	}
}

// A figure the tool could not read is printed as absent. It is never rendered as a zero,
// because a zero here is indistinguishable from a measurement.
func TestRenderSaysUnavailableRatherThanZero(t *testing.T) {
	r := &Report{PID: 1, CommittedTotal: 8 << 20, PageSize: 4096}
	r.add("MEM_PRIVATE", "READWRITE", memCommit, 8<<20)
	r.finish()
	out := r.render()
	if !strings.Contains(out, "UNAVAILABLE") {
		t.Errorf("render did not report the missing figure as unavailable:\n%s", out)
	}
	if !strings.Contains(out, "never as 0") {
		t.Errorf("render did not say how it reports the missing figure:\n%s", out)
	}
	if strings.Contains(out, "commit charge (PagefileUsage) 0.00 MB") {
		t.Errorf("render printed the missing figure as a zero:\n%s", out)
	}
}

func TestRenderPrintsPrivateUsageWhenItWasRead(t *testing.T) {
	r := &Report{PID: 1, CommitChargeBytes: 12 << 20, CommitChargeKnown: true, PageSize: 4096}
	r.add("MEM_IMAGE", "WRITECOPY", memCommit, 4<<20)
	r.add("MEM_PRIVATE", "READWRITE", memCommit, 12<<20)
	r.finish()
	out := r.render()
	if !strings.Contains(out, "benchpeer reports") {
		t.Errorf("render does not tie the figure to the other tool:\n%s", out)
	}
	if !strings.Contains(out, "+4.00 MB") {
		t.Errorf("render does not show the image/DLL difference:\n%s", out)
	}
}

// A total filled in before the walk must not survive it, or every figure in the report is
// counted twice.
func TestFinishRecomputesTotalsRatherThanAccumulatingThem(t *testing.T) {
	r := &Report{PID: 1, CommittedTotal: 16 << 20, ReservedTotal: 99 << 20}
	r.add("MEM_IMAGE", "WRITECOPY", memCommit, 4<<20)
	r.add("MEM_PRIVATE", "READWRITE", memCommit, 12<<20)
	r.add("MEM_PRIVATE", "NOACCESS", memReserve, 1024<<20)
	r.finish()

	if got, want := r.CommittedTotal, uint64(16<<20); got != want {
		t.Errorf("committed total = %d, want %d", got, want)
	}
	if got, want := r.ReservedTotal, uint64(1024<<20); got != want {
		t.Errorf("reserved total = %d, want %d", got, want)
	}
}

func TestReportMarshalsToJSON(t *testing.T) {
	r := &Report{PID: 99, CommittedTotal: 4096, PageSize: 4096, CommitChargeKnown: true, CommitChargeBytes: 2048}
	r.add("MEM_IMAGE", "EXEC", memCommit, 4096)
	r.finish()
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{`"pid":99`, `"type":"MEM_IMAGE"`, `"commit_charge_known":true`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("JSON missing %s: %s", want, b)
		}
	}
}

// The walk is exercised against this test binary, which is a real process with a real
// image and a real Go heap. It is the same trick cmd/proxy/identity_test.go uses: there is
// no other process to point at, and a tool that cannot read its own address space is not
// a tool.
func TestWalkReadsThisProcess(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("the walk is Win32-only")
	}
	rep, err := walk(os.Getpid())
	if err != nil {
		t.Fatalf("walk(self): %v", err)
	}
	if rep.PID != os.Getpid() {
		t.Errorf("pid = %d, want %d", rep.PID, os.Getpid())
	}
	if rep.CommittedTotal == 0 {
		t.Fatal("committed total is 0 for a running process")
	}
	if rep.PageSize == 0 {
		t.Error("page size is 0")
	}

	var image, private uint64
	for _, b := range rep.Buckets {
		switch b.Type {
		case "MEM_IMAGE":
			image += b.Committed
		case "MEM_PRIVATE":
			private += b.Committed
		}
	}
	if image == 0 {
		t.Error("no committed MEM_IMAGE region: the executable was not attributed")
	}
	if private == 0 {
		t.Error("no committed MEM_PRIVATE region: the heap was not attributed")
	}
	if rep.ReservedTotal == 0 {
		t.Error("no reserved-only region: address space is being reported as cost")
	}

	// The Go runtime reserves address space and commits a small part of it. A report
	// where the reserved figure is not larger than the committed one would mean the two
	// are being conflated, which is the mistake this tool exists to prevent.
	if rep.ReservedTotal <= rep.CommittedTotal {
		t.Errorf("reserved (%.2f MB) is not larger than committed (%.2f MB); the two are probably conflated",
			float64(rep.ReservedTotal)/mib, float64(rep.CommittedTotal)/mib)
	}
}

// An absent PID must be an error and not an empty report. A report of nothing for a
// process that does not exist reads as "this process costs nothing".
func TestWalkOnAbsentPIDIsAnError(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("the walk is Win32-only")
	}
	rep, err := walk(0x7FFFFFF0)
	if err == nil {
		t.Fatalf("walk of an absent pid returned a report: %+v", rep)
	}
	if rep != nil {
		t.Errorf("walk returned a report alongside the error: %+v", rep)
	}
}

func TestWalkOnNonWindowsPlatformIsUnavailableNotZero(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("this asserts the !windows build")
	}
	rep, err := walk(os.Getpid())
	if err == nil {
		t.Fatalf("walk succeeded on %s: %+v", runtime.GOOS, rep)
	}
	if !strings.Contains(err.Error(), "Win32-only") {
		t.Errorf("error does not say the figure is absent: %v", err)
	}
}


