package main

import (
	"os"
	"path/filepath"
	"testing"
)

// A capture has to be attributable without guessing. The file names are not
// usable: BattleServerOrig.exe has held our build (2026-09-28) and
// BattleServerReimpl.exe the original, and on 2026-08-21 it was the other way
// round. So identity is decided by reading the file.
//
// The markers are chosen so this does not depend on a runtime log line: the Go
// build markers are always embedded in a Go binary and never in the MSVC
// original, which is what makes them usable offline on old binaries too.
func TestIdentifyExeReadsContentNotName(t *testing.T) {
	dir := t.TempDir()

	// A Go binary: this test binary itself.
	ours := filepath.Join(dir, "BattleServerOrig.exe") // deliberately the "wrong" name
	copyFile(t, os.Args[0], ours)
	if got, err := IdentifyExe(ours); err != nil {
		t.Fatalf("IdentifyExe: %v", err)
	} else if got.Identity != Ours {
		t.Errorf("a Go binary named BattleServerOrig.exe identified as %q, want %q (%s)",
			got.Identity, Ours, got.Why)
	}

	// Not a Go binary and not an MSVC one either: nothing may be concluded.
	junk := filepath.Join(dir, "whatever.exe")
	if err := os.WriteFile(junk, []byte("not a pe file"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := IdentifyExe(junk)
	if err != nil {
		t.Fatalf("IdentifyExe: %v", err)
	}
	if got.Identity != Unknown {
		t.Errorf("a junk file identified as %q, want %q", got.Identity, Unknown)
	}
	if got.SHA256 == "" || got.Size == 0 {
		t.Error("the fingerprint must be recorded even when the identity is unknown")
	}

	// A missing file must be an error, never a silent Unknown.
	if _, err := IdentifyExe(filepath.Join(dir, "nope.exe")); err == nil {
		t.Error("a missing binary must return an error")
	}
}

func TestWriteRolesRefusesToGuess(t *testing.T) {
	dir := t.TempDir()
	ours := filepath.Join(dir, "BattleServerOrig.exe")
	copyFile(t, os.Args[0], ours)
	junk := filepath.Join(dir, "other.exe")
	if err := os.WriteFile(junk, []byte("not a pe file"), 0o644); err != nil {
		t.Fatal(err)
	}

	// One side unknown: the manifest must say so, because a capture like that
	// cannot be compared and must not look decided.
	roles, err := WriteRoles(dir, ours, junk, "main_log.csv", "secondary_log.csv")
	if err != nil {
		t.Fatalf("WriteRoles: %v", err)
	}
	if roles.Warning == "" {
		t.Error("a manifest with an unknown side must carry a warning")
	}
	if roles.Primary.OursSide != true || roles.Secondary.OursSide != false {
		t.Errorf("is_ours wrong: primary=%v secondary=%v", roles.Primary.OursSide, roles.Secondary.OursSide)
	}
	if _, err := os.Stat(filepath.Join(dir, "roles.json")); err != nil {
		t.Errorf("roles.json was not written: %v", err)
	}
	if roles.Describe() == "" {
		t.Error("Describe must produce the one line that goes into the log")
	}

	// Both sides identical: also unusable, and also has to be flagged.
	roles2, err := WriteRoles(dir, ours, ours, "main_log.csv", "secondary_log.csv")
	if err != nil {
		t.Fatalf("WriteRoles: %v", err)
	}
	if roles2.Warning == "" {
		t.Error("two identical sides must carry a warning")
	}
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, data, 0o755); err != nil {
		t.Fatal(err)
	}
}
