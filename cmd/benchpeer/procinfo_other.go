//go:build !windows

package main

import (
	"errors"
	"os/exec"
)

// This tool exists to compare two Windows binaries on process-level figures, and
// every figure it prints comes from the Win32 process APIs. Off Windows there is
// nothing to report, and returning zeros would be worse than refusing: a zero
// working set and a zero thread count are both legitimate readings on Windows
// and neither can be produced here.

type counter struct {
	WorkingSet uint64
	Private    uint64
	Handles    uint32
	Threads    uint32
}

type systemTimes struct{ idle, kernel, user, total uint64 }

type ticks uint64

func hide(*exec.Cmd) {}

// errNoWin32 is returned instead of a zero reading. The callers already report a
// counter error as the run's error and a thread-count error as "unavailable",
// which is the refusal this file exists to give.
var errNoWin32 = errors.New("the process counters are read with the Win32 APIs, which this platform does not have")

func counters(int) (counter, error) { return counter{}, errNoWin32 }

func threadCount(int) (uint32, error) { return 0, errNoWin32 }

func processCPUTime(int) ticks   { return 0 }
func systemCPUTime() systemTimes { return systemTimes{} }
