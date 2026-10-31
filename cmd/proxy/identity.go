package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Identity is what a BattleServer binary actually IS, decided by reading the
// file. The file name is worthless on its own and has been wrong: a file called
// BattleServerOrig.exe has held our build (2026-09-28) while BattleServerReimpl.exe
// held the original, and the reverse was true on 2026-08-21. Every capture has to
// record which is which, because guessing it cost three wrong conclusions about
// the StartObserve byte.
type Identity string

const (
	// Ours is the reimplementation: a Go binary. Detected by the Go build markers
	// the toolchain always embeds, which the original (an MSVC C++ binary) lacks.
	Ours Identity = "ours"
	// Original is the shipped MSVC binary.
	Original Identity = "original"
	// Unknown means neither marker was found, so nothing may be concluded.
	Unknown Identity = "unknown"
)

var (
	goMarkers   = []string{"Go build ID:", "runtime.main", "go1."}
	msvcMarkers = []string{"Microsoft Visual C++ runtime", "MSVCP", "VCRUNTIME"}
)

// ExeInfo is the fingerprint of one BattleServer binary.
type ExeInfo struct {
	Path     string   `json:"path"`
	Identity Identity `json:"identity"`
	SHA256   string   `json:"sha256"`
	Size     int64    `json:"size"`
	Why      string   `json:"why"`
}

// IdentifyExe fingerprints a binary. It never trusts the file name.
func IdentifyExe(path string) (ExeInfo, error) {
	info := ExeInfo{Path: path, Identity: Unknown}
	data, err := os.ReadFile(path)
	if err != nil {
		info.Why = "unreadable: " + err.Error()
		return info, err
	}
	sum := sha256.Sum256(data)
	info.SHA256 = hex.EncodeToString(sum[:])
	info.Size = int64(len(data))
	body := string(data)

	if hasAny(body, goMarkers) {
		info.Identity = Ours
		info.Why = "contains the Go build markers " + strings.Join(goMarkers, ", ")
		return info, nil
	}
	if hasAny(body, msvcMarkers) {
		info.Identity = Original
		info.Why = "contains the MSVC markers " + strings.Join(msvcMarkers, ", ")
		return info, nil
	}
	info.Why = "no Go and no MSVC marker: nothing can be concluded"
	return info, nil
}

func hasAny(body string, markers []string) bool {
	for _, m := range markers {
		if strings.Contains(body, m) {
			return true
		}
	}
	return false
}

// Side is one half of a capture: which binary served it, and which log holds it.
type Side struct {
	// Target is 1 (primary, main_log_NN.csv) or 2 (secondary, secondary_log_NN.csv).
	Target   int     `json:"target"`
	LogFile  string  `json:"log_file"`
	Exe      ExeInfo `json:"exe"`
	OursSide bool    `json:"is_ours"`
	Verified bool    `json:"identity_verified"`
}

func nowStamp() string { return time.Now().Format("2006-01-02 15:04:05") }

// Roles is the manifest written next to every capture so the two logs can never
// be confused again.
type Roles struct {
	Capture   string    `json:"capture"`
	WrittenAt time.Time `json:"written_at"`
	Method    string    `json:"method"`
	Primary   Side      `json:"primary"`
	Secondary Side      `json:"secondary"`
	Warning   string    `json:"warning,omitempty"`
}

// WriteRoles identifies both binaries and writes roles.json next to the capture.
// The launcher calls it before starting anything, so a capture can never exist
// without a record of who was who.
func WriteRoles(dir string, exe1, exe2, mainLog, secondaryLog string) (Roles, error) {
	i1, err1 := IdentifyExe(exe1)
	i2, err2 := IdentifyExe(exe2)
	roles := Roles{
		Capture:   filepath.Base(dir),
		WrittenAt: time.Now(),
		Method:    "binary content (Go build markers => ours, MSVC markers => original); file names are not trusted",
		Primary: Side{
			Target: 1, LogFile: mainLog, Exe: i1,
			OursSide: i1.Identity == Ours, Verified: i1.Identity != Unknown,
		},
		Secondary: Side{
			Target: 2, LogFile: secondaryLog, Exe: i2,
			OursSide: i2.Identity == Original, Verified: i2.Identity != Unknown,
		},
	}
	switch {
	case err1 != nil || err2 != nil:
		roles.Warning = fmt.Sprintf("could not identify: primary=%v secondary=%v", err1, err2)
	case roles.Primary.Exe.Identity == roles.Secondary.Exe.Identity:
		roles.Warning = "both sides have the SAME identity: this capture cannot be used for comparison"
	case roles.Primary.Exe.Identity == Unknown || roles.Secondary.Exe.Identity == Unknown:
		roles.Warning = "at least one side is unknown: this capture cannot be used for comparison"
	}

	body, err := json.MarshalIndent(roles, "", "  ")
	if err != nil {
		return roles, err
	}
	return roles, os.WriteFile(filepath.Join(dir, "roles.json"), append(body, '\n'), 0o644)
}

// Describe returns the one line to print at startup, so the log itself says who
// is who instead of leaving it to be inferred later.
func (r Roles) Describe() string {
	return fmt.Sprintf(
		"CAPTURE %s | main_log=%s is %s (%s) | secondary_log=%s is %s (%s)",
		r.Capture, r.Primary.LogFile, r.Primary.Exe.Identity, short(r.Primary.Exe.SHA256),
		r.Secondary.LogFile, r.Secondary.Exe.Identity, short(r.Secondary.Exe.SHA256))
}

func short(sha string) string {
	if len(sha) > 16 {
		return sha[:16]
	}
	return sha
}
