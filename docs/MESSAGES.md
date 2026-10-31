# RLink::SessionMessage Catalogue

## General Description

BattleServer.exe uses more than **30 message types** organised in the `RLink::SessionMessage` hierarchy. Each message has a specific type ID and is transmitted between clients and server over bsPort or webSocketPort.

**Source**: `ClientLib\rlink\networking\` (network messages) and `cohoreplaytool\Src\` (replays)

---

## Class Hierarchy

```
RLink::SessionMessage (base)
  ├── Message (common methods)
  ├── PeerMessage
  │   ├── AddPeerMessage
  │   ├── RemovePeerMessage
  │   └── DropoutPeerMessage
  ├── GameDataMessage
  ├── GameDataAndChecksumMessage
  ├── MultiGameDataMessage
  ├── EndOfGameDataMessage
  ├── ProxyMessage
  ├── PingMessage
  ├── NudgeMessage
  ├── LobbyConnectMessage
  ├── LobbyRelayMessage
  ├── LobbyPingMessage
  ├── RequestJoinMessage
  ├── RequestJoinDeniedMessage
  ├── RouteCheckMessage
  ├── SeekMessage
  ├── SeekResponseMessage
  ├── IntegrityCheckMessage
  ├── GameVoiceMessage
  ├── RelayBroadcastMessage
  ├── SetObservableParamatersMessage
  ├── StartObserveRequestMessage
  ├── StartObserveResponseMessage
  ├── StopObserveMessage
  └── ActivateObserverMessage
      └── ActivateObserverResponseMessage
```

---

## Complete Message Catalogue

### Main Session Messages

| Type ID | Class Name | Hex ID | Description |
|---------|-----------------|--------|-------------|
| `0x09` | `GameDataMessage` | 9 | Game state data |
| `0x10` | `ProxyMessage` | 16 | Proxy between peers |
| `0x17` | `GameDataAndChecksumMessage` | 23 | Game data + MD5 checksum |
| `0x0F` | `PeerMessage` | 15 | Base for peer messages |
| `0x11` | `PingMessage` | 17 | General ping/keepalive |
| `0x14` | `LobbyPingMessage` | 20 | Lobby ping |
| `0x15` | `LobbyConnectMessage` | 21 | Initial connection with the lobby |
| `0x13` | `LobbyRelayMessage` | 19 | Data relay via the lobby |
| `0x21` | `RelayBroadcastMessage` | 33 | Presence broadcast |

### Peer Messages

| Type ID | Class Name | Hex ID | Description |
|---------|-----------------|--------|-------------|
| `0x01` | `RequestJoinMessage` | 1 | Request to join the game |
| `0x02` | `RequestJoinDeniedMessage` | 2 | Join request denied |
| `0x03` | `AddPeerMessage` | 3 | New peer notification |
| `0x04` | `RemovePeerMessage` | 4 | Peer removed notification |
| `0x05` | `DropoutPeerMessage` | 5 | Peer leaving the match |

### Observer Messages

| Type ID | Class Name | Hex ID | Description |
|---------|-----------------|--------|-------------|
| `0x18` | `StartObserveRequestMessage` | 24 | Request to observe |
| `0x19` | `StartObserveResponseMessage` | 25 | Response to an observe request |
| `0x1A` | `ActivateObserverMessage` | 26 | Activate observer mode |
| `0x1B` | `StopObserveMessage` | 27 | Deactivate observer mode |
| `0x1C` | `StartObserveRequestMessage (Vx)` | 28 | Vx variant of StartObserve (same handler as `0x18`) |
| `0x1D` | `SetObservableParamatersMessage` | 29 | Configure observable parameters |
| `0x1E` | `ActivateObserverResponseMessage` | 30 | Response to observer activation |

### Advanced Game Messages

| Type ID | Class Name | Hex ID | Description |
|---------|-----------------|--------|-------------|
| `0x06` | `IntegrityCheckMessage` | 6 | Integrity check |
| `0x07` | `GameVoiceMessage` | 7 | In-game voice data |
| `0x08` | `MultiGameDataMessage` | 8 | Several game data in one |
| `0x0A` | `EndOfGameDataMessage` | 10 | End of game data |
| `0x0B` | `NudgeMessage` | 11 | Notification/nudge message |
| `0x0C` | `RouteCheckMessage` | 12 | Route check |
| `0x0D` | `SeekMessage` | 13 | Match search |
| `0x0E` | `SeekResponseMessage` | 14 | Search response |
| `0x1F` | `LeaveSimulationMessage` | 31 | Leave the simulation |

---

## Details of the Main Messages

### RequestJoinMessage (0x01)

Request to join a game session.

| Field | Type | Description |
|-------|------|-------------|
| `peerID` | byte | ID of the requesting peer |
| `sessionID` | longlong | Target session ID |
| `protocolVersion` | short | Protocol version |

### AddPeerMessage (0x03)

Notification that a new peer has joined.

| Field | Type | Description |
|-------|------|-------------|
| `peerID` | byte | ID of the new peer |
| `peerInfo` | bytes | Peer information |

### GameDataMessage (0x09)

Game state data (the most frequent).

| Field | Type | Description |
|-------|------|-------------|
| `messageType` | byte | Subtype (0xFE = replay data) |
| `dataLength` | ushort | Payload length |
| `payload` | byte[] | Binary state data |

### ProxyMessage (0x10)

Data relay between peers through the server.

| Field | Type | Description |
|-------|------|-------------|
| `sourcePeerID` | byte | Origin peer ID |
| `destinationPeerIDs` | byte[] | Destination peer IDs |
| `data` | byte[] | Sub-packet data |

### LobbyConnectMessage (0x15)

Initial connection message with the lobby server.

| Field | Type | Description |
|-------|------|-------------|
| `titleID` | short | Game ID (default: 0xFFFF) |
| `sessionID` | longlong | Session ID (default: -1) |
| `secret` | ulonglong | Session secret (default: -1) |
| `dataReceivedCount` | int | Data counter (default: -1) |
| `protocolVersion` | short | Protocol version (default: 5) |
| `flags` | int | Connection flags |

### RelayBroadcastMessage (0x21)

Server presence broadcast (sent over UDP).

| Field | Type | Description |
|-------|------|-------------|
| `serverName` | string | Server name |
| `serverGUID` | string | Unique GUID (UUID format) |
| `publicIP` | string | Public IP |
| `publicPort` | ushort | Public TCP port |
| `outOfBandPort` | ushort | OOB TCP port |
| `relayBroadcastPort` | ushort | UDP broadcast port |
| `simulationPeriod` | ushort | Simulation period |

---

## Typical Message Flow

### Joining a Match

```
Client                                    Server
   │                                          │
   │─── RequestJoinMessage (0x01) ───────────→│
   │                                          │
   │←── AddPeerMessage (0x03) ────────────────│  (to all peers)
   │                                          │
   │─── GameDataMessage (0x09) ──────────────→│  (state send)
   │←── GameDataMessage (0x09) ───────────────│  (receiving from others)
   │                                          │
   │─── PingMessage (0x11) ──────────────────→│  (keepalive)
   │←── PingMessage (0x11) ───────────────────│  (response)
