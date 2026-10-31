package loadgen

import "runtime/metrics"

// ResourceStats snapshots process-wide CPU and heap counters via
// runtime/metrics. It measures the whole process, so in the in-process harness
// the load engine and the server are measured together, which is the same
// figure a production process would report for that traffic.
type ResourceStats struct {
	CPUUsedSeconds float64
	HeapBytes      float64
	AllocsBytes    float64
	AllocsObjs     float64
}

// ReadResourceStats reads the current process counters.
func ReadResourceStats() ResourceStats {
	var s ResourceStats
	var cpuIdle float64
	descs := metrics.All()
	samples := make([]metrics.Sample, 0, len(descs))
	for _, d := range descs {
		samples = append(samples, metrics.Sample{Name: d.Name})
	}
	metrics.Read(samples)
	for _, smp := range samples {
		var v float64
		switch smp.Value.Kind() {
		case metrics.KindUint64:
			v = float64(smp.Value.Uint64())
		case metrics.KindFloat64:
			v = smp.Value.Float64()
		default:
			continue
		}
		switch smp.Name {
		case "/cpu/classes/total:cpu-seconds":
			s.CPUUsedSeconds = v
		case "/cpu/classes/idle:cpu-seconds":
			cpuIdle = v
		case "/memory/classes/heap/objects:bytes":
			s.HeapBytes = v
		case "/gc/heap/allocs:bytes":
			s.AllocsBytes = v
		case "/gc/heap/allocs:objects":
			s.AllocsObjs = v
		}
	}
	// Keep only the busy fraction so the report shows actual work, not the
	// accumulated idle time.
	s.CPUUsedSeconds -= cpuIdle
	return s
}

// Diff returns the delta between two snapshots.
func (a ResourceStats) Diff(b ResourceStats) ResourceStats {
	return ResourceStats{
		CPUUsedSeconds: b.CPUUsedSeconds - a.CPUUsedSeconds,
		HeapBytes:      b.HeapBytes - a.HeapBytes,
		AllocsBytes:    b.AllocsBytes - a.AllocsBytes,
		AllocsObjs:     b.AllocsObjs - a.AllocsObjs,
	}
}
