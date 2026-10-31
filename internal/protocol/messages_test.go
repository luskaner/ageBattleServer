package protocol

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestGameDataMessageSerialize(t *testing.T) {
	msg := &GameDataMessage{
		InnerType: 0xFE,
		Payload:   []byte{0x01, 0x02, 0x03},
	}
	data := msg.Serialize()

	if data[0] != 0xFE {
		t.Errorf("expected inner type 0xFE, got 0x%02x", data[0])
	}
	if data[1] != 0x03 || data[2] != 0x00 {
		t.Errorf("expected little-endian length bytes 0x03 0x00, got 0x%02x 0x%02x", data[1], data[2])
	}
}

func TestGameDataMessageDeserialize(t *testing.T) {
	msg := &GameDataMessage{}
	err := msg.Deserialize([]byte{0xFE, 0x02, 0x00, 0xAB, 0xCD})
	if err != nil {
		t.Fatalf("deserialize error: %v", err)
	}
	if msg.InnerType != 0xFE {
		t.Errorf("expected inner type 0xFE, got 0x%02x", msg.InnerType)
	}
	if !bytes.Equal(msg.Payload, []byte{0xAB, 0xCD}) {
		t.Errorf("payload mismatch: %v", msg.Payload)
	}
}

func TestGameDataMessageIsReplayData(t *testing.T) {
	msg := &GameDataMessage{InnerType: 0xFE}
	if !msg.IsReplayData() {
		t.Error("expected 0xFE to be replay data")
	}

	msg.InnerType = 0x01
	if msg.IsReplayData() {
		t.Error("expected 0x01 to NOT be replay data")
	}
}

func TestAppendString(t *testing.T) {
	buf := []byte{0x01, 0x02}
	buf = AppendStringLE(buf, "test")

	if len(buf) != 8 {
		t.Errorf("expected length 8, got %d", len(buf))
	}
	if buf[2] != 0x04 || buf[3] != 0x00 {
		t.Errorf("expected little-endian length bytes 0x04 0x00, got 0x%02x 0x%02x", buf[2], buf[3])
	}
	if string(buf[4:]) != "test" {
		t.Errorf("expected 'test', got '%s'", string(buf[4:]))
	}
}

func TestAppendUint16LE(t *testing.T) {
	buf := []byte{0xAA, 0xBB}
	buf = AppendUint16LE(buf, 0x1234)

	if len(buf) != 4 {
		t.Fatalf("expected length 4, got %d", len(buf))
	}
	if buf[0] != 0xAA || buf[1] != 0xBB {
		t.Errorf("preexisting bytes modified: %v", buf[:2])
	}
	if buf[2] != 0x34 || buf[3] != 0x12 {
		t.Errorf("expected little-endian 0x34 0x12, got 0x%02x 0x%02x", buf[2], buf[3])
	}

	buf = AppendUint16LE(buf, 0x0000)
	if len(buf) != 6 || buf[4] != 0x00 || buf[5] != 0x00 {
		t.Errorf("zero value not appended correctly: %v", buf[4:])
	}
}

func TestGameDataMessageDeserializeEmptyPayload(t *testing.T) {
	msg := &GameDataMessage{}
	if err := msg.Deserialize([]byte{0xFE, 0x00, 0x00}); err != nil {
		t.Fatalf("deserialize error: %v", err)
	}
	if msg.InnerType != 0xFE {
		t.Errorf("InnerType = 0x%02x, want 0xFE", msg.InnerType)
	}
	if len(msg.Payload) != 0 {
		t.Errorf("Payload = %v, want empty", msg.Payload)
	}
}

func TestGameDataMessageDeserializeTooShort(t *testing.T) {
	msg := &GameDataMessage{}
	for _, data := range [][]byte{nil, {}, {0xFE}, {0xFE, 0x00}} {
		if err := msg.Deserialize(data); err == nil {
			t.Errorf("expected error for %v", data)
		}
	}
}

func TestGameDataMessageDeserializeTruncated(t *testing.T) {
	msg := &GameDataMessage{}
	err := msg.Deserialize([]byte{0xFE, 0x05, 0x00, 0xAB, 0xCD})
	if err == nil {
		t.Fatal("expected error for truncated payload")
	}
}

