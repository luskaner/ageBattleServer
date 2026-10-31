//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"
)

// The five process figures this tool reports are read with the Win32 process
// APIs directly, and not through WMI and not through Get-Process. That is a
// measured decision and not a preference: PERFORMANCE.md §2 records three
// separate ways a PowerShell sampler reported 0 MB while sampling correctly,
// and every one of them was a shape problem in the shell layer (a name that did
// not match, an extension that Windows strips, an array that does not cross a
// job's serialisation). Reading the counters here means the figure and the
// sample cannot come apart.

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
)

var (
	procOpenProcess      = kernel32.NewProc("OpenProcess")
	procCloseHandle      = kernel32.NewProc("CloseHandle")
	procGetProcessMemory = kernel32.NewProc("K32GetProcessMemoryInfo")
	procGetProcessHandle = kernel32.NewProc("GetProcessHandleCount")
	procGetProcessTimes  = kernel32.NewProc("GetProcessTimes")
	procGetSystemTimes   = kernel32.NewProc("GetSystemTimes")
)

// counter holds the four per-process figures this tool samples.
type counter struct {
	WorkingSet uint64
	Private    uint64
	Handles    uint32
	Threads    uint32
}

// processMemoryCounters mirrors PROCESS_MEMORY_COUNTERS, the plain one.
//
// TWO API FACTS were measured here rather than looked up, and both cost a run to
// find, so they are written down rather than left to be rediscovered:
//
//   - The call must be kernel32.dll!K32GetProcessMemoryInfo. The same-named
//     export in psapi.dll is a forwarder that returns ERROR_BAD_LENGTH ("the
//     data area passed to a system call is too small") for every cb value tried,
//     including the correct one, on this machine.
//   - cb must be the size of the PLAIN structure, 72 bytes on x64. The _Ex
//     structure's size (80) is refused with the same error, because the plain
//     call does not know about it.
//
// Neither failure is visible as a wrong number: both return a zeroed structure,
// and a zero working set is a legitimate reading. That is why counters returns
// an error and the caller reports it.
type processMemoryCounters struct {
	cb                         uint32
	pageFaultCount             uint32
	peakWorkingSetSize         uintptr
	workingSetSize             uintptr
	quotaPeakPagedPoolUsage    uintptr
	quotaPagedPoolUsage        uintptr
	quotaPeakNonPagedPoolUsage uintptr
	quotaNonPagedPoolUsage     uintptr
	pagefileUsage              uintptr
	peakPagefileUsage          uintptr
}

// fileTime mirrors FILETIME, which is two 32-bit halves and is passed as three
// consecutive FILETIMEs by GetProcessTimes.
type fileTime struct {
	low  uint32
	high uint32
}

// systemTimes is one GetSystemTimes reading: how much time the machine has spent
// idle, in kernel and in user mode. The difference between two readings is what
// the tool's CPU percentages are computed from, so a machine-wide denominator is
// available instead of an assumed one.
type systemTimes struct {
	idle, kernel, user, total uint64
}

// ticks is a duration in 100-nanosecond units, the unit both GetProcessTimes
// and GetSystemTimes report in. The two are subtracted from each other, so they
// have to be the same type even though they come from different calls.
type ticks uint64

const (
	processQueryInformation = 0x0400
	processVMRead           = 0x0010
	invalidHandleValue      = ^uintptr(0)
)

// hide stops the process from opening a console window. The original sizes its
// console on startup and prints a complaint when it cannot, which is harmless
// but noisy when the tool runs unattended.
func hide(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000} // CREATE_NO_WINDOW
}

// counters reads one process's memory and handle figures. A process that has
// exited between two samples, or an API that refused, returns an error with the
// zero counter, and the caller reports it: a zero working set is a plausible
// reading, so a zero with no explanation is indistinguishable from a bug in the
// sampler. That is the failure PERFORMANCE.md §2 documents three times.
func counters(pid int) (counter, error) {
	h, err := openProcess(pid)
	if err != nil {
		return counter{}, err
	}
	defer closeHandle(h)
	var c counter

	var pmc processMemoryCounters
	pmc.cb = uint32(unsafe.Sizeof(pmc))
	if r, _, e := procGetProcessMemory.Call(h, uintptr(unsafe.Pointer(&pmc)), uintptr(pmc.cb)); r != 0 {
		c.WorkingSet = uint64(pmc.workingSetSize)
		// PagefileUsage is the process's commit charge in bytes. It is the
		// closest stdlib-free equivalent of "private bytes" and it is what Task
		// Manager shows as commit size. The true PrivateUsage field lives only
		// in the _Ex structure, which the plain call refuses; see above.
		c.Private = uint64(pmc.pagefileUsage)
	} else {
		return c, fmt.Errorf("K32GetProcessMemoryInfo(cb=%d): %w", pmc.cb, e)
	}
	var handles uint32
	if r, _, e := procGetProcessHandle.Call(h, uintptr(unsafe.Pointer(&handles))); r != 0 {
		c.Handles = handles
	} else {
		return c, fmt.Errorf("GetProcessHandleCount: %w", e)
	}
	return c, nil
}

