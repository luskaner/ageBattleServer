package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	MsgRequestJoin             = 0x01
	MsgRequestJoinDenied       = 0x02
	MsgAddPeer                 = 0x03
	MsgRemovePeer              = 0x04
	MsgDropoutPeer             = 0x05
	MsgIntegrityCheck          = 0x06
	MsgGameVoice               = 0x07
	MsgMultiGameData           = 0x08
	MsgGameData                = 0x09
	MsgEndOfGameData           = 0x0A
	MsgNudge                   = 0x0B
	MsgRouteCheck              = 0x0C
	MsgSeek                    = 0x0D
	MsgSeekResponse            = 0x0E
	MsgPeer                    = 0x0F
	MsgProxy                   = 0x10
	MsgPing                    = 0x11
	MsgLobbyRelay              = 0x13
	MsgLobbyPing               = 0x14
	MsgLobbyConnect            = 0x15
	MsgStartObserve            = 0x18
	MsgStartObserveResp        = 0x19
	MsgActivateObserver        = 0x1A
	MsgStopObserve             = 0x1B
	MsgSetObservableParameters = 0x1D
	MsgLeaveSimulation         = 0x1F
	MsgRelayBroadcast          = 0x21
	// MsgRelayBroadcastLegacy is the Age1-era presence broadcast type (0x20):
	// same body layout as MsgRelayBroadcast but only two trailing u16 fields
	// (port + 0, since Age1 has no websocket port).
	MsgRelayBroadcastLegacy = 0x20
	MsgGameDataChecksum     = 0x17
	// MsgActivateObserverResp is the ActivateObserverResponseMessage wire type (30
	// in the WS protocol): body = [status u8][time i64].
	MsgActivateObserverResp = 0x1E
	// MsgEndOfGameDataObserver is the EndOfGameDataMessage wire type (34 in the
	// WS protocol) sent to observers once they are caught up. MsgEndOfGameData
	// (0x0A) is the raw-protocol type and is not used for the observer stream.
	MsgEndOfGameDataObserver = 0x22
)

// MessageTypeName returns the human-readable name of a wire message type. It
// references every message identifier the protocol defines, including those this
// server never emits (RequestJoinDenied, RemovePeer, DropoutPeer), so the full
// set stays referenced and available for logging.
func MessageTypeName(t byte) string {
	switch t {
	case MsgRequestJoin:
		return "RequestJoin"
	case MsgRequestJoinDenied:
		return "RequestJoinDenied"
	case MsgAddPeer:
		return "AddPeer"
	case MsgRemovePeer:
		return "RemovePeer"
	case MsgDropoutPeer:
		return "DropoutPeer"
	case MsgIntegrityCheck:
		return "IntegrityCheck"
	case MsgGameVoice:
		return "GameVoice"
	case MsgMultiGameData:
		return "MultiGameData"
	case MsgGameData:
		return "GameData"
	case MsgEndOfGameData:
		return "EndOfGameData"
	case MsgNudge:
		return "Nudge"
	case MsgRouteCheck:
		return "RouteCheck"
	case MsgSeek:
		return "Seek"
	case MsgSeekResponse:
		return "SeekResponse"
	case MsgPeer:
		return "Peer"
	case MsgProxy:
		return "Proxy"
	case MsgPing:
		return "Ping"
	case MsgLobbyRelay:
		return "LobbyRelay"
	case MsgLobbyPing:
		return "LobbyPing"
	case MsgLobbyConnect:
		return "LobbyConnect"
	case MsgStartObserve:
		return "StartObserve"
	case MsgStartObserveResp:
		return "StartObserveResp"
	case MsgActivateObserver:
		return "ActivateObserver"
	case MsgActivateObserverResp:
		return "ActivateObserverResp"
	case MsgStopObserve:
		return "StopObserve"
	case MsgSetObservableParameters:
		return "SetObservableParameters"
	case MsgLeaveSimulation:
		return "LeaveSimulation"
	case MsgRelayBroadcast:
		return "RelayBroadcast"
	case MsgRelayBroadcastLegacy:
		return "RelayBroadcastLegacy"
	case MsgGameDataChecksum:
		return "GameDataChecksum"
	case MsgEndOfGameDataObserver:
		return "EndOfGameDataObserver"
	default:
		return fmt.Sprintf("0x%02X", t)
	}
}

