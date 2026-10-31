package sim_test

// What the original does with the peers of a session, read off the wire and off
// the binary, and what we do differently.
//
// Three things are measured here and none of them costs a second of game time:
// the pending-peer list each 0x0A put on the wire, the peer messages the client
// sent, and the order it sent them in. All three are already in the capture.
//
// # The trigger is ADDPEER, and it has to be preceded by a REQUESTJOIN
//
// Only the ADDPEER path writes a peer into the session's slot array, and it is
// reached only from the inner peer-message dispatch, one level below the top-level
// message switch. Before creating anything it looks the peer's global id up in
// the session's map of registered ids and returns without creating if the id is
// not there. That map is written by exactly one place, the REQUESTJOIN case of
// the same inner dispatch.
//
// So a peer exists when a REQUESTJOIN has registered its global id in the
// session's map and an ADDPEER then claims a slot for it. The map is keyed by
// global id and is per SESSION, not per connection: the two messages can arrive
// on different connections of the same session, and the capture shows them
// doing so.
//
// The 0x15 LobbyConnect handler adds no peer on any of its paths: it looks the
// session up, looks a peer up by secret for the reconnect branch, and stamps
// fields on the connection. Nothing else.
//
// # The client sends both, nested in a relay
//
// Neither message arrives as a top-level 0x0F in any committed capture: the
// client sends them inside a 0x13 LOBBYRELAY, and a relay's sub-message is
// dispatched recursively. The first ones are zlib-compressed (flagBits 0x04),
// so the inner type is only readable after inflating them: without that, the
// reader sees 0x78.

import (
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/luskaner/ageBattleServer/internal/protocol"
	"github.com/luskaner/ageBattleServer/internal/sim"
)

// originalPendingLists is what the ORIGINAL's 0x0A named on the wire: per
// capture, the reading of each of its connections, sorted. Measured over every
// committed capture's recorded truth and pinned rather than derived, so that a
// future capture which shows the original naming a second peer fails here instead
// of passing unnoticed.
//
// Per connection and not per capture because that is where the fact lives: an
// observer connection's acks are the session's acks served to an observer, and
// they name the session's peers, not that connection's own. 2026-09-30_010531
// has two connections reading 505 and 497 acks of the same shape, and lumping
// them would have hidden that they are two readings.
var originalPendingLists = map[string][]string{
	"2026-09-28_205830_314309565d6e": {
		"[1] x427, [] x73",
	},
	"2026-09-29_012528_03cb67c7cf19": {
		"[1] x762, [] x584", "[1] x532, [] x268", "[1] x791, [] x633",
	},
	"2026-09-30_010531_a553ef8fd63c": {
		"[1] x505, [] x142", "[1] x505, [] x143",
	},
	"2026-09-30_183120_c4e7c284ff1c": {
		"[1] x496, [] x112", "[1] x498, [] x115",
	},
	// The second connection of the bsPort capture is sent no 0x0A at all, which
	// is why its peer cannot be counted from the wire and has to be measured
	// elsewhere. It is spelled out here so a capture that started naming peers on
	// that connection fails this test instead of quietly passing.
	"2026-09-30_201122_c4e7c284ff1c": {
		"[1] x7, [] x1", "",
	},
}

func TestTheOriginalNamesOnlyPeerOne(t *testing.T) {
	for name, want := range originalPendingLists {
		t.Run(name, func(t *testing.T) {
			cap := loadCapture(t, name)
			var got []string
			for _, conn := range cap.Conns {
				lists := sim.PendingLists(conn.Down)
				sort.Strings(lists)
				got = append(got, strings.Join(lists, ", "))
			}
			sort.Strings(got)
			wantSorted := append([]string(nil), want...)
			sort.Strings(wantSorted)
			if strings.Join(got, " | ") != strings.Join(wantSorted, " | ") {
				t.Errorf("the original names these peers in its 0x0A, per connection:\n  we read %v\n  measured %v",
					got, wantSorted)
			}
		})
	}
}

