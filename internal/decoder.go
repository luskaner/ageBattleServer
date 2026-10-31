package internal

import (
	"encoding/binary"
)

// ParseMessage labels a captured frame for the proxy's CSV output.
//
// The capture tool sees two framings and does not tell this function which one
// it got: bsPort and out-of-band frames are [u16 BE length][type][body] (length
// covers the type byte plus the body) while WebSocket payloads are the bare
// [type][body]. So it tries the length-prefixed form first and falls back to the
// bare one. That is good enough for a log label; the authoritative framing
// lives in the transports (server.Client, server.WsClient) and in the proxy's
// own framer.
func ParseMessage(data []byte) (msgType byte, body []byte, valid bool) {
	if len(data) < 1 {
		return 0, nil, false
	}

	if len(data) >= 3 {
		declared := int(binary.BigEndian.Uint16(data[:2]))
		if declared == len(data)-2 && declared < 10000 {
			return data[2], data[3 : 3+declared-1], true
		}
	}

	return data[0], data[1:], true
}
