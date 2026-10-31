package sim

import (
	"strings"
	"testing"

	"github.com/luskaner/ageBattleServer/internal/server"
)

// The TLS decision is a property of the port, and it is the one that a whole
// capture was misread by.
//
// 2026-09-30_201122_c4e7c284ff1c is a bsPort capture taken with -sslCert and
// -sslKey in force. The replay decided TLS once for the whole run, dialled TLS into
// the bsPort, which this server builds WITHOUT the pair
// (internal/server/server.go:195, and :199 is the listener that does get it), read
// the ClientHello as an RLink length and answered nothing. The report said
// ours=0 on both connections against 24 and 2 messages from the original, with no
// must-match divergence and nothing anywhere saying why.
//
// Two things had to come out of it and only one did. The decision is now taken by
// this function, and the other half is the declaration: a replay that talks
// plaintext to a port the recording did not describe has to say so, because a
// server that answered nothing looks exactly the same in the connection table as
// one that was sent something it did not understand.

// connsWith is a capture whose connections all carry the given transports.
func connsWith(transports ...string) []*Conn {
	out := make([]*Conn, 0, len(transports))
	for _, tr := range transports {
		out = append(out, &Conn{ClientPort: len(out) + 1, Transport: tr})
	}
	return out
}

func TestReplayTLSIsDecidedByThePortAndSaysWhenItIsNot(t *testing.T) {
	certed := &server.Config{SSLCert: "cert.pem", SSLKey: "key.pem"}

	t.Run("a WebSocket-only capture with the pair in force is TLS and says nothing", func(t *testing.T) {
		useTLS, dev := replayTLS(certed, connsWith(FramingWS, FramingWS))
		if !useTLS {
			t.Error("a capture whose only connections are WebSocket ones and whose command line names the pair " +
				"is the case where the harness has to dial TLS: the first real capture produced 547 messages from the " +
				"original and none from us for exactly this reason")
		}
		if len(dev) != 0 {
			t.Errorf("nothing is being changed on purpose here, so there is nothing to declare: %v", dev)
		}
	})

	t.Run("a capture with a plaintext transport dials it in plaintext and says so", func(t *testing.T) {
		useTLS, dev := replayTLS(certed, connsWith(FramingRLink))
		if !useTLS {
			t.Error("the certificate pair is in force, so the WebSocket port is still TLS; only the bsPort is plaintext")
		}
		if len(dev) != 1 {
			t.Fatalf("a bsPort capture under a certificate pair has to be declared, got %d deviation(s)", len(dev))
		}
		d := dev[0]
		if d.Field != "bsPort over TLS" {
			t.Errorf("deviation field = %q, want %q", d.Field, "bsPort over TLS")
		}
		for _, want := range []string{"server.go:195", ":199", "answered nothing"} {
			if !strings.Contains(d.Why, want) {
				t.Errorf("the reason has to say %q, got: %s", want, d.Why)
			}
		}
	})

	t.Run("no pair in force is plaintext and nothing to declare", func(t *testing.T) {
		for _, cfg := range []*server.Config{
			{},
			{SSLCert: "cert.pem"},
			{SSLKey: "key.pem"},
		} {
			useTLS, dev := replayTLS(cfg, connsWith(FramingRLink, FramingWS))
			if useTLS {
				t.Errorf("%+v: a half-configured pair is not a pair, and dialling TLS with one half of it is the failure", cfg)
			}
			if len(dev) != 0 {
				t.Errorf("%+v: with no pair there is no change on purpose to declare, got %v", cfg, dev)
			}
		}
	})

	t.Run("a connection the reader could not classify is not a plaintext one", func(t *testing.T) {
		// Transport "" is what finalise leaves on a connection that holds no
		// recognisable frame, most likely the out-of-band port. It is not evidence
		// that the battle port was plaintext, and declaring on it would be a claim
		// the capture cannot support.
		useTLS, dev := replayTLS(certed, connsWith(""))
		if !useTLS {
			t.Error("an unclassified connection says nothing about the ports, so the pair still decides")
		}
		if len(dev) != 0 {
			t.Errorf("got %d deviation(s) for a connection with no transport at all: %v", len(dev), dev)
		}
	})
}

// TestReplayTLSOnTheBsPortCaptureIsWhatTheReportSays ties the function to the
// capture it exists for: the bsPort capture really does name the pair, and really
// does hold a connection that is not WebSocket, so a replay of it has to declare
// the plaintext port.
func TestReplayTLSOnTheBsPortCaptureIsWhatTheReportSays(t *testing.T) {
	cap, err := Load("testdata/2026-09-30_201122_c4e7c284ff1c")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	cfg, dropped, err := ConfigFromArgs(cap.Args())
	if err != nil {
		t.Fatalf("ConfigFromArgs: %v", err)
	}
	if len(dropped) != 0 {
		t.Fatalf("flags dropped: %v", dropped)
	}
	useTLS, dev := replayTLS(cfg, cap.Conns)
	if !useTLS {
		t.Error("this capture's command line names -sslCert and -sslKey and the files exist on this machine, " +
			"so the WebSocket port is TLS. prepareConfig blanks the pair only when the files are not here")
	}
	if len(dev) != 1 {
		t.Fatalf("this capture's connections are all on the bsPort, so the declaration is due: %v", cap.Conns)
	}
	if dev[0].Running != "plaintext" {
		t.Errorf("the bsPort is dialled in plaintext, got %q", dev[0].Running)
	}
	// The whole point of the property: no connection of this capture is WebSocket.
	for _, c := range cap.Conns {
		if c.Transport == FramingWS {
			t.Errorf("client port %d came out as WebSocket on a bsPort capture, which would hide the bug", c.ClientPort)
		}
	}
}
