# WebSocket Port (webSocketPort) — Documentation

## General Description

The WebSocket port (`-webSocketPort`) is a **TCP port with TLS** that allows client connections over WebSocket (wss://). It uses the **RFC 6455** protocol (the modern standard; earlier documentation claimed Hixie-76, which is **wrong** — verified in proxy captures) and it is implemented in `WebSocketManager.cpp`.

---

## Configuration Parameter

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `-webSocketPort` | uint | — | TCP port for WebSocket connections (wss://) |
| `-sslCert` | path | — | Path to the SSL certificate (required for wss://) |
| `-sslKey` | path | — | Path to the SSL private key (required for wss://) |

---

## Protocol

### TCP with TLS (WebSocket Secure)

```c
// FUN_1400d8c80 - WebSocketManager initialisation
// Create the WebSocketManager with SSL
WebSocketManager *wm = new WebSocketManager(1, hasSSL);  // 1 = WebSocket
```

**Implementation**: `websocketpp` with the **RFC 6455** protocol

### RFC 6455 Handshake

The modern standard:
- A single `Sec-WebSocket-Key` in the request
- The `Sec-WebSocket-Accept` response = SHA-1(Key + GUID `258EAFA5-E914-47DA-95CA-C5AB0DC85B11`)
- Binary frames masked from the client (mask **zero** in this game), unmasked from the server

```
Client                                          Server (webSocketPort)
   │                                                   │
   │─── GET / HTTP/1.1 ───────────────────────────────→│
   │     Upgrade: WebSocket                             │
   │     Connection: Upgrade                            │
   │     Sec-WebSocket-Key: <base64 16 bytes>           │
   │     Sec-WebSocket-Version: 13                      │
   │                                                   │
   │←── HTTP/1.1 101 Switching Protocols ──────────────│
   │     Upgrade: WebSocket                             │
   │     Connection: Upgrade                            │
   │     Sec-WebSocket-Accept: <base64 SHA-1>           │
   │                                                   │
   │═══ WebSocket connection established (TLS encrypted) ══│
```

*Evidence in captures: the handshake with `Sec-WebSocket-Key`/`Sec-WebSocket-Accept` (not MD5) and the `82 89 00 00 00 00` frames (binary opcode, masked with a zero mask) from the client. The captures and their results are in [CAPTURES.md](CAPTURES.md); the measurements of this handshake are in `SESSION-JOURNAL.md` (§`_4`, 2026-09-28).*

---

## Initialising the WebSocket Port

### Argument Parsing (FUN_1400d1310)

```c
// FUN_1400d1310.c, line 721:
iVar11 = _stricmp(pcVar22, "webSocketPort");
if (iVar11 == 0) {
    if (*_Str != '\0') {
        iVar11 = atoi(_Str);
        // Stores into the high word of local_348
        local_348 = CONCAT22((short)iVar11, (undefined2)local_348);
    }
}
```

### WebSocketManager Initialisation (FUN_1400d8c80)

```c
// FUN_1400d8c80 - WebSocket socket initialisation
void InitializeWebSocketSocket(longlong server, longlong params) {
    // 1. Get the WebSocket port
    short wsPort = *(short *)(params + 2);  // param_2[1]
    
    // 2. Check whether SSL is configured
    longlong sslCertPath = *(longlong *)(params + 0x158);
    
    // 3. Create the WebSocketManager (type 1, with SSL if applicable)
    WebSocketManager *wm = new WebSocketManager(1, sslCertPath != 0);
    server->webSocketManager = wm;  // Server+0x188
    
    // 4. Load the SSL certificate if it exists
    if (sslCertPath != 0) {
        char *certPath = (char *)(params + 0x148);
        char *keyPath = (char *)(params + 0x168);
        wm->LoadSSLCertificate(certPath, keyPath);
    }
    
    // 5. Create the TCP socket
    void *socket = wm->CreateSocket(0);  // TCP
    server->webSocketListenSocket = socket;  // Server+0x198
    
    // 6. Bind to the WebSocket port
    wm->Bind(socket, wsPort);
    
    // 7. Listen with backlog 1000
    wm->Listen(socket, 1000);
}
```

### Decompiled Code (FUN_1400d8c80)

```c
void FUN_1400d8c80(longlong param_1, longlong param_2) {
    // Get the WebSocket port
    uVar1 = *(undefined2 *)(param_2 + 2);  // param_2[1] = webSocketPort
    
    // Check SSL
    lVar2 = *(longlong *)(param_2 + 0x158);  // sslCert path
    
    // Create WebSocketManager(1, hasSSL)
    uVar5 = FUN_14009be00(1, lVar2 != 0);
    *(undefined8 *)(param_1 + 0x188) = uVar5;  // Server+0x188
    
    // Load SSL if it exists
    if (lVar2 != 0) {
        // Load the certificate
        cVar4 = FUN_14009bf20(
            *(undefined8 *)(param_1 + 0x188),
            sslCertPath, sslKeyPath
        );
        if (cVar4 == '\0') {
            _CxxThrowException("Unable to load SSL certificate.");
        }
    }
    
    // Create the TCP socket
    uVar5 = (**(code **)(**(longlong **)(param_1 + 0x188) + 8))(
        *(longlong **)(param_1 + 0x188), 0
    );
    *(undefined8 *)(param_1 + 0x198) = uVar5;  // Server+0x198
    
    // Bind to the port
    cVar4 = (**(code **)(**(longlong **)(param_1 + 0x188) + 0x38))(
        *(longlong **)(param_1 + 0x188), uVar5, uVar1
    );
    if (cVar4 == '\0') {
        // Error: "web SockBind port %u failed"
        _CxxThrowException(...);
    }
    
    // Listen with backlog 1000
    cVar4 = (**(code **)(**(longlong **)(param_1 + 0x188) + 0x40))(
        *(longlong **)(param_1 + 0x188), 
        *(undefined8 *)(param_1 + 0x198), 
        1000
    );
    if (cVar4 == '\0') {
        // Error: "web SockListen port %u failed"
        _CxxThrowException(...);
    }
}
```

---

## WebSocketManager (FUN_1400a06e0)

### Structure

```c
// WebSocketManager (size: 0xC0 bytes)
struct WebSocketManager {
    void *vftable;                  // +0x00
    void *lock;                     // +0x08
    void *cond;                     // +0x10
    char serverName[16];            // +0x18
    size_t serverNameSize;          // +0x28
    size_t serverNameCapacity;      // +0x30
    void *connectionList;           // +0x38
    void *messageHandlers[4];       // +0x48
    void *sslContext;               // +0x68
    void *listenSocket;             // +0x70
    char buffer[1000];              // +0x78
    // ...
};
```

### Connection Creation (FUN_1400a1330)

```c
// WebSocketManager::SockAccept - Accepts a new WebSocket connection
longlong *WebSocketManager::SockAccept(
    WebSocketManager *this,
    void *listenSocket,
    void **outSocket
) {
    // 1. Accept the TCP connection
    int result = SockAccept(&this->sslContext, *(void **)(listenSocket + 8), outSocket);
    
    if (result == 0 && *outSocket != NULL) {
        // 2. Create the WebSocketInfo (size: 0x1B0 bytes)
        WebSocketInfo *wsInfo = new WebSocketInfo();
        
        // 3. Configure the WebSocket connection
        wsInfo->wsManager = &this->sslContext;
        wsInfo->socket = *outSocket;
        wsInfo->state = 3;  // Connecting
        
        // 4. Start the SSL Accept thread if applicable
        if (this->sslContext) {
            wsInfo->sslAcceptThread = new SSLAcceptThread();
            wsInfo->sslAcceptThread->Start();
        }
        
        return wsInfo;
    }
    return NULL;
}
```

---

## WebSocket Message Processing

### Receiving (FUN_1400a35a0)

```c
// WebSocketManager::SockPacketReceive - Receives data over WebSocket
ulonglong WebSocketManager::SockPacketReceive(
    WebSocketInfo *wsInfo,
    void *socket,
    char *data,
    int size
) {
    // 1. The RFC 6455 frame has already been unwrapped by websocketpp
    //    (binary: 0x82 [len] [data], masked from the client)
    char *payload = data;  // data = payload already unmasked
    int payloadSize = size;

    // 2. Copy the payload
    memcpy(wsInfo->buffer, payload, payloadSize);
    wsInfo->bufferSize = payloadSize;

    // 3. Process it as an RLink message
    //    (the same binary protocol as bsPort)
    ProcessMessage(wsInfo, payload, payloadSize);

    return payloadSize;
}
```

> **Note**: the game uses a **zero mask** in its frames, so the data arrives unaltered. In real captures: `82 89 00 00 00 00` = binary frame (FIN+opcode 0x2), 9 bytes of payload, 4-byte mask all zeros.

### Sending (FUN_1400a24d0)

```c
// WebSocketManager::SockPacketSend - Sends data over WebSocket
void WebSocketManager::SockPacketSend(
    WebSocketInfo *wsInfo,
    char *data,
    int size
) {
// 1. Build the RFC 6455 frame (binary, unmasked)
//    Format: 0x82 [len 7/16/64 bits] [data]
//    (websocketpp handles the encoding on send)
    char *frame = data;
    
    // 2. Send over the socket (with TLS if applicable)
    if (wsInfo->sslContext) {
        // SockSend with SSL
        wsInfo->wsManager->SockSendSSL(wsInfo->socket, frame, size);
    } else {
        // SockSend without SSL
        wsInfo->wsManager->SockSend(wsInfo->socket, frame, size);
    }
}
```

---

## WebSocket Protocol Messages

### The same messages as bsPort

The WebSocket protocol carries **the same binary messages** as the bsPort port:

| Type ID | Message | Description |
|---------|---------|-------------|
| `0x09` | `GameDataMessage` | Game state data |
| `0x10` | `ProxyMessage` | Proxy between peers |
| `0x17` | `GameDataAndChecksumMessage` | Game data + checksum |
| `0x18` | `StartObserveRequestMessage` | Request to observe |
| `0x1A` | `ActivateObserverMessage` | Activate observer mode |
| `0x1B` | `StopObserveMessage` | Deactivate observer mode |
| `0x1D` | `SetObservableParametersMessage` | Configure parameters |
| `0x1F` | `LeaveSimulationMessage` | Leave the simulation |
| `0x13` | `LobbyRelayMessage` | Relay via the lobby |
| `0x14` | `LobbyPingMessage` | Lobby ping |
| `0x15` | `LobbyConnectMessage` | Initial lobby connection |

### Difference: Packaging

```
bsPort (direct TCP):
  [RLink::Packet] → [TCP Socket] → [Network]

webSocketPort (WebSocket):
  [RLink::Packet] → [RFC 6455 frame: 0x82 + len + data] → [TLS] → [TCP Socket] → [Network]
```

---

## Server Offsets

| Offset | Field | Type | Description |
|--------|-------|------|-------------|
| `+0x188` | WebSocketManager | `ptr` | WebSocket manager |
| `+0x198` | WebSocketListenSocket | `ptr` | Listening TCP socket |
| `+0x232` | publicPort | `ushort` | Public port (can be bsPort or webSocketPort) |

---

## Associated Metrics

| Metric | Description |
|---------|-------------|
| `WsConnections` | Active WebSocket connections |
| `WsConnectionErrors` | WebSocket connection errors |
| `SslAcceptCpuUsage` | CPU used by the SSL handshake |

---

## Lifecycle of a WebSocket Connection

```
Client                                    Server (webSocketPort)
   │                                              │
   │──────── SYN ────────────────────────────────→│
   │                                              │
   │←──────── SYN-ACK ────────────────────────────│
   │                                              │
   │──────── ACK ────────────────────────────────→│  ← TCP connection established
   │                                              │
   │──────── Client Hello (TLS) ─────────────────→│
   │                                              │
   │←──────── Server Hello + Certificate ─────────│
   │                                              │
 │──────── Key Exchange ───────────────────────→│  ← TLS established
   │                                              │
   │──────── GET / HTTP/1.1 ─────────────────────→│
   │          Upgrade: WebSocket                   │
   │          Sec-WebSocket-Key                    │
   │                                              │
   │←──────── HTTP/1.1 101 ──────────────────────│  ← WebSocket established
   │          Sec-WebSocket-Accept (SHA-1)         │
   │                                              │
   │═══ [0x82 + len + GameDataMessage] ══════════│  ← Game data
   │                                              │
   │═══ [0x82 + len + ADDPEER] ═════════════════─│
   │                                              │
   │═══ FIN ═════════════════════════════════════│  ← Disconnection
```

---

## Involved Functions

### WebSocketManager

| Function | Address | Purpose |
|---------|-----------|---------|
| `FUN_1400a06e0` | `0x1400a06e0` | WebSocketManager constructor |
| `FUN_1400a1330` | `0x1400a1330` | SockAccept (accept a connection) |
| `FUN_1400a35a0` | `0x1400a35a0` | SockPacketReceive (receive data) |
| `FUN_1400a24d0` | `0x1400a24d0` | SockPacketSend (send data) |
| `FUN_1400a5d80` | `0x1400a5d80` | SockSendPing (send a ping) |
| `FUN_1400a4030` | `0x1400a4030` | SockSend (send generic data) |
| `FUN_1400a43a0` | `0x1400a43a0` | SockReceive (receive generic data) |

### Initialisation

| Function | Address | Purpose |
|---------|-----------|---------|
| `FUN_1400d1310` | `0x1400d1310` | Parsing of the `-webSocketPort` argument |
| `FUN_1400d8c80` | `0x1400d8c80` | WebSocketManager initialisation |
| `FUN_1400d6790` | `0x1400d6790` | Configuration in the Server constructor |

---

## Summary

| Characteristic | Detail |
|----------------|---------|
| **Protocol** | TCP + TLS + WebSocket (RFC 6455) |
| **Purpose** | Game connections via the web browser |
| **SSL** | Required (wss://) |
| **Handshake** | RFC 6455 (Sec-WebSocket-Key → Accept, SHA-1) |
| **Frame format** | 0x82 [len] [payload] (masked from the client, unmasked from the server) |
| **Messages** | The same as bsPort (GameData, Proxy, etc.) |
| **Socket backlog** | 1000 |
| **Source** | `ClientLib/rlink/networking/WebSocketManager.cpp` |

---

## See Also

- [bsPort.md](bsPort.md) — Main game port
- [WEBSOCKET-BSPORT-COMPARISON.md](WEBSOCKET-BSPORT-COMPARISON.md) — Detailed comparison
- [oob.md](oob.md) — Out-of-Band ping service
- [relay.md](relay.md) — Presence broadcast
- [README.md](README.md) — General project description
