package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/luskaner/ageBattleServer/internal/defaults"
	"github.com/luskaner/ageBattleServer/internal/logger"
	"github.com/luskaner/ageBattleServer/internal/server"
)

// LogModName and LogModLevel hold the -logmod/-logmodlevel flag values; the
// command line reads them back to configure per-module logging.
var (
	LogModName  string
	LogModLevel string
)

func ParseFlags() *server.Config {
	cfg := &server.Config{}

	flag.StringVar(&cfg.Name, "name", "My Server", "Readable server name")
	bsPort, _ := strconv.ParseUint(defaults.DefaultBsPort, 10, 32)
	flag.UintVar(&cfg.BsPort, "bsport", uint(bsPort), "Port for incoming game connections (TCP)")
	flag.UintVar(&cfg.WebSocketPort, "websocketport", 0, "WebSocket listen port")
	flag.UintVar(&cfg.OutOfBandPort, "outofbandport", 27212, "WebSocket ping OOB port (TCP, 0 = disabled)")
	relayPort := uint(0)
	flag.UintVar(&relayPort, "relaybroadcastport", 0, "Port for presence broadcast (UDP, default 9999 and 8888, 0 = disabled)")
	flag.StringVar(&cfg.LobbyHost, "lobbyhost", "localhost", "Heartbeat/lobby server host")
	flag.UintVar(&cfg.LobbyPort, "lobbyport", 0, "Heartbeat/lobby server port (0 = disabled)")
	flag.StringVar(&cfg.PublicIP, "publicip", "", "Server public IP")
	publicPort, _ := strconv.ParseUint(defaults.DefaultPublicPort, 10, 32)
	flag.UintVar(&cfg.PublicPort, "publicport", uint(publicPort), "Public port")
	flag.StringVar(&cfg.SSLCert, "sslcert", "", "SSL certificate path")
	flag.StringVar(&cfg.SSLKey, "sslkey", "", "SSL private key path")
	flag.StringVar(&cfg.UploadVia, "uploadvia", "", "Upload method: lobby")
	flag.StringVar(&cfg.ReplayHost, "replayhost", "", "Host for replay upload")
	flag.UintVar(&cfg.ReplayPort, "replayport", 0, "Port for replay upload")
	flag.StringVar(&cfg.Login, "login", "", "DB user (mysql/odbc)")
	flag.StringVar(&cfg.Password, "password", "", "DB password (mysql/odbc)")
	flag.StringVar(&cfg.ReplayDBName, "replaydbname", "ww_1_relayhistory", "DB name")
	flag.StringVar(&cfg.OdbcDriver, "odbcdriver", "MysqlTest", "ODBC driver")
	flag.StringVar(&cfg.Region, "region", "global", "Server region")
	flag.UintVar(&cfg.ObserverMax, "observermax", 512, "Max concurrent observers")
	flag.UintVar(&cfg.ObserverUpdateIntervalMS, "observerupdateintervalms", 8, "Observer update interval (ms)")
	flag.UintVar(&cfg.ObsUpdateMaxLoadPct, "obsupdatemaxloadpct", 0, "Max CPU percentage for updates")
	flag.UintVar(&cfg.ObserverSocketSkipMS, "observersocketskipms", 8, "Socket skip interval (ms)")
	flag.Float64Var(&cfg.SimulationPeriod, "simulationperiod", 0, "Simulation period (ms)")
	// The value is divided by 8 on the way in and used as
	// size = ceil(peers * f) with peers floored at 8, where f = value/8. The
	// default is f = 1.0, i.e. one page per peer. The store is always finite: a
	// store that never rolled over would leave the first retained frame at 0 and
	// StartObserve would refuse every observer forever. So 0 means the protocol's
	// default of one page per peer, not "unlimited".
	flag.UintVar(&cfg.MaxPagesPerMatch, "maxpagespermatch", 0, "Max pages per session, scaled by peers/8 (0 = one page per peer)")
	// -maxPagesToStorePerSession is the name the upstream server spells this flag
	// with. Flag parsing is strict (an unknown flag exits 2), so without this
	// alias a launcher that passes that name would be rejected. Both names set the
	// same field; if both are given, the last wins.
	flag.UintVar(&cfg.MaxPagesPerMatch, "maxpagestostorepersession", 0, "Alias of -maxPagesPerMatch, the upstream name for this flag")
	flag.UintVar(&cfg.MaxSessions, "maxsessions", 0, "Max active sessions (0 = no limit)")
	flag.UintVar(&cfg.PageSize, "pagesize", 65536, "Page size")
	flag.BoolVar(&cfg.AllowTimeoutOverride, "allowtimeoutoverride", false, "Allow timeout override")
	flag.StringVar(&cfg.HeartbeatSecurityKey, "heartbeatsecuritykey", "", "Heartbeat security key")
	flag.StringVar(&cfg.ClusterName, "clustername", "", "Cluster identifier")
	flag.BoolVar(&cfg.EnableStats, "enablestats", false, "Enable stats (Munin)")
	flag.StringVar(&cfg.MuninLog, "muninlog", "", "Munin log file")
	flag.UintVar(&cfg.MuninReportDelay, "muninreportdelay", 60000, "Munin report interval (ms)")
	flag.StringVar(&cfg.LogLevel, "loglevel", "debug", "Global logging level (none, error, warning, info, debug)")
	flag.StringVar(&LogModName, "logmod", "", "Specific module for logging")
	flag.StringVar(&LogModLevel, "logmodlevel", "debug", "Module logging level (none, error, warning, info, debug)")
	flag.StringVar(&cfg.ServerLog, "serverlog", "", "(deprecated) Server log file")
	flag.UintVar(&cfg.LogFileCount, "logfilecount", 0, "(deprecated) Number of log files")
	flag.UintVar(&cfg.LogFileSize, "logfilesize", 0, "(deprecated) Maximum log file size")

	var help bool
	flag.BoolVar(&help, "help", false, "Show help")
	flag.BoolVar(&help, "h", false, "Show help")
	flag.BoolVar(&help, "?", false, "Show help")

	if err := flag.CommandLine.Parse(normalizeArgs(os.Args[1:])); err != nil {
		logger.Error("Main", "parseFlags", "%v", err)
		os.Exit(2)
	}

	// The relay broadcast targets both default ports unless -relaybroadcastPort
	// is given explicitly, in which case only that single port is used (0 = disabled).
	relaySet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "relaybroadcastport" {
			relaySet = true
		}
	})
	if relaySet {
		if relayPort != 0 {
			cfg.RelayBroadcastPorts = []uint{relayPort}
		}
	} else {
		primary, _ := strconv.ParseUint(defaults.DefaultRelayBroadcastPort, 10, 32)
		secondary, _ := strconv.ParseUint(defaults.SecondaryRelayBroadcastPort, 10, 32)
		cfg.RelayBroadcastPorts = []uint{uint(primary), uint(secondary)}
	}

	if help {
		printUsage()
		os.Exit(0)
	}

	if cfg.ServerLog != "" || cfg.LogFileCount > 0 || cfg.LogFileSize > 0 {
		logger.Warning("Main", "parseFlags", "serverLog, logFileCount and logFileSize are deprecated and ignored")
	}

	switch cfg.UploadVia {
	case "", "lobby":
	case "mysql", "odbc":
		logger.Error("Main", "parseFlags", "uploadVia=%s not supported, only 'lobby' is available", cfg.UploadVia)
		os.Exit(1)
	default:
		logger.Error("Main", "parseFlags", "invalid uploadVia=%s, only 'lobby' is supported", cfg.UploadVia)
		os.Exit(1)
	}

	var observerUpdateIntervalSet, observerSocketSkipSet, simulationPeriodSet bool
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "observerupdateintervalms":
			observerUpdateIntervalSet = true
		case "observersocketskipms":
			observerSocketSkipSet = true
		case "simulationperiod":
			simulationPeriodSet = true
		}
	})
	if (observerUpdateIntervalSet || observerSocketSkipSet) && simulationPeriodSet && cfg.SimulationPeriod == 0 {
		fmt.Println("Usage: -simulationPeriod shouldn't be 0 when observerSocketSkipMS or observerUpdateIntervalMS is given.")
		os.Exit(1)
	}

	// ObserverUpdateIntervalMS is converted from milliseconds to
	// simulation ticks only when the flag was actually given. With the flag absent
	// the value stays the one it was initialised with, 8 TICKS, and not 8 ms. The
	// two are not the same thing: 8 ms at a 125 ms period is 0 ticks, and the
	// observer update strides the observer list by this number, so 0 would be the
	// protocol's own infinite loop and 1 is eight visits a second where the
	// recorded run makes one. Capture 2026-09-29_012528_03cb67c7cf19 measures an
	// observer served 200 frames a second, once every ninth tick.
	cfg.ObserverUpdateIntervalSet = observerUpdateIntervalSet

	return cfg
}

