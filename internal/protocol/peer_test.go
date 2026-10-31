package protocol

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// Byte-exact tests for the 0x0F peer-message envelope and the inner messages,
// built from the field order each deserialiser reads:
//
//	AddPeer (ADDPEER)                    [06][blob][blob][u8][u8][u8][u64] exact
//	Remove/Dropout (REMOVEPEER, DROPOUT)  [type][u32][blob][u8] exact
//	IntegrityCheck (INTEGRITYCHECK)      [03][blob][u8][u8] exact
//	The 0x0F envelope                    [0f][15-byte header][inner] exact
//
// Every length-prefixed field is a u16 little-endian count followed by that
// many bytes: the original reads it with a raw 16-bit load and copies.

func blob(s string) []byte {
	b := make([]byte, 2+len(s))
	binary.LittleEndian.PutUint16(b, uint16(len(s)))
	copy(b[2:], s)
	return b
}

func TestParsePeerHeader(t *testing.T) {
	raw := []byte{
		0x04,                   // flagBits (compressed)
		0x11, 0x22, 0x33, 0x44, // packetNumber
		0x55, 0x66, 0x77, 0x88, // startACKReliableNonSequential
		0x99, 0xAA, // ackBits
		0xBB, 0xCC, 0xDD, 0xEE, // lastReliableSequentialID
	}
	if len(raw) != PeerHeaderLen {
		t.Fatalf("test vector is %d bytes, want %d", len(raw), PeerHeaderLen)
	}
	h, err := ParsePeerHeader(raw)
	if err != nil {
		t.Fatalf("ParsePeerHeader: %v", err)
	}
	if h.FlagBits != 0x04 {
		t.Errorf("FlagBits = %#x, want 0x04", h.FlagBits)
	}
	if h.PacketNumber != 0x44332211 {
		t.Errorf("PacketNumber = %#x, want 0x44332211", h.PacketNumber)
	}
	if h.StartACKReliableNonSequential != 0x88776655 {
		t.Errorf("StartACKReliableNonSequential = %#x, want 0x88776655", h.StartACKReliableNonSequential)
	}
	if h.AckBits != 0xAA99 {
		t.Errorf("AckBits = %#x, want 0xAA99 (little-endian)", h.AckBits)
	}
	if h.LastReliableSequentialID != 0xEEDDCCBB {
		t.Errorf("LastReliableSequentialID = %#x, want 0xEEDDCCBB", h.LastReliableSequentialID)
	}
}

func TestPeerHeaderRoundtrip(t *testing.T) {
	h := PeerHeader{
		FlagBits:                      FlagCompressed,
		PacketNumber:                  0xDEADBEEF,
		StartACKReliableNonSequential: 0x01020304,
		AckBits:                       0x0506,
		LastReliableSequentialID:      0x0708090A,
	}
	raw := h.Serialize()
	if len(raw) != PeerHeaderLen {
		t.Fatalf("Serialize length = %d, want %d", len(raw), PeerHeaderLen)
	}
	got, err := ParsePeerHeader(raw)
	if err != nil {
		t.Fatalf("ParsePeerHeader: %v", err)
	}
	if got != h {
		t.Errorf("roundtrip = %+v, want %+v", got, h)
	}
}

func TestParsePeerHeaderTooShort(t *testing.T) {
	if _, err := ParsePeerHeader(make([]byte, PeerHeaderLen-1)); err == nil {
		t.Error("expected an error for a header shorter than 15 bytes")
	}
}

func TestParsePeerEnvelope(t *testing.T) {
	header := PeerHeader{PacketNumber: 7, AckBits: 3}.Serialize()
	inner := append([]byte{PeerMsgAddPeer}, blob("10.0.0.1:27012")...)
	env, err := ParsePeerEnvelope(append(header, inner...))
	if err != nil {
		t.Fatalf("ParsePeerEnvelope: %v", err)
	}
	if env.Header.PacketNumber != 7 || env.Header.AckBits != 3 {
		t.Errorf("header = %+v", env.Header)
	}
	if env.InnerType != PeerMsgAddPeer {
		t.Errorf("InnerType = 0x%02X, want 0x%02X", env.InnerType, PeerMsgAddPeer)
	}
	if !bytes.Equal(env.Inner, inner) {
		t.Errorf("Inner = %x, want %x", env.Inner, inner)
	}
}

// A 16-byte message (header only, no inner part) is what the original accepts
// when nothing follows the header.
func TestParsePeerEnvelopeHeaderOnly(t *testing.T) {
	env, err := ParsePeerEnvelope(PeerHeader{}.Serialize())
	if err != nil {
		t.Fatalf("ParsePeerEnvelope: %v", err)
	}
	if len(env.Inner) != 0 || env.InnerType != 0 {
		t.Errorf("expected an empty inner part, got type 0x%02X and %d bytes", env.InnerType, len(env.Inner))
	}
}

