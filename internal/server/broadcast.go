package server

import "github.com/luskaner/ageBattleServer/internal/protocol"

// broadcastWebSocketPort is what the presence broadcast advertises as the
// WebSocket port. Zero is deliberate: the field exists in the payload
// (RelayBroadcastMessage, relay.md:117) but the server does not advertise its
// browser/watch port to the LAN. A discovery client that wants the WebSocket
// port has to be configured with it out of band. Keep the zero.
const broadcastWebSocketPort = 0

func (s *Server) buildBroadcastMessage() []byte {
	return protocol.BuildRelayBroadcastMessage(
		s.relayBroadcaster.GUID(),
		s.config.Name,
		uint16(s.config.PublicPort),
		broadcastWebSocketPort,
		uint16(s.config.OutOfBandPort),
	)
}

// buildBroadcastMessages returns the set of presence broadcasts for the given
// destination port (see protocol.BuildRelayBroadcastMessages for the per-port
// format rules: legacy on 8888, current+legacy on 9999, current elsewhere).
func (s *Server) buildBroadcastMessages(port uint) [][]byte {
	return protocol.BuildRelayBroadcastMessages(
		s.relayBroadcaster.GUID(),
		s.config.Name,
		uint16(s.config.PublicPort),
		broadcastWebSocketPort,
		uint16(s.config.OutOfBandPort),
		port,
	)
}
