# The presence beacon (`-relayBroadcastPort`)

> **What this is.** The UDP announcement the server makes of its own existence, so a client
> on the LAN can find it without being told an address. **It is original behaviour**: the
> flag is parsed by the original's own argument parser, the socket is built by the original's
> own `Server` constructor and the packet is sent by the original's own process loop. Every
> claim below cites `src/`.
>
> **Read this section first if you only remember one thing**: the beacon is **UDP, and the
> harness does not capture UDP**, so none of the behaviour here is covered by a capture and
> none of it is covered by the byte-for-byte comparison. What *is* covered is stated at the
> bottom, in [What the harness checks about the beacon](#what-the-harness-checks-about-the-beacon).

---

## 1. The configuration

| flag | original default | what it does in the original |
|---|---|---|
| `-relayBroadcastPort` | **9999**, one port | the destination UDP port of the broadcast |

The original reads the value into the **high word** of one `int` that also carries
`-outOfBandPort` in the low word (`main_1400d1310.cpp:685`), and that word is initialised to
`0x270F6A4C` (`main_1400d1310.cpp:216`) — `HIWORD` = `0x270F` = **9999** for the beacon,
`LOWORD` = `0x6A4C` = **27212** for the out-of-band port. The help string is at
`main_1400d1310.cpp:696-697`:

```
-relayBroadcastPort	Port to broadcast server presence, must also pass same arg/value to client
```

Two things follow from that, and both are easy to get wrong:

- **There is one port, not two, and there is no way to switch the beacon off.** The original
  has no `0` case, no second destination and no disable flag; the value goes straight into
  `SockSetBroadcast`. There is no branch on it anywhere.
- **The default is 9999 because of that same initialiser**, not because of anything in the
  beacon code. Nothing in `src/` contains the string `9999`; the number exists only as
  `0x270F` inside `0x270F6A4C`.

### The one flag that decides the beacon for our side

Ours is different, and the difference is load-bearing. `internal/config/config.go:31`
registers the flag with a **default of 0** and resolves it in three branches
(`internal/config/config.go:90-106`):

| invocation | destinations |
|---|---|
| flag **absent** | `9999` **and** `8888` |
| `-relaybroadcastport N`, `N != 0` | `N` only |
| `-relaybroadcastport 0` | none — **disabled** |

So on our side `0` *does* disable it, and the *absent* case means two ports. That is a
divergence and it is registered as one:
[`UNINTENTIONAL-DIVERGENCES.md`](UNINTENTIONAL-DIVERGENCES.md) 13. It is not settled there
yet, and it is not settled here: this file records what each side does.

---

## 2. The destination

The original uses **one** address: the limited broadcast `255.255.255.255`. It is the
**only** occurrence of that string in the entire decompilation, and it is in the `Server`
constructor (`sub_1400D6790_1400d6790.cpp:336-340`):

```c
if ( (**(unsigned __int8 (__fastcall **)(_QWORD, _QWORD, _QWORD, const char *, _QWORD))(**(_QWORD **)(a1 + 384) + 80LL))(
         *(_QWORD *)(a1 + 384),
         *(_QWORD *)(a1 + 416),
         "255.255.255.255",
         *(unsigned __int16 *)(a3 + 6)) == 0 )
```

`a3 + 6` is the `u16` `relayBroadcastPort` out of the command parameters, and `a1 + 416`
(`0x1A0`) is the broadcast socket. Both failures of that call and of the flag call below it
throw, with the message `"broadcast socket SockSetBroadcast port %u failed"`
(`:342-347`, `:355-360`). The parameters reach the constructor from
`main_1400d1310` through `sub_1400D2F40_1400d2f40.cpp:64,82`.

Ours does **not** use only `255.255.255.255`. `internal/net/udp.go:114-126` enumerates the
IPv4 broadcast address of every interface that is up **and** carries the broadcast flag, and
uses `255.255.255.255` only as the fallback for when no interface qualifies
(`internal/net/udp.go:68-72`). Same divergence register, same entry.

### What the constructor does, in order

| step | evidence | note |
|---|---|---|
| make sure the raw socket manager exists at `Server+0x180` | `sub_1400D6790:326-331` → `sub_1400D8B70_1400d8b70.cpp:21-48` | `sub_1400D8B70` is the **bsPort** binder (`sub_14009BE00(0, 0)` then bind/listen on `*a2`); the beacon shares its manager |
| `CreateSocket(1)` — UDP — into `Server+0x1A0` | `sub_1400D6790:332-335` | argument `1`, not `0`, is what makes it a datagram socket |
| `SockSetBroadcast(sock, "255.255.255.255", relayBroadcastPort)` | `sub_1400D6790:336-348` | throws on failure |
| `SockSetBroadcastFlag(sock, 1)` — `SO_BROADCAST` on | `sub_1400D6790:349-361` | throws on failure |
| `CoCreateGuid(Server+0x1A8)` then `UuidToStringA` into `Server+0x1B8` | `sub_1400D6790:362-376` | `UuidToStringA` failing throws too; the string is `rpc_string_freeA`'d |

The **guid is minted per process run**, at construction, and never changes afterwards. A
discovery client that has seen the server before still sees a different GUID after a restart.

---

## 3. The packet

### The type byte

`33`, decimal — `0x21`. It appears in four places and they agree: the factory stores it as a
`DWORD` (`sub_1400F28B0_1400f28b0.cpp:39`), the serialiser writes it
(`RLink__SessionMessage__RelayBroadcastMessage__vmethod_18_1400e06d0.cpp:35`), the
deserialiser reads it
(`RLink__SessionMessage__RelayBroadcastMessage__vmethod_20_1400dd890.cpp:23`) and the
process loop sets it (`sub_1400D81E0_1400d81e0.cpp:264`, `v54 = 33`). The class is named in
the binary: `RLink::SessionMessage::RelayBroadcastMessage`. It is also the type the
original's **dispatcher** knows (`sub_1400ED400_1400ed400.cpp:1226-1230`, `case 0x21u`).

### The layout

The class is `0x78` bytes (`sub_1400F28B0:19`) and holds **three** `std::string` members and
three `u16`. MSVC lays a `std::string` out as buffer/size/capacity, and the factory's
initialiser order puts them at:

| member | object | size | capacity |
|---|---|---|---|
| string A | `+0x10` | `+0x20` | `+0x28` |
| string B | `+0x30` | `+0x40` | `+0x48` |
| string C | `+0x50` | `+0x60` | `+0x68` |
| `u16` | `+0x70` | | |
| `u16` | `+0x72` | | |
| `u16` | `+0x74` | | |

Only **two** of the three strings go on the wire. Both the serialiser and the deserialiser
walk them in the same order, and they agree with each other because the deserialiser names
the offsets as plain byte arithmetic:

- `vmethod_18:37` takes `a1 + 10` (byte `+0x50`) and its length from `a1[12]` (byte `+0x60`);
- `vmethod_18:54` takes `a1 + 6` (byte `+0x30`) and its length from `a1[8]` (byte `+0x40`);
- `vmethod_20:25` reads into `(a1 + 80)` = `+0x50`, and `vmethod_20:27` into `(a1 + 48)` =
  `+0x30`;
- the ports go out at `+0x70`, then `+0x74`, then `+0x72` (`vmethod_18:78,89,100`;
  `vmethod_20:34,40,45`).

Which string is which follows from the process loop, which fills them in declaration order:
`Server::Process` assigns the **server name** first and the **guid** second
(`sub_1400D81E0:265-271` then `:272-278`), so the name is the `+0x30` member and the guid
the `+0x50` one. Since `+0x50` is written first, **the guid is the first string on the wire**
and the name the second. The `+0x10` member is constructed and never serialised — that is
where `publicIP` would live, and `publicIP` indeed never reaches the payload.

### The serialised form

```
byte    content
0       0x21
1-2     guidLength   u16, little-endian
3-N     serverGUID   guidLength bytes
N+1-N+2 nameLength   u16, little-endian
N+3-M   serverName   nameLength bytes
M+1-M+2 publicPort       u16, little-endian
M+3-M+4 webSocketPort    u16, little-endian
M+5-M+6 outOfBandPort    u16, little-endian
```

Every multi-byte value is **little-endian**, and that is not a choice: RLink writes the raw
in-memory bytes and never calls `htons`/`ntohs` on this path. The lengths are written
through the packet's own two-byte writer (`vmethod_18:42`, `:59`) and the port fields with
`__vtable + 64` as a bare 2-byte write (`vmethod_18:78,89,100`), which is what makes the
endianness raw memory order rather than network order.

A worked example, for `-name "My Server"` and the original's own defaults
(`-publicPort` from the command line, `-webSocketPort` absent, `-outOfBandPort` 27212), with
the guid `12345678-90ab-cdef-1234-567890abcdef`:

```
21                                   messageType
24 00                                guidLength = 36
31 32 33 34 35 36 37 38 2D 39 30 61  "12345678-90a"
62 2D 63 64 65 66 2D 31 32 33 34 35  "b-cdef-12345"
36 37 38 39 30 61 62 63 64 65 66      "67890abcdef"
09 00                                nameLength = 9
4D 79 20 53 65 72 76 65 72            "My Server"
<2 bytes>                            publicPort
00 00                                webSocketPort = 0
6A 6C                                outOfBandPort = 27212
```

Total for this configuration: `1 + 2 + 36 + 2 + 9 + 6` = **56 bytes**, variable with the name.

### The fields, and where they come from

| # | field | value in the payload | read from |
|---|---|---|---|
| 1 | message type | `0x21` | constant, `sub_1400D81E0:264` |
| 2 | guid | `UuidToStringA` of a fresh `CoCreateGuid` | `Server+0x1B8`, `sub_1400D81E0:272-278` |
| 3 | server name | `-name`, default `"My Server"` | `Server+0x280`, `sub_1400D81E0:265-271` |
| 4 | public port | `-publicPort` | `Server+0x232`, `sub_1400D81E0:279` |
| 5 | WebSocket port | `-webSocketPort` | `Server+0x234`, `sub_1400D81E0:280` |
| 6 | out-of-band port | `-outOfBandPort`, default 27212 | `Server+0x236`, `sub_1400D81E0:281` |

`Server+0x280` is the `-name` string: the parser initialises it to `"My Server"`
(`main_1400d1310.cpp:336`) and overwrites it in the `-name` arm (`:671-680`), and the
constructor copies the parameter into `Server+640` = `0x280`
(`sub_1400D6790:149`). The guid is the one built at `sub_1400D6790:362-376`.

**What is *not* in the payload**, and this is worth stating because the flag names invite the
opposite conclusion:

- **`publicIP` does not travel.** A discovery client takes the **source address of the UDP
  packet** it received. `publicIP` is parsed and stored and never serialised.
- **`relayBroadcastPort` does not travel.** It only chooses the destination port, which the
  receiver already knows, because it had to be listening there to receive anything.
- **`simulationPeriod` does not travel**, and neither does anything else.
- **There are no spare bytes.** The deserialiser finishes with `sub_1400EF360`
  (`vmethod_20:46`), which requires the packet to be exactly consumed, so a longer payload is
  rejected rather than tolerated.

---

## 4. The cadence: every 10 000 ms

`Server::Process` (`sub_1400D81E0_1400d81e0.cpp:251-252`):

```c
if ( v3 - *(_QWORD *)(a1 + 520) <= 0x2710 )
  return sub_1400DC440(v94);
```

- `a1 + 520` is `0x208`, the last-broadcast stamp.
- **`0x2710` is 10000**, decimal. Not `0x2711`, and not "about ten seconds": the test is
  `<= 10000`, so the first send happens on the first pass where **more** than 10 000 ms have
  gone by, and then every pass after that.
- The stamp is written after the send (`:336`), so it is the **send** time, not the schedule.

The whole body is: build the message, serialise it with `vmethod_18` (`:327`), **send once**
through the raw manager's vtable `+144` with the broadcast socket (`:328-332`), stamp
(`:336`). One call, one destination, one packet.

---

## 5. What ours does that the original does not

All three are on the wire — a client listening on UDP can tell every one of them — and all
three are registered together in
[`UNINTENTIONAL-DIVERGENCES.md`](UNINTENTIONAL-DIVERGENCES.md) 13, which is open. They are
repeated here because this is the document somebody reads when the beacon misbehaves.

| # | ours | the original | where |
|---|---|---|---|
| 1 | **two** default ports, `9999` and `8888` | **one**, `9999` | `internal/config/config.go:103-105` |
| 2 | every qualifying interface's broadcast address, `255.255.255.255` only as fallback | `255.255.255.255`, always | `internal/net/udp.go:114-126` |
| 3 | sends a **legacy `0x20`** message as well: alone on `8888`, and in addition to `0x21` on `9999` | **no `0x20` exists.** The message factory's switch has cases `0x1D, 0x1E, 0x1F, 0x21, 0x22` (`sub_1400ED400:1211-1234`) and a `default` that logs `"Invalid Packet type passed in to create message"` (`:1250-1258`). There is no case for `0x20`, so the original cannot produce or accept it | `internal/protocol/messages.go:170-182`, `BuildLegacyRelayBroadcastMessage` at `:156-164` |
| 4 | `-relaybroadcastport 0` disables the beacon | no disable; the value is the destination port whatever it is | `internal/config/config.go:98-101` |

Number 3 is the one to be careful about: the `0x20` payload
(`BuildLegacyRelayBroadcastMessage`) is `[0x20][guidLen][guid][nameLen][name][port][wsPort]`
— the same shape as `0x21` with the out-of-band port missing. It is modelled on the Age1-era
announcement. **Nothing in `src/` corresponds to it.** Its justification is outside this
repository and is not demonstrated here, so under rule 1 of
[`AGENTS.md`](AGENTS.md) it is an open question, not a documented behaviour of the original.

There is a fifth difference, and it is the only one that is **settled**: the WebSocket port
field always goes out as `0` on our side (`internal/server/broadcast.go:10`), whatever
`-webSocketPort` is. That one has a written reason and lives in
[`INTENTIONAL-DIVERGENCES.md`](INTENTIONAL-DIVERGENCES.md) 1.1. Note that `0` is also the
original's own default for that field, so with a default configuration the two sides agree
byte for byte.

---

## 6. What the harness checks about the beacon

**Nothing, and this is a limit of the instrumentation, not a pass.** The proxy records TCP
and WebSocket streams; the beacon is UDP
([`CAPTURES.md`](CAPTURES.md), "The ceiling of what the harness can ask"). So:

- the payload bytes are verified against `src/`, by the citations in this file, and by
  nothing else;
- no capture in `internal/sim/testdata/` contains a beacon, so a regression in
  `buildBroadcastMessage` would not fail a test;
- this is filed as a ceiling, not as a divergence:
  [`CAPTURES.md`](CAPTURES.md), "The ceiling of what the harness can ask".

If you want the beacon covered, what it needs is a UDP listener in the harness that records
the datagrams into the capture, and then the same byte-for-byte comparison the other ports
get. That is a change to `cmd/proxy` and to `internal/sim`, and it is listed in
[`PENDING.md`](PENDING.md).

---

## 7. Functions, by address

| address | file in `src/` | what it does |
|---|---|---|
| `0x1400d1310` | `main_1400d1310.cpp` | parses `-relayBroadcastPort` (`:682-686`); default at `:216`; help at `:696-697` |
| `0x1400d2f40` | `sub_1400D2F40_1400d2f40.cpp` | passes the command parameters to the constructor (`:64,82`) |
| `0x1400d6790` | `sub_1400D6790_1400d6790.cpp` | `Server` constructor: the UDP socket, `255.255.255.255`, the guid (`:326-376`) |
| `0x1400d8b70` | `sub_1400D8B70_1400d8b70.cpp` | the raw socket manager and the bsPort bind the beacon shares |
| `0x1400d81e0` | `sub_1400D81E0_1400d81e0.cpp` | `Server::Process`: the 10 000 ms gate, the send (`:251-336`) |
| `0x1400dd890` | `RLink__SessionMessage__RelayBroadcastMessage__vmethod_20_1400dd890.cpp` | deserialiser; the exact field order |
| `0x1400e06d0` | `RLink__SessionMessage__RelayBroadcastMessage__vmethod_18_1400e06d0.cpp` | serialiser |
| `0x1400ed400` | `sub_1400ED400_1400ed400.cpp` | the message factory; `case 0x21u` at `:1226`, and the `default` that has no `0x20` |
| `0x1400ef250` | | writes the type byte into the packet |
| `0x1400ef2e0` | | reads the type byte out of the packet |
| `0x1400ef360` | | the "packet exactly consumed" check |
| `0x1400f28b0` | `sub_1400F28B0_1400f28b0.cpp` | the factory: object size, member offsets, type 33 |
| `0x14009be00` | | `SocketManager` construction; `(0, 0)` is raw, `(1, ...)` is the WebSocket one |

---

## 8. See also

- [`MESSAGES.md`](MESSAGES.md) — the message table, where `0x21` sits among the others
- [`oob.md`](oob.md) — the out-of-band service whose port is field 6 of the beacon
- [`bsPort.md`](bsPort.md) — the port whose value is field 4, and the socket manager the
  beacon shares
- [`LOBBY-PROTOCOL.md`](LOBBY-PROTOCOL.md) — the heartbeat. **Not** the same thing as the
  beacon, and not to be confused with it: the beacon is one-way UDP to the LAN, the lobby
  connection is a TCP client/server conversation
- [`README.md`](README.md) — the index