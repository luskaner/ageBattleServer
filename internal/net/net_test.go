package net

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestGenerateGUID(t *testing.T) {
	g1 := GenerateGUID()
	g2 := GenerateGUID()

	if g1 == g2 {
		t.Error("GUIDs should be unique")
	}
	if len(g1) != 36 {
		t.Errorf("expected GUID length 36, got %d", len(g1))
	}
}

func TestTCPListenerStartStop(t *testing.T) {
	ln := NewTCPListener("127.0.0.1:0", func(conn Conn) {
		_ = conn.Close()
	})

	if err := ln.Start(t.Context()); err != nil {
		t.Fatalf("Start error: %v", err)
	}

	ln.Stop()
}

func TestTCPListenerAcceptConnection(t *testing.T) {
	accepted := make(chan struct{}, 1)

	ln := NewTCPListener("127.0.0.1:0", func(conn Conn) {
		accepted <- struct{}{}
		_ = conn.Close()
	})

	if err := ln.Start(t.Context()); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer ln.Stop()

	addr := ln.listener.Addr().String()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Dial error: %v", err)
	}
	defer func() { _ = conn.Close() }()

	select {
	case <-accepted:
	case <-time.After(time.Second):
		t.Error("connection not accepted")
	}
}

func TestTCPListenerGetPort(t *testing.T) {
	ln := NewTCPListener("127.0.0.1:0", func(conn Conn) {
		_ = conn.Close()
	})

	if port := ln.GetPort(); port != 0 {
		t.Errorf("GetPort before Start = %d, want 0", port)
	}

	if err := ln.Start(t.Context()); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer ln.Stop()

	if port := ln.GetPort(); port == 0 {
		t.Error("GetPort after Start = 0, want valid port")
	}
}

func TestTCPListenerStartError(t *testing.T) {
	ln := NewTCPListener("not a valid addr", func(conn Conn) {})
	if err := ln.Start(context.Background()); err == nil {
		t.Fatal("expected error for invalid listen address")
	}
}

func TestTCPListenerStopWithoutStart(t *testing.T) {
	ln := NewTCPListener("127.0.0.1:0", func(conn Conn) {})
	ln.Stop()
}

// TestTCPListenerRemovesClosedConn verifies the conn registry drops a connection
// as soon as it is closed (P7): a slice-based registry kept every closed conn
// alive until Stop().
func TestTCPListenerRemovesClosedConn(t *testing.T) {
	ln := NewTCPListener("127.0.0.1:0", func(conn Conn) { _ = conn.Close() })
	if err := ln.Start(t.Context()); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer ln.Stop()

	conn, err := net.Dial("tcp", ln.listener.Addr().String())
	if err != nil {
		t.Fatalf("Dial error: %v", err)
	}
	defer func() { _ = conn.Close() }()

	deadline := time.Now().Add(time.Second)
	for {
		ln.mu.Lock()
		n := len(ln.conns)
		ln.mu.Unlock()
		if n == 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("closed connection still tracked: %d", n)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// BenchmarkTCPListenerConnRetention shows what a connect+disconnect cycle
// leaves behind in the listener's connection registry. Before P7 every closed
// connection stayed in the tracking slice until Stop(), so the tracked count
// and the retained heap grew with each connection ever accepted; after, Close
// removes the connection and only the live set is retained. The conns param is
// fixed per sub-benchmark (it is the load, not the iteration count).
func BenchmarkTCPListenerConnRetention(b *testing.B) {
	for _, n := range []int{100, 1000} {
		b.Run(fmt.Sprintf("conns=%d", n), func(b *testing.B) {
			accepted := make(chan struct{}, n)
			ln := NewTCPListener("127.0.0.1:0", func(conn Conn) {
				_ = conn.Close()
				accepted <- struct{}{}
			})
			if err := ln.Start(context.Background()); err != nil {
				b.Fatal(err)
			}
			defer ln.Stop()
			addr := ln.listener.Addr().String()

			clients := make([]net.Conn, 0, n)
			for range n {
				c, err := net.Dial("tcp", addr)
				if err != nil {
					b.Fatal(err)
				}
				clients = append(clients, c)
			}
			// The handler closes each server-side conn right after accept.
			for range n {
				<-accepted
			}
			for _, c := range clients {
				_ = c.Close()
			}

			var ms runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&ms)

			b.ReportMetric(float64(len(ln.conns)), "tracked-conns")
			b.ReportMetric(float64(ms.HeapInuse), "heap-inuse-B")
		})
	}
}

func TestUDPBroadcasterStartStop(t *testing.T) {
	b := NewUDPBroadcaster([]uint{0}, func(_ uint) [][]byte {
		return [][]byte{{0x01, 0x02}}
	})

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	if err := b.Start(ctx); err != nil {
		t.Fatalf("Start error: %v", err)
	}

	b.Stop()
}

func TestUDPBroadcasterSendsDatagram(t *testing.T) {
	receiver, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		t.Fatalf("ListenUDP: %v", err)
	}
	defer receiver.Close()

	want := []byte{0x21, 0xAA, 0xBB}
	b := NewUDPBroadcaster([]uint{uint(receiver.LocalAddr().(*net.UDPAddr).Port)}, func(uint) [][]byte {
		return [][]byte{want}
	})
	if err := b.Start(t.Context()); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer b.Stop()

	b.broadcast()

	_ = receiver.SetReadDeadline(time.Now().Add(time.Second))
	got := make([]byte, 64)
	n, _, err := receiver.ReadFromUDP(got)
	if err != nil {
		t.Fatalf("ReadFromUDP: %v", err)
	}
	if string(got[:n]) != string(want) {
		t.Fatalf("received datagram %x, want %x", got[:n], want)
	}
}

