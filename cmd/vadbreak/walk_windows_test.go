//go:build windows

package main

import (
	"os"
	"runtime"
	"syscall"
	"testing"
)

// These tests live behind the windows build tag because they call into the Win32 walk
// (pageSize, OpenProcess, the commit charge), which exists only in walk_windows.go. A
// runtime skip in main_test.go cannot stop those references from failing to compile
// everywhere else.

// dwPageSize sits at offset 4 of SYSTEM_INFO, after two WORDs. Reading it through a
// pointer-sized field lands on lpMinimumApplicationAddress, which on 64-bit Windows is
// 0x10000, and the walk then steps over 64 KB at a time wherever it cannot describe a
// region: the totals still print and a region goes missing, so nothing looks wrong. This
// is the assertion that catches it.
func TestPageSizeIsReadFromTheRightOffset(t *testing.T) {
	want := uintptr(4096)
	if runtime.GOARCH == "arm64" {
		want = 65536
	}
	if got := pageSize(); got != want {
		t.Errorf("page size = %d, want %d for %s/%s; SYSTEM_INFO.dwPageSize is at offset 4 and a pointer-sized read of it yields lpMinimumApplicationAddress",
			got, want, runtime.GOOS, runtime.GOARCH)
	}
}

// The commit charge is the figure benchpeer reports, so the two tools can be reconciled.
// Reading a quota field instead would return a number three orders of magnitude smaller
// and would still look like a measurement.
func TestCommitChargeIsTheFigureBenchpeerReports(t *testing.T) {
	h, _, err := procOpenProcess.Call(processQueryInformation|processVMSize, 0, uintptr(os.Getpid()))
	if h == 0 {
		t.Fatalf("OpenProcess(self): %v", err)
	}
	defer procCloseHandle.Call(h)

	got, ok := commitChargeBytes(syscall.Handle(h))
	if !ok {
		t.Fatal("commit charge reported as unavailable for this process")
	}
	// A Go test binary is tens of megabytes of committed memory, and a quota field would
	// be well under one. The bound is loose on purpose: what it rules out is the wrong
	// field, not a change in size.
	if got < 8<<20 {
		t.Errorf("commit charge = %d bytes (%.2f MB), which is too small to be a commit charge and looks like a quota field",
			got, float64(got)/mib)
	}
}
