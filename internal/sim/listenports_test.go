package sim

import (
	"strings"
	"testing"

	"github.com/luskaner/ageBattleServer/internal/server"
)

// WHAT THIS FILE IS
//
// The listen ports of a replay, and why they are not a number this harness picks.
//
// The recorded ports belong to the run that made the capture, so a replay cannot
// bind them. The first version of this asked the OS for a free port, released it
// and let the server bind that number later, and that lost the race: a listener
// bind on Windows is exclusive, so whatever owns the port between the release
// and the bind keeps it and Start fails, and the whole replay dies before a
// single message is compared:
//
//	--- FAIL: TestSim_synthetic (0.04s)
//	    sim: start server: ws listener: listen :31681: listen tcp :31681: bind:
//	    Only one use of each socket address is allowed
//
// Measured on this machine: 1 in 8 runs of `go test -short ./internal/sim`
// under load, and 0 in 30 runs of TestSim_synthetic on its own under the same
// load, which is why it looked like a timing flake in a test that has no timing
// assertion in it (SESSION-JOURNAL.md 38).
//
// The fix is not a retry and not a wider tolerance: the harness asks for
// OS-assigned ports (Config.EphemeralListenPorts), the server binds ":0", and the
// replay reads the two ports back with Server.BoundPorts. There is no window
// because the port is never free. What is left to assert is that the wiring is
// what it says it is, which is what this file does.

// TestPrepareConfigAsksForOSAssignedPorts is the harness half: the ports are not
// chosen here at all. What the replay has to report about them is asserted in
// TestReplayRunsOnThePortsTheServerBound, which is where the numbers are known.
func TestPrepareConfigAsksForOSAssignedPorts(t *testing.T) {
	cfg := &server.Config{BsPort: 15001, WebSocketPort: 15003, OutOfBandPort: 15002}
	run, dev := prepareConfig(cfg)

	if run.EphemeralListenPorts != server.EphemeralBoth {
		t.Errorf("the replay config asks for OS-assigned listen ports on %08b, want both (%08b), so the server is "+
			"going to bind the recorded numbers and the replay will fail to start as soon as they are in use",
			run.EphemeralListenPorts, server.EphemeralBoth)
	}
	if run.BsPort != 0 || run.WebSocketPort != 0 {
		t.Errorf("the replay config still carries listen ports (%d/%d). A non-zero port is a port this harness picked "+
			"and then has to keep, which is the race; the real ones come from Server.BoundPorts after Start",
			run.BsPort, run.WebSocketPort)
	}
	if run.OutOfBandPort != 0 {
		t.Errorf("the out-of-band port is %d, want 0: it is a separate service and no battle message travels through it", run.OutOfBandPort)
	}
	// The rest of the changes still have to be declared, or "only the ports are
	// chosen by the OS" would be a change nobody can see.
	var oob bool
	for _, d := range dev {
		if d.Field == "outOfBandPort" {
			oob = true
		}
	}
	if !oob {
		t.Errorf("prepareConfig no longer declares the out-of-band port it switches off:\n%v", dev)
	}
}

// TestReplayRunsOnThePortsTheServerBound is the end of the wiring: a real replay
// against a real capture, and the report has to name two ports that are not the
// recorded ones. The ports themselves belong to the report, not to this test;
// what is asserted is that the deviation exists and carries two numbers, and
// that neither is a recorded one.
//
// It is the check that would catch a harness which went back to picking a port:
// the deviation would name a number the harness chose, and that number would
// exist before Start instead of after it.
func TestReplayRunsOnThePortsTheServerBound(t *testing.T) {
	const dir = "testdata/synthetic"
	if skipLongCapture(t, dir) {
		return
	}
	rep, err := Check(dir, Options{})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	var line string
	for _, d := range rep.Running {
		if d.Field == "listen ports" {
			line = d.Running
		}
	}
	if line == "" {
		t.Fatalf("the replay reported no listen-port deviation, so nothing says which ports it listened on:\n%v", rep.Running)
	}
	t.Logf("listen ports: %s", line)
	for _, recorded := range []string{"15001", "15003"} {
		if strings.Contains(line, recorded) {
			t.Errorf("the replay reports listen port %s, which is the port the capture recorded. That belongs to the "+
				"run that made the capture, not to this process, and binding it is what used to fail", recorded)
		}
	}
	if !strings.Contains(line, "webSocketPort") || !strings.Contains(line, "bsPort") {
		t.Errorf("the listen-port deviation has to name both transports, got %q", line)
	}
	var why string
	for _, d := range rep.Running {
		if d.Field == "listen ports" {
			why = d.Why
		}
	}
	if !strings.Contains(why, "OS assigns") {
		t.Errorf("the listen-port deviation does not say that the OS assigns the ports, which is the whole reason the "+
			"replay is not binding the recorded ones: %q", why)
	}
}