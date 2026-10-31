package main

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	internalnet "github.com/luskaner/ageBattleServer/internal/net"
	"github.com/luskaner/ageBattleServer/internal/protocol"
)

// ProcessManager controls the execution and lifecycle of both executable binaries
type ProcessManager struct {
	cmd1       *exec.Cmd
	cmd2       *exec.Cmd
	publicPort int
	oobPort    int
	relayPorts []int
	guid       string
	logOut     io.Writer
	args1      []string
	args2      []string
	broadcastC func()

	// target1 and target2 are where each child is expected to listen its bsPort,
	// recovered from the -bsPort each one was given. The proxy dials both as soon
	// as a client connects (tcp_proxy.go:199 and :211), so a child that has not
	// bound its port yet produces "Failed to connect to Target 2" and silently
	// misses that connection's duplicate. Start waits for each child instead.
	target1 string
	target2 string
	// readyTimeout bounds each wait. It is a field so tests can shrink it; zero
	// means childReadyTimeout.
	readyTimeout time.Duration
}

// NewProcessManager creates a new manager for running child processes
// Args1 and Args2 are the exact command lines handed to each child. They are
// recorded in the capture metadata so a simulation test can rebuild the same
// server: -simulationPeriod, -pageSize and friends change behaviour, so a test
// that guessed them would be testing a configuration nobody ever ran.
func (pm *ProcessManager) Args1() []string { return pm.args1 }
func (pm *ProcessManager) Args2() []string { return pm.args2 }
func NewProcessManager(cfg Config, passedArgs []string, logOut io.Writer) *ProcessManager {
	var args1 []string
	if cfg.Exe1ExtraArgs != "" {
		args1 = append(args1, strings.Fields(cfg.Exe1ExtraArgs)...)
	}
	args1 = append(args1, passedArgs...)

	var args2 []string
	if cfg.Exe2ExtraArgs != "" {
		args2 = append(args2, strings.Fields(cfg.Exe2ExtraArgs)...)
	}
	args2 = append(args2, passedArgs...)

	// Relay broadcast targets both default ports when nothing is configured;
	// an explicit value (env or -relayBroadcastPort) allows a single port,
	// and 0 disables the broadcast. -1 means "not configured".
	relayPorts := []int{9999, 8888}
	if cfg.RelayBroadcastPort != -1 {
		if cfg.RelayBroadcastPort == 0 {
			relayPorts = nil
		} else {
			relayPorts = []int{cfg.RelayBroadcastPort}
		}
	}

	return &ProcessManager{
		cmd1:       exec.Command(cfg.Exe1Path, args1...),
		cmd2:       exec.Command(cfg.Exe2Path, args2...),
		publicPort: cfg.ProxyPort,
		oobPort:    cfg.ProxyOOBPort,
		relayPorts: relayPorts,
		guid:       internalnet.GenerateGUID(),
		logOut:     logOut,
		args1:      args1,
		args2:      args2,
		target1:    listenAddr(cfg.Target1Host, argInt(args1, "bsPort")),
		target2:    listenAddr(cfg.Target2Host, argInt(args2, "bsPort")),
	}
}

// argInt returns the integer value that follows flag in args, or 0 when the flag
// is absent or has no usable value. The launcher passes "-bsPort 14001" to both
// children, which is the port each one will listen on.
func argInt(args []string, flag string) int {
	for i, a := range args {
		if strings.EqualFold(strings.TrimLeft(a, "-"), flag) && i+1 < len(args) {
			if n, err := strconv.Atoi(args[i+1]); err == nil {
				return n
			}
		}
	}
	return 0
}

// listenAddr joins a host and a port, returning "" when either is unusable so the
// caller can skip waiting rather than probe something meaningless.
func listenAddr(host string, port int) string {
	if host == "" || port <= 0 || port > 65535 {
		return ""
	}
	return net.JoinHostPort(host, strconv.Itoa(port))
}

