package replay

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestNewReplayUploadManager(t *testing.T) {
	mgr := NewReplayUploadManager("localhost", 80, "lobby")
	if mgr == nil {
		t.Fatal("NewReplayUploadManager() returned nil")
	}
	if mgr.host != "localhost" {
		t.Errorf("expected host localhost, got %s", mgr.host)
	}
	if mgr.port != 80 {
		t.Errorf("expected port 80, got %d", mgr.port)
	}
}

func TestConvertToBase64(t *testing.T) {
	mgr := NewReplayUploadManager("localhost", 80, "lobby")

	input := []byte("test data for compression")
	result, err := mgr.convertToBase64(input)
	if err != nil {
		t.Fatalf("convertToBase64 error: %v", err)
	}

	if len(result) == 0 {
		t.Error("expected non-empty result")
	}

	if len(result) >= len(input) {
		t.Log("warning: compressed+encoded data larger than input")
	}
}

func TestConvertToBase64Roundtrip(t *testing.T) {
	mgr := NewReplayUploadManager("localhost", 80, "lobby")

	input := bytes.Repeat([]byte("replay payload "), 100)
	encoded, err := mgr.convertToBase64(input)
	if err != nil {
		t.Fatalf("convertToBase64 error: %v", err)
	}

	raw, err := base64.StdEncoding.DecodeString(string(encoded))
	if err != nil {
		t.Fatalf("base64 decode error: %v", err)
	}

	r, err := zlib.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("zlib reader error: %v", err)
	}
	defer func() { _ = r.Close() }()
	decompressed, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("zlib decompress error: %v", err)
	}

	if !bytes.Equal(decompressed, input) {
		t.Error("roundtrip data mismatch")
	}
}