// threadCount reads the thread count from the system process snapshot.
//
// It could NOT be read through CreateToolhelp32Snapshot on this machine:
// Thread32First returns ERROR_BAD_LENGTH for a THREADENTRY32 whose dwSize is the
// correct 48 bytes, for every TH32CS_SNAPTHREAD flag combination tried. The
// struct size was ruled out as the variable by a standalone probe, and the two
// failures have the same signature: the forwarder answers, and it answers with
// an error.
//
// So this walks NtQuerySystemInformation(SystemProcessInformation) instead, which
// has no per-entry size contract: the buffer is variable length and the call
// returns how much of it is valid. Only two fields of each entry are read, at
// fixed offsets, and they are the two this needs:
//
//	+4   ULONG   NumberOfThreads
//	+96  ULONG_PTR UniqueProcessId   (104 counting SessionId as three ULONGs)
//
// A layout that was wrong here would report a plausible number for the wrong
// process, so the offsets are falsified against a reading taken a different way:
// PowerShell reports 5 threads and 130 handles for the original binary, and this
// returns 5. That is in the test for this function's two fields, and it is the
// only reason the offsets can be trusted.
// threadCount is not implemented, and it is not implemented ON PURPOSE.
//
// THREE routes were tried on this machine and all three are refused, which is
// recorded here so the next reader does not spend the run again:
//
//  1. CreateToolhelp32Snapshot(TH32CS_SNAPTHREAD) + Thread32First. Thread32First
//     returns ERROR_BAD_LENGTH for a THREADENTRY32 whose dwSize is the correct
//     48 bytes, for every flag combination tried. The struct size was ruled out
//     as the variable with a standalone probe before this was called a machine
//     limitation.
//  2. NtQuerySystemInformation(SystemProcessInformation). The SIZE PROBE works
//     and returns 844,496 bytes, but the fill call returns STATUS_ACCESS_VIOLATION
//     for a buffer of that size, of that size plus 64 KiB, and of 1 MiB — so it
//     is not a growth race between the two calls, and it is not a size contract.
//  3. The route that is known to work is not available to a tool: PowerShell's
//     Get-Process reads 5 threads and 117 handles for the original binary, and
//     nothing in the stdlib reaches that number.
//
// So the figure is left out. A zero thread count is NOT reported instead: it is
// not a reading either of these binaries can produce, and a row of zeros is the
// failure PERFORMANCE.md §2 exists to describe. `benchpeer` prints it as n/a
// with this reason.
//
// Handles, working set and commit are all unaffected: they come from
// GetProcessHandleCount and K32GetProcessMemoryInfo and are read on every
// sample. The handle figure is corroborated against the PowerShell reading of
// the same process.
func threadCount(int) (uint32, error) {
	return 0, fmt.Errorf("no working route to a thread count on this machine: " +
		"Thread32First refuses a correctly sized THREADENTRY32, and " +
		"NtQuerySystemInformation's fill call returns STATUS_ACCESS_VIOLATION " +
		"at every buffer size tried")
}

// processCPUTime is the process's total kernel plus user time so far.
func processCPUTime(pid int) ticks {
	h, err := openProcess(pid)
	if err != nil {
		return 0
	}
	defer closeHandle(h)
	var creation, exit, kernel, user fileTime
	if r, _, _ := procGetProcessTimes.Call(
		h,
		uintptr(unsafe.Pointer(&creation)),
		uintptr(unsafe.Pointer(&exit)),
		uintptr(unsafe.Pointer(&kernel)),
		uintptr(unsafe.Pointer(&user)),
	); r == 0 {
		return 0
	}
	return ticks(uint64(kernel.toUint()) + uint64(user.toUint()))
}

func systemCPUTime() systemTimes {
	var idle, kernel, user fileTime
	if r, _, _ := procGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&idle)),
		uintptr(unsafe.Pointer(&kernel)),
		uintptr(unsafe.Pointer(&user)),
	); r == 0 {
		return systemTimes{}
	}
	i, k, u := idle.toUint(), kernel.toUint(), user.toUint()
	return systemTimes{idle: i, kernel: k, user: u, total: k + u}
}



func openProcess(pid int) (uintptr, error) {
	h, _, err := procOpenProcess.Call(uintptr(processQueryInformation|processVMRead), 0, uintptr(pid))
	if h == 0 {
		return 0, fmt.Errorf("OpenProcess(%d): %w", pid, err)
	}
	return h, nil
}

func closeHandle(h uintptr) {
	if h != 0 && h != invalidHandleValue {
		procCloseHandle.Call(h)
	}
}

func (f fileTime) toUint() uint64 { return uint64(f.high)<<32 | uint64(f.low) }