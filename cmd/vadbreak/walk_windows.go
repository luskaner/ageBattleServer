//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

const processQueryInformation = 0x0400
const processVMSize = 0x0008

type memBasicInformation struct {
	BaseAddress       uintptr
	AllocationBase    uintptr
	AllocationProtect uint32
	_                 uint32
	RegionSize        uintptr
	State             uint32
	Protect           uint32
	Type              uint32
	_                 uint32
}

var (
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	procOpenProcess  = kernel32.NewProc("OpenProcess")
	procCloseHandle  = kernel32.NewProc("CloseHandle")
	procVirtualQuery = kernel32.NewProc("VirtualQueryEx")
	procSystemInfo   = kernel32.NewProc("GetNativeSystemInfo")
	procProcessMem   = kernel32.NewProc("K32GetProcessMemoryInfo")

	// The counters come from kernel32 and not from the psapi forwarder, which refuses
	// every cb on this machine, and cb must be the plain structure's 72 bytes rather
	// than a size derived from a pointer-width field. Both facts were measured, and
	// both failures are invisible as a wrong number because each returns a zeroed
	// structure, which is why PrivateUsageKnown is carried separately from the value.
	// Kept in sync with cmd/benchpeer/procinfo_windows.go.
	errNoPrivateUsage = fmt.Errorf("K32GetProcessMemoryInfo returned a zeroed structure")
)

type processMemoryCounters struct {
	cb                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
}

const sizeOfProcessMemoryCounters = 72

// systemInfo mirrors the head of SYSTEM_INFO: wProcessorArchitecture and wReserved are
// two WORDs and dwPageSize is the DWORD at offset 4. Reading dwPageSize through a pointer-
// sized field would land on lpMinimumApplicationAddress instead, which on 64-bit Windows is
// 0x10000, and the walk would then step over 64 KB at a time wherever it could not describe
// a region, which is how a report ends up missing the region it was looking for.
type systemInfo struct {
	_         uint16
	_         uint16
	pageSize  uint32
	_         uintptr
	_         uintptr
	_         uintptr
}

func pageSize() uintptr {
	var si systemInfo
	procSystemInfo.Call(uintptr(unsafe.Pointer(&si)))
	if si.pageSize == 0 {
		return 4096
	}
	return uintptr(si.pageSize)
}

// commitChargeBytes reads the figure benchpeer reports, so one run of this tool and one
// run of benchpeer are talking about the same number and the two can be reconciled.
//
// It is PagefileUsage, the process's commit charge, and not PrivateUsage: PrivateUsage
// lives in PROCESS_MEMORY_COUNTERS_EX and not in the plain 72-byte structure, and the
// counters for this process have to come from kernel32 with cb set to the plain structure's
// 72 bytes. The column benchpeer labels "commit / private" is therefore the commit charge.
// The quota fields are a third figure again and are not this one.
func commitChargeBytes(handle syscall.Handle) (uint64, bool) {
	var pmc processMemoryCounters
	pmc.cb = sizeOfProcessMemoryCounters
	ret, _, _ := procProcessMem.Call(uintptr(handle), uintptr(unsafe.Pointer(&pmc)), sizeOfProcessMemoryCounters)
	if ret == 0 {
		return 0, false
	}
	if pmc.PagefileUsage == 0 {
		return 0, false
	}
	return uint64(pmc.PagefileUsage), true
}

// walk totals the committed and reserved-only bytes of a process's address space by
// region type and protection.
//
// Two limits bound the walk and both are properties of the request, not of the target: a
// 64-bit Windows address space is walked up to 2^47, and a region that VirtualQueryEx
// declines to describe is stepped over one granule at a time rather than ending the walk,
// because ending early would silently under-report.
func walk(pid int) (*Report, error) {
	handle, _, err := procOpenProcess.Call(processQueryInformation|processVMSize, 0, uintptr(pid))
	if handle == 0 {
		return nil, fmt.Errorf("OpenProcess(%d): %w", pid, err)
	}
	defer procCloseHandle.Call(handle)

	rep := &Report{PID: pid, PageSize: uint64(pageSize())}
	if v, ok := commitChargeBytes(syscall.Handle(handle)); ok {
		rep.CommitChargeBytes = v
		rep.CommitChargeKnown = true
	}

	gran := pageSize()
	var addr uintptr
	const limit = uintptr(1) << 47
	for addr < limit {
		var mbi memBasicInformation
		ret, _, _ := procVirtualQuery.Call(handle, addr, uintptr(unsafe.Pointer(&mbi)), unsafe.Sizeof(mbi))
		if ret == 0 || mbi.RegionSize == 0 {
			addr += gran
			continue
		}
		rep.add(regionTypeName(mbi.Type), protectionName(mbi.Protect), mbi.State, uint64(mbi.RegionSize))
		addr = mbi.BaseAddress + mbi.RegionSize
	}
	rep.finish()
	return rep, nil
}