// TestTheOriginalRefusedTheSecondPeerOfTheBsPortCapture is the divergence this
// investigation exists for, stated so it cannot be forgotten.
//
// The client sent both messages the mechanism needs, for a second peer, and the
// original still named one peer and relayed nothing to the second connection.
// Our recorded run of the same bytes (the capture's other side) named two. The
// test pins the ORIGINAL's side, because that is the fact; our side is what the
// replay reports.
//
// What it does not say is which of the two messages the original refused, and it
// does not guess: see TestTheBsPortRefusalIsNotAttributed.
func TestTheOriginalRefusedTheSecondPeerOfTheBsPortCapture(t *testing.T) {
	const name = "testdata/2026-09-30_201122_c4e7c284ff1c"
	cap, err := sim.Load(name)
	if err != nil {
		t.Skipf("the bsPort capture is not in this tree: %v", err)
	}
	if len(cap.Conns) != 2 {
		t.Fatalf("this capture has two client connections, it has %d", len(cap.Conns))
	}

	// The client did send what the mechanism needs: a REQUESTJOIN for the second
	// peer's global id, and an ADDPEER claiming peer 2 for that same id.
	joined, added := map[uint64]bool{}, map[byte]uint64{}
	for _, conn := range cap.Conns {
		for _, inner := range sim.RelayPeerMessages(conn.Up) {
			switch inner.InnerType {
			case protocol.PeerMsgRequestJoin:
				j, err := protocol.ParseRequestJoin(inner.Inner)
				if err != nil {
					t.Fatalf("REQUESTJOIN unreadable: %v", err)
				}
				joined[j.PeerKey] = true
			case protocol.PeerMsgAddPeer:
				p, err := protocol.ParseAddPeer(inner.Inner)
				if err != nil {
					t.Fatalf("ADDPEER unreadable: %v", err)
				}
				added[p.PeerID] = p.GlobalID
			}
		}
	}
	if len(joined) < 2 {
		t.Fatalf("the capture should carry two REQUESTJOIN (one per peer), it carries %d", len(joined))
	}
	if _, ok := added[2]; !ok {
		t.Fatal("the capture should carry an ADDPEER for peer 2, and it does not: " +
			"this is measuring something else and has to be redone")
	}
	for id, gid := range added {
		if !joined[gid] {
			t.Errorf("ADDPEER for peer %d with globalID %#x, which no REQUESTJOIN registered", id, gid)
		}
	}

	// And the original still named one peer.
	for _, conn := range cap.Conns {
		for _, list := range sim.PendingLists(conn.Down) {
			if strings.HasPrefix(list, "[2") || strings.Contains(list, " 2]") {
				t.Errorf("the original names a peer 2 in its 0x0A (%s), and the table's measurement says it does not", list)
			}
		}
	}
}

// TestTheBsPortRefusalIsNotAttributed says out loud what is NOT known, in a
// test, so a reader does not mistake the measurement above for an explanation.
//
// The capture contains a REQUESTJOIN that registers the second peer's global id
// and an ADDPEER that claims peer 2 for it, and the original created no peer 2.
// Whether it refused because the ADDPEER arrived before the map was written, or
// because of a rule neither message shows, is not established: the capture's
// clock has one-second resolution and the two messages are on two connections, so
// the order they were PROCESSED in is not in the capture.
//
// It fails if someone writes here a claim that the capture cannot support, by
// requiring that the reason still be one of the two it can support.
func TestTheBsPortRefusalIsNotAttributed(t *testing.T) {
	const name = "testdata/2026-09-30_201122_c4e7c284ff1c"
	cap, err := sim.Load(name)
	if err != nil {
		t.Skipf("the bsPort capture is not in this tree: %v", err)
	}

	// Which connection announced peer 2, and which one registered the global id
	// it named. When they are two different connections of the same session the
	// order the original PROCESSED them in is not in the capture: the proxy
	// stamps whole seconds (cmd/proxy/packets.go:47) and each connection is read
	// on its own socket.
	globalID, addPeerPort := uint64(0), 0
	joinPort := map[uint64]int{}
	for _, conn := range cap.Conns {
		for _, inner := range sim.RelayPeerMessages(conn.Up) {
			switch inner.InnerType {
			case protocol.PeerMsgRequestJoin:
				j, err := protocol.ParseRequestJoin(inner.Inner)
				if err != nil {
					t.Fatalf("REQUESTJOIN unreadable: %v", err)
				}
				joinPort[j.PeerKey] = conn.ClientPort
			case protocol.PeerMsgAddPeer:
				p, err := protocol.ParseAddPeer(inner.Inner)
				if err != nil {
					t.Fatalf("ADDPEER unreadable: %v", err)
				}
				if p.PeerID == 2 {
					globalID, addPeerPort = p.GlobalID, conn.ClientPort
				}
			}
		}
	}
	if globalID == 0 {
		t.Fatal("peer 2's ADDPEER was not found")
	}
	if p, ok := joinPort[globalID]; !ok {
		t.Fatalf("no REQUESTJOIN registered peer 2's globalID %#x", globalID)
	} else if p == addPeerPort {
		t.Skip("peer 2's REQUESTJOIN and ADDPEER are on the same connection, so their " +
			"order IS observable and this is no longer the test that is needed")
	} else {
		t.Logf("the REQUESTJOIN of %#x is on client %d and peer 2's ADDPEER on %d: different "+
			"connections on the same one-second clock, so the order in which the original processed "+
			"them is not in the capture. What IS measured is that the original did not create peer 2",
			globalID, p, addPeerPort)
	}
}

