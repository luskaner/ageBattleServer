package sim

// The pending-peer rule, on its own, with the two readings spelled out.
//
// The rule compares what each side's 0x0A said about the session's peers. It is
// the rule that turned the bsPort divergence from a paragraph into a measurement
// (SESSION-JOURNAL.md), and it is here tested without a capture because it is
// pure: it reads two lists of messages and decides whether they agree.

import (
	"bytes"
	"testing"

	"github.com/luskaner/ageBattleServer/internal/protocol"
)

// ackWithPending is one 0x0A body carrying a pending list and nothing else:
// [FE][0 subs][pending count][pending ids][no entries]. It is the heartbeat
// shape, which is where a peer list shows up with no data around it.
func ackWithPending(pending ...byte) *Msg {
	body := []byte{0xFE, 0x00, byte(len(pending))}
	body = append(body, pending...)
	body = append(body, 0x00, 0x00)
	return &Msg{Type: protocol.MsgEndOfGameData, Body: body}
}

func connOf(port int, msgs ...*Msg) *Conn {
	return &Conn{ClientPort: port, Down: msgs}
}

// TestTheRuleIsQuietWhenTheTwoSidesNameTheSamePeers: the rule must not fire on
// the four WebSocket captures' shape, or it would be noise on every capture.
func TestTheRuleIsQuietWhenTheTwoSidesNameTheSamePeers(t *testing.T) {
	rep := &Report{}
	theirs := connOf(1, ackWithPending(1), ackWithPending(1), ackWithPending())
	ours := connOf(1, ackWithPending(1), ackWithPending(1), ackWithPending())
	pendingPeers(rep, theirs, ours)
	if len(rep.Informative) != 0 {
		t.Errorf("both sides name the same peers and the rule said something: %+v", rep.Informative[0])
	}
}

// TestTheRuleFailsWithTheDivergenceReintroduced is the falsification: the bsPort
// shape, ours naming two peers where the original named one. Without the rule in
// the comparison this finding does not exist.
func TestTheRuleFailsWithTheDivergenceReintroduced(t *testing.T) {
	rep := &Report{}
	theirs := connOf(1, ackWithPending(1), ackWithPending(1), ackWithPending())
	ours := connOf(1, ackWithPending(1, 2), ackWithPending(1, 2), ackWithPending())
	pendingPeers(rep, theirs, ours)
	if len(rep.Informative) != 1 {
		t.Fatalf("the rule should have said something and said nothing (%d findings)", len(rep.Informative))
	}
	f := rep.Informative[0]
	if f.Rule != RulePendingPeers {
		t.Errorf("the rule that fired is %q, not %q", f.Rule, RulePendingPeers)
	}
	if !bytes.Contains([]byte(f.Ours), []byte("[1 2]")) {
		t.Errorf("our side should read [1 2] and reads %q", f.Ours)
	}
	if !bytes.Contains([]byte(f.Theirs), []byte("[1]")) {
		t.Errorf("the original's side should read [1] and reads %q", f.Theirs)
	}
	if f.ClientPort != 1 {
		t.Errorf("the finding carries port %d and should carry 1", f.ClientPort)
	}
	// It must be informative, not must-match: promoting it would fail the
	// bsPort capture over a question two connections cannot answer.
	if f.MustMatch {
		t.Error("the rule has been marked must-match: with two connections one cannot tell " +
			"'one peer per connection' from 'one peer in total', and real failures get invented")
	}
}

// TestTheRuleSpeaksWhenOneSideSentNoAckAtAll: on the bsPort capture the original
// sends the second connection no 0x0A at all, and we send it eight. Both sides
// speaking is not a condition of the rule firing, and a side that never spoke is
// something the finding has to say rather than skip.
func TestTheRuleSpeaksWhenOneSideSentNoAckAtAll(t *testing.T) {
	rep := &Report{}
	theirs := connOf(2)
	ours := connOf(2, ackWithPending(1, 2), ackWithPending(1, 2))
	pendingPeers(rep, theirs, ours)
	if len(rep.Informative) != 1 {
		t.Fatalf("one side with no 0x0A and another with the peer list should give a finding, there are %d",
			len(rep.Informative))
	}
	if got := rep.Informative[0].Theirs; got == "" {
		t.Error("the original's side appears as empty instead of as 'I name no peer'")
	}
}

// TestTheRuleStaysQuietWhenOnlyTheCountsDiffer: the 2026-09-29_012528 shape. Both
// sides name {1} and differ in how many acks carried each list, which is
// grouping and not peers. A rule that compared the counts would report a
// difference of peers on a capture where there is none.
func TestTheRuleStaysQuietWhenOnlyTheCountsDiffer(t *testing.T) {
	rep := &Report{}
	theirs := connOf(1,
		ackWithPending(1), ackWithPending(1), ackWithPending(1),
		ackWithPending(), ackWithPending())
	ours := connOf(1,
		ackWithPending(1), ackWithPending(1), ackWithPending(1), ackWithPending(1),
		ackWithPending(1), ackWithPending(1), ackWithPending(1), ackWithPending(1),
		ackWithPending())
	pendingPeers(rep, theirs, ours)
	if len(rep.Informative) != 0 {
		t.Errorf("both sides name the same peer and only the number of acks that carried it changes: %+v",
			rep.Informative[0])
	}
}

// TestTheRuleStaysQuietWithNoAcksOnEitherSide: a connection where neither side
// sent a 0x0A is not a divergence of peers, it is a connection with no acks in
// it, and RuleMessageCount already has something to say about it.
func TestTheRuleStaysQuietWithNoAcksOnEitherSide(t *testing.T) {
	rep := &Report{}
	pendingPeers(rep, connOf(3), connOf(3))
	if len(rep.Informative) != 0 {
		t.Errorf("neither side sent a 0x0A and the rule said something: %+v", rep.Informative[0])
	}
}

// TestTheReadingIsPerAckAndNotPerConnection: the counts have to be per ack, or a
// side that sent the same list twice reads as one message and a difference in
// grouping hides inside the reading.
func TestTheReadingIsPerAckAndNotPerConnection(t *testing.T) {
	got := PendingLists([]*Msg{ackWithPending(1), ackWithPending(1), ackWithPending()})
	want := []string{"[1] x2", "[] x1"}
	if len(got) != len(want) {
		t.Fatalf("the reading is %v and should be %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entry %d: %q instead of %q", i, got[i], want[i])
		}
	}
}