func TestParsePeerEnvelopeTooShort(t *testing.T) {
	if _, err := ParsePeerEnvelope(make([]byte, PeerHeaderLen-1)); err == nil {
		t.Error("expected an error for an envelope shorter than the header")
	}
}

func TestParseAddPeer(t *testing.T) {
	raw := []byte{PeerMsgAddPeer}
	raw = append(raw, blob("192.168.1.50:27012")...)
	raw = append(raw, blob("extra-blob")...)
	raw = append(raw, 0x02)                                         // peerID
	raw = append(raw, 0x01)                                         // isHost
	raw = append(raw, 0x00)                                         // isLocal
	raw = binary.LittleEndian.AppendUint64(raw, 0x1122334455667788) // globalID

	m, err := ParseAddPeer(raw)
	if err != nil {
		t.Fatalf("ParseAddPeer: %v", err)
	}
	if m.AddressList != "192.168.1.50:27012" {
		t.Errorf("AddressList = %q", m.AddressList)
	}
	if string(m.Extra) != "extra-blob" {
		t.Errorf("Extra = %q, want %q", m.Extra, "extra-blob")
	}
	if m.PeerID != 2 || !m.IsHost || m.IsLocal {
		t.Errorf("peerID=%d isHost=%v isLocal=%v", m.PeerID, m.IsHost, m.IsLocal)
	}
	if m.GlobalID != 0x1122334455667788 {
		t.Errorf("GlobalID = %#x", m.GlobalID)
	}
}

func TestParseAddPeerRejects(t *testing.T) {
	good := []byte{PeerMsgAddPeer}
	good = append(good, blob("addr")...)
	good = append(good, blob("")...)
	good = append(good, 0x01, 0x00, 0x00)
	good = binary.LittleEndian.AppendUint64(good, 1)

	if _, err := ParseAddPeer(good[:len(good)-1]); err == nil {
		t.Error("expected an error when globalID is truncated")
	}
	trailing := append(append([]byte{}, good...), 0xFF)
	if _, err := ParseAddPeer(trailing); err == nil {
		t.Error("expected an error for a trailing byte (the original requires exact consumption)")
	}
	wrongType := append([]byte{}, good...)
	wrongType[0] = PeerMsgRemovePeer
	if _, err := ParseAddPeer(wrongType); err == nil {
		t.Error("expected an error for the wrong inner type")
	}
}

func TestParsePeerRemoval(t *testing.T) {
	for _, typ := range []byte{PeerMsgRemovePeer, PeerMsgDropoutPeer} {
		raw := []byte{typ}
		raw = binary.LittleEndian.AppendUint32(raw, 0xCAFEBABE) // reasonCode
		raw = append(raw, blob("kicked")...)                    // reason
		raw = append(raw, 0x03)                                 // peerID

		m, err := ParsePeerRemoval(raw, typ)
		if err != nil {
			t.Fatalf("ParsePeerRemoval(0x%02X): %v", typ, err)
		}
		if m.ReasonCode != 0xCAFEBABE {
			t.Errorf("type 0x%02X ReasonCode = %#x", typ, m.ReasonCode)
		}
		if m.Reason != "kicked" {
			t.Errorf("type 0x%02X Reason = %q", typ, m.Reason)
		}
		if m.PeerID != 3 {
			t.Errorf("type 0x%02X PeerID = %d", typ, m.PeerID)
		}
	}
}

func TestParsePeerRemovalWrongType(t *testing.T) {
	raw := []byte{PeerMsgDropoutPeer}
	raw = binary.LittleEndian.AppendUint32(raw, 0)
	raw = append(raw, 0x00, 0x01)
	if _, err := ParsePeerRemoval(raw, PeerMsgRemovePeer); err == nil {
		t.Error("expected an error when the type is not the expected one")
	}
}

func TestParseIntegrityCheck(t *testing.T) {
	raw := []byte{PeerMsgIntegrityCheck}
	raw = append(raw, blob("peers")...)
	raw = append(raw, 0x02) // proposedHostID
	raw = append(raw, 0x00) // migrateState

	m, err := ParseIntegrityCheck(raw)
	if err != nil {
		t.Fatalf("ParseIntegrityCheck: %v", err)
	}
	if string(m.Peers) != "peers" {
		t.Errorf("Peers = %q", m.Peers)
	}
	if m.ProposedHostID != 2 {
		t.Errorf("ProposedHostID = %d, want 2", m.ProposedHostID)
	}
	if m.MigrateState != 0 {
		t.Errorf("MigrateState = %d, want 0", m.MigrateState)
	}

	// migrateState 1 must not be mistaken for 0: it is the state that does not
	// move the host.
	raw[len(raw)-1] = 1
	m, err = ParseIntegrityCheck(raw)
	if err != nil {
		t.Fatalf("ParseIntegrityCheck: %v", err)
	}
	if m.MigrateState != 1 {
		t.Errorf("MigrateState = %d, want 1", m.MigrateState)
	}
}

