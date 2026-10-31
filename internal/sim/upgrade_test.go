package sim

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

// The WebSocket upgrade response is the one part of the exchange that used to be
// thrown away, and one header the original sends was living in it. These tests
// are about the rule that now looks at it, and they are written so that a rule
// which compared nothing, or compared the wrong thing, fails.

// The bytes the original answered 2026-09-30_010531 with, in the two TCP records
// the proxy logged them in (86 bytes and then 70).
var (
	originalResponsePart1 = []byte("HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: ICX+Yqv66")
	originalResponsePart2 = []byte("kxgM0FcWaLWlFLwTAI=\r\nServer: WebSocket++/0.8.2\r\nUpgrade: websocket\r\n\r\n")

	// originalRequest is the client's GET from the same capture. Its key is the
	// 26 A's of the capture, and ICX+Yqv66kxgM0FcWaLWlFLwTAI= is the RFC 6455
	// derivation of it.
	originalRequest = []byte("GET / HTTP/1.1\r\nConnection: Upgrade\r\nHost: 172.23.16.1:64602\r\n" +
		"Sec-WebSocket-Key: AAAAAAAAAAAAAAAAAAAAAA==\r\nSec-WebSocket-Version: 13\r\nUpgrade: websocket\r\n" +
		"User-Agent: WebSocket++/0.8.2\r\n\r\n")

	// replayKey is the fixed key the harness sends (replay.go wsHandshake), so
	// that its own Sec-WebSocket-Accept is reproducible.
	replayKey = base64.StdEncoding.EncodeToString([]byte("battleserver-sim0"))

	// oursBeforeFix is what gws wrote before the Server header was added.
	oursBeforeFixPart1 = []byte("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSock")
	oursBeforeFixPart2 = []byte("et-Accept: ICX+Yqv66kxgM0FcWaLWlFLwTAI=\r\n\r\n")
)

// handshake builds a connection with an upgrade exchange on it.
func handshake(request, response []byte) *Conn {
	return &Conn{ClientPort: 1, Transport: FramingWS, upHandshake: request, downHandshake: response}
}

func ruleVerdict(t *testing.T, want, got *Conn) (*Report, bool) {
	t.Helper()
	rep := &Report{started: time.Now(), origin: time.Now()}
	upgradeResponse(rep, want, got)
	return rep, len(rep.MustMatch) > 0
}

// TestUpgradeResponseCatchesTheMissingServerHeader is the falsification of the
// claim in SESSION-JOURNAL.md 25 that this was "our WebSocket ping 0x65 where the
// original sends a pong 0x6B". It was never a ping: 0x65 is the 'e' of
// "Sec-WebSocket-Accept" in OUR response and 0x6B is the 'k' of the accept key
// in the original's, both continuation records of the same 101 answer. What the
// bytes really say is that one side sends a Server header and the other does not,
// and that is what the rule has to say.
func TestUpgradeResponseCatchesTheMissingServerHeader(t *testing.T) {
	want := handshake(originalRequest, append(append([]byte{}, originalResponsePart1...), originalResponsePart2...))
	got := handshake(originalRequest, append(append([]byte{}, oursBeforeFixPart1...), oursBeforeFixPart2...))

	rep, fatal := ruleVerdict(t, want, got)
	if !fatal {
		t.Fatal("the two responses carry a different set of headers (the original sends Server and we do not) and the rule said nothing")
	}
	f := rep.MustMatch[0]
	if f.Rule != RuleUpgradeResponse {
		t.Fatalf("finding is under %q, want %q", f.Rule, RuleUpgradeResponse)
	}
	if !strings.Contains(f.Detail, "Server") {
		t.Fatalf("the finding does not name the header that differs: %s", f.Detail)
	}
}

// TestUpgradeResponsePassesOnceTheHeaderIsThere is the other direction: the same
// comparison with the Server header added must be clean of must-match findings,
// even though the order still differs. A rule that failed on the order would be
// failing on a byte no client reads, and one that passed with the header missing
// would be the rule above with its mouth shut.
func TestUpgradeResponsePassesOnceTheHeaderIsThere(t *testing.T) {
	want := handshake(originalRequest, append(append([]byte{}, originalResponsePart1...), originalResponsePart2...))
	ours := "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + expectedAccept(replayKey) + "\r\nServer: WebSocket++/0.8.2\r\n\r\n"
	oursReq := "GET / HTTP/1.1\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: " + replayKey + "\r\nSec-WebSocket-Version: 13\r\n\r\n"

	rep, fatal := ruleVerdict(t, want, handshake([]byte(oursReq), []byte(ours)))
	if fatal {
		t.Fatalf("the header sets match and each accept is the derivation of its own key, so the verdict must not be fatal: %s", rep.MustMatch[0].Detail)
	}
	if len(rep.Informative) == 0 {
		t.Fatal("the order still differs and nothing said so: a difference that is not compared has to be reported as one")
	}
	if !strings.Contains(rep.Informative[0].Detail, "not must-match") {
		t.Fatalf("the finding does not say why the order is not compared: %s", rep.Informative[0].Detail)
	}
}

