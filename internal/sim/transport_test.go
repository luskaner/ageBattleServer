package sim

import "testing"

// The transport of a connection is decided by the CONTENT of its bytes.
//
// The msg_type column cannot decide it and is not read: the proxy labels a
// WebSocket frame 0x82 and an RLink row with the message's own type, so a bsPort
// capture has not one 0x82 row in it (measured on 2026-09-30_201122_c4e7c284ff1c:
// 0x0A x2533, 0x13 x250, 0x09 x64, 0x14 x40, 0x17 x34, 0x15 x4, 0x0F x3 and no
// 0x82 at all). Two bytes cannot decide it either, which is what these tests are
// about.

// rlinkMessage is one RLink message as it is logged: [u16 BE length][type][body],
// where the length counts the type byte.
func rlinkMessage(msgType byte, body []byte) []byte {
	n := len(body) + 1
	return append([]byte{byte(n >> 8), byte(n), msgType}, body...)
}

// wsFrame is one WebSocket frame as the SERVER writes it, so it carries no mask.
func wsFrame(payload []byte) []byte {
	return append([]byte{0x82, byte(len(payload))}, payload...)
}

// TestAnRLinkMessageThatStartsLikeAWebSocketHeadIsStillRLink is the collision the
// two framings have, written down with the arithmetic.
//
// A WebSocket head is a FIN bit plus opcode 2, which is 0x82. An RLink message
// starts with its length in big endian, so a message of 33,024..33,279 or
// 33,536..34,047 bytes ALSO begins 0x82 (0x81 or 0x82 as its high byte). Believing
// the first byte would frame a real bsPort capture as WebSocket and produce a
// stream of nonsense that still looks like a capture.
//
// The frame here is the exact case: a message whose declared length is 0x8200, so
// its first two bytes are 0x82 0x00. Read as WebSocket that is a binary frame of
// zero payload, and the byte after it is a message type, which cannot start a
// frame: the WebSocket reading breaks two bytes in.
func TestAnRLinkMessageThatStartsLikeAWebSocketHeadIsStillRLink(t *testing.T) {
	body := make([]byte, 0x8200-1)
	for i := range body {
		body[i] = byte(i)
	}
	stream := rlinkMessage(0x09, body)
	if stream[0] != 0x82 {
		t.Fatalf("the fixture has to start 0x82 for this to be the case it claims: %02x", stream[0])
	}
	if got := wsFrame(nil)[0]; got != stream[0] {
		t.Fatalf("a WebSocket frame head is %02x and the RLink message starts %02x: the collision is not the one under test", got, stream[0])
	}

	transport, note := detectTransport(stream, nil, false)
	if transport != FramingRLink {
		t.Fatalf("an RLink stream has to be read as RLink, got %q (note: %q)", transport, note)
	}
	if note != "" {
		t.Errorf("a stream one framing reads and the other breaks is not a doubt, it is a decision: %s", note)
	}
	msgs, left, err := scanBytes(stream, transport)
	if err != nil {
		t.Fatalf("scanBytes: %v", err)
	}
	if left != 0 {
		t.Errorf("%d bytes left over: the framing did not consume the message", left)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected the one message, got %d", len(msgs))
	}
	if msgs[0].Type != 0x09 || len(msgs[0].Body) != len(body) {
		t.Errorf("the message did not come out whole: type 0x%02X, body %d bytes, wanted 0x09 and %d",
			msgs[0].Type, len(msgs[0].Body), len(body))
	}
}

// TestBothDirectionsAreWalkedBeforeATransportIsDecided: one direction alone is not
// enough to identify a connection. The reply of the case above is a plain RLink
// stream, and a connection is one framing for both directions: if the up stream
// had been taken at its word the down stream would have been framed as WebSocket
// too, and both would have been nonsense.
func TestBothDirectionsAreWalkedBeforeATransportIsDecided(t *testing.T) {
	head := make([]byte, 0x8200-1)
	stream := rlinkMessage(0x0A, head)
	reply := rlinkMessage(0x15, []byte{0x2a, 0x00, 0x00, 0x00})

	transport, note := detectTransport(stream, reply, false)
	if transport != FramingRLink {
		t.Fatalf("transport = %q (note: %q), want %q", transport, note, FramingRLink)
	}
	msgs, _, err := scanBytes(reply, transport)
	if err != nil || len(msgs) != 1 || msgs[0].Type != 0x15 {
		t.Errorf("the reply has to frame as the same transport: %d messages, err %v", len(msgs), err)
	}
}

// TestWebSocketWithoutTheHandshakeIsStillWebSocket: the upgrade is the strongest
// evidence when it is there, and it is not always there. A capture cut mid-handshake
// has no GET and no HTTP/1.1 in it, so the decision has to come out of the frames
// themselves.
func TestWebSocketWithoutTheHandshakeIsStillWebSocket(t *testing.T) {
	stream := append(wsFrame([]byte{0x15, 0x01}), wsFrame([]byte{0x0A, 0x02})...)
	transport, note := detectTransport(stream, stream, false)
	if transport != FramingWS {
		t.Fatalf("transport = %q (note: %q), want %q", transport, note, FramingWS)
	}
	msgs, left, err := scanBytes(stream, transport)
	if err != nil || left != 0 || len(msgs) != 2 {
		t.Fatalf("two frames expected, got %d (left %d, err %v)", len(msgs), left, err)
	}
	if msgs[0].Type != 0x15 || msgs[1].Type != 0x0A {
		t.Errorf("the payloads did not come out: %v", msgs)
	}
}

// TestTheUpgradeDecidesOnItsOwn: an HTTP upgrade in the capture is a fact about the
// capture and not an inference, so it is what the framing is decided by, whatever
// the frames after it happen to look like.
func TestTheUpgradeDecidesOnItsOwn(t *testing.T) {
	transport, note := detectTransport(streamBytes(nil), []byte("HTTP/1.1 101 Switching Protocols\r\n\r\n"), true)
	if transport != FramingWS || note != "" {
		t.Errorf("an upgraded connection is WebSocket with no doubt about it: got %q (%q)", transport, note)
	}
}

// TestSomethingThatFramesAsNeitherIsNotTheBattleProtocol: the out-of-band port is a
// separate TLS service and its records start with 0x16, which starts neither
// framing. A connection like that must not be handed to the comparison as if it
// were traffic, and it must not be given a framing it does not speak.
func TestSomethingThatFramesAsNeitherIsNotTheBattleProtocol(t *testing.T) {
	tlsRecord := []byte{0x16, 0x03, 0x01, 0x00, 0x2f, 0x01, 0x00, 0x00, 0x2b}
	if transport, _ := detectTransport(tlsRecord, tlsRecord, false); transport != "" {
		t.Errorf("a TLS record is not a battle frame, got transport %q", transport)
	}
	// Nor is a stream too short to hold a frame head: there is nothing to decide on.
	if transport, _ := detectTransport([]byte{0x15}, nil, false); transport != "" {
		t.Errorf("a single byte cannot carry a frame head, got transport %q", transport)
	}
}