func TestUDPBroadcasterBroadcastNilMsg(t *testing.T) {
	b := NewUDPBroadcaster([]uint{0}, nil)
	b.broadcast()

	b2 := NewUDPBroadcaster([]uint{0}, func(_ uint) [][]byte {
		return nil
	})
	b2.broadcast()
}

func TestUDPBroadcasterBroadcastNoConn(t *testing.T) {
	b := NewUDPBroadcaster([]uint{0}, func(_ uint) [][]byte {
		return [][]byte{{0x01}}
	})

	b.broadcast()
}

func TestUDPBroadcasterBroadcastAfterStop(t *testing.T) {
	b := NewUDPBroadcaster([]uint{0}, func(_ uint) [][]byte {
		return [][]byte{{0x01}}
	})

	b.Stop()
	b.broadcast()
}

func TestUDPBroadcasterStopWithoutStart(t *testing.T) {
	b := NewUDPBroadcaster([]uint{0}, nil)
	b.Stop()
}

// TestTCPListenerAcceptErrorContinues closes the underlying listener without
// Stop: Accept errors while the listener is not marked closed, exercising the
// retry branch of the accept loop.
func TestTCPListenerAcceptErrorContinues(t *testing.T) {
	ln := NewTCPListener("127.0.0.1:0", func(conn Conn) { _ = conn.Close() })

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	if err := ln.Start(ctx); err != nil {
		t.Fatalf("Start error: %v", err)
	}

	ln.mu.Lock()
	_ = ln.listener.Close()
	ln.mu.Unlock()

	time.Sleep(50 * time.Millisecond)

	cancel()
	ln.Stop()
}

// TestTCPListenerCtxCancel cancels the context without stopping the listener:
// the accept loop must exit through the ctx.Done branch.
func TestTCPListenerCtxCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	ln := NewTCPListener("127.0.0.1:0", func(conn Conn) { _ = conn.Close() })
	if err := ln.Start(ctx); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	addr := fmt.Sprintf("127.0.0.1:%d", ln.GetPort())

	cancel()

	// Cancelling the context alone must close the listener and unblock the
	// accept loop: no explicit Stop() is required.
	deadline := time.Now().Add(2 * time.Second)
	for {
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err != nil {
			break // listener closed
		}
		_ = conn.Close()
		if time.Now().After(deadline) {
			t.Fatal("listener still accepting after ctx cancellation")
		}
		time.Sleep(20 * time.Millisecond)
	}
	ln.Stop()
}

func TestUDPBroadcasterStartDialError(t *testing.T) {
	b := NewUDPBroadcaster([]uint{99999}, func(_ uint) [][]byte { return [][]byte{{0x01}} })
	if err := b.Start(t.Context()); err == nil {
		t.Fatal("expected dial error for port 99999")
	}
}

// TestUDPBroadcasterTicker uses a short interval so the ticker case of the
// broadcast loop fires while the context is alive.
func TestUDPBroadcasterTicker(t *testing.T) {
	var mu sync.Mutex
	count := 0

	b := NewUDPBroadcaster([]uint{0}, func(_ uint) [][]byte {
		mu.Lock()
		count++
		mu.Unlock()
		return [][]byte{{0x01}}
	})
	b.interval = 20 * time.Millisecond

	ctx, cancel := context.WithTimeout(t.Context(), 150*time.Millisecond)
	defer cancel()

	if err := b.Start(ctx); err != nil {
		t.Fatalf("Start error: %v", err)
	}
	defer b.Stop()

	deadline := time.Now().Add(time.Second)
	for {
		mu.Lock()
		got := count
		mu.Unlock()
		if got > 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("broadcast ticker never fired")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestUDPBroadcasterBroadcastWrites(t *testing.T) {
	recv, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("ListenUDP error: %v", err)
	}
	defer func() { _ = recv.Close() }()

	send, err := net.DialUDP("udp", nil, recv.LocalAddr().(*net.UDPAddr))
	if err != nil {
		t.Fatalf("DialUDP error: %v", err)
	}
	defer func() { _ = send.Close() }()

	b := NewUDPBroadcaster([]uint{0}, func(_ uint) [][]byte {
		return [][]byte{{0xDE, 0xAD}}
	})
	b.conns = []connAddr{{conn: send}}

	b.broadcast()

	_ = recv.SetReadDeadline(time.Now().Add(time.Second))
	buf := make([]byte, 16)
	n, _, err := recv.ReadFromUDP(buf)
	if err != nil {
		t.Fatalf("read from broadcaster error: %v", err)
	}
	if n != 2 || buf[0] != 0xDE || buf[1] != 0xAD {
		t.Errorf("received %v (%d bytes), want {DE AD}", buf[:n], n)
	}
}