// TestUpgradeResponseDoesNotCompareTheTwoAcceptStrings is the correction that the
// capture forced. The value is base64(SHA1(key+GUID)), the replay sends its own
// key, and so the two strings are different for a correct server. Comparing them
// is a rule that fails on correctness, and the first version of this file did
// exactly that: the run came out with 2 must-match divergences saying the
// original answered ICX+Yqv66kxgM0FcWaLWlFLwTAI= and we answered
// +HSqxxdw1hvHfQALlH+e14eO1X4=, which are the correct derivations of two
// different keys.
func TestUpgradeResponseDoesNotCompareTheTwoAcceptStrings(t *testing.T) {
	want := handshake(originalRequest, append(append([]byte{}, originalResponsePart1...), originalResponsePart2...))
	oursReq := "GET / HTTP/1.1\r\nUpgrade: websocket\r\nSec-WebSocket-Key: " + replayKey + "\r\n\r\n"
	ours := "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + expectedAccept(replayKey) + "\r\nServer: WebSocket++/0.8.2\r\n\r\n"

	rep, fatal := ruleVerdict(t, want, handshake([]byte(oursReq), []byte(ours)))
	if fatal {
		t.Fatalf("the two accept values are the correct derivations of two different keys, which is not a difference: %s", rep.MustMatch[0].Detail)
	}
}

// TestUpgradeResponseCatchesAWrongDerivation is the rule that replaces it: each
// side's accept has to be the derivation of the key THAT side's request carried,
// so a server that computes it wrong fails even though it saw its own key.
func TestUpgradeResponseCatchesAWrongDerivation(t *testing.T) {
	want := handshake(originalRequest, append(append([]byte{}, originalResponsePart1...), originalResponsePart2...))
	oursReq := "GET / HTTP/1.1\r\nUpgrade: websocket\r\nSec-WebSocket-Key: " + replayKey + "\r\n\r\n"
	ours := "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: ICX+Yqv66kxgM0FcWaLWlFLwTAI=\r\nServer: WebSocket++/0.8.2\r\n\r\n"

	rep, fatal := ruleVerdict(t, want, handshake([]byte(oursReq), []byte(ours)))
	if !fatal {
		t.Fatal("our accept is the original's accept, which is the derivation of the ORIGINAL's key and not of the one our request carried, and the rule said nothing")
	}
	if !strings.Contains(rep.MustMatch[0].Detail, "derivation") {
		t.Fatalf("the finding does not say what was wrong: %s", rep.MustMatch[0].Detail)
	}
}

// TestExpectedAcceptIsTheRFCValue pins the derivation itself, so a change to the
// GUID or to the hashing cannot quietly make every side pass.
func TestExpectedAcceptIsTheRFCValue(t *testing.T) {
	if got := expectedAccept("AAAAAAAAAAAAAAAAAAAAAA=="); got != "ICX+Yqv66kxgM0FcWaLWlFLwTAI=" {
		t.Fatalf("expectedAccept = %q, want the value the capture records for that key", got)
	}
}

// TestUpgradeResponseCatchesAMissingAccept: a 101 without Sec-WebSocket-Accept is
// what the client checks first, and it refuses the upgrade.
func TestUpgradeResponseCatchesAMissingAccept(t *testing.T) {
	want := handshake(originalRequest, append(append([]byte{}, originalResponsePart1...), originalResponsePart2...))
	ours := "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nServer: WebSocket++/0.8.2\r\n\r\n"
	_, fatal := ruleVerdict(t, want, handshake([]byte("GET / HTTP/1.1\r\nSec-WebSocket-Key: "+replayKey+"\r\n\r\n"), []byte(ours)))
	if !fatal {
		t.Fatal("a 101 with no Sec-WebSocket-Accept was accepted")
	}
}

// TestUpgradeResponseSaysNothingWhenThereIsNoHandshake is the "two empty things"
// guard. A bsPort connection has no HTTP upgrade on either side, and a rule that
// claimed a match there would be reporting a verdict about bytes that do not
// exist.
func TestUpgradeResponseSaysNothingWhenThereIsNoHandshake(t *testing.T) {
	rep, fatal := ruleVerdict(t, &Conn{ClientPort: 1, Transport: FramingRLink}, &Conn{ClientPort: 1, Transport: FramingRLink})
	if fatal || len(rep.Informative) != 0 {
		t.Fatalf("a connection with no upgrade on either side produced %d must-match and %d informative findings; it should produce none",
			len(rep.MustMatch), len(rep.Informative))
	}
}