```

### Observer

```
Observer                                 Server
   │                                          │
   │─── StartObserveRequestMessage (0x18) ───→│
   │                                          │
   │←── StartObserveResponseMessage (0x19) ───│
   │                                          │
   │─── ActivateObserverMessage (0x1A) ──────→│
   │                                          │
   │←── ActivateObserverResponseMessage (0x1E)│
   │                                          │
   │←── GameDataMessage (0x09) ───────────────│  (receives data as an observer)
   │                                          │
   │─── StopObserveMessage (0x1B) ───────────→│
```

---

## Processing in Server::ProcessPacket

Incoming messages are processed in `FUN_1400de050` (Server::ProcessPacket):

```c
switch (messageType) {
    case 0x09:  // GameDataMessage
        // Extracts messageType, dataLength, payload
        // If messageType == 0xFE → Session::OnSyncMessageFromPeer
        break;
    case 0x10:  // ProxyMessage
        // Extracts sourcePeerID, destinationPeerIDs, data
        // Sends to PackReplayData
        break;
    case 0x17:  // GameDataAndChecksumMessage
        // Processes the checksum and forwards the sub-packet
        break;
    case 0x18:  // StartObserveRequestMessage
        // Records the observer request
        break;
    case 0x1A:  // ActivateObserverMessage
        // Activates observer mode
        break;
    case 0x1B:  // StopObserveMessage
        // Deactivates observer mode
        break;
    case 0x1D:  // SetObservableParamatersMessage
        // Configures the observable parameters
        break;
    case 0x1F:  // LeaveSimulationMessage
        // Processes leaving the simulation
        break;
    case 0x14:  // LobbyPingMessage
        // Keepalive response
        break;
    case 0x15:  // LobbyConnectMessage
        // Session registration/creation
        break;
}
```

---

## Source Files per Message

Each message has multiple vtable methods implemented in separate files:

```
src/
  ├── RLink__SessionMessage__GameDataMessage__vmethod_00_1400dca90.cpp
  ├── RLink__SessionMessage__GameDataMessage__vmethod_08_1400e0210.cpp
  ├── RLink__SessionMessage__GameDataMessage__vmethod_18_1400e0510.cpp
  ├── RLink__SessionMessage__GameDataMessage__vmethod_20_1400dd6d0.cpp
  ├── RLink__SessionMessage__GameDataMessage__vmethod_28_1400ddcb0.cpp
  └── ...
```

### Common Vtable Offset

| Offset | Typical Function |
|--------|----------------|
| `0x00` | Constructor / Destructor |
| `0x08` | Serialize |
| `0x10` | Deserialize |
| `0x18` | GetType / GetSize |
| `0x20` | Process |
| `0x28` | Copy/Clone |

---

## See Also

- [bsPort.md](bsPort.md) — Main game port
- [webSocketPort.md](webSocketPort.md) — WebSocket port
- [LOBBY-PROTOCOL.md](LOBBY-PROTOCOL.md) — Lobby messages
- [replays.md](replays.md) — Replay data
- [README.md](README.md) — General description