func AppendStringLE(buf []byte, s string) []byte {
	buf = binary.LittleEndian.AppendUint16(buf, uint16(len(s)))
	return append(buf, s...)
}

func AppendUint16LE(buf []byte, v uint16) []byte {
	return binary.LittleEndian.AppendUint16(buf, v)
}

// Default relay broadcast ports: 9999 and 8888 (see internal/defaults).
const (
	RelayBroadcastPortDefault = 9999
	RelayBroadcastPortLegacy  = 8888
)

// BuildRelayBroadcastMessage serializes the current presence broadcast (type
// 0x21): GUID, server name, public port, websocket port, out-of-band port.
func BuildRelayBroadcastMessage(guid, name string, publicPort, wsPort, oobPort uint16) []byte {
	msg := make([]byte, 0, 64)
	msg = append(msg, MsgRelayBroadcast)
	msg = AppendStringLE(msg, guid)
	msg = AppendStringLE(msg, name)
	msg = AppendUint16LE(msg, publicPort)
	msg = AppendUint16LE(msg, wsPort)
	msg = AppendUint16LE(msg, oobPort)
	return msg
}

// BuildLegacyRelayBroadcastMessage serializes the Age1-era presence broadcast
// (type 0x20): GUID, server name, the public (bs) port and the websocket port
// (0 for Age1, which has none).
func BuildLegacyRelayBroadcastMessage(guid, name string, port, wsPort uint16) []byte {
	msg := make([]byte, 0, 48)
	msg = append(msg, MsgRelayBroadcastLegacy)
	msg = AppendStringLE(msg, guid)
	msg = AppendStringLE(msg, name)
	msg = AppendUint16LE(msg, port)
	msg = AppendUint16LE(msg, wsPort)
	return msg
}

// BuildRelayBroadcastMessages returns the presence broadcasts for a given
// destination port. The default relay ports are special-cased: 8888 carries
// only the legacy Age1 format, 9999 carries both the current (0x21) and the
// legacy (0x20) format, and any other port carries only the current format.
func BuildRelayBroadcastMessages(guid, name string, publicPort, wsPort, oobPort uint16, broadcastPort uint) [][]byte {
	switch broadcastPort {
	case RelayBroadcastPortLegacy:
		return [][]byte{BuildLegacyRelayBroadcastMessage(guid, name, publicPort, wsPort)}
	case RelayBroadcastPortDefault:
		return [][]byte{
			BuildRelayBroadcastMessage(guid, name, publicPort, wsPort, oobPort),
			BuildLegacyRelayBroadcastMessage(guid, name, publicPort, wsPort),
		}
	default:
		return [][]byte{BuildRelayBroadcastMessage(guid, name, publicPort, wsPort, oobPort)}
	}
}

type LobbyConnectInfo struct {
	SessionID            uint64
	Secret               uint64
	ProfileID            uint64
	DataReceivedCount    uint32
	TimeoutOverride      uint32
	TitleID              uint16
	ClientMessageVersion uint16
	ReconnectState       byte
}