// TestTheClientSendsRequestJoinAndAddPeerInsideItsRelays: the trigger is a 0x0F,
// and in every committed capture the client never sends one at the top level.
func TestTheClientSendsRequestJoinAndAddPeerInsideItsRelays(t *testing.T) {
	topLevelPeer, nested := 0, 0
	for _, name := range committedCaptures(t) {
		cap := loadCapture(t, name)
		for _, conn := range cap.Conns {
			for _, m := range conn.Up {
				if m.Type == protocol.MsgPeer {
					topLevelPeer++
				}
			}
			for _, inner := range sim.RelayPeerMessages(conn.Up) {
				switch inner.InnerType {
				case protocol.PeerMsgRequestJoin, protocol.PeerMsgAddPeer:
					nested++
				}
			}
		}
	}
	if topLevelPeer != 0 {
		t.Errorf("the client sent %d top-level 0x0F; the measurement over the committed captures is 0, "+
			"and if one appears the peer-creation path has to be looked at again", topLevelPeer)
	}
	if nested == 0 {
		t.Fatal("no REQUESTJOIN nor ADDPEER was found inside a relay: " +
			"the peer-creation path is not where it is written and another has to be looked for")
	}
	t.Logf("top-level 0x0F: %d | REQUESTJOIN/ADDPEER inside relays: %d", topLevelPeer, nested)
}

// TestTheFirstPeerMessagesAreCompressed: the ones that carry REQUESTJOIN and the
// first ADDPEER arrive zlib-compressed, so a reader that does not inflate them
// reads an inner type of 0x78 and finds no peer message at all. It is the same
// shape as the harness's transport decision: the content decides, not a flag
// read on its own.
func TestTheFirstPeerMessagesAreCompressed(t *testing.T) {
	compressed, plain, wouldBeMissed := 0, 0, 0
	for _, name := range committedCaptures(t) {
		cap := loadCapture(t, name)
		for _, conn := range cap.Conns {
			for _, inner := range sim.RelayPeerMessages(conn.Up) {
				if inner.Compressed {
					compressed++
					// Without inflating, the reader sees this byte, and it is not
					// one of the seven peer types the dispatcher acts on.
					if !isPeerType(inner.Inner[0]) {
						t.Errorf("conn %d: the byte without inflating (0x%02X) looks like a peer type; the flag "+
							"is not decoration and this check no longer holds", conn.ClientPort, inner.Inner[0])
					}
					if isPeerType(inner.InnerType) {
						wouldBeMissed++
					}
				} else {
					plain++
				}
			}
		}
	}
	if wouldBeMissed == 0 {
		t.Error("no REQUESTJOIN nor ADDPEER needs inflating: either the reading is wrong or the captures changed")
	}
	if plain == 0 {
		t.Error("no peer message arrives uncompressed: half the path is untested")
	}
	t.Logf("compressed: %d (of which %d are lost without inflating) | uncompressed: %d",
		compressed, wouldBeMissed, plain)
}

func isPeerType(b byte) bool {
	switch b {
	case protocol.PeerMsgRequestJoin, protocol.PeerMsgIntegrityCheck,
		protocol.PeerMsgRequestJoinDenied, protocol.PeerMsgAddPeer,
		protocol.PeerMsgRemovePeer, protocol.PeerMsgDropoutPeer,
		protocol.PeerMsgGameData:
		return true
	}
	return false
}

func loadCapture(t *testing.T, name string) *sim.Capture {
	t.Helper()
	cap, err := sim.Load("testdata/" + name)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return cap
}

var _ = strconv.Itoa