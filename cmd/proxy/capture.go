package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// The capture layout, hardcoded and relative to the working directory:
//
//	captures/2026-09-28_034112_65de70a/
//	    main.log          <- T1, the client <-> primary flow
//	    secondary.log     <- T2, the client <-> secondary flow
//	    proxy.log         <- the proxy's own log
//	    metadata.json     <- commit, dirty, timestamp, and who each side was
//	    COMMIT.txt        <- the same provenance as plain text
//
// The names never vary and never carry a number. Which log is ours is decided by
// CONTENT and recorded in metadata.json, so the file names carry no meaning and
// there is nothing left to confuse: it used to be main_log_NN.csv and
// secondary_log_NN.csv side by side in one directory, with NN chosen by the
// manager and the two binaries swapped from capture to capture.
const (
	MainLogName      = "main.log"
	SecondaryLogName = "secondary.log"
	ProxyLogName     = "proxy.log"
	MetadataName     = "metadata.json"
	CommitName       = "COMMIT.txt"
	CapturesRoot     = "captures"
)

// LogMeta is deliberately one field. Reading a capture only ever needs to answer
// "is this our traffic or the original's", and the file name does not say, so the
// answer has to be recorded somewhere.
//
// It is a tri-state and not a bool on purpose: "original" and "we could not tell"
// are different facts, and collapsing them with a bool would make a capture whose
// binary could not be identified look decided. It is not. When the answer is
// Unknown the capture cannot be used for comparison, and that has to be visible.
//
// The sha256, the size and the path of each binary are NOT repeated here: they
// belong to the capture, not to each log, and they already live in COMMIT.txt and
// in CaptureMeta.Exes.
type LogMeta struct {
	File   string   `json:"file"`
	Writer Identity `json:"writer"` // "ours" | "original" | "unknown"
}

// CaptureMeta is metadata.json: the provenance of the run plus, once, who each
// side was. The per-log files only repeat the writer so a log can be read on its
// own.
type CaptureMeta struct {
	FormatVersion int    `json:"format_version"`
	StartedAt     string `json:"started_at"`
	Commit        string `json:"commit"`
	Dirty         string `json:"dirty"`
	Note          string `json:"note,omitempty"`
	Primary       struct {
		Writer  Identity `json:"writer"`
		ExePath string   `json:"exe_path"`
		SHA256  string   `json:"sha256"`
	} `json:"primary"`
	Secondary struct {
		Writer  Identity `json:"writer"`
		ExePath string   `json:"exe_path"`
		SHA256  string   `json:"sha256"`
	} `json:"secondary"`
	// Args are the exact command lines each side was launched with. A simulation
	// test needs them to rebuild the same server: -simulationPeriod, -pageSize and
	// the rest change behaviour, so a test that guessed them would be testing a
	// configuration nobody ran.
	PrimaryArgs   []string `json:"primary_args"`
	SecondaryArgs []string `json:"secondary_args"`
	Warning       string   `json:"warning,omitempty"`
}

// CaptureDir is the one place the layout is defined: captures/<date>_<commit>.
func CaptureDir() string {
	name := time.Now().Format("2006-01-02_150405") + "_" + shortCommit(BuildCommit)
	return filepath.Join(CapturesRoot, name)
}

// NewCaptureDir creates the directory and returns it.
func NewCaptureDir() (string, error) {
	dir := CaptureDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return dir, err
	}
	return dir, nil
}

// WriteCaptureMeta writes metadata.json, the two <name>.meta.json sidecars and
// COMMIT.txt. It runs before any logging starts, so a capture always exists as a
// complete, self-describing unit even if the run dies immediately.
func WriteCaptureMeta(dir string, roles Roles, args1, args2 []string) error {
	started := nowStamp()
	meta := CaptureMeta{
		FormatVersion: 1,
		StartedAt:     started,
		Commit:        BuildCommit,
		Dirty:         BuildDirty,
		Warning:       roles.Warning,
	}
	meta.Primary.Writer = roles.Primary.Exe.Identity
	meta.Primary.ExePath = roles.Primary.Exe.Path
	meta.Primary.SHA256 = roles.Primary.Exe.SHA256
	meta.Secondary.Writer = roles.Secondary.Exe.Identity
	meta.Secondary.ExePath = roles.Secondary.Exe.Path
	meta.Secondary.SHA256 = roles.Secondary.Exe.SHA256
	meta.PrimaryArgs = args1
	meta.SecondaryArgs = args2
	if BuildCommit == "dev" {
		meta.Note = "built without -ldflags: commit unknown. Use scripts/deploy.ps1."
	}

	// One sidecar per log, same base name, so a log can be read on its own.
	for _, l := range []LogMeta{
		{File: MainLogName, Writer: roles.Primary.Exe.Identity},
		{File: SecondaryLogName, Writer: roles.Secondary.Exe.Identity},
	} {
		sidecar := l.File[:len(l.File)-len(filepath.Ext(l.File))] + ".meta.json"
		if err := writeJSON(filepath.Join(dir, sidecar), l); err != nil {
			return err
		}
	}
	if err := writeJSON(filepath.Join(dir, MetadataName), meta); err != nil {
		return err
	}
	return WriteCommitFile(dir, roles.Primary.Exe, roles.Secondary.Exe)
}

func writeJSON(path string, v any) error {
	body, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0o644)
}

// DescribeCapture is the one line for proxy.log, so a capture can be identified
// from that file alone.
func DescribeCapture(dir string, roles Roles) string {
	return fmt.Sprintf("capture %s | %s | main.log=%s secondary.log=%s",
		dir, StampTo(), roles.Primary.Exe.Identity, roles.Secondary.Exe.Identity)
}
