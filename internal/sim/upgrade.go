package sim

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
)

// The WebSocket upgrade response is bytes the original puts on the wire before a
// single battle message exists, and this harness used to throw them away. That is
// not a neutral choice: a difference in them is invisible, and one was there.
//
// What the capture actually holds, measured on 2026-09-30_010531_a553ef8fd63c and
// again on 2026-09-30_183120_c4e7c284ff1c, both connections of both:
//
//	original: HTTP/1.1 101 / Connection: Upgrade / Sec-WebSocket-Accept: ... /
//	          Server: WebSocket++/0.8.2 / Upgrade: websocket
//	ours:     HTTP/1.1 101 / Upgrade: websocket / Connection: Upgrade /
//	          Sec-WebSocket-Accept: ... / Server: WebSocket++/0.8.2
//
// Two differences, and they are not the same kind of thing:
//
//   - The header SET differed: the original sends a `Server` header we did not.
//     That is must-match. It is one header, it is on the wire in every capture
//     with a webSocketPort, and reproducing it is a constant
//     (internal/server/ws_listener.go).
//   - The header ORDER differs. That is informative and not must-match, and the
//     reason is not "the order felt unimportant": RFC 6455 gives the response no
//     ordering requirement at all, the original's own client is websocketpp (its
//     request carries `User-Agent: WebSocket++/0.8.2`) and websocketpp collects
//     the response headers into a map, and gws writes its own two fixed headers
//     before anything a caller can add. A rule that failed on the order would be
//     failing on a byte no client reads.
//
// Sec-WebSocket-Accept is NOT compared between the two sides, and the reason is
// worth stating because the first version of this rule got it wrong and the
// capture caught it: the value is base64(SHA1(key + GUID)), so it is only
// comparable when both sides saw the SAME key. The proxy does duplicate every
// client write to the two children of a run, so within one run they match, but
// the REPLAY sends its own key (replay.go wsHandshake, a fixed one so a failing
// replay is reproducible) and so its Accept is a different string from a correct
// derivation of a different input. Comparing the two strings would be a rule that
// fails on a correct server.
//
// What is compared instead is each side's own derivation: the Accept on each side
// has to be the RFC 6455 value of the key that side's own request carried. That
// is checkable, it is what the client checks, and it also proves the harness joined
// the two records of a split response correctly (a partial read would fail it).

const websocketGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// RuleUpgradeResponse is the name the report prints for this comparison. It is
// declared with the other rule names in compare.go.

// upgradeHeaders splits an HTTP upgrade response into its status line and its
// header names, in order. It returns ok=false when the bytes are not an HTTP
// response, which is what a capture with no upgrade on this connection looks like.
func upgradeHeaders(b []byte) (status string, names []string, ok bool) {
	if !bytes.HasPrefix(b, []byte("HTTP/1.")) {
		return "", nil, false
	}
	end := bytes.Index(b, []byte("\r\n\r\n"))
	if end < 0 {
		return "", nil, false
	}
	lines := strings.Split(string(b[:end]), "\r\n")
	if len(lines) == 0 {
		return "", nil, false
	}
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		name, _, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		names = append(names, strings.TrimSpace(name))
	}
	return lines[0], names, true
}

// headerValue returns the value of one header, case-insensitively, and whether it
// was there.
func headerValue(b []byte, name string) (string, bool) {
	end := bytes.Index(b, []byte("\r\n\r\n"))
	if end < 0 {
		return "", false
	}
	for _, line := range strings.Split(string(b[:end]), "\r\n")[1:] {
		k, v, found := strings.Cut(line, ":")
		if found && strings.EqualFold(strings.TrimSpace(k), name) {
			return strings.TrimSpace(v), true
		}
	}
	return "", false
}

