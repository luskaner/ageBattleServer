package main

import (
	"errors"
	"flag"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/luskaner/ageBattleServer/internal/config"
)

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
}

func runHelper(t *testing.T, testCase string) (int, string) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable error: %v", err)
	}
	cmd := exec.Command(exe, "-test.run=^TestHelperProcess$")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1", "GO_HELPER_CASE="+testCase)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return 0, string(out)
	}
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		t.Fatalf("unexpected error type: %v", err)
	}
	return ee.ExitCode(), string(out)
}
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)
	os.Args = []string{"battleserver"}
	resetFlags()
	switch os.Getenv("GO_HELPER_CASE") {
	case "help":
		os.Args = append(os.Args, "-h")
	case "help-long":
		os.Args = append(os.Args, "--help")
	case "uploadvia-mysql":
		os.Args = append(os.Args, "-uploadvia", "mysql")
	case "uploadvia-invalid":
		os.Args = append(os.Args, "-uploadvia", "bogus")
	case "simulation-period-zero":
		os.Args = append(os.Args, "-observersocketskipms", "8", "-simulationperiod", "0")
	case "main-bad-loglevel":
		os.Args = append(os.Args, "-loglevel", "bogus")
		main()
		_, _ = os.Stdout.WriteString("main did not exit\n")
		os.Exit(99)
	case "main-bad-modlevel":
		os.Args = append(os.Args, "-logmod", "Server", "-logmodlevel", "bogus")
		main()
		_, _ = os.Stdout.WriteString("main did not exit\n")
		os.Exit(99)
	case "main-new-fails":
		os.Args = append(os.Args, "-logmod", "Server", "-logmodlevel", "debug", "-websocketport", "27013", "-sslcert", "/nonexistent/cert.pem", "-sslkey", "/nonexistent/key.pem")
		main()
		_, _ = os.Stdout.WriteString("main did not exit\n")
		os.Exit(99)
	default:
		_, _ = os.Stdout.WriteString("unknown helper case\n")
		os.Exit(99)
	}
	config.ParseFlags()
	_, _ = os.Stdout.WriteString("did-not-exit\n")
	os.Exit(99)
}

func TestParseFlagsExitHelp(t *testing.T) {
	code, out := runHelper(t, "help")
	if code != 0 {
		t.Errorf("exit code for -h = %d, want 0", code)
	}
	if !strings.Contains(out, "Usage: battleserver") {
		t.Error("expected usage output on -h")
	}
}

func TestParseFlagsExitHelpLong(t *testing.T) {
	code, out := runHelper(t, "help-long")
	if code != 0 {
		t.Errorf("exit code for --help = %d, want 0", code)
	}
	if !strings.Contains(out, "Usage: battleserver") {
		t.Error("expected usage output on --help")
	}
}

func TestParseFlagsExitUploadViaMysql(t *testing.T) {
	code, _ := runHelper(t, "uploadvia-mysql")
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestParseFlagsExitUploadViaInvalid(t *testing.T) {
	code, _ := runHelper(t, "uploadvia-invalid")
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestParseFlagsExitSimulationPeriodZero(t *testing.T) {
	code, out := runHelper(t, "simulation-period-zero")
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	if !strings.Contains(out, "simulationPeriod shouldn't be 0") {
		t.Errorf("expected conflict message, got %q", out)
	}
}

func TestMainExitBadLogLevel(t *testing.T) {
	if code, _ := runHelper(t, "main-bad-loglevel"); code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestMainExitBadModLevel(t *testing.T) {
	if code, _ := runHelper(t, "main-bad-modlevel"); code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}

func TestMainExitNewFails(t *testing.T) {
	if code, _ := runHelper(t, "main-new-fails"); code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
}