// TestConvertToBase64PoolReuse exercises the pooled zlib writer across
// sequential and concurrent uploads: each call must Reset and reuse the
// compressor while producing a byte-identical, decompressible stream.
func TestConvertToBase64PoolReuse(t *testing.T) {
	mgr := NewReplayUploadManager("localhost", 80, "lobby")

	var wg sync.WaitGroup
	errs := make(chan error, 64)
	for g := range 8 {
		wg.Go(func() {
			for i := range 8 {
				payload := bytes.Repeat([]byte{byte(g*10 + i)}, 512+i)
				encoded, err := mgr.convertToBase64(payload)
				if err != nil {
					errs <- err
					continue
				}
				raw, err := base64.StdEncoding.DecodeString(string(encoded))
				if err != nil {
					errs <- err
					continue
				}
				r, err := zlib.NewReader(bytes.NewReader(raw))
				if err != nil {
					errs <- err
					continue
				}
				decompressed, err := io.ReadAll(r)
				_ = r.Close()
				if err != nil {
					errs <- err
					continue
				}
				if !bytes.Equal(decompressed, payload) {
					errs <- fmt.Errorf("roundtrip mismatch (g=%d i=%d)", g, i)
				}
			}
		})
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}

func TestOnAddReplayData(t *testing.T) {
	mgr := NewReplayUploadManager("localhost", 80, "lobby")

	info := &Info{
		SessionID:  123,
		SequenceID: 0,
		Data:       []byte("test replay data"),
		RefCount:   1,
	}

	mgr.OnAddReplayData(info)

	select {
	case received := <-mgr.sendQueue:
		if received.SessionID != 123 {
			t.Errorf("expected session ID 123, got %d", received.SessionID)
		}
		// Compression happens on the upload loop; the queue carries the raw
		// page so the game-message path never pays for deflate.
		if !bytes.Equal(received.Data, []byte("test replay data")) {
			t.Errorf("expected raw page in queue, got %q", received.Data)
		}
	default:
		t.Error("expected data in send queue")
	}
}

func TestOnAddReplayDataQueueFull(t *testing.T) {
	mgr := NewReplayUploadManager("localhost", 80, "lobby")

	for i := 0; i < cap(mgr.sendQueue); i++ {
		mgr.OnAddReplayData(&Info{SessionID: int64(i), Data: []byte("x")})
	}

	before := len(mgr.pool)
	mgr.OnAddReplayData(&Info{SessionID: 999, Data: []byte("y")})

	if len(mgr.sendQueue) != cap(mgr.sendQueue) {
		t.Errorf("queue length = %d, want %d (full, no items consumed)", len(mgr.sendQueue), cap(mgr.sendQueue))
	}

	if len(mgr.pool) != before+1 {
		t.Errorf("pool size = %d, want %d", len(mgr.pool), before+1)
	}
}

func TestPackReplayData(t *testing.T) {
	mgr := NewReplayUploadManager("localhost", 80, "lobby")

	payload := []byte("packed payload")
	mgr.PackReplayData(777, payload)

	var first *Info
	select {
	case first = <-mgr.sendQueue:
		if first.SessionID != 777 {
			t.Errorf("SessionID = %d, want 777", first.SessionID)
		}
		if first.SequenceID != 1 {
			t.Errorf("SequenceID = %d, want 1 (first upload of the session)", first.SequenceID)
		}
		if first.RefCount != 1 {
			t.Errorf("RefCount = %d, want 1", first.RefCount)
		}
		if first.DataLength != uint32(len(first.Data)) {
			t.Errorf("DataLength = %d, want %d", first.DataLength, len(first.Data))
		}
	default:
		t.Fatal("expected data in send queue")
	}

	// Mutating the caller's buffer must not touch the already-queued page:
	// PackReplayData snapshots at pack time.
	payload[0] = 0xFF
	if first.Data[0] != 'p' {
		t.Errorf("queued page was mutated by the caller: %q", first.Data)
	}

	// A later pack of the mutated buffer copies the new content.
	mgr.PackReplayData(1, payload)
	select {
	case second := <-mgr.sendQueue:
		if second.SessionID != 1 {
			t.Errorf("SessionID = %d, want 1", second.SessionID)
		}
		if second.Data[0] != 0xFF {
			t.Error("second pack should copy the mutated payload")
		}
	default:
		t.Error("expected second data in send queue")
	}
}

func TestOnProxyData(t *testing.T) {
	mgr := NewReplayUploadManager("localhost", 80, "lobby")
	mgr.OnProxyData(4242, []byte("proxy payload"))

	select {
	case received := <-mgr.sendQueue:
		if received.SessionID != 4242 {
			t.Errorf("SessionID = %d, want 4242 (the match the data belongs to)", received.SessionID)
		}
		if string(received.Data) != "proxy payload" {
			t.Errorf("Data = %q, want %q", received.Data, "proxy payload")
		}
	default:
		t.Error("expected data in send queue")
	}
}

// TestOnProxyDataPerMatchSequences covers the reason the session ID exists: the
// replay host files pages by matchid+sequenceid, so two matches must get their
// own match id and their own sequence counter instead of interleaving under one
// key.
func TestOnProxyDataPerMatchSequences(t *testing.T) {
	received := make(chan string, 4)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
		received <- r.URL.RawQuery
	}))
	defer srv.Close()
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse server URL: %v", err)
	}
	port, _ := strconv.Atoi(u.Port())

	mgr := NewReplayUploadManager("127.0.0.1", uint(port), "lobby")
	if err := mgr.Start(t.Context()); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer mgr.Stop()

	mgr.OnProxyData(11, []byte("a"))
	mgr.OnProxyData(22, []byte("b"))
	mgr.OnProxyData(11, []byte("c"))

	want := map[string]bool{
		"matchid=11&sequenceid=1": false,
		"matchid=22&sequenceid=1": false,
		"matchid=11&sequenceid=2": false,
	}
	for range want {
		select {
		case q := <-received:
			if _, ok := want[q]; !ok {
				t.Fatalf("unexpected query %q", q)
			}
			want[q] = true
		case <-time.After(3 * time.Second):
			t.Fatalf("timed out waiting for replay requests, got %d of %d", countTrue(want), len(want))
		}
	}
	for q, seen := range want {
		if !seen {
			t.Errorf("never saw query %q", q)
		}
	}
}

func countTrue(m map[string]bool) int {
	n := 0
	for _, v := range m {
		if v {
			n++
		}
	}
	return n
}

func TestPool(t *testing.T) {
	mgr := NewReplayUploadManager("localhost", 80, "lobby")

	info := mgr.getFromPool()
	if info == nil {
		t.Fatal("getFromPool returned nil")
	}

	info.SessionID = 999
	mgr.returnToPool(info)

	info2 := mgr.getFromPool()
	if info2.SessionID != 999 {
		t.Errorf("expected pooled info to have session ID 999, got %d", info2.SessionID)
	}
}

func TestPoolCapacity(t *testing.T) {
	mgr := NewReplayUploadManager("localhost", 80, "lobby")

	for i := range 150 {
		mgr.returnToPool(&Info{SessionID: int64(i)})
	}

	if len(mgr.pool) != 100 {
		t.Errorf("pool size = %d, want 100", len(mgr.pool))
	}
}