func TestGameDataMessageDeserializeLongLength(t *testing.T) {
	msg := &GameDataMessage{}
	err := msg.Deserialize([]byte{0xFE, 0xFF, 0xFF, 0xAB})
	if err == nil {
		t.Fatal("expected error for 0xFFFF length")
	}
}

func TestGameDataMessageDeserializeExactFit(t *testing.T) {
	data := []byte{0xFE, 0x04, 0x00, 0x01, 0x02, 0x03, 0x04}
	msg := &GameDataMessage{}
	if err := msg.Deserialize(data); err != nil {
		t.Fatalf("deserialize error: %v", err)
	}
	if !bytes.Equal(msg.Payload, []byte{0x01, 0x02, 0x03, 0x04}) {
		t.Errorf("Payload = %v, want {01 02 03 04}", msg.Payload)
	}
}

func TestGameDataMessageSerializeRoundtrip(t *testing.T) {
	msg := &GameDataMessage{
		InnerType: 0xAB,
		Payload:   []byte{1, 2, 3, 4, 5},
	}
	serialized := msg.Serialize()
	var decoded GameDataMessage
	if err := decoded.Deserialize(serialized); err != nil {
		t.Fatalf("deserialize error: %v", err)
	}
	if decoded.InnerType != msg.InnerType {
		t.Errorf("InnerType = 0x%02x, want 0x%02x", decoded.InnerType, msg.InnerType)
	}
	if !bytes.Equal(decoded.Payload, msg.Payload) {
		t.Errorf("Payload = %v, want %v", decoded.Payload, msg.Payload)
	}
}

func TestMessageConstants(t *testing.T) {
	checks := []struct {
		name string
		got  byte
		want byte
	}{
		{"MsgGameData", MsgGameData, 0x09},
		{"MsgPing", MsgPing, 0x11},
		{"MsgLobbyConnect", MsgLobbyConnect, 0x15},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = 0x%02x, want 0x%02x", c.name, c.got, c.want)
		}
	}
}

func TestParseLobbyConnect(t *testing.T) {
	body := []byte{
		0x66, 0x55, 0x44, 0x33, 0x22, 0x11, // sessionID 0x112233445566 (u48 LE)
		0xBB, 0xAA, // titleID 0xAABB (u16 LE)
		0xFF, 0xEE, 0xDD, 0xCC, 0xBB, 0xAA, 0x99, 0x88, // secret
		0x00,       // reconnectState 0
		0x05, 0x00, // clientMessageVersion 5
		0xBE, 0xBA, 0xFE, 0xCA, // timeoutOverride 0xCAFEBABE (u32 LE)
		0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, // profileID
	}

	info, ok := ParseLobbyConnect(body)
	if !ok {
		t.Fatal("expected parse to succeed")
	}
	if info.SessionID != 0x112233445566 {
		t.Errorf("SessionID = %#x, want 0x112233445566", info.SessionID)
	}
	if info.TitleID != 0xAABB {
		t.Errorf("TitleID = %#x, want 0xAABB", info.TitleID)
	}
	if info.Secret != 0x8899AABBCCDDEEFF {
		t.Errorf("Secret = %#x, want 0x8899AABBCCDDEEFF", info.Secret)
	}
	if info.ReconnectState != 0 {
		t.Errorf("ReconnectState = %d, want 0", info.ReconnectState)
	}
	if info.ClientMessageVersion != 5 {
		t.Errorf("ClientMessageVersion = %d, want 5", info.ClientMessageVersion)
	}
	if info.TimeoutOverride != 0xCAFEBABE {
		t.Errorf("TimeoutOverride = %#x, want 0xCAFEBABE", info.TimeoutOverride)
	}
	if info.ProfileID != 0x1122334455667788 {
		t.Errorf("ProfileID = %#x, want 0x1122334455667788", info.ProfileID)
	}
}

