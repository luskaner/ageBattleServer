package main

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"
)

// The secondary CSV must carry BOTH directions of the Target 2 connection. It
// used to hold only server->client, which made everything our own server received
// invisible: a message we dropped was indistinguishable from one we never saw.
// That is what blocked the diagnosis of the 0x14 echo in capture _26.
func TestSecondaryCSVRecordsBothDirectionsOfTarget2(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secondary.csv")
	l := NewCSVLogger(path)

	// Target 2 -> client, as the relay logs it.
	l.LogSecondary("127.0.0.1", 15001, "10.0.0.5", 44444, []byte{0xAA, 0x01})
	// client -> Target 2, the direction that used to be missing.
	l.LogSecondary("10.0.0.5", 44444, "127.0.0.1", 15001, []byte{0xBB, 0x02})

	l.Close()
	waitForCSV(t, path, 2)

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = f.Close() }()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(rows) != 3 { // header + 2
		t.Fatalf("rows = %d, want 3 (header + 2)", len(rows))
	}

	// Match on the endpoints, not the payload: decodePacket reinterprets the
	// bytes as a protocol packet, so the hex column is not the raw frame.
	type ep struct{ srcH, srcP, dstH, dstP string }
	got := map[ep]int{}
	for _, r := range rows[1:] {
		got[ep{r[1], r[2], r[3], r[4]}]++
	}
	down := ep{"127.0.0.1", "15001", "10.0.0.5", "44444"}
	up := ep{"10.0.0.5", "44444", "127.0.0.1", "15001"}
	if got[down] != 1 {
		t.Errorf("rows for T2->client (%v) = %d, want 1; all rows: %v", down, got[down], got)
	}
	if got[up] != 1 {
		t.Errorf("rows for client->T2 (%v) = %d, want 1; all rows: %v", up, got[up], got)
	}
}

// waitForCSV polls until the file has at least n data rows, so the test does not
// race the logger's background writer.
func waitForCSV(t *testing.T, path string, n int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		f, err := os.Open(path)
		if err == nil {
			rows, rerr := csv.NewReader(f).ReadAll()
			_ = f.Close()
			if rerr == nil && len(rows) >= n+1 {
				return
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("CSV %s did not reach %d rows in time", filepath.Base(path), n)
}

// Guard the endpoints the way handleConnection passes them, so a signature change
// that swaps host/port or drops the payload fails here.
func TestLogSecondarySignatureKeepsHostPortOrder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "s.csv")
	l := NewCSVLogger(path)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		l.LogSecondary("1.2.3.4", 1111, "5.6.7.8", 2222, []byte{0x01})
	}()
	wg.Wait()
	l.Close()
	waitForCSV(t, path, 1)

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = f.Close() }()
	rows, _ := csv.NewReader(f).ReadAll()
	if len(rows) < 2 {
		t.Fatal("no data row")
	}
	if _, err := strconv.Atoi(rows[1][2]); err != nil {
		t.Errorf("src_port %q is not numeric", rows[1][2])
	}
	if rows[1][2] != "1111" || rows[1][4] != "2222" {
		t.Errorf("ports = %s / %s, want 1111 / 2222", rows[1][2], rows[1][4])
	}
}