// TestUpgradeResponseSaysNothingWhenOnlyTheOriginalUpgraded is the other half of
// the same guard: a capture whose truth side went through a TLS-terminating proxy
// records a decrypted handshake and a replay that went out in plaintext has none.
// Comparing them would report a difference that is really the harness's own
// configuration, which is already on the deviation list.
func TestUpgradeResponseSaysNothingWhenOnlyTheOriginalUpgraded(t *testing.T) {
	want := handshake(originalRequest, append(append([]byte{}, originalResponsePart1...), originalResponsePart2...))
	rep, fatal := ruleVerdict(t, want, &Conn{ClientPort: 1, Transport: FramingWS})
	if fatal {
		t.Fatalf("one side upgraded and the other did not, and that was called a must-match: %s", rep.MustMatch[0].Detail)
	}
	if len(rep.Informative) != 1 || !strings.Contains(rep.Informative[0].Detail, "not compared") {
		t.Fatalf("the asymmetry was not reported as such: %+v", rep.Informative)
	}
}

// TestUpgradeHeadersReadsSplitRecords: the original's response arrives in two
// records and the reader used to read the tail as a message whose type was the
// first letter of the accept key. Joining them first is what makes the header set
// readable at all, and this is the shape that proves it.
func TestUpgradeHeadersReadsSplitRecords(t *testing.T) {
	joined := append(append([]byte{}, originalResponsePart1...), originalResponsePart2...)
	status, names, ok := upgradeHeaders(joined)
	if !ok {
		t.Fatal("a response in two records was not recognised as one")
	}
	if status != "HTTP/1.1 101 Switching Protocols" {
		t.Fatalf("status = %q", status)
	}
	want := []string{"Connection", "Sec-WebSocket-Accept", "Server", "Upgrade"}
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Fatalf("headers = %v, want %v", names, want)
	}
	if v, ok := headerValue(joined, "Sec-WebSocket-Accept"); !ok || v != "ICX+Yqv66kxgM0FcWaLWlFLwTAI=" {
		t.Fatalf("Sec-WebSocket-Accept = %q found=%v, want the whole value across both records", v, ok)
	}
}

// TestStripHandshakeKeepsTheBytesItRemoved is what makes the rule possible at
// all: the handshake has to leave the message stream (it cannot start a frame)
// and it has to stay on the connection (nothing else looks at it).
func TestStripHandshakeKeepsTheBytesItRemoved(t *testing.T) {
	body := []byte{0x82, 0x02, 0x0A, 0x00}
	chunks := []chunk{{data: originalResponsePart1}, {data: originalResponsePart2}, {data: body}}

	rest, upgraded, _, kept := stripHandshake(chunks, true)
	if !upgraded {
		t.Fatal("the response was not recognised")
	}
	if len(rest) != 1 || string(streamBytes(rest)) != string(body) {
		t.Fatalf("the message stream after the handshake is %q, want just the battle frame", streamBytes(rest))
	}
	joined := append(append([]byte{}, originalResponsePart1...), originalResponsePart2...)
	if string(kept) != string(joined) {
		t.Fatalf("kept %d bytes of handshake, want %d", len(kept), len(joined))
	}
}

// TestFixtureUpgradeIsRFCDerived keeps the synthetic fixture honest at the level
// it can be, and it reads it off disk rather than off the variables that wrote
// it: the fixture is a committed capture directory and the rule reads THAT, so
// checking the bytes in a variable would only check that the writer and the
// writer agree.
//
// The fixture's "original" must answer with the header set the real original
// sends. It used to answer with three headers and the order gws writes, because
// the rule that reads this did not exist: the upgrade response was thrown away, so
// the fixture could hold any answer at all and nothing would notice.
// ruleUpgradeResponse caught it on the first run and named the missing header,
// which is exactly what a fixture is for.
func TestFixtureUpgradeIsRFCDerived(t *testing.T) {
	cap, err := Load("testdata/synthetic")
	if err != nil {
		t.Fatal(err)
	}
	var saw bool
	for _, c := range cap.Conns {
		_, names, ok := upgradeHeaders(c.downHandshake)
		if !ok {
			continue
		}
		saw = true
		for _, want := range []string{"Connection", "Sec-WebSocket-Accept", "Server", "Upgrade"} {
			if len(missingHeaders([]string{want}, names)) > 0 {
				t.Fatalf("the fixture's original does not send %s; it sends %v", want, names)
			}
		}
		key, ok := headerValue(c.upHandshake, "Sec-WebSocket-Key")
		if !ok {
			t.Fatal("the fixture's client request carries no key, so the accept value cannot be derived")
		}
		accept, ok := headerValue(c.downHandshake, "Sec-WebSocket-Accept")
		if !ok {
			t.Fatal("the fixture's original carries no accept")
		}
		if derived := expectedAccept(key); derived != accept {
			t.Fatalf("the fixture's accept is %s and the RFC 6455 derivation of its own key is %s", accept, derived)
		}
	}
	if !saw {
		t.Fatal("the fixture holds no WebSocket upgrade at all, so this test is checking nothing")
	}
}

// TestUpgradeHeadersIgnoresNonHTTP is the transport-agnostic part: an RLink
// connection's first bytes must not be read as a status line.
func TestUpgradeHeadersIgnoresNonHTTP(t *testing.T) {
	if _, _, ok := upgradeHeaders([]byte{0x02, 0x05, 0x0A, 0x00}); ok {
		t.Fatal("an RLink message was read as an HTTP response")
	}
}