func ParseLobbyConnect(body []byte) (LobbyConnectInfo, bool) {
	var info LobbyConnectInfo
	// Minimum fixed part: [sessionID 48b][titleID 16b][secret u64][reconnectState u8]
	// = 8 + 8 + 1 = 17 bytes (the first 8 cover sessionID + titleID).
	if len(body) < 17 {
		return info, false
	}
	info.SessionID = binary.LittleEndian.Uint64(body[0:8]) & 0xFFFFFFFFFFFF // 48-bit session ID
	info.TitleID = binary.LittleEndian.Uint16(body[6:8])
	info.Secret = binary.LittleEndian.Uint64(body[8:16])
	info.ReconnectState = body[16]
	// The data-received count is -1 unless the message actually carries one, and
	// that default is not cosmetic: the field is serialised as -1 for every
	// reconnect state but 2, and it is read as the first frame of the reconnect
	// resend range. A state 1 or state 3+ connect therefore asks for a range that
	// starts below zero, which the range check rejects as invalid and nothing is
	// resent. Leaving it at 0 would instead resend the whole retained window.
	info.DataReceivedCount = ^uint32(0)
	pos := 17
	if info.ReconnectState == 2 {
		// Reconnect state 2 carries the dataReceivedCount (u32) before the
		// version fields.
		if len(body) < 21 {
			return info, false
		}
		info.DataReceivedCount = binary.LittleEndian.Uint32(body[17:21])
		pos = 21
	}
	if len(body) < pos+2 {
		info.ClientMessageVersion = 0
		info.TimeoutOverride = 0
		return info, len(body) == pos
	}
	info.ClientMessageVersion = binary.LittleEndian.Uint16(body[pos : pos+2])
	if len(body) < pos+6 {
		return info, false
	}
	info.TimeoutOverride = binary.LittleEndian.Uint32(body[pos+2 : pos+6])
	if info.ClientMessageVersion >= 5 {
		// Client message version >= 5 appends the profile ID (u64).
		if len(body) < pos+14 {
			return info, false
		}
		info.ProfileID = binary.LittleEndian.Uint64(body[pos+6 : pos+14])
		return info, len(body) == pos+14
	}
	// Versions < 5 end after the timeout override (pos+6).
	return info, len(body) == pos+6
}

// BuildLobbyConnect serialises a LobbyConnectMessage body (everything after the
// 0x15 type byte), using the field order and byte order confirmed on the wire:
// the session ID and title ID share one u64 (title ID in the high 16 bits),
// followed by the secret, the reconnect state, the data-received count only when
// the state is 2, then the client message version, the flags, and the trailing
// 8-byte field only when the version is 5 or above. ParseLobbyConnect is the
// exact inverse, so Build -> Parse round-trips.
//
// The server's own lobby registration uses this with titleID 0xFFFF, sessionID
// and secret -1 and no data count, which are the message's own defaults.
func BuildLobbyConnect(titleID uint16, sessionID int64, secret uint64, reconnectState byte, dataReceivedCount int32, version uint16, flags uint32, profileID uint64) []byte {
	body := make([]byte, 0, 31)
	body = binary.LittleEndian.AppendUint64(body, uint64(sessionID)&0xFFFFFFFFFFFF|uint64(titleID)<<48)
	body = binary.LittleEndian.AppendUint64(body, secret)
	body = append(body, reconnectState)
	if reconnectState == 2 {
		body = binary.LittleEndian.AppendUint32(body, uint32(dataReceivedCount))
	}
	body = binary.LittleEndian.AppendUint16(body, version)
	body = binary.LittleEndian.AppendUint32(body, flags)
	if version >= 5 {
		body = binary.LittleEndian.AppendUint64(body, profileID)
	}
	return body
}

type GameDataMessage struct {
	InnerType byte
	Payload   []byte
}

func (m *GameDataMessage) Serialize() []byte {
	// GameData body: [inner type u8][len u16 LE][payload].
	buf := make([]byte, 3+len(m.Payload))
	buf[0] = m.InnerType
	binary.LittleEndian.PutUint16(buf[1:3], uint16(len(m.Payload)))
	copy(buf[3:], m.Payload)
	return buf
}

func (m *GameDataMessage) Deserialize(data []byte) error {
	if len(data) < 3 {
		return errors.New("gamedata too short")
	}
	m.InnerType = data[0]
	length := binary.LittleEndian.Uint16(data[1:3])
	if len(data) < 3+int(length) {
		return errors.New("gamedata truncated")
	}
	m.Payload = data[3 : 3+length]
	return nil
}

