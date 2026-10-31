//go:build !windows

package main

import (
	"fmt"
	"runtime"
)

// The walk is VirtualQueryEx, which is a Win32 call. On any other platform the figure is
// absent, and it is reported as absent: a zero here would be a number that reads as a
// measurement and is not one.
func walk(pid int) (*Report, error) {
	return nil, fmt.Errorf("the address-space walk is Win32-only; this build is %s/%s", runtime.GOOS, runtime.GOARCH)
}