func TestParseLobbyConnectReconnectState2(t *testing.T) {
	body := []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // sessionID 0 + titleID 0
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // secret 0
		0x02,                   // reconnectState 2
		0x10, 0x00, 0x00, 0x00, // dataReceivedCount 16
		0x05, 0x00, // clientMessageVersion 5
		0x01, 0x00, 0x00, 0x00, // timeoutOverride 1
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // profileID
	}

	info, ok := ParseLobbyConnect(body)
	if !ok {
		t.Fatal("expected parse to succeed")
	}
	if info.ReconnectState != 2 {
		t.Errorf("ReconnectState = %d, want 2", info.ReconnectState)
	}
	if info.DataReceivedCount != 16 {
		t.Errorf("DataReceivedCount = %d, want 16", info.DataReceivedCount)
	}
	if info.ClientMessageVersion != 5 {
		t.Errorf("ClientMessageVersion = %d, want 5", info.ClientMessageVersion)
	}
	if info.TimeoutOverride != 1 {
		t.Errorf("TimeoutOverride = %d, want 1", info.TimeoutOverride)
	}
}

// The data-received count is present in the message ONLY when the reconnect state
// is 2, and the parser writes -1 for every other state. That default is
// load-bearing: the reconnect resend range starts at that field, so a zero
// default would ask for a range starting at frame 0 and resend the whole
// retained window on a connect that carries no resume point at all.
func TestParseLobbyConnectDefaultsTheDataCountToMinusOne(t *testing.T) {
	for _, state := range []byte{0, 1, 3, 4, 0xFF} {
		body := make([]byte, 23)
		body[16] = state
		info, ok := ParseLobbyConnect(body)
		if !ok {
			t.Fatalf("reconnect state %d: parse failed", state)
		}
		if got := int32(info.DataReceivedCount); got != -1 {
			t.Errorf("reconnect state %d: DataReceivedCount = %d, want -1", state, got)
		}
	}
	// State 2 carries the field and the parser must read what the body says,
	// including a negative count the client chose to send.
	body := make([]byte, 27)
	body[16] = 2
	binary.LittleEndian.PutUint32(body[17:21], 0xFFFFFFFF)
	info, ok := ParseLobbyConnect(body)
	if !ok {
		t.Fatal("state 2: parse failed")
	}
	if got := int32(info.DataReceivedCount); got != -1 {
		t.Errorf("state 2 with a body of 0xFFFFFFFF: DataReceivedCount = %d, want -1", got)
	}
	binary.LittleEndian.PutUint32(body[17:21], 1234)
	info, ok = ParseLobbyConnect(body)
	if !ok {
		t.Fatal("state 2: parse failed")
	}
	if info.DataReceivedCount != 1234 {
		t.Errorf("state 2: DataReceivedCount = %d, want 1234", info.DataReceivedCount)
	}
}

func TestParseLobbyConnectErrors(t *testing.T) {
	if _, ok := ParseLobbyConnect([]byte{0x01, 0x02, 0x03}); ok {
		t.Error("expected failure for short body")
	}

	minimal := make([]byte, 17)
	if info, ok := ParseLobbyConnect(minimal); !ok {
		t.Error("expected minimal 17-byte body to parse")
	} else if info.ClientMessageVersion != 0 || info.TimeoutOverride != 0 {
		t.Errorf("expected zeros without version fields, got version=%d timeoutOverride=%d", info.ClientMessageVersion, info.TimeoutOverride)
	}

	truncated := make([]byte, 19)
	truncated[16] = 1 // reconnectState, clientMessageVersion fits but timeoutOverride missing
	if _, ok := ParseLobbyConnect(truncated); ok {
		t.Error("expected failure when timeoutOverride is missing")
	}

	valid := []byte{
		0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0xBB, 0xAA,
		0xFF, 0xEE, 0xDD, 0xCC, 0xBB, 0xAA, 0x99, 0x88,
		0x00,
		0x05, 0x00,
		0xBE, 0xBA, 0xFE, 0xCA,
		0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11,
	}
	if _, ok := ParseLobbyConnect(append(valid, 0x00)); ok {
		t.Error("expected failure for trailing bytes")
	}
}

// TestParseLobbyConnectReconnectState2TooShort covers the guard inside the
// reconnectState==2 branch: a 17-20 byte body with state 2 is rejected.
func TestParseLobbyConnectReconnectState2TooShort(t *testing.T) {
	for _, n := range []int{17, 20} {
		body := make([]byte, n)
		body[16] = 2
		if _, ok := ParseLobbyConnect(body); ok {
			t.Errorf("len=%d with reconnectState 2 should be rejected", n)
		}
	}
}

