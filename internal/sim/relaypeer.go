package sim

import (
	"bytes"
	"compress/zlib"
	"io"
	"strconv"
	"strings"

	"github.com/luskaner/ageBattleServer/internal/protocol"
)

// Reading a relay's peer message, and a connection's peer lists, for the tools
// and the tests that ask "what did this side say about the session's peers".
//
// The relay is where the client sends them: neither REQUESTJOIN nor ADDPEER
// arrives as a top-level 0x0F in any committed capture, and both arrive inside a
// 0x13 LOBBYRELAY, whose inner message is dispatched as if it had arrived on its
// own. The first ones are zlib-compressed (flagBits 0x04), so the inner type is
// only readable after inflating them.

// RelayPeerMessage is one peer message the client nested in a LobbyRelay.
type RelayPeerMessage struct {
	// ToPeerID and FromPeerID are the relay header, which is the relay's own
	// addressing and not part of the peer message.
	ToPeerID   byte
	FromPeerID byte
	// Compressed is the PEERHEADER's flagBits 0x04, kept because it is the fact
	// that made the inner type unreadable without inflating the payload.
	Compressed bool
	// InnerType and Inner are the message inside the envelope, inflated when it
	// was compressed.
	InnerType byte
	Inner     []byte
}

// RelayPeerMessages returns the peer messages the given client->server messages
// carry inside their relays, in order. A relay that carries no 0x0F, or whose
// payload this decoder cannot read, is skipped: it is not a finding here, and a
// decoder that reported its own failures as peers would be inventing them.
func RelayPeerMessages(msgs []*Msg) []RelayPeerMessage {
	var out []RelayPeerMessage
	for _, m := range msgs {
		if m.Type != protocol.MsgLobbyRelay || len(m.Body) < 3 || m.Body[2] != protocol.MsgPeer {
			continue
		}
		env, err := protocol.ParsePeerEnvelope(m.Body[3:])
		if err != nil {
			continue
		}
		raw := env.Inner
		if env.Header.FlagBits&protocol.FlagCompressed == 0 {
			out = append(out, RelayPeerMessage{
				ToPeerID: m.Body[0], FromPeerID: m.Body[1],
				InnerType: innerType(raw), Inner: raw,
			})
			continue
		}
		zr, err := zlib.NewReader(bytes.NewReader(raw))
		if err != nil {
			continue
		}
		// The cap is the ceiling on what one relay may inflate to, so a capture
		// that carries a hostile or corrupt payload cannot make a test allocate
		// without limit (AGENTS.md, RAM budget).
		inflated, err := io.ReadAll(io.LimitReader(zr, maxInflatedPeerMessage))
		_ = zr.Close()
		if err != nil {
			continue
		}
		out = append(out, RelayPeerMessage{
			ToPeerID: m.Body[0], FromPeerID: m.Body[1], Compressed: true,
			InnerType: innerType(inflated), Inner: inflated,
		})
	}
	return out
}

// maxInflatedPeerMessage bounds one inflated relay payload. The largest one in
// the committed captures is 2134 bytes; the ceiling is two orders of magnitude
// above that and is here so the decoder cannot be made to allocate by a capture.
const maxInflatedPeerMessage = 1 << 20

func innerType(b []byte) byte {
	if len(b) == 0 {
		return 0
	}
	return b[0]
}

// PendingLists returns the distinct pending-peer lists the given messages named,
// in first-seen order, each with the number of acks that carried it. It reads
// only 0x0A messages and ignores every other type, so it can be pointed at a
// whole connection in either direction.
func PendingLists(msgs []*Msg) []string {
	p := readPendingPeers(msgs)
	out := make([]string, 0, len(p.seen))
	for i, s := range p.seen {
		out = append(out, s+" x"+strconv.Itoa(p.counts[i]))
	}
	return out
}

// PendingListString renders one connection's reading in one line, for a log or an
// error message. It says which side's stream it read, because "the pending list
// is [1]" is not a statement without it.
func PendingListString(ours, theirs []*Msg) string {
	return "ours: " + strings.Join(PendingLists(ours), ", ") +
		" | original: " + strings.Join(PendingLists(theirs), ", ")
}