func TestStartDialError(t *testing.T) {
	mgr := NewReplayUploadManager("127.0.0.1", 1, "lobby")
	if err := mgr.Start(context.Background()); err == nil {
		t.Fatal("expected dial error for closed port")
	}
}

func TestStartAndPushReplayData(t *testing.T) {
	type receivedRequest struct {
		request *http.Request
		body    []byte
	}
	received := make(chan receivedRequest, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read replay body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		received <- receivedRequest{request: r, body: body}
	}))
	defer srv.Close()
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse server URL: %v", err)
	}
	port, _ := strconv.Atoi(u.Port())

	mgr := NewReplayUploadManager("127.0.0.1", uint(port), "lobby")
	ctx := t.Context()

	if err := mgr.Start(ctx); err != nil {
		t.Fatalf("Start error: %v", err)
	}

	mgr.PackReplayData(4242, []byte("hello replay"))

	select {
	case got := <-received:
		req := got.request
		if req.Method != http.MethodPost || req.URL.Path != "/game/replay/setReplayData" {
			t.Errorf("unexpected request: %s %s", req.Method, req.URL.Path)
		}
		if q := req.URL.RawQuery; q != "matchid=4242&sequenceid=1" {
			t.Errorf("unexpected query: %q", q)
		}
		compressed, err := base64.StdEncoding.DecodeString(string(got.body))
		if err != nil {
			t.Fatalf("decode replay body: %v", err)
		}
		zr, err := zlib.NewReader(bytes.NewReader(compressed))
		if err != nil {
			t.Fatalf("open replay body: %v", err)
		}
		decoded, err := io.ReadAll(zr)
		_ = zr.Close()
		if err != nil {
			t.Fatalf("decompress replay body: %v", err)
		}
		if !bytes.Equal(decoded, []byte("hello replay")) {
			t.Errorf("unexpected decoded replay body: %q", decoded)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for replay request")
	}

	mgr.Stop()
}

func TestPushReplayDataNoConn(t *testing.T) {
	mgr := NewReplayUploadManager("localhost", 80, "lobby")
	info := &Info{SessionID: 1, Data: []byte("x")}
	mgr.pushReplayData(info)
	if mgr.inflight.Load() != 0 {
		t.Errorf("inflight = %d, want 0", mgr.inflight.Load())
	}
}

func TestPushReplayDataWriteError(t *testing.T) {
	a, b := net.Pipe()
	defer func() { _ = a.Close() }()

	mgr := NewReplayUploadManager("localhost", 80, "lobby")
	mgr.conn = a
	_ = b.Close()

	info := &Info{SessionID: 1, Data: []byte("x"), DataLength: 1}
	mgr.pushReplayData(info)
	if mgr.inflight.Load() != 0 {
		t.Errorf("inflight = %d, want 0", mgr.inflight.Load())
	}
}

func TestPushReplayDataReconnect(t *testing.T) {
	var received chan *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
		received <- r
	}))
	defer srv.Close()
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse server URL: %v", err)
	}
	port, _ := strconv.Atoi(u.Port())

	received = make(chan *http.Request, 4)
	mgr := NewReplayUploadManager("127.0.0.1", uint(port), "lobby")
	if err := mgr.Start(t.Context()); err != nil {
		t.Fatalf("Start error: %v", err)
	}

	expectPage := func(wantSeq string) {
		t.Helper()
		select {
		case req := <-received:
			if q := req.URL.RawQuery; q != "matchid=4242&sequenceid="+wantSeq {
				t.Errorf("unexpected request query: %q (want sequenceid=%s)", q, wantSeq)
			}
		case <-time.After(3 * time.Second):
			t.Fatalf("timed out waiting for replay request (sequenceid=%s)", wantSeq)
		}
	}

	mgr.PackReplayData(4242, []byte("first"))
	expectPage("1")

	// Drop the connection; the next upload must reconnect and still succeed.
	mgr.dropConn()

	mgr.PackReplayData(4242, []byte("second"))
	expectPage("2")

	mgr.Stop()
}

func TestProcessLoopCancel(t *testing.T) {
	mgr := NewReplayUploadManager("localhost", 80, "lobby")

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		mgr.processLoop(ctx)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("processLoop did not stop after context cancellation")
	}
}

func TestStop(t *testing.T) {
	mgr := NewReplayUploadManager("localhost", 80, "lobby")
	mgr.Stop()

	a, b := net.Pipe()
	defer func() { _ = a.Close() }()
	defer func() { _ = b.Close() }()
	mgr.conn = a
	mgr.Stop()
	if mgr.conn != nil {
		t.Error("conn should be nil after Stop")
	}
}