// TestParseLobbyConnectVersion5TooShort covers the version>=5 guard: a body
// with enough room for the version and timeoutOverride but not the 8-byte
// profile ID is rejected.
func TestParseLobbyConnectVersion5TooShort(t *testing.T) {
	for _, n := range []int{19, 30} {
		body := make([]byte, n)
		body[17] = 5 // clientMessageVersion = 5
		if _, ok := ParseLobbyConnect(body); ok {
			t.Errorf("len=%d with version 5 should be rejected", n)
		}
	}
}

// TestParseLobbyConnectVersion4 covers the version<5 path: the message ends
// after the timeout override, so an exact-length body parses and a trailing
// byte is rejected.
func TestParseLobbyConnectVersion4(t *testing.T) {
	body := []byte{
		0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0xBB, 0xAA, // sessionID + titleID
		0xFF, 0xEE, 0xDD, 0xCC, 0xBB, 0xAA, 0x99, 0x88, // secret
		0x00,       // reconnectState 0
		0x04, 0x00, // clientMessageVersion 4
		0xBE, 0xBA, 0xFE, 0xCA, // timeoutOverride
	}
	info, ok := ParseLobbyConnect(body)
	if !ok {
		t.Fatal("expected exact-length version 4 body to parse")
	}
	if info.ClientMessageVersion != 4 {
		t.Errorf("ClientMessageVersion = %d, want 4", info.ClientMessageVersion)
	}
	if info.TimeoutOverride != 0xCAFEBABE {
		t.Errorf("TimeoutOverride = %#x, want 0xCAFEBABE", info.TimeoutOverride)
	}
	if info.ProfileID != 0 {
		t.Errorf("ProfileID = %#x, want 0 for version < 5", info.ProfileID)
	}

	if _, ok := ParseLobbyConnect(append(body, 0x00)); ok {
		t.Error("expected failure for trailing byte on version 4 body")
	}

	truncated := append([]byte{}, body[:22]...) // drop the last timeout byte
	if _, ok := ParseLobbyConnect(truncated); ok {
		t.Error("expected failure when timeoutOverride is incomplete")
	}
}

// TestBuildLobbyConnectRoundtrip pins the serialiser against the parser: the
// server's own lobby registration is built with BuildLobbyConnect, so a
// disagreement between the two would make the message unreadable by this
// codebase. Both must agree with the LobbyConnect field order and byte order.
func TestBuildLobbyConnectRoundtrip(t *testing.T) {
	cases := []struct {
		name                                       string
		titleID                                    uint16
		sessionID                                  int64
		secret                                     uint64
		state                                      byte
		count                                      int32
		version                                    uint16
		flags                                      uint32
		profileID                                  uint64
		wantLen                                    int
		wantDataCount                              int32
		wantReconnectState, wantCountParsedPresent bool
	}{
		{
			name: "server registration defaults", titleID: 0xFFFF, sessionID: -1,
			secret: 0xFFFFFFFFFFFFFFFF, version: 5,
			wantLen: 31,
		},
		{
			name: "game client, version 5", titleID: 24, sessionID: 1014582557,
			secret: 6155435443241933260, state: 1, version: 5, flags: 7, profileID: 0x1122334455667788,
			wantLen: 31,
		},
		{
			name: "reconnect carries the data count", titleID: 3, sessionID: 42,
			secret: 99, state: 2, count: -1, version: 5,
			wantLen: 35, wantDataCount: -1, wantReconnectState: true, wantCountParsedPresent: true,
		},
		{
			name: "version 4 has no trailing field", titleID: 1, sessionID: 2,
			secret: 3, version: 4, flags: 0xCAFEBABE,
			wantLen: 23,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := BuildLobbyConnect(tc.titleID, tc.sessionID, tc.secret, tc.state, tc.count, tc.version, tc.flags, tc.profileID)
			if len(body) != tc.wantLen {
				t.Fatalf("body length = %d, want %d", len(body), tc.wantLen)
			}

			info, ok := ParseLobbyConnect(body)
			if !ok {
				t.Fatal("ParseLobbyConnect rejected a body produced by BuildLobbyConnect")
			}
			if info.TitleID != tc.titleID {
				t.Errorf("TitleID = %#x, want %#x", info.TitleID, tc.titleID)
			}
			// The session ID is a 48-bit field, so -1 comes back as 0xFFFFFFFFFFFF
			// rather than as a negative number.
			if wantSession := uint64(tc.sessionID) & 0xFFFFFFFFFFFF; info.SessionID != wantSession {
				t.Errorf("SessionID = %#x, want %#x", info.SessionID, wantSession)
			}
			if info.Secret != tc.secret {
				t.Errorf("Secret = %#x, want %#x", info.Secret, tc.secret)
			}
			if info.ReconnectState != tc.state {
				t.Errorf("ReconnectState = %d, want %d", info.ReconnectState, tc.state)
			}
			if tc.wantCountParsedPresent && info.DataReceivedCount != uint32(tc.wantDataCount) {
				t.Errorf("DataReceivedCount = %d, want %d", info.DataReceivedCount, uint32(tc.wantDataCount))
			}
			if info.ClientMessageVersion != tc.version {
				t.Errorf("ClientMessageVersion = %d, want %d", info.ClientMessageVersion, tc.version)
			}
			if info.TimeoutOverride != tc.flags {
				t.Errorf("TimeoutOverride = %#x, want %#x", info.TimeoutOverride, tc.flags)
			}
			wantProfile := tc.profileID
			if tc.version < 5 {
				wantProfile = 0
			}
			if info.ProfileID != wantProfile {
				t.Errorf("ProfileID = %#x, want %#x", info.ProfileID, wantProfile)
			}
		})
	}
}

