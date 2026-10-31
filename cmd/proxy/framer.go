package main

import "encoding/binary"

// frameMode selects the framing used to reassemble a captured byte stream.
type frameMode int

const (
	frameRLink frameMode = iota
	frameWebSocket
)

// maxWSPayload bounds accepted WebSocket payload lengths: game frames are u16
// framed upstream, so anything beyond 1 MiB is a desync, not a real frame.
const maxWSPayload = 1 << 20

// frameAssembler reassembles a byte stream into complete protocol frames so
// segmented TCP reads never produce partial CSV records. Complete frames are
// returned as they become available; incomplete tails stay buffered.
//
// The returned slices alias the assembler's internal buffer and stay valid
// only until the next feed call (consumers log them synchronously).
type frameAssembler struct {
	mode frameMode
	buf  []byte
}

func newFrameAssembler(mode frameMode) *frameAssembler {
	return &frameAssembler{mode: mode, buf: make([]byte, 0, 4096)}
}

// feed appends data and returns the complete frames now available. Client
// WebSocket payloads are unmasked in place so logged rows carry the real
// message bytes; the relayed stream is untouched (the pump forwards raw reads).
func (a *frameAssembler) feed(data []byte) [][]byte {
	a.buf = append(a.buf, data...)
	var out [][]byte
	for {
		n := a.frameLen()
		if n < 0 {
			// Impossible framing (bad length/opcode): flush everything raw so
			// the capture resynchronizes on the next real frame instead of
			// stalling forever.
			out = append(out, a.buf)
			a.buf = nil
			break
		}
		if n == 0 {
			break
		}
		if a.mode == frameWebSocket {
			unmaskWS(a.buf[:n])
		}
		out = append(out, a.buf[:n])
		a.buf = a.buf[n:]
	}
	// Leading data that cannot start a frame (e.g. a WebSocket handshake or a
	// raw TLS record when MITM is off) must not stall reassembly: emit it as a
	// raw record so later frames still get captured.
	if len(a.buf) > 0 && !a.startsWithFrame() {
		out = append(out, a.buf)
		a.buf = nil
	}
	return out
}

// unmaskWS unmasks the payload of a masked WebSocket frame in place.
func unmaskWS(frame []byte) {
	if len(frame) < 2 || frame[1]&0x80 == 0 {
		return
	}
	payloadLen := int(frame[1] & 0x7F)
	idx := 2
	switch payloadLen {
	case 126:
		if len(frame) < 4 {
			return
		}
		payloadLen = int(binary.BigEndian.Uint16(frame[2:4]))
		idx = 4
	case 127:
		if len(frame) < 10 {
			return
		}
		payloadLen = int(binary.BigEndian.Uint64(frame[2:10]))
		idx = 10
	}
	key := frame[idx : idx+4]
	idx += 4
	for i := 0; i < payloadLen && idx+i < len(frame); i++ {
		frame[idx+i] ^= key[i%4]
	}
}

// frameLen returns the total length of a complete frame at the start of the
// buffer, 0 when more data is needed, or -1 when the head cannot be a frame
// under any continuation of the stream (framing error).
func (a *frameAssembler) frameLen() int {
	switch a.mode {
	case frameWebSocket:
		return wsFrameLen(a.buf)
	default:
		return rlinkFrameLen(a.buf)
	}
}

// startsWithFrame reports whether the buffered head could be the start of a
// frame, meaning more data should be awaited instead of flushing raw.
func (a *frameAssembler) startsWithFrame() bool {
	switch a.mode {
	case frameWebSocket:
		if len(a.buf) < 2 {
			return true // header incomplete: wait
		}
		switch a.buf[0] & 0x0F {
		case 0, 1, 2, 8, 9, 10: // continuation/text/binary/close/ping/pong
			return true
		}
		return false
	default:
		// RLink: any two bytes are a candidate length field. A declared length
		// of zero is impossible and already reported as -1 by frameLen.
		return true
	}
}

// rlinkFrameLen returns the total length of a complete RLink frame
// ([u16 BE length][type][body]), 0 if incomplete, or -1 when the declared
// length is impossible (a zero-length body cannot occur on this protocol).
func rlinkFrameLen(buf []byte) int {
	if len(buf) < 2 {
		return 0
	}
	declared := int(binary.BigEndian.Uint16(buf[:2]))
	if declared == 0 {
		return -1
	}
	total := 2 + declared
	if total > len(buf) {
		return 0
	}
	return total
}

// wsFrameLen returns the total byte length of a complete WebSocket frame at the
// start of buf, 0 when more data is needed, or -1 when the head cannot start a
// frame. Valid opcodes are continuation(0), text(1), binary(2), close(8),
// ping(9) and pong(10).
func wsFrameLen(buf []byte) int {
	if len(buf) < 2 {
		return 0
	}
	switch buf[0] & 0x0F {
	case 0, 1, 2, 8, 9, 10:
	default:
		return -1
	}
	payloadLen := int(buf[1] & 0x7F)
	idx := 2
	switch payloadLen {
	case 126:
		if len(buf) < 4 {
			return 0
		}
		payloadLen = int(binary.BigEndian.Uint16(buf[2:4]))
		idx = 4
	case 127:
		if len(buf) < 10 {
			return 0
		}
		hi := binary.BigEndian.Uint32(buf[2:6])
		if hi != 0 {
			return -1
		}
		payloadLen = int(binary.BigEndian.Uint32(buf[6:10]))
		idx = 10
	}
	if payloadLen > maxWSPayload {
		return -1
	}
	if buf[1]&0x80 != 0 { // masked
		idx += 4
	}
	total := idx + payloadLen
	if total > len(buf) {
		return 0
	}
	return total
}
