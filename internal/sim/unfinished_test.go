package sim

import (
	"errors"
	"strings"
	"testing"
)

// WHAT THIS FILE IS
//
// A replay that did not finish has to stop the run, and this is the function that
// decides it.
//
// The distinction is between two things that look identical in the connection
// table: a server that answered nothing, and a harness that collected nothing.
// Reporting the first when the second happened is how a load spike on the machine
// running the tests turns into three must-match findings about the server, and it
// is the shape of failure SESSION-JOURNAL.md 24 and 27 are about.
//
// It is not a relaxation. Before this, the run failed too, with a verdict and a
// crop verdict attached to it. Now it fails with the reason, which is the only
// thing that changed: what is compared, and how strictly, is untouched.

// TestUnfinishedReplayIsAnError pins the whole contract: no failures is no error,
// and one failure is an error that names what did not finish and does NOT claim a
// verdict.
func TestUnfinishedReplayIsAnError(t *testing.T) {
	if err := unfinishedReplay(nil); err != nil {
		t.Errorf("a replay with every connection driven returned %v. Nothing failed, so there is nothing to report", err)
	}
	if err := unfinishedReplay([]error{}); err != nil {
		t.Errorf("an empty failure list returned %v", err)
	}

	err := unfinishedReplay([]error{errors.New("client port 51002 (webSocketPort): websocket upgrade response: i/o timeout")})
	if err == nil {
		t.Fatal("a connection the harness could not drive returned no error, so the run goes on to report the empty " +
			"side of it as a divergence of the server")
	}
	msg := err.Error()
	for _, want := range []string{
		"the replay did not complete",
		"websocket upgrade response: i/o timeout",
		"webSocketPort",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("the error does not mention %q, and the error is the only thing a reader gets: %q", want, msg)
		}
	}
	for _, unwanted := range []string{"must-match", "divergence", "crop"} {
		if strings.Contains(msg, unwanted) {
			t.Errorf("the error says %q, which is a verdict about the server or the crop. The harness knows it did not "+
				"finish, and it knows nothing else: %q", unwanted, msg)
		}
	}
	// The cause has to survive wrapping: it is the thing that says WHAT did not
	// finish, and errors.Is is how a caller can tell one run's failure from
	// another's.
	cause := errors.New("websocket upgrade response: i/o timeout")
	wrapped := unfinishedReplay([]error{fmtWrap("client port 51002 (webSocketPort)", cause)})
	if !errors.Is(wrapped, cause) {
		t.Errorf("the cause is not reachable with errors.Is: %v", wrapped)
	}
}

// fmtWrap is errors.Wrapf without the dependency: what replay puts around each
// connection's failure before collecting it.
func fmtWrap(msg string, err error) error { return &wrapped{msg: msg, err: err} }

type wrapped struct {
	msg string
	err error
}

func (w *wrapped) Error() string { return w.msg + ": " + w.err.Error() }
func (w *wrapped) Unwrap() error { return w.err }