// TestBuildLobbyConnectMatchesCapturedBytes checks the byte layout against the
// 32-byte LobbyConnect seen on the wire in captures _13 (the 0x15 echo, 0x20
// payload): the body is everything after the type byte, so the packed u64 holds
// the low 48 bits of the session ID and the title ID in the high 16.
func TestBuildLobbyConnectMatchesCapturedBytes(t *testing.T) {
	body := BuildLobbyConnect(0xFFFF, -1, 0xFFFFFFFFFFFFFFFF, 0, 0, 5, 0, 0)
	if len(body)+1 != 32 {
		t.Fatalf("message length = %d, want 32 (1 type byte + 31 body bytes)", len(body)+1)
	}
	// sessionID -1 masked to 48 bits, with titleID 0xFFFF in the top half.
	want := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	if !bytes.Equal(body[:8], want) {
		t.Errorf("packed sessionID+titleID = %x, want %x", body[:8], want)
	}
	if !bytes.Equal(body[8:16], want) {
		t.Errorf("secret = %x, want %x", body[8:16], want)
	}
	if body[16] != 0 {
		t.Errorf("reconnectState = %d, want 0", body[16])
	}
	if !bytes.Equal(body[17:19], []byte{0x05, 0x00}) {
		t.Errorf("clientMessageVersion = %x, want 05 00 (little-endian 5)", body[17:19])
	}
}

// TestSerializeRoundtrips covers the RLink-framed messages: each Serialize must
// emit [type u8][len u16 BE][payload] with the length field intact (a previous
// implementation wrote the length then immediately overwrote it with append),
// and Deserialize (fed the body without the frame header) must reproduce the
// original fields.
func TestSerializeRoundtrips(t *testing.T) {
	gv := &GameVoiceMessage{VoiceData: []byte{0xDE, 0xAD, 0xBE, 0xEF}}
	ser := gv.Serialize()
	if ser[0] != MsgGameVoice {
		t.Errorf("GameVoice type = 0x%02x, want 0x%02x", ser[0], MsgGameVoice)
	}
	if ser[1] != 0x00 || ser[2] != 0x05 {
		t.Errorf("GameVoice length field = %02x %02x, want 00 05", ser[1], ser[2])
	}
	var gotGV GameVoiceMessage
	if err := gotGV.Deserialize(ser[3:]); err != nil {
		t.Fatalf("GameVoice deserialize: %v", err)
	}
	if !bytes.Equal(gotGV.VoiceData, gv.VoiceData) {
		t.Errorf("GameVoice roundtrip mismatch: %x", gotGV.VoiceData)
	}

	multi := &MultiGameDataMessage{
		SubMessages: []GameDataMessage{
			{InnerType: 0xFE, Payload: []byte{0x01, 0x02, 0x03}},
			{InnerType: 0xAB, Payload: []byte{0xAA, 0xBB}},
		},
	}
	ser = multi.Serialize()
	if ser[0] != MsgMultiGameData {
		t.Errorf("MultiGameData type = 0x%02x, want 0x%02x", ser[0], MsgMultiGameData)
	}
	var gotMulti MultiGameDataMessage
	if err := gotMulti.Deserialize(ser[3:]); err != nil {
		t.Fatalf("MultiGameData deserialize: %v", err)
	}
	if len(gotMulti.SubMessages) != 2 {
		t.Fatalf("MultiGameData count = %d, want 2", len(gotMulti.SubMessages))
	}
	if gotMulti.SubMessages[0].InnerType != multi.SubMessages[0].InnerType ||
		gotMulti.SubMessages[1].InnerType != multi.SubMessages[1].InnerType ||
		!bytes.Equal(gotMulti.SubMessages[0].Payload, multi.SubMessages[0].Payload) ||
		!bytes.Equal(gotMulti.SubMessages[1].Payload, multi.SubMessages[1].Payload) {
		t.Errorf("MultiGameData roundtrip mismatch: %+v", gotMulti.SubMessages)
	}

	nudge := &NudgeMessage{}
	ser = nudge.Serialize()
	if ser[0] != MsgNudge {
		t.Errorf("Nudge type = 0x%02x, want 0x%02x", ser[0], MsgNudge)
	}
	if ser[1] != 0x00 || ser[2] != 0x01 {
		t.Errorf("Nudge length field = %02x %02x, want 00 01", ser[1], ser[2])
	}
	if err := (&NudgeMessage{}).Deserialize(ser[3:]); err != nil {
		t.Fatalf("Nudge deserialize: %v", err)
	}

	rc := &RouteCheckMessage{SessionID: 0x1122334455667788, PeerID: 0x42, RelayCount: 0x1234}
	ser = rc.Serialize()
	if ser[0] != MsgRouteCheck {
		t.Errorf("RouteCheck type = 0x%02x, want 0x%02x", ser[0], MsgRouteCheck)
	}
	var gotRC RouteCheckMessage
	if err := gotRC.Deserialize(ser[3:]); err != nil {
		t.Fatalf("RouteCheck deserialize: %v", err)
	}
	if gotRC != *rc {
		t.Errorf("RouteCheck roundtrip mismatch: %+v", gotRC)
	}

	seek := &SeekMessage{TargetSessionID: 0xFEEDBEEF}
	ser = seek.Serialize()
	if ser[0] != MsgSeek {
		t.Errorf("Seek type = 0x%02x, want 0x%02x", ser[0], MsgSeek)
	}
	var gotSeek SeekMessage
	if err := gotSeek.Deserialize(ser[3:]); err != nil {
		t.Fatalf("Seek deserialize: %v", err)
	}
	if gotSeek != *seek {
		t.Errorf("Seek roundtrip mismatch: %+v", gotSeek)
	}

	sr := &SeekResponseMessage{SessionID: 0xCAFEBABE}
	ser = sr.Serialize()
	if ser[0] != MsgSeekResponse {
		t.Errorf("SeekResponse type = 0x%02x, want 0x%02x", ser[0], MsgSeekResponse)
	}
	var gotSR SeekResponseMessage
	if err := gotSR.Deserialize(ser[3:]); err != nil {
		t.Fatalf("SeekResponse deserialize: %v", err)
	}
	if gotSR != *sr {
		t.Errorf("SeekResponse roundtrip mismatch: %+v", gotSR)
	}

}

// BenchmarkSerialize measures the Serialize allocation count for the message
// types. Before C9 each method built the payload and then a second frame
// buffer (two allocations); after, each writes directly into a single frame.
func BenchmarkSerialize(b *testing.B) {
	gv := &GameVoiceMessage{VoiceData: make([]byte, 64)}
	rc := &RouteCheckMessage{SessionID: 12345, PeerID: 2, RelayCount: 7}
	nudge := &NudgeMessage{}
	seek := &SeekMessage{TargetSessionID: 99}
	sr := &SeekResponseMessage{SessionID: 99}
	gd := &GameDataMessage{InnerType: 0xFE, Payload: make([]byte, 64)}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = gv.Serialize()
		_ = rc.Serialize()
		_ = nudge.Serialize()
		_ = seek.Serialize()
		_ = sr.Serialize()
		_ = gd.Serialize()
	}
}