// ReadGameData reads a GameData body (everything after the type byte) field by
// field: the tag, the u16 little-endian payload length, then the payload. A field
// that does not fit stops the read and stays at its zero value, and so does every
// field after it: a missing tag reads as 0 and a payload that runs past the end
// reads as empty. exact reports whether the body was consumed to its last byte.
//
// The two readers below make the two uses explicit. A top-level 0x09 (and the
// 0x09 a 0x17 carries) counts only when exact, see ParseGameData. A GameData
// inside a peer message is taken as far as it was read, because the server acts
// on its tag even when the rest of it is malformed.
func ReadGameData(body []byte) (m GameDataMessage, exact bool) {
	if len(body) < 1 {
		return m, false
	}
	m.InnerType = body[0]
	if len(body) < 3 {
		return m, false
	}
	n := int(binary.LittleEndian.Uint16(body[1:3]))
	if 3+n > len(body) {
		return m, false
	}
	if n > 0 {
		m.Payload = body[3 : 3+n]
	}
	return m, 3+n == len(body)
}

// ParseGameData parses a top-level GameData body, which must be consumed
// exactly: a declared length that is not the actual one is an error, so a
// malformed upload is dropped instead of reaching the next 0x0A.
func ParseGameData(body []byte) (GameDataMessage, error) {
	m, exact := ReadGameData(body)
	if !exact {
		return GameDataMessage{}, fmt.Errorf("gamedata: body of %d bytes does not match its declared length", len(body))
	}
	return m, nil
}

// IsReplayData reports whether the GameData carries the replay page stream:
// inner type 0xFE is the replay peer ID (uploaded by the host, E2E00).
func (m *GameDataMessage) IsReplayData() bool {
	return m.InnerType == 0xFE
}

// The six messages below (0x07 GameVoice, 0x08 MultiGameData, 0x0B Nudge,
// 0x0C RouteCheck, 0x0D Seek, 0x0E SeekResponse) have parsers and serializers
// exercised by the round-trip tests, but the dispatcher has no case for any of
// them: they are only ever logged as an unknown message type. The server
// therefore has no handler for them, and these definitions exist to document the
// wire format, not because anything sends or receives them.
type GameVoiceMessage struct {
	VoiceData []byte
}

func (m *GameVoiceMessage) Serialize() []byte {
	frame := make([]byte, 4+len(m.VoiceData))
	frame[0] = MsgGameVoice
	binary.BigEndian.PutUint16(frame[1:3], uint16(1+len(m.VoiceData)))
	frame[3] = 11
	copy(frame[4:], m.VoiceData)
	return frame
}

func (m *GameVoiceMessage) Deserialize(data []byte) error {
	if len(data) < 2 {
		return errors.New("game voice too short")
	}
	if data[0] != 11 {
		return errors.New("game voice inner type must be 11 (0x0B)")
	}
	m.VoiceData = data[1:]
	return nil
}

type MultiGameDataMessage struct {
	SubMessages []GameDataMessage
}

func (m *MultiGameDataMessage) Serialize() []byte {
	// The payload is [8][count] followed by each sub-message's game-data body
	// ([inner type][len u16 LE][payload]), so its size is computable up front
	// and the frame fits in a single allocation.
	total := 2
	for _, sub := range m.SubMessages {
		total += 3 + len(sub.Payload)
	}
	frame := make([]byte, 3+total)
	frame[0] = MsgMultiGameData
	binary.BigEndian.PutUint16(frame[1:3], uint16(total))
	body := frame[3:]
	body[0] = 8
	body[1] = byte(len(m.SubMessages))
	off := 2
	for _, sub := range m.SubMessages {
		body[off] = sub.InnerType
		binary.LittleEndian.PutUint16(body[off+1:off+3], uint16(len(sub.Payload)))
		copy(body[off+3:], sub.Payload)
		off += 3 + len(sub.Payload)
	}
	return frame
}

