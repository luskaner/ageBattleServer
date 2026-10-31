package sim

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/luskaner/ageBattleServer/internal/protocol"
)

// The pending-peer list of the 0x0A, read per side. It is the session's own
// account of which peers its tick has not heard from, and it is the only place
// the wire says how many peers a session has.
//
// # What the reference server builds it from
//
// The tick walks one record per peer, the game-data queue. Each record names a
// peer, the walk looks that peer up in the session's peer table, and a record
// whose peer is missing or is not participating is SKIPPED, so its id never
// reaches the wire. A record that survives and has nothing queued contributes its
// id, and one whose peer is past its sync deadline contributes nothing and gets
// closed.
//
// The walk does not stop at the first record it skips: the jump it takes is the
// bottom of the loop, not a break. The list is therefore the peers that were
// quiet, in the order the records are in, which is the order the peers joined in,
// and a peer's queued data dies with it.
//
// # What makes it worth comparing
//
// It is direct evidence of peer creation, and it costs no game time: the 0x0A
// stream is already decoded by the harness for the observer rules, so reading
// this off it adds nothing to a replay.
//
// Measured on every committed capture (testdata, 2026-10-01): the reference
// server's list is `[1]` on every connection of every capture, including the two
// with three client connections (2026-09-29_012528_03cb67c7cf19 has three) and
// the bsPort one (2026-09-30_201122_c4e7c284ff1c has two, and ours reads `[1,2]`
// there against `[1]`).
//
// # What it does NOT decide
//
// It is informative, not must-match. With two connections, "one peer per
// connection" and "one peer in total" put the same bytes on the wire for the
// player, and the difference only shows on the SECOND connection's own stream,
// which the reference server leaves empty. That is a real divergence and it is
// reported, but promoting it to must-match would fail a capture over a question
// the evidence answers only indirectly.
const (
	// RulePendingPeers is the pending-peer list each side's 0x0A names. It is the
	// session's own count of its peers on the wire, so a difference in it is a
	// difference in how many peers each side gave the session.
	RulePendingPeers = "pending-peers"
)

// peerRead is what one side's 0x0A stream says about the session's peers: the
// distinct pending lists it carried, and how many acks carried each.
type peerRead struct {
	// seen is every distinct list, in first-seen order, so the report can show
	// what the side actually said rather than a set that hides the ordering.
	seen []string
	// counts is how many acks carried each entry of seen.
	counts []int
	// frames is how many 0x0A this decoder read on this side.
	frames int
	// unreadable counts the acks whose shape this decoder does not understand.
	unreadable int
}

// add records one ack's pending list. The receiver is a pointer because the
// appends have to reach the caller: with a value receiver the reading comes back
// empty and every capture "agrees" with every other one.
func (p *peerRead) add(list []byte) {
	k := renderPending(list)
	for i, s := range p.seen {
		if s == k {
			p.counts[i]++
			return
		}
	}
	p.seen = append(p.seen, k)
	p.counts = append(p.counts, 1)
}

func (p peerRead) String() string {
	if p.frames == 0 {
		return "no 0x0A on this side, so this side never named a peer"
	}
	out := ""
	for i, s := range p.seen {
		if i > 0 {
			out += ", "
		}
		out += fmt.Sprintf("%s on %d ack(s)", s, p.counts[i])
	}
	if p.unreadable > 0 {
		out += fmt.Sprintf(", and %d ack(s) this decoder does not understand", p.unreadable)
	}
	return out + " (an empty list means every peer carried data in that ack, not that the session had no peers)"
}

// renderPending is the compact form of one list, which is what a test pins and
// what the report shows. The long form (why an empty list is not an absence) is
// in peerRead.String.
func renderPending(list []byte) string {
	if len(list) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(list))
	for _, p := range list {
		parts = append(parts, strconv.Itoa(int(p)))
	}
	return "[" + strings.Join(parts, " ") + "]"
}

