# OOB Port (Out-of-Band) — Documentation

## General Description

The OOB port (`-outOfBandPort`) is a **TCP** port that runs a dedicated ping server (**PingServer** / **OutOfBandServer**). It lets clients query the server's latency and status outside the main game channel.

---

## Configuration Parameter

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `-outOfBandPort` | uint | — | TCP port for the OOB ping service |

---

## Protocol

### TCP (Transmission Control Protocol)

```c
// The PingServer listens on TCP, not UDP
// FUN_1400d3f60 - PingServer initialisation
lVar2 = (**(code **)(**(longlong **)(param_1 + 0x2c) + 8))(*(longlong **)(param_1 + 0x2c), 0);
// CreateSocket(0) = TCP socket
```

**Source**: the address-based decompilation in `src/` (the original source name
`OutOfBandServer.cpp` is not present as a local file).

---

## PingServer Architecture

### Server Creation (FUN_1400d3f60)

```
PingServer::PingServer(short port)
  │
  ├→ SocketManager *sm = new SocketManager(0, 0);  // TCP manager
  ├→ void *listenSocket = sm->CreateSocket(0);       // TCP socket
  ├→ sm->Bind(listenSocket, port);                   // Bind to outOfBandPort
  ├→ sm->Listen(listenSocket, 100);                  // Listen with backlog 100
  │
  └→ _beginthreadex(..., PingServer::ThreadFunction, this, ...);
      └→ Starts a dedicated thread
```

### PingServer Structure

```
PingServer
  +0x18   serverName        string      Server name
  +0x28   listenSocket      void*       Listening TCP socket
  +0x2c   socketManager     SocketManager*  Socket manager
  +0x30   threadHandle      uintptr_t   Thread handle
  +0x34   threadID          int         Thread ID
  +0x48   stopFlag          byte        Stop flag
  +0x50   connectionSocket  void*       Active connection socket
  +0x58   sendSocketManager SocketManager*  Manager for sending
```

### Thread Function (FUN_1400d4200)

```c
// PingServer::ThreadFunction - Main loop
void PingServer::ThreadFunction(PingServer *server) {
    while (!server->stopFlag) {
        // Receive connections and process packets
        int bytesReceived = server->socketManager->RecvAll(
            server->connectionSocket, buffer, bufferSize
        );
        
        // Process each connection
        for (auto& socket : connectedSockets) {
            char result = server->socketManager->ProcessPacket(
                socket, buffer
            );
            if (result) {
                // Send the ping response
                PingServer::ProcessPacket(server, socket, packet);
            }
        }
        
        // Update CPU metrics every 5000ms
        if (elapsed > 5000) {
            UpdateCpuUsageMetrics();
        }
    }
}
```

---

## OOB Service Messages

### Incoming Messages (Client → Server)

The server accepts two kinds of ping message:

| Type ID | Message | Structure | Size |
|---------|---------|------------|--------|
| `0x11` | `PingMessage` | `RLink::SessionMessage::PingMessage` | 0x40 bytes |
| `0x14` | `LobbyPingMessage` | `RLink::SessionMessage::LobbyPingMessage` | 0x38 bytes |

### PingMessage Structure (0x11)

Total size: 0x40 bytes (64 bytes)

```
Offset      Field               Size      Type
─────────────────────────────────────────────────────────
+0x00       vftable             8 bytes     pointer
+0x08       messageType         4 bytes     uint32  (0x11 = 17)
+0x10       data                48 bytes    byte[]  (client data)
```

### LobbyPingMessage Structure (0x14)

Total size: 0x38 bytes (56 bytes)

```
Offset      Field               Size      Type
─────────────────────────────────────────────────────────
+0x00       vftable             8 bytes     pointer
+0x08       messageType         4 bytes     uint32  (0x14 = 20)
+0x10       data                40 bytes    byte[]  (client data)
```

---

## Packet Processing (FUN_1400d4460)

### Source code: `OutOfBandServer.cpp` (original source name)

```c
// PingServer::ProcessPacket
bool PingServer::ProcessPacket(
    PingServer *server,
    void *socket,
    RLink::Packet *packet
) {
    // 1. Reset the packet buffer
    packet->Reset(0xFFFF);
    
    // 2. Receive data from the socket
    ushort dataSize = packet->GetSize();
    void *data = packet->GetData(0);
    int received = server->socketManager->RecvSome(
        socket, data, dataSize
    );
    
    if (received < 1) return false;
    
    // 3. Adjust the packet size
    packet->SetSize(received);
    
    // 4. Check the message type
    uint msgType;
    packet->ReadBytes(&msgType, 4);
    
    if (msgType != 0x14) {  // Not a LobbyPingMessage
        return true;  // Ignore silently
    }
    
    // 5. Send the response (echo/pong)
    char result = server->socketManager->Send(
        socket, data, received
    );
    
    if (!result) {
        Log("PingServer::ProcessPacket: SockSend failed; %s");
    }
    
    return result;
}
```

---

## Server Response

### Behaviour: **Echo (reflection)**

