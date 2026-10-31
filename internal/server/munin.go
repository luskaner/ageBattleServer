package server

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/luskaner/ageBattleServer/internal/logger"
)

func (s *Server) initMunin() error {
	f, err := os.OpenFile(s.config.MuninLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	s.muninFile = f
	s.muninWriter = bufio.NewWriter(f)
	s.muninTicker = time.NewTicker(time.Duration(s.config.MuninReportDelay) * time.Millisecond)

	logger.Info("Server", "Munin", "Writing metrics to %s every %dms", s.config.MuninLog, s.config.MuninReportDelay)
	return nil
}

func (s *Server) muninLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.muninTicker.C:
			s.writeMuninMetrics()
		}
	}
}

func (s *Server) writeMuninMetrics() {
	if s.muninWriter == nil {
		return
	}

	s.muninMu.Lock()
	defer s.muninMu.Unlock()

	prefix := "bsmunin"
	if s.config.ClusterName != "" {
		prefix = fmt.Sprintf("bsmunin_%s", s.config.ClusterName)
	}

	counters := s.metricsCollector.GetAllCounters()
	gauges := s.metricsCollector.GetAllGauges()

	_, _ = s.muninWriter.WriteString(fmt.Sprintf("multigraph %s_connections\n", prefix))
	_, _ = s.muninWriter.WriteString(fmt.Sprintf("tcp_connections.value %f\n", counters["TcpConnections"]))
	_, _ = s.muninWriter.WriteString(fmt.Sprintf("ws_connections.value %f\n", counters["WsConnections"]))
	_, _ = s.muninWriter.WriteString(fmt.Sprintf("ping_connections.value %f\n", counters["PingConnections"]))

	_, _ = s.muninWriter.WriteString(fmt.Sprintf("multigraph %s_errors\n", prefix))
	_, _ = s.muninWriter.WriteString(fmt.Sprintf("tcp_connection_errors.value %f\n", counters["TcpConnectionErrors"]))
	_, _ = s.muninWriter.WriteString(fmt.Sprintf("ws_connection_errors.value %f\n", counters["WsConnectionErrors"]))
	_, _ = s.muninWriter.WriteString(fmt.Sprintf("ping_connection_errors.value %f\n", counters["PingConnectionErrors"]))

	_, _ = s.muninWriter.WriteString(fmt.Sprintf("multigraph %s_traffic\n", prefix))
	_, _ = s.muninWriter.WriteString(fmt.Sprintf("bytes_received.value %f\n", counters["BytesReceived"]))
	_, _ = s.muninWriter.WriteString(fmt.Sprintf("bytes_sent.value %f\n", counters["BytesSent"]))

	_, _ = s.muninWriter.WriteString(fmt.Sprintf("multigraph %s_performance\n", prefix))
	_, _ = s.muninWriter.WriteString(fmt.Sprintf("clients.value %f\n", gauges["Clients"]))
	_, _ = s.muninWriter.WriteString(fmt.Sprintf("server_cpu_usage.value %f\n", gauges["ServerCpuUsage"]))

	_ = s.muninWriter.Flush()
}
