package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	// Pass command-line arguments to child processes, except the listen ports
	// (the proxy takes them over) and the certificate (used as the MITM pair)
	passedArgs, bsPort, wsPort, oobPort, relayPort, argCert, argKey := splitServerPorts(os.Args[1:])

	// Load configuration
	cfg := loadConfig()
	if bsPort > 0 {
		cfg.ProxyPort = bsPort
	} else {
		// No -bsPort given (or 0): the proxy takes an OS-assigned port so it never
		// collides with a fixed listener on the LAN. It is 0 here and becomes the
		// real number at bind time below, before anything advertises it. Asking the
		// OS for a number by binding ":0" and letting go would leave the port owned
		// by nobody until the proxy binds it again, and on Windows a listener bind
		// is exclusive: whatever takes it in that window keeps it and the capture
		// dies before it records a frame.
		cfg.ProxyPort = 0
	}
	if wsPort > 0 {
		cfg.ProxyWSPort = wsPort
	}
	if oobPort > 0 {
		cfg.ProxyOOBPort = oobPort
	}
	if relayPort != -1 {
		cfg.RelayBroadcastPort = relayPort
	}
	if argCert != "" {
		cfg.MitmCert = argCert
	}
	if argKey != "" {
		cfg.MitmKey = argKey
	}
	// The proxy owns the LAN broadcast: children must never advertise
	// their (fixed) ports directly, or the game would bypass the proxy.
	passedArgs = append(passedArgs, "-relayBroadcastPort", "0")

	// In MITM mode, make sure a certificate pair exists and hand it to both
	// target binaries so they can serve TLS to the proxy. The pair comes from
	// the command line (-sslCert/-sslKey) when given, otherwise from the env.
	if cfg.MitmTLS {
		if err := ensureCertFiles(cfg.MitmCert, cfg.MitmKey); err != nil {
			log.Fatalf("[FATAL] Could not prepare MITM certificate: %v", err)
		}
		passedArgs = append(passedArgs, "-sslCert", cfg.MitmCert, "-sslKey", cfg.MitmKey)
		log.Printf("[SYSTEM] MITM TLS enabled: %s / %s", cfg.MitmCert, cfg.MitmKey)
	} else if argCert != "" && argKey != "" {
		// Not in MITM mode: restore the certificate pair the real servers use
		passedArgs = append(passedArgs, "-sslCert", argCert, "-sslKey", argKey)
	}

	// The capture layout is HARDCODED and relative to the working directory:
	// captures/<fecha>_<commit>/{main.log, secondary.log, proxy.log, metadata.json}.
	// It deliberately ignores MAIN_LOG_PATH / SECONDARY_LOG_PATH: the manager used
	// to choose the names, which is how main_log_NN.csv and secondary_log_NN.csv
	// ended up side by side with NN coming from outside and the two binaries
	// swapped from capture to capture.
	captureDir, err := NewCaptureDir()
	if err != nil {
		log.Fatalf("[FATAL] could not create the capture directory %s: %v", captureDir, err)
	}
	cfg.MainLogPath = filepath.Join(captureDir, MainLogName)
	cfg.SecondaryLogPath = filepath.Join(captureDir, SecondaryLogName)
	cfg.LogFilePath = filepath.Join(captureDir, ProxyLogName)
	if ignored := os.Getenv("MAIN_LOG_PATH"); ignored != "" {
		log.Printf("[CAPTURE] ignoring MAIN_LOG_PATH=%s and SECONDARY_LOG_PATH=%s: the layout is fixed",
			ignored, os.Getenv("SECONDARY_LOG_PATH"))
	}

	// Open the log file: proxy logs and child output go to the terminal and the file
	logFile, err := os.OpenFile(cfg.LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("[FATAL] Could not open log file %s: %v", cfg.LogFilePath, err)
	}
	defer func(logFile *os.File) {
		_ = logFile.Close()
	}(logFile)
	log.SetOutput(io.MultiWriter(os.Stderr, logFile))
	log.Printf("[SYSTEM] Logging to terminal and %s", cfg.LogFilePath)

	// Initialize CSV Loggers: a background goroutine appends each complete frame
	// to its CSV progressively, so a hard kill only loses the frames still in the
	// in-memory queue (<= 4096) rather than the whole capture. The main
	// logger records the primary flow (game <-> Target 1); the secondary logger
	// records Target 2 responses that are drained.
	logger := NewCSVLogger(cfg.MainLogPath)
	defer func() { _ = logger.Close() }()
	secondaryLogger := NewCSVLogger(cfg.SecondaryLogPath)
	defer func() { _ = secondaryLogger.Close() }()

	// The proxy is built and bound HERE, before the port is advertised and before
	// the children start, because the port the first listener binds is what the
	// relay broadcast tells the game and what both children are handed. Building
	// it further down, after pm.Start(), is what made that number a guess.
	proxy, err := NewTCPProxyServer(cfg, logger, secondaryLogger)
	if err != nil {
		log.Fatalf("[FATAL] Could not initialize proxy: %v", err)
	}
	if err := proxy.Bind(); err != nil {
		log.Fatalf("[FATAL] Server error: %v", err)
	}
	// From here cfg.ProxyPort is the port that is listening, whether the operator
	// named it or the OS assigned it at bind time.
	cfg.ProxyPort = proxy.ClientPort()

	if bsPort > 0 || wsPort > 0 || oobPort > 0 {
		log.Printf("[SYSTEM] Taking over server ports from command line: bs=%d ws=%d oob=%d", cfg.ProxyPort, cfg.ProxyWSPort, cfg.ProxyOOBPort)
	}
	if bsPort <= 0 {
		log.Printf("[SYSTEM] No -bsPort given, using OS-assigned port %d", cfg.ProxyPort)
	}

	// Identify both binaries by CONTENT and write the capture metadata before
	// anything starts, so which log is ours and which is the original is recorded
	// automatically instead of being inferred afterwards from file names or from a
	// log banner. The file names have already been wrong in both directions, and
	// guessing it is what produced three wrong conclusions about the StartObserve
	// byte.
	// Built before launching anything, because the capture records the exact
	// command lines each side was started with.
	pm := NewProcessManager(cfg, passedArgs, logFile)
	if roles, err := WriteRoles(captureDir, cfg.Exe1Path, cfg.Exe2Path, MainLogName, SecondaryLogName); err != nil {
		fmt.Fprintf(logFile, "[IDENTITY] could not identify the binaries: %v\n", err)
	} else {
		if err := WriteCaptureMeta(captureDir, roles, pm.Args1(), pm.Args2()); err != nil {
			fmt.Fprintf(logFile, "[IDENTITY] could not write the capture metadata: %v\n", err)
		}
		fmt.Fprintf(logFile, "[CAPTURE] %s\n", DescribeCapture(captureDir, roles))
		if roles.Warning != "" {
			fmt.Fprintf(logFile, "[IDENTITY] WARNING: %s\n", roles.Warning)
		}
	}

	// Initialize and launch child executables
	pm.Start()

	// Handle system termination signals (CTRL+C, SIGTERM)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\n[SYSTEM] Shutting down proxy server...")
		proxy.Stop()
		pm.Terminate()
		_ = logger.Close()
		_ = secondaryLogger.Close()
		os.Exit(0)
	}()

	// Start Proxy Server: the listeners are already bound, so this only blocks
	// until Stop.
	if err := proxy.Start(); err != nil {
		// os.Exit skips every deferred close, so the children and the captures
		// are torn down here.
		proxy.Stop()
		pm.Terminate()
		_ = logger.Close()
		_ = secondaryLogger.Close()
		log.Fatalf("[FATAL] Server error: %v", err)
	}
	// Start blocks until Stop; the signal handler path performs its own
	// cleanup, and the deferred closes are idempotent for the normal path.
}
