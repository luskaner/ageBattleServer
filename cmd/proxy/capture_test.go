package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The layout is hardcoded and the names never vary. If this test has to be
// edited to accommodate a new name, the names have stopped being a contract.
func TestCaptureLayoutIsFixed(t *testing.T) {
	dir, err := NewCaptureDir()
	if err != nil {
		t.Fatalf("NewCaptureDir: %v", err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	if !strings.Contains(filepath.ToSlash(dir), "captures/") {
		t.Errorf("capture dir must be under ./captures/, got %q", dir)
	}
	// <fecha>_<commit>: the date, then an underscore, then the short commit.
	leaf := filepath.Base(dir)
	parts := strings.Split(leaf, "_")
	if len(parts) != 3 {
		t.Fatalf("capture dir name must be <date>_<time>_<commit>, got %q", leaf)
	}
	if len(parts[0]) != 10 || parts[0][4] != '-' || parts[0][7] != '-' {
		t.Errorf("first part must be a date, got %q", parts[0])
	}
	if len(parts[1]) != 6 {
		t.Errorf("second part must be a time, got %q", parts[1])
	}
	if parts[2] == "" {
		t.Error("third part must be the commit")
	}
}

func TestWriteCaptureMetaProducesTheContract(t *testing.T) {
	dir, err := NewCaptureDir()
	if err != nil {
		t.Fatalf("NewCaptureDir: %v", err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	ours := filepath.Join(dir, "ours.exe")
	copyFile(t, os.Args[0], ours)
	roles := Roles{
		Primary:   Side{Target: 1, LogFile: MainLogName, Exe: mustIdentify(t, ours)},
		Secondary: Side{Target: 2, LogFile: SecondaryLogName, Exe: ExeInfo{Identity: Original, SHA256: "ff", Path: "orig.exe", Size: 1}},
	}
	args1 := []string{"-relaybroadcastPort", "0", "-bsPort", "14001", "-simulationPeriod", "125"}
	args2 := []string{"-relaybroadcastPort", "0", "-bsPort", "15001", "-simulationPeriod", "125"}
	if err := WriteCaptureMeta(dir, roles, args1, args2); err != nil {
		t.Fatalf("WriteCaptureMeta: %v", err)
	}

	// metadata.json
	body, err := os.ReadFile(filepath.Join(dir, MetadataName))
	if err != nil {
		t.Fatalf("%s missing: %v", MetadataName, err)
	}
	var meta CaptureMeta
	if err := json.Unmarshal(body, &meta); err != nil {
		t.Fatalf("%s is not valid json: %v", MetadataName, err)
	}
	if meta.Commit != BuildCommit {
		t.Errorf("metadata commit = %q, want %q", meta.Commit, BuildCommit)
	}
	// Here the PRIMARY is the Go binary, so main.log is ours even though the name
	// says nothing about it. That is the property worth pinning: the file name
	// never implies the side, the recorded writer does.
	if meta.Primary.Writer != Ours {
		t.Errorf("primary writer = %q, want %q", meta.Primary.Writer, Ours)
	}
	if meta.Secondary.Writer != Original {
		t.Errorf("secondary writer = %q, want %q", meta.Secondary.Writer, Original)
	}
	// The binary details belong to the capture, once, not repeated per log.
	if meta.Primary.SHA256 == "" {
		t.Error("metadata.json must record the sha256 of each binary once")
	}
	// The launch args are what let a simulation test rebuild the same server.
	// -simulationPeriod and -pageSize change behaviour, so a test that guessed
	// them would be testing a configuration nobody ever ran.
	if len(meta.PrimaryArgs) != 6 || meta.PrimaryArgs[4] != "-simulationPeriod" {
		t.Errorf("primary args not recorded: %v", meta.PrimaryArgs)
	}
	if len(meta.SecondaryArgs) != 6 {
		t.Errorf("secondary args not recorded: %v", meta.SecondaryArgs)
	}

	// One metadata sidecar per log, same base name, one field each.
	for _, want := range []struct {
		base   string
		writer Identity
	}{
		{"main", Ours},
		{"secondary", Original},
	} {
		p := filepath.Join(dir, want.base+".meta.json")
		body, err := os.ReadFile(p)
		if err != nil {
			t.Errorf("sidecar %s missing: %v", want.base, err)
			continue
		}
		var l LogMeta
		if err := json.Unmarshal(body, &l); err != nil {
			t.Errorf("%s is not valid json: %v", want.base, err)
			continue
		}
		if l.File != want.base+".log" {
			t.Errorf("%s describes %q", want.base, l.File)
		}
		if l.Writer != want.writer {
			t.Errorf("%s writer = %q, want %q", want.base, l.Writer, want.writer)
		}
	}
	// A bool would have said "original" for a binary we could not identify, which
	// is a different and much more dangerous claim.
	if _, err := os.Stat(filepath.Join(dir, CommitName)); err != nil {
		t.Errorf("%s missing: %v", CommitName, err)
	}
}

func mustIdentify(t *testing.T, p string) ExeInfo {
	t.Helper()
	i, err := IdentifyExe(p)
	if err != nil {
		t.Fatalf("IdentifyExe: %v", err)
	}
	return i
}