// expectedAccept is the RFC 6455 Sec-WebSocket-Accept for a key.
func expectedAccept(key string) string {
	h := sha1.New()
	h.Write([]byte(key))
	h.Write([]byte(websocketGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// upgradeResponse compares the two sides' HTTP 101 answers. See the file comment
// for what is must-match and why, and in particular for why the two accept
// strings are not compared with each other.
//
// A connection with no upgrade on either side says nothing and claims nothing: a
// rule that compared two absent handshakes would be the "comparing two empty
// things" failure this harness has already paid for three times.
func upgradeResponse(rep *Report, want, got *Conn) {
	theirStatus, theirNames, theirsOK := upgradeHeaders(want.downHandshake)
	ourStatus, ourNames, oursOK := upgradeHeaders(got.downHandshake)
	switch {
	case !theirsOK && !oursOK:
		return
	case theirsOK != oursOK:
		side := "we"
		if theirsOK {
			side = "the original"
		}
		rep.Informative = append(rep.Informative, Finding{
			Rule:       RuleUpgradeResponse,
			ClientPort: want.ClientPort,
			Detail: fmt.Sprintf("%s answered the WebSocket upgrade and the recorded bytes of the other side's answer are not readable, "+
				"so this exchange is not compared. That is not a verdict about the server: the replay keeps the response it read "+
				"in wsHandshake (replay.go) and an empty one there means the upgrade did not happen, which the connection's own "+
				"notes would say", side),
		})
		return
	}

	lines := []string{
		"status: original " + theirStatus + " | ours " + ourStatus,
		"header order: original " + strings.Join(theirNames, ", ") + " | ours " + strings.Join(ourNames, ", "),
	}

	// Each side's accept against the key that side's own request carried.
	theirAccept, theySaid := headerValue(want.downHandshake, "Sec-WebSocket-Accept")
	ourAccept, weSaid := headerValue(got.downHandshake, "Sec-WebSocket-Accept")
	lines = append(lines, "Sec-WebSocket-Accept: original "+theirAccept+" | ours "+ourAccept)
	lines = append(lines, "the two strings differ by construction and are not compared with each other: the value is "+
		"base64(SHA1(key+GUID)), and the replay sends its own key so a failing replay stays reproducible. "+
		"Each side is checked against the key its own request carried instead")

	for _, side := range []struct {
		who      string
		key      []byte
		accept   string
		said     bool
		readable bool
	}{
		{"the original", want.upHandshake, theirAccept, theySaid, theirsOK},
		{"ours", got.upHandshake, ourAccept, weSaid, oursOK},
	} {
		if !side.readable {
			continue
		}
		if !side.said {
			rep.mustMatch(Finding{
				Rule:       RuleUpgradeResponse,
				ClientPort: want.ClientPort,
				Detail: fmt.Sprintf("%s answered the WebSocket upgrade without a Sec-WebSocket-Accept header, which is what the client "+
					"checks first and what refuses the upgrade", side.who),
				Lines: lines,
			})
			return
		}
		key, ok := headerValue(side.key, "Sec-WebSocket-Key")
		if !ok {
			rep.Informative = append(rep.Informative, Finding{
				Rule:       RuleUpgradeResponse,
				ClientPort: want.ClientPort,
				Detail: fmt.Sprintf("%s answered the upgrade but its own request carries no Sec-WebSocket-Key, so the accept value cannot "+
					"be derived and the derivation is not compared on that side", side.who),
				Lines: lines,
			})
			continue
		}
		if derived := expectedAccept(key); derived != side.accept {
			rep.mustMatch(Finding{
				Rule:       RuleUpgradeResponse,
				ClientPort: want.ClientPort,
				Detail: fmt.Sprintf("%s answered Sec-WebSocket-Accept %s and the RFC 6455 derivation of the key its own request carried (%s) is %s",
					side.who, side.accept, key, derived),
				Lines: lines,
			})
			return
		}
	}

	missing := missingHeaders(theirNames, ourNames)
	extra := missingHeaders(ourNames, theirNames)
	if len(missing) > 0 || len(extra) > 0 {
		rep.mustMatch(Finding{
			Rule:       RuleUpgradeResponse,
			ClientPort: want.ClientPort,
			Detail: fmt.Sprintf("the WebSocket upgrade response carries a different set of headers: the original sends %s that we do not, "+
				"and we send %s that it does not", renderHeaders(missing), renderHeaders(extra)),
			Lines: lines,
		})
		return
	}

	rep.Informative = append(rep.Informative, Finding{
		Rule:       RuleUpgradeResponse,
		ClientPort: want.ClientPort,
		Detail: "the WebSocket upgrade response carries the same headers on both sides, and each side's Sec-WebSocket-Accept is the " +
			"RFC 6455 derivation of the key its own request carried. The order differs and is not must-match: RFC 6455 gives the " +
			"response no ordering requirement, the client is websocketpp (its request carries User-Agent: WebSocket++/0.8.2) and it " +
			"collects the headers into a map",
		Lines: lines,
	})
}

// missingHeaders returns the names of want that other does not have, sorted.
func missingHeaders(want, other []string) []string {
	have := make(map[string]bool, len(other))
	for _, n := range other {
		have[strings.ToLower(n)] = true
	}
	var out []string
	for _, n := range want {
		if !have[strings.ToLower(n)] {
			out = append(out, n)
		}
	}
	sort.Strings(out)
	return out
}

func renderHeaders(names []string) string {
	if len(names) == 0 {
		return "none"
	}
	return strings.Join(names, ", ")
}
