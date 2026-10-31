package loadgen

import (
	"fmt"
	"strings"
	"time"
)

// FormatReport renders a human-readable summary of a run.
func FormatReport(rep Report) string {
	s := rep.Stats
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, "=== load report ===\n")
	_, _ = fmt.Fprintf(&sb, "matches=%d players/match=%d observers/match=%d duration=%s\n",
		rep.Profile.Matches, rep.Profile.Players, rep.Profile.Observers, rep.Elapsed.Round(time.Millisecond))
	el := rep.Elapsed.Seconds()
	if el <= 0 {
		el = 1
	}
	_, _ = fmt.Fprintf(&sb, "client->server: %d msgs, %.1f msgs/s, %.2f MB/s\n",
		s.SentMsgs, float64(s.SentMsgs)/el, float64(s.SentBytes)/(1<<20)/el)
	_, _ = fmt.Fprintf(&sb, "server->client: %d msgs, %.1f msgs/s, %.2f MB/s\n",
		s.RecvMsgs, float64(s.RecvMsgs)/el, float64(s.RecvBytes)/(1<<20)/el)
	if s.PingCount > 0 {
		_, _ = fmt.Fprintf(&sb, "ping rtt: avg=%s min=%s max=%s (%d samples)\n",
			time.Duration(s.PingSumNS/s.PingCount).Round(time.Microsecond),
			time.Duration(s.PingMinNS).Round(time.Microsecond),
			time.Duration(s.PingMax).Round(time.Microsecond), s.PingCount)
	}
	_, _ = fmt.Fprintf(&sb, "errors=%d\n", s.Errors)
	if rep.CPUUsedSeconds > 0 || rep.HeapBytes != 0 {
		_, _ = fmt.Fprintf(&sb, "cpu=%.1fs (%.0f%% of 1 core) heap-delta=%.1fMB allocs=%.0f objects (%.1fMB)\n",
			rep.CPUUsedSeconds, rep.CPUUsedSeconds/el*100,
			rep.HeapBytes/(1<<20), rep.AllocsObjs, rep.AllocsBytes/(1<<20))
	}
	if rep.ServerBytesReceived > 0 || rep.ServerBytesSent > 0 {
		_, _ = fmt.Fprintf(&sb, "server counters: received=%.2fMB sent=%.2fMB (%.1f MB/s in, %.1f MB/s out)\n",
			rep.ServerBytesReceived/(1<<20), rep.ServerBytesSent/(1<<20),
			rep.ServerBytesReceived/(1<<20)/el, rep.ServerBytesSent/(1<<20)/el)
	}
	return sb.String()
}
