package metrics

import (
	"sync"
	"testing"
)

func TestNewCollector(t *testing.T) {
	c := NewCollector()
	if c == nil {
		t.Fatal("NewCollector() returned nil")
	}
}

func TestAddToCounter(t *testing.T) {
	c := NewCollector()
	c.AddToCounter("test", 1.5)
	c.AddToCounter("test", 2.5)

	if c.GetCounter("test") != 4.0 {
		t.Errorf("expected 4.0, got %f", c.GetCounter("test"))
	}
}

func TestSetCounter(t *testing.T) {
	c := NewCollector()
	c.SetCounter("test", 10.0)
	c.SetCounter("test", 5.0)

	if c.GetCounter("test") != 5.0 {
		t.Errorf("expected 5.0, got %f", c.GetCounter("test"))
	}
}

func TestSetGauge(t *testing.T) {
	c := NewCollector()
	c.SetGauge("cpu", 50.0)
	c.SetGauge("cpu", 75.0)

	if c.GetGauge("cpu") != 75.0 {
		t.Errorf("expected 75.0, got %f", c.GetGauge("cpu"))
	}
}

func TestGetAllCounters(t *testing.T) {
	c := NewCollector()
	c.AddToCounter("a", 1.0)
	c.AddToCounter("b", 2.0)

	counters := c.GetAllCounters()
	if len(counters) != 4 {
		t.Errorf("expected 4 counters, got %d", len(counters))
	}
	if counters["a"] != 1.0 {
		t.Errorf("expected a=1.0, got %f", counters["a"])
	}
	if counters["BytesReceived"] != 0.0 || counters["BytesSent"] != 0.0 {
		t.Error("byte counters should default to 0")
	}
}

func TestByteCounters(t *testing.T) {
	c := NewCollector()
	c.AddToCounter("BytesReceived", 100.0)
	c.AddToCounter("BytesReceived", 50.0)
	c.AddToCounter("BytesSent", 42.0)

	if c.GetCounter("BytesReceived") != 150.0 {
		t.Errorf("expected 150.0, got %f", c.GetCounter("BytesReceived"))
	}
	if c.GetCounter("BytesSent") != 42.0 {
		t.Errorf("expected 42.0, got %f", c.GetCounter("BytesSent"))
	}
	if c.GetCounter("TcpConnections") != 0.0 {
		t.Errorf("expected 0.0, got %f", c.GetCounter("TcpConnections"))
	}

	counters := c.GetAllCounters()
	if counters["BytesReceived"] != 150.0 || counters["BytesSent"] != 42.0 {
		t.Error("GetAllCounters should fold in the atomic byte counters")
	}
}

func TestByteCounterReset(t *testing.T) {
	c := NewCollector()
	c.AddToCounter("BytesReceived", 100.0)
	c.AddToCounter("BytesSent", 100.0)

	c.Reset()

	if c.GetCounter("BytesReceived") != 0.0 || c.GetCounter("BytesSent") != 0.0 {
		t.Error("byte counters should be 0 after reset")
	}
}

func TestGetAllGauges(t *testing.T) {
	c := NewCollector()
	c.SetGauge("cpu", 50.0)
	c.SetGauge("mem", 100.0)

	gauges := c.GetAllGauges()
	if len(gauges) != 2 {
		t.Errorf("expected 2 gauges, got %d", len(gauges))
	}
	if gauges["cpu"] != 50.0 {
		t.Errorf("expected cpu=50.0, got %f", gauges["cpu"])
	}
}

func TestReset(t *testing.T) {
	c := NewCollector()
	c.AddToCounter("test", 1.0)
	c.SetGauge("cpu", 50.0)

	c.Reset()

	if c.GetCounter("test") != 0.0 {
		t.Error("counter should be 0 after reset")
	}
	if c.GetGauge("cpu") != 0.0 {
		t.Error("gauge should be 0 after reset")
	}
}

func BenchmarkCollectorAddBytes(b *testing.B) {
	c := NewCollector()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.AddToCounter("BytesReceived", 1.0)
		}
	})
}

func BenchmarkCollectorAddBytesLocked(b *testing.B) {
	var mu sync.Mutex
	var total float64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			total += 1.0
			mu.Unlock()
		}
	})
}