// readPendingPeers takes one side's 0x0A stream apart for its pending lists.
func readPendingPeers(msgs []*Msg) peerRead {
	var p peerRead
	for _, m := range msgs {
		if m.Type != protocol.MsgEndOfGameData {
			continue
		}
		p.frames++
		f, ok := decodeAck(m.Body)
		if !ok {
			p.unreadable++
			continue
		}
		p.add(f.pending)
	}
	return p
}

// pendingPeers reports the peer ids each side named, when the SETS differ.
//
// It compares the sets and not the counts, and that is the whole design: how
// many acks carried each list is grouping, which two servers are free to differ
// in (the tick emits an ack per tick, not per input) and which RuleTickCadence
// and RuleMessageCount already own. How many peers the session has is not
// grouping. On 2026-09-29_012528_03cb67c7cf19 both sides name {1} and only the
// counts differ ([1] on 791 acks against our 1273, [] on 633 against our 151),
// and a rule that compared counts would have called that a
// divergence of peers on a capture where there is none.
func pendingPeers(rep *Report, want, got *Conn) {
	ours := readPendingPeers(got.Down)
	theirs := readPendingPeers(want.Down)
	if ours.frames == 0 && theirs.frames == 0 {
		return
	}
	if samePeerSet(ours, theirs) {
		return
	}
	rep.Informative = append(rep.Informative, Finding{
		Rule:       RulePendingPeers,
		ClientPort: want.ClientPort,
		Detail: "the 0x0A names a different set of peers on each side, which is the session telling " +
			"the client how many peers it has",
		Ours:   ours.String(),
		Theirs: theirs.String(),
		Lines: []string{
			"the list is the session's own account of the peers whose data this ack does not carry, built by " +
				"walking one record per peer and skipping the records whose peer is missing or not participating",
			"a slot is claimed by a peer message of type 6 (ADDPEER) whose global id a REQUESTJOIN of type 2 already " +
				"registered: the peer is not created at all when that id is not in the session's registered-id set, and " +
				"the set is written only by REQUESTJOIN",
			"so a peer that never sent REQUESTJOIN, or sent it and never got an ADDPEER, is not in the list. " +
				"Whether a connection that only sent 0x15 LOBBYCONNECT should be a peer is what this is measuring, " +
				"and the 0x15 handler adds no peer on any of its paths",
			"the counts above are not part of the verdict: how many acks carried each list is grouping, which " +
				"[tick-cadence] and [message-count] already own",
		},
	})
}

// samePeerSet reports whether the two sides named the same peer ids, whatever
// the counts were. It is deliberately not a string comparison of the readings:
// the counts move with the tick, and a reader that moved with them would report
// a difference of peers on every capture whose two sides grouped differently.
func samePeerSet(a, b peerRead) bool {
	if a.frames == 0 || b.frames == 0 {
		// One side never spoke about peers. That is a difference, not an absence
		// of one: the reference server sends the bsPort second connection no 0x0A
		// at all, and "it named nobody" is not the same claim as "we named two".
		return false
	}
	if a.unreadable > 0 || b.unreadable > 0 {
		// A list this decoder did not read is not evidence of anything, so a
		// side with unreadable acks cannot be said to agree with the other.
		return false
	}
	return a.named().equal(b.named())
}

// peerSet is the sorted set of peer ids a side named, with 0 for the acks whose
// list was empty (every peer carried data), which is a fact about the ack and
// not about the peers.
type peerSet map[byte]bool

func (s peerSet) equal(other peerSet) bool {
	if len(s) != len(other) {
		return false
	}
	for k := range s {
		if !other[k] {
			return false
		}
	}
	return true
}

func (p peerRead) named() peerSet {
	out := peerSet{}
	for _, s := range p.seen {
		for _, id := range parsePending(s) {
			out[id] = true
		}
	}
	return out
}

// parsePending reads back the ids out of a rendered list. It exists so the set
// and the rendering cannot drift apart: both go through renderPending, and this
// is the other half of that one function.
func parsePending(rendered string) []byte {
	inner := strings.TrimSuffix(strings.TrimPrefix(rendered, "["), "]")
	if inner == "" {
		return nil
	}
	var out []byte
	for _, part := range strings.Split(inner, " ") {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 || n > 255 {
			continue
		}
		out = append(out, byte(n))
	}
	return out
}