func printUsage() {
	fmt.Println("Usage: battleserver [OPTION]...")
	fmt.Println()
	fmt.Println("Network Options:")
	fmt.Println("  -name string                     Readable server name (default \"My Server\")")
	fmt.Printf("  -bsPort uint                     Port for game connections (TCP) (default %s)\n", defaults.DefaultBsPort)
	fmt.Println("  -webSocketPort uint              WebSocket port")
	fmt.Println("  -outOfBandPort uint              WebSocket ping OOB port (TCP, default 27212, 0 = disabled)")
	fmt.Printf("  -relaybroadcastPort uint         Port presence broadcast (UDP, default %s and %s, 0 = disabled)\n", defaults.DefaultRelayBroadcastPort, defaults.SecondaryRelayBroadcastPort)
	fmt.Println("  -lobbyHost string                Heartbeat/lobby server host (default \"localhost\")")
	fmt.Println("  -lobbyPort uint                  Heartbeat/lobby server port (0 = disabled)")
	fmt.Println("  -publicIP string                 Server public IP")
	fmt.Printf("  -publicPort uint                 Public port (default %s)\n", defaults.DefaultPublicPort)
	fmt.Println()
	fmt.Println("SSL/TLS Options:")
	fmt.Println("  -sslCert string                  SSL certificate path")
	fmt.Println("  -sslKey string                   SSL private key path")
	fmt.Println()
	fmt.Println("Replay Options:")
	fmt.Println("  -uploadVia string                Upload method (only 'lobby')")
	fmt.Println("  -replayHost string                Host for replay upload")
	fmt.Println("  -replayPort uint                 Port for replay upload")
	fmt.Println("  -login string                    DB user (mysql/odbc)")
	fmt.Println("  -password string                 DB password (mysql/odbc)")
	fmt.Println("  -replayDBName string             DB name (default \"ww_1_relayhistory\")")
	fmt.Println("  -odbcDriver string               ODBC driver (default \"MysqlTest\")")
	fmt.Println()
	fmt.Println("Session Options:")
	fmt.Println("  -observerMax uint                Max observers (default 512)")
	fmt.Println("  -observerUpdateIntervalMS uint   Observer update interval (ms)")
	fmt.Println("  -obsUpdateMaxLoadPct uint        Max CPU percentage for updates")
	fmt.Println("  -observerSocketSkipMS uint       Socket skip interval (ms)")
	fmt.Println("  -simulationPeriod float          Simulation period (ms)")
	fmt.Println("  -maxPagesPerMatch uint           Max pages per session, scaled by peers/8 (0 = original default)")
	fmt.Println("  -maxPagesToStorePerSession uint  Alias of -maxPagesPerMatch (the upstream name)")
	fmt.Println("  -maxSessions uint               Max active sessions (0 = no limit)")
	fmt.Println("  -pageSize uint                   Page size (default 65536)")
	fmt.Println("  -allowTimeoutOverride            Allow timeout override")
	fmt.Println("  -heartbeatSecurityKey string     Heartbeat security key")
	fmt.Println()
	fmt.Println("Observer Options:")
	fmt.Println("  -region string                   Region (default \"global\")")
	fmt.Println("  -clusterName string              Cluster identifier")
	fmt.Println()
	fmt.Println("Logging Options:")
	fmt.Println("  -logLevel string                 Global logging level (none, error, warning, info, debug)")
	fmt.Println("  -logMod string                   Specific module for logging")
	fmt.Println("  -logModLevel string              Module logging level (none, error, warning, info, debug)")
	fmt.Println()
	fmt.Println("Monitoring Options:")
	fmt.Println("  -enableStats                   Enable Munin stats (required with -muninLog)")
	fmt.Println("  -muninLog string                 Munin log file (ignored without -enableStats)")
	fmt.Println("  -muninReportDelay uint         Munin report interval (ms, default 60000)")
	fmt.Println()
	fmt.Println("Deprecated Options:")
	fmt.Println("  -serverLog string                (deprecated)")
	fmt.Println("  -logFileCount uint               (deprecated)")
	fmt.Println("  -logFileSize uint                (deprecated)")
	fmt.Println()
	fmt.Println("Help:")
	fmt.Println("  -help, -h, -?                    Show this help")
}

func normalizeArgs(args []string) []string {
	result := make([]string, len(args))
	copy(result, args)
	for i := range result {
		if result[i] == "--" {
			break
		}
		if strings.HasPrefix(result[i], "-") && !strings.HasPrefix(result[i], "--") {
			// Lowercase only the flag name: a value passed with -flag=value
			// (case-sensitive: secrets, file paths) must keep its case.
			name, value, hasValue := strings.Cut(result[i], "=")
			result[i] = "-" + strings.ToLower(strings.TrimLeft(name, "-"))
			if hasValue {
				result[i] += "=" + value
			}
		}
	}
	return result
}
