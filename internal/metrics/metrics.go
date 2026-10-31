package metrics

import (
	"maps"
	"sync"
	"sync/atomic"
)

// Collector tracks named counters and gauges. The byte traffic counters are
// updated on every network frame (BytesReceived on each inbound message), so
// they use atomic fields instead of the map mutex; the map covers the rare
// named counters/gauges.
type Collector struct {
	mu       sync.RWMutex
	counters map[string]float64
	gauges   map[string]float64
	// bytesReceived and bytesSent are the hot traffic counters; they bypass
	// the mutex and are folded into GetAllCounters.
	bytesReceived atomic.Uint64
	bytesSent     atomic.Uint64
}

func NewCollector() *Collector {
	return &Collector{
		counters: make(map[string]float64),
		gauges:   make(map[string]float64),
	}
}

func (m *Collector) AddToCounter(name string, value float64) {
	switch name {
	case "BytesReceived":
		m.bytesReceived.Add(uint64(value))
		return
	case "BytesSent":
		m.bytesSent.Add(uint64(value))
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name] += value
}

func (m *Collector) SetCounter(name string, value float64) {
	switch name {
	case "BytesReceived":
		m.bytesReceived.Store(0)
		m.bytesReceived.Add(uint64(value))
		return
	case "BytesSent":
		m.bytesSent.Store(0)
		m.bytesSent.Add(uint64(value))
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name] = value
}

func (m *Collector) GetCounter(name string) float64 {
	switch name {
	case "BytesReceived":
		return float64(m.bytesReceived.Load())
	case "BytesSent":
		return float64(m.bytesSent.Load())
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.counters[name]
}

func (m *Collector) SetGauge(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gauges[name] = value
}

func (m *Collector) GetGauge(name string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.gauges[name]
}

func (m *Collector) GetAllCounters() map[string]float64 {
	m.mu.RLock()
	result := make(map[string]float64, len(m.counters)+2)
	maps.Copy(result, m.counters)
	m.mu.RUnlock()
	result["BytesReceived"] = float64(m.bytesReceived.Load())
	result["BytesSent"] = float64(m.bytesSent.Load())
	return result
}

func (m *Collector) GetAllGauges() map[string]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]float64, len(m.gauges))
	maps.Copy(result, m.gauges)
	return result
}

func (m *Collector) Reset() {
	m.mu.Lock()
	m.counters = make(map[string]float64)
	m.gauges = make(map[string]float64)
	m.mu.Unlock()
	m.bytesReceived.Store(0)
	m.bytesSent.Store(0)
}