func TestPeerTypeName(t *testing.T) {
	cases := map[byte]string{
		PeerMsgRequestJoin:       "REQUESTJOIN",
		PeerMsgIntegrityCheck:    "INTEGRITYCHECK",
		PeerMsgRequestJoinDenied: "REQUESTJOINDENIED",
		PeerMsgAddPeer:           "ADDPEER",
		PeerMsgRemovePeer:        "REMOVEPEER",
		PeerMsgDropoutPeer:       "DROPOUT",
		PeerMsgGameData:          "GAMEDATA",
		0x01:                     "0x01",
	}
	for typ, want := range cases {
		if got := PeerTypeName(typ); got != want {
			t.Errorf("PeerTypeName(0x%02X) = %q, want %q", typ, got, want)
		}
	}
}

func TestParseRequestJoin(t *testing.T) {
	raw := []byte{PeerMsgRequestJoin}
	raw = append(raw, blob("player")...)
	raw = append(raw, blob("opaque")...)
	// Packed exactly like LobbyConnectMessage: title ID in the high 16 bits,
	// session ID in the low 48.
	raw = binary.LittleEndian.AppendUint64(raw, 0x0018_112233445566)
	raw = binary.LittleEndian.AppendUint64(raw, 0xAABBCCDDEEFF0011)

	m, err := ParseRequestJoin(raw)
	if err != nil {
		t.Fatalf("ParseRequestJoin: %v", err)
	}
	if m.Name != "player" {
		t.Errorf("Name = %q", m.Name)
	}
	if string(m.Extra) != "opaque" {
		t.Errorf("Extra = %q", m.Extra)
	}
	if m.TitleID != 0x18 {
		t.Errorf("TitleID = %#x, want 0x18", m.TitleID)
	}
	if m.SessionID != 0x112233445566 {
		t.Errorf("SessionID = %#x, want 0x112233445566", m.SessionID)
	}
	if m.PeerKey != 0xAABBCCDDEEFF0011 {
		t.Errorf("PeerKey = %#x", m.PeerKey)
	}
}

func TestParseRequestJoinRejects(t *testing.T) {
	good := []byte{PeerMsgRequestJoin}
	good = append(good, blob("n")...)
	good = append(good, blob("")...)
	good = binary.LittleEndian.AppendUint64(good, 0)
	good = binary.LittleEndian.AppendUint64(good, 1)

	if _, err := ParseRequestJoin(good[:len(good)-1]); err == nil {
		t.Error("expected an error when the peer key is truncated")
	}
	if _, err := ParseRequestJoin(append(append([]byte{}, good...), 0x00)); err == nil {
		t.Error("expected an error for a trailing byte")
	}
}

func TestParseRequestJoinDenied(t *testing.T) {
	raw := []byte{PeerMsgRequestJoinDenied}
	raw = append(raw, blob("opaque")...)
	raw = binary.LittleEndian.AppendUint32(raw, 0x04030201) // reasonCode
	raw = append(raw, blob("session full")...)
	raw = binary.LittleEndian.AppendUint64(raw, 0x0102030405060708)

	m, err := ParseRequestJoinDenied(raw)
	if err != nil {
		t.Fatalf("ParseRequestJoinDenied: %v", err)
	}
	if string(m.Extra) != "opaque" {
		t.Errorf("Extra = %q", m.Extra)
	}
	if m.ReasonCode != 0x04030201 {
		t.Errorf("ReasonCode = %#x", m.ReasonCode)
	}
	if m.Reason != "session full" {
		t.Errorf("Reason = %q", m.Reason)
	}
	if m.PeerKey != 0x0102030405060708 {
		t.Errorf("PeerKey = %#x", m.PeerKey)
	}

	if _, err := ParseRequestJoinDenied(append(append([]byte{}, raw...), 0xFF)); err == nil {
		t.Error("expected an error for a trailing byte")
	}
	bad := append([]byte{}, raw...)
	bad[0] = PeerMsgAddPeer
	if _, err := ParseRequestJoinDenied(bad); err == nil {
		t.Error("expected an error for the wrong inner type")
	}
}