func (m *MultiGameDataMessage) Deserialize(data []byte) error {
	if len(data) < 2 {
		return errors.New("multi game data too short")
	}
	if data[0] != 8 {
		return errors.New("multi game data inner type must be 8 (0x08)")
	}
	count := int(data[1])
	if len(data) < 2+count*3 {
		return errors.New("multi game data truncated")
	}
	m.SubMessages = make([]GameDataMessage, count)
	offset := 2
	for i := range count {
		if offset >= len(data) {
			return fmt.Errorf("multi game data truncated at sub-message %d", i)
		}
		var sub GameDataMessage
		if err := sub.Deserialize(data[offset:]); err != nil {
			return fmt.Errorf("sub-message %d: %w", i, err)
		}
		m.SubMessages[i] = sub
		offset += 3 + int(binary.LittleEndian.Uint16(data[offset+1:offset+3]))
	}
	return nil
}

type NudgeMessage struct{}

func (m *NudgeMessage) Serialize() []byte {
	frame := make([]byte, 4)
	frame[0] = MsgNudge
	binary.BigEndian.PutUint16(frame[1:3], 1)
	frame[3] = 13
	return frame
}

func (m *NudgeMessage) Deserialize(data []byte) error {
	if len(data) < 1 {
		return errors.New("nudge too short")
	}
	if data[0] != 13 {
		return errors.New("nudge inner type must be 13 (0x0D)")
	}
	return nil
}

type RouteCheckMessage struct {
	SessionID  uint64
	PeerID     byte
	RelayCount uint16
}

func (m *RouteCheckMessage) Serialize() []byte {
	frame := make([]byte, 15)
	frame[0] = MsgRouteCheck
	binary.BigEndian.PutUint16(frame[1:3], 12)
	frame[3] = 12
	binary.LittleEndian.PutUint64(frame[4:12], m.SessionID)
	frame[12] = m.PeerID
	binary.LittleEndian.PutUint16(frame[13:15], m.RelayCount)
	return frame
}

func (m *RouteCheckMessage) Deserialize(data []byte) error {
	if len(data) < 12 {
		return errors.New("route check too short")
	}
	if data[0] != 12 {
		return errors.New("route check inner type must be 12 (0x0C)")
	}
	m.SessionID = binary.LittleEndian.Uint64(data[1:9])
	m.PeerID = data[9]
	m.RelayCount = binary.LittleEndian.Uint16(data[10:12])
	return nil
}

type SeekMessage struct {
	TargetSessionID uint64
}

func (m *SeekMessage) Serialize() []byte {
	frame := make([]byte, 12)
	frame[0] = MsgSeek
	binary.BigEndian.PutUint16(frame[1:3], 9)
	frame[3] = 1
	binary.LittleEndian.PutUint64(frame[4:12], m.TargetSessionID)
	return frame
}

func (m *SeekMessage) Deserialize(data []byte) error {
	if len(data) < 9 {
		return errors.New("seek too short")
	}
	if data[0] != 1 {
		return errors.New("seek inner type must be 1")
	}
	m.TargetSessionID = binary.LittleEndian.Uint64(data[1:9])
	return nil
}

type SeekResponseMessage struct {
	SessionID uint64
}

func (m *SeekResponseMessage) Serialize() []byte {
	frame := make([]byte, 12)
	frame[0] = MsgSeekResponse
	binary.BigEndian.PutUint16(frame[1:3], 9)
	frame[3] = 4
	binary.LittleEndian.PutUint64(frame[4:12], m.SessionID)
	return frame
}

func (m *SeekResponseMessage) Deserialize(data []byte) error {
	if len(data) < 9 {
		return errors.New("seek response too short")
	}
	if data[0] != 4 {
		return errors.New("seek response inner type must be 4")
	}
	m.SessionID = binary.LittleEndian.Uint64(data[1:9])
	return nil
}
