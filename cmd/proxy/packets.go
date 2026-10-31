package main

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/luskaner/ageBattleServer/internal"
	"github.com/luskaner/ageBattleServer/internal/protocol"
)

// PacketInfo contains the decoded information of a packet
type PacketInfo struct {
	Timestamp string
	SrcHost   string
	SrcPort   string
	DestHost  string
	DestPort  string
	MsgType   byte
	TypeName  string
	BodyHex   string
	BodyLen   int
	DataB64   string
	B64Len    int
}

// decodePacket decodes a complete captured frame and builds its CSV row.
func decodePacket(data []byte, srcHost, srcPort, destHost, destPort string) *PacketInfo {
	// Raw TCP/WebSocket payload is not base64: store it as base64 for the CSV
	b64Str := base64.StdEncoding.EncodeToString(data)

	// Decode the battle protocol message: type byte + body
	msgType, body, _ := internal.ParseMessage(data)
	// protocol.MessageTypeName is the single source of truth for the type names
	// (it also knows 0x20/0x21 RelayBroadcast and 0x22, which the proxy's own
	// copy of the table mislabelled).
	typeName := protocol.MessageTypeName(msgType)

	bodyHex := ""
	bodyLen := 0
	if body != nil && len(body) > 0 {
		bodyHex = fmt.Sprintf("%x", body)
		bodyLen = len(body)
	}

	return &PacketInfo{
		Timestamp: time.Now().Format(time.RFC3339),
		SrcHost:   srcHost,
		SrcPort:   srcPort,
		DestHost:  destHost,
		DestPort:  destPort,
		MsgType:   msgType,
		TypeName:  typeName,
		BodyHex:   bodyHex,
		BodyLen:   bodyLen,
		DataB64:   b64Str,
		B64Len:    len(data),
	}
}

func (p *PacketInfo) row() []string {
	return []string{
		p.Timestamp,
		p.SrcHost,
		p.SrcPort,
		p.DestHost,
		p.DestPort,
		fmt.Sprintf("0x%02X", p.MsgType),
		p.TypeName,
		fmt.Sprintf("%d", p.BodyLen),
		p.BodyHex,
		p.DataB64,
	}
}

var csvHeader = []string{
	"timestamp", "src_host", "src_port", "dest_host", "dest_port",
	"msg_type", "type_name", "body_len", "body_hex", "data_b64",
}