The PingServer implements an **echo** behaviour:
1. It receives data from the client
2. It sends the same data back

```
Client                                    Server (outOfBandPort)
   │                                              │
   │─── PingMessage (0x11) or LobbyPingMessage ───→│
   │    [client_timestamp + payload]                │
   │                                              │
   │←── Echo (the same data received) ────────────│
   │                                              │
   │  [Computes latency: now - client_timestamp]    │
```

### Typical Request Data

The client typically sends:
- Send timestamp (for the latency calculation)
- Session identifier (optional)
- Diagnostic payload (optional)

### Response Data

The server responds with:
- **The same bytes received** (a full echo)
- No modification or extra processing

---

## OOB Service Metrics

### Performance Counters

| Metric | Variable | Description |
|---------|----------|-------------|
| `PingConnections` | Counter | Accepted active ping connections |
| `PingConnectionErrors` | Counter | Errors on ping connections |
| `OobServiceCpuUsage` | Percentage | CPU used by the PingServer thread |

### Metrics Update (FUN_1400d4200)

```c
// Update metrics every 5000ms
if (elapsedMilliseconds > 5000) {
    int cpuUsage = CalculateCpuUsage();
    UpdateMetric(OobServiceCpuUsage, cpuUsage);
}
```

---

## PingServer Lifecycle

### Initialisation (FUN_1400d3f60)

```c
void StartPingServer(short port) {
    PingServer *server = new PingServer();
    
    // 1. Create the TCP SocketManager
    server->socketManager = new SocketManager(0, 0);
    
    // 2. Load the SSL certificate (if configured)
    if (sslCertPath != NULL) {
        server->socketManager->LoadSSLCertificate(
            sslCertPath, sslKeyPath
        );
    }
    
    // 3. Create the listening TCP socket
    server->listenSocket = server->socketManager->CreateSocket(0);
    
    // 4. Bind to the OOB port
    server->socketManager->Bind(server->listenSocket, port);
    
    // 5. Listen with backlog 100
    server->socketManager->Listen(server->listenSocket, 100);
    
    // 6. Start the thread
    server->threadHandle = _beginthreadex(
        NULL, 0, 
        PingServer::ThreadFunction, 
        server, 0, 
        &server->threadID
    );
}
```

### Shutdown (FUN_1400d40e0 / FUN_1400d3e10)

```c
void StopPingServer(PingServer *server) {
    // 1. Signal the stop
    LOCK();
    server->stopFlag = 1;
    UNLOCK();
    
    // 2. Close the listening socket
    if (server->listenSocket != NULL) {
        server->socketManager->Close(server->listenSocket, "PingServer::Stop");
        server->listenSocket = NULL;
    }
    
    // 3. Wait for the thread to finish
    if (server->threadID != 0) {
        _Thrd_join(&server->threadHandle, NULL);
        server->threadHandle = 0;
        server->threadID = 0;
    }
    
    // 4. Free the SocketManager
    delete server->socketManager;
}
```

---

## Relation to `-relayBroadcastPort`

The `outOfBandPort` is included as **data** inside the broadcast message sent by `-relayBroadcastPort`:

```c
// FUN_1400d81e0 - Server::Process
local_1c4 = *(ushort *)(param_1 + 0x234);  // Loads outOfBandPort
// Included in RelayBroadcastMessage.outOfBandPort (offset +0x72)
```

That lets clients that receive the broadcast know which port to send pings to.

---

## Server Offsets

| Offset | Field | Type | Description |
|--------|-------|------|-------------|
| `+0x234` | outOfBandPort | `ushort` | PingServer's TCP port |
| `+0x236` | relayBroadcastPort | `ushort` | UDP broadcast port |

---

## Usage Example

### Server Configuration

```bash
BattleServer.exe ^
  -name "My Server" ^
  -bsPort 27012 ^
  -outOfBandPort 27013 ^
  -relayBroadcastPort 27014
```

### Client Flow

```
1. Client receives the UDP broadcast on port 27014:
   "My Server" + publicIP + publicPort + outOfBandPort(27013) + ...

2. Client opens a TCP connection to outOfBandPort (27013)

3. Client sends a PingMessage with timestamp T1

4. Server responds with an echo (the same data)

5. Client receives the response at T2
   Latency = T2 - T1
```

---

## Summary

| Characteristic | Detail |
|----------------|---------|
| **Protocol** | TCP |
| **Type** | Ping server (echo) |
| **Messages** | PingMessage (0x11), LobbyPingMessage (0x14) |
| **Response** | Echo (reflection of the received data) |
| **Dedicated thread** | Yes (_beginthreadex) |
| **Metrics** | PingConnections, PingConnectionErrors, OobServiceCpuUsage |
| **Source** | Address-based decompilation under `src/` (the original source name is `OutOfBandServer.cpp`) |

---

## See Also

- [relay.md](relay.md) — Server presence broadcast
- [README.md](README.md) — General project description
- [LOBBY-PROTOCOL.md](LOBBY-PROTOCOL.md) — Heartbeat system
