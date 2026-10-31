// Command vadbreak reports where a running process's memory is committed.
//
// benchpeer measures how much a process costs. It reports one number, the commit charge,
// and that number has no attribution: it cannot say whether the cost is the heap, the
// executable, or the DLLs, and those three have different remedies. This tool walks a
// process's address space and totals the committed bytes by region type and protection,
// which is the attribution benchpeer does not carry.
//
// It exists because a single figure was read as one cause. A figure without an
// attribution is comparable and not diagnosable.
//
//	go build -o vadbreak ./cmd/vadbreak
//	vadbreak -pid 12345
//	vadbreak -pid 12345 -json
//
// Windows only. On any other platform the walk reports UNAVAILABLE and exits non-zero
// rather than printing a zero, because "measured as zero" and "could not be measured" are
// different facts and a comparison built on the second one is worthless.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
)

type Bucket struct {
	Type      string `json:"type"`
	Protect   string `json:"protect"`
	Committed uint64 `json:"committed_bytes"`
	Reserved  uint64 `json:"reserved_bytes"`
	Regions   int    `json:"regions"`
}

type Report struct {
	PID               int      `json:"pid"`
	CommittedTotal    uint64   `json:"committed_total_bytes"`
	ReservedTotal     uint64   `json:"reserved_total_bytes"`
CommitChargeBytes uint64   `json:"commit_charge_bytes"`
	CommitChargeKnown bool     `json:"commit_charge_known"`
	PageSize          uint64   `json:"page_size"`
	Buckets           []Bucket `json:"buckets"`
	Note              string   `json:"note,omitempty"`
}

// MEM_IMAGE, MEM_MAPPED and MEM_PRIVATE from winnt.h. The three states a region of
// interest can be in.
const (
	memCommit  = 0x1000
	memReserve = 0x2000

	memPrivate = 0x20000
	memMapped  = 0x40000
	memImage   = 0x1000000
)

// regionTypeName is the attribution this tool exists to produce. MEM_IMAGE is the
// executable and its sections, MEM_PRIVATE is memory the process asked the OS for, and
// MEM_MAPPED is somebody else's file, which for a server process means the DLLs.
//
// A dirty image page is reported as MEM_IMAGE with write-copy protection, and that is the
// case that decides most of the commit charge for a Go binary: a page the loader copied in
// order to write to it. It is not the heap and the garbage collector cannot return it.
func regionTypeName(t uint32) string {
	switch t {
	case memImage:
		return "MEM_IMAGE"
	case memPrivate:
		return "MEM_PRIVATE"
	case memMapped:
		return "MEM_MAPPED"
	}
	return fmt.Sprintf("MEM_UNKNOWN_0x%x", t)
}

// protectionName keeps only the low base-protection byte, so the guard and modifier bits
// above it do not fragment one category into several rows. The values are winnt.h's:
// NOACCESS 1, READONLY 2, READWRITE 4, WRITECOPY 8, EXECUTE 0x10, EXECUTE_READ 0x20,
// EXECUTE_READWRITE 0x40, EXECUTE_WRITECOPY 0x80.
//
// WRITECOPY is kept as its own name because a dirty image page is the case the tool
// exists to attribute, and a region the loader has not written to yet carries NOACCESS.
func protectionName(p uint32) string {
	switch p & 0xff {
	case 0x01:
		return "NOACCESS"
	case 0x02:
		return "READONLY"
	case 0x04:
		return "READWRITE"
	case 0x08:
		return "WRITECOPY"
	case 0x10:
		return "EXECUTE"
	case 0x20:
		return "EXECUTE_READ"
	case 0x40:
		return "EXECUTE_READWRITE"
	case 0x80:
		return "EXECUTE_WRITECOPY"
	}
	return fmt.Sprintf("PROT_UNKNOWN_0x%x", p&0xff)
}

func (r *Report) add(typ, prot string, state uint32, size uint64) {
	switch state {
	case memCommit, memReserve:
	default:
		return
	}
	for i := range r.Buckets {
		if r.Buckets[i].Type != typ || r.Buckets[i].Protect != prot {
			continue
		}
		if state == memCommit {
			r.Buckets[i].Committed += size
			r.Buckets[i].Regions++
		} else {
			r.Buckets[i].Reserved += size
		}
		return
	}
	b := Bucket{Type: typ, Protect: prot}
	if state == memCommit {
		b.Committed = size
		b.Regions = 1
	} else {
		b.Reserved = size
	}
	r.Buckets = append(r.Buckets, b)
}

// finish orders the rows and recomputes both totals from them. It recomputes rather than
// accumulates, so a Report whose totals were filled in before the walk cannot end up
// carrying them twice.
func (r *Report) finish() {
	sort.Slice(r.Buckets, func(i, j int) bool {
		if r.Buckets[i].Committed != r.Buckets[j].Committed {
			return r.Buckets[i].Committed > r.Buckets[j].Committed
		}
		if r.Buckets[i].Reserved != r.Buckets[j].Reserved {
			return r.Buckets[i].Reserved > r.Buckets[j].Reserved
		}
		return r.Buckets[i].Type < r.Buckets[j].Type
	})
	r.CommittedTotal = 0
	r.ReservedTotal = 0
	for _, b := range r.Buckets {
		r.CommittedTotal += b.Committed
		r.ReservedTotal += b.Reserved
	}
}

const mib = 1024.0 * 1024.0

func (r *Report) render() string {
	out := fmt.Sprintf("pid %d   committed %.2f MB   reserved-only %.2f MB   page %d\n",
		r.PID, float64(r.CommittedTotal)/mib, float64(r.ReservedTotal)/mib, r.PageSize)

if r.CommitChargeKnown {
		out += fmt.Sprintf("  commit charge (PagefileUsage)      %.2f MB   -- the figure benchpeer reports\n",
			float64(r.CommitChargeBytes)/mib)
		diff := int64(r.CommittedTotal) - int64(r.CommitChargeBytes)
		out += fmt.Sprintf("  committed - commit charge          = %+.2f MB   (image and DLL pages are not private)\n",
			float64(diff)/mib)
	} else {
		out += "  commit charge UNAVAILABLE -- reported as absent, never as 0\n"
	}

	out += "\n  TYPE          PROT          COMMITTED   RESERVED-ONLY  REGIONS\n"
	for _, b := range r.Buckets {
		if b.Committed == 0 && b.Reserved == 0 {
			continue
		}
		out += fmt.Sprintf("  %-12s %-12s %8.2f MB %8.2f MB %8d\n",
			b.Type, b.Protect, float64(b.Committed)/mib, float64(b.Reserved)/mib, b.Regions)
	}
	if r.Note != "" {
		out += "\n  NOTE: " + r.Note + "\n"
	}
	return out
}

func main() {
	pid := flag.Int("pid", 0, "process id to walk; 0 means this process")
	asJSON := flag.Bool("json", false, "emit JSON instead of a table")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: vadbreak [-pid N] [-json]\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	target := *pid
	if target == 0 {
		target = os.Getpid()
	}

	rep, err := walk(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "vadbreak: UNAVAILABLE: %v\n", err)
		os.Exit(2)
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(rep); err != nil {
			fmt.Fprintf(os.Stderr, "vadbreak: %v\n", err)
			os.Exit(1)
		}
		return
	}
	fmt.Print(rep.render())
}