// Start launches both secondary processes passing command line arguments directly
func (pm *ProcessManager) Start() {
	if len(pm.relayPorts) > 0 {
		ports := make([]uint, len(pm.relayPorts))
		for i, p := range pm.relayPorts {
			ports[i] = uint(p)
		}
		ctx, cancel := context.WithCancel(context.Background())
		pm.broadcastC = cancel
		broadcaster := internalnet.NewUDPBroadcaster(ports, func(port uint) [][]byte {
			return protocol.BuildRelayBroadcastMessages(
				pm.guid,
				"BattleServerProxy",
				uint16(pm.publicPort),
				0,
				uint16(pm.oobPort),
				port,
			)
		})
		if err := broadcaster.Start(ctx); err != nil {
			log.Printf("[LAUNCHER ERROR] Could not start relay broadcaster: %v", err)
		}
	}
	// Connect child process standard output/error to system output and log file
	pm.cmd1.Stdout = io.MultiWriter(os.Stdout, pm.logOut)
	pm.cmd1.Stderr = io.MultiWriter(os.Stderr, pm.logOut)

	pm.cmd2.Stdout = io.MultiWriter(os.Stdout, pm.logOut)
	pm.cmd2.Stderr = io.MultiWriter(os.Stderr, pm.logOut)

	log.Printf("[LAUNCHER] Starting Process 1: %s %v", pm.cmd1.Path, pm.cmd1.Args[1:])
	if err := pm.cmd1.Start(); err != nil {
		log.Printf("[LAUNCHER ERROR] Could not start %s: %v", pm.cmd1.Path, err)
	} else {
		// The proxy is already listening by the time the children start, and it
		// dials both targets as soon as any client connects. Starting the second
		// child before the first has bound its port races that dial, which is what
		// produced "Failed to connect to Target 2" on every launch. Wait for the
		// port instead.
		pm.waitForPort("Process 1", pm.target1)
	}

	log.Printf("[LAUNCHER] Starting Process 2: %s %v", pm.cmd2.Path, pm.cmd2.Args[1:])
	if err := pm.cmd2.Start(); err != nil {
		log.Printf("[LAUNCHER ERROR] Could not start %s: %v", pm.cmd2.Path, err)
	} else {
		// Same reasoning for the window between this child binding its port and
		// the first client arriving: the proxy is live throughout.
		pm.waitForPort("Process 2", pm.target2)
	}
}

// childReadyTimeout is how long a child gets to bind its bsPort before the
// launcher stops waiting and carries on. Startup must not hang forever because
// of a child that is slow or broken; the failure surfaces as the proxy's
// "Failed to connect to Target" warning instead.
const childReadyTimeout = 10 * time.Second

// childReadyPoll is the gap between readiness probes.
const childReadyPoll = 25 * time.Millisecond

// waitForPort blocks until addr accepts a TCP connection, the timeout expires, or
// addr is empty (nothing to wait for). It probes with a real dial and closes the
// probe immediately: the children treat an empty connection as noise, and a
// half-open probe would be worse than none.
func (pm *ProcessManager) waitForPort(name, addr string) {
	if addr == "" {
		return
	}
	timeout := pm.readyTimeout
	if timeout <= 0 {
		timeout = childReadyTimeout
	}
	deadline := time.Now().Add(timeout)
	for {
		conn, err := net.DialTimeout("tcp", addr, 250*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			log.Printf("[LAUNCHER] %s is listening on %s", name, addr)
			return
		}
		if !time.Now().Before(deadline) {
			log.Printf("[LAUNCHER WARNING] %s did not listen on %s within %s; continuing anyway",
				name, addr, timeout)
			return
		}
		time.Sleep(childReadyPoll)
	}
}

// Terminate cleanly shuts down both secondary processes
func (pm *ProcessManager) Terminate() {
	log.Println("[LAUNCHER] Stopping child processes...")
	if pm.broadcastC != nil {
		pm.broadcastC()
	}

	procs := []struct {
		name string
		cmd  *exec.Cmd
	}{
		{"Process 1", pm.cmd1},
		{"Process 2", pm.cmd2},
	}

	for _, p := range procs {
		if p.cmd != nil && p.cmd.Process != nil {
			// SIGTERM has no console-handler effect on Windows child GUI/console
			// processes launched this way, so go straight to Kill and save the
			// grace wait for platforms where the signal would be delivered.
			if err := p.cmd.Process.Kill(); err != nil && runtime.GOOS != "windows" {
				log.Printf("[LAUNCHER] %s kill: %v", p.name, err)
			}
			_ = p.cmd.Wait()
			log.Printf("[LAUNCHER] %s terminated.", p.name)
		}
	}
}
