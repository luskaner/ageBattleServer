# Captures

> **What this file covers.** How a capture is recorded (the procedure and the deployment),
> the matrix of what is left to record, and **the results, capture by capture, of what has
> already been verified**. The second half of this file is the reference: what does not
> appear in the results table is not verified.
>
> **Status.** Current, and it is **the only** one that describes the current layout of a capture. The
> old format (`main_log_*.csv`) is not described anywhere live: the proxy writes
> `main.log` / `secondary.log` plus `metadata.json` and `*.meta.json`, and the validation is
> done by the `internal/sim` harness.
>
> **What this file is NOT.** It is not the list of divergences: those are in
> [`INTENTIONAL-DIVERGENCES.md`](INTENTIONAL-DIVERGENCES.md) and
> [`UNINTENTIONAL-DIVERGENCES.md`](UNINTENTIONAL-DIVERGENCES.md). It is not the count
> of what is missing: that is in [`PENDING.md`](PENDING.md). And it is not the journal of
> how we got here: that is in [`SESSION-JOURNAL.md`](SESSION-JOURNAL.md), which is
> history.

Each capture is a folder in `./captures/<date>_<commit>/`, created by the proxy
at start-up. The names carry no number and mean nothing: who is who is
said by `metadata.json`, decided by reading the binary, not by the file name.

```
captures/2026-09-28_040213_56c1309/
    main.log          T1, client <-> primary
    secondary.log     T2, client <-> secondary
    proxy.log         the proxy
    metadata.json     commit, dirty, and who each side is
    main.meta.json    { "file": "main.log", "writer": "original" }
    secondary.meta.json
    COMMIT.txt        the same provenance, in plain text
```

`writer` is a **tri-state** (`ours` / `original` / `unknown`) and not a boolean on
purpose: "it is the original" and "it could not be determined" are different facts, and
with a boolean an unidentified capture would look decided. If any side comes out
`unknown`, that capture **is not usable for comparing**.

## Before recording

```
.\scripts\deploy.ps1            # main = ORIGINAL, secondary = OURS
.\scripts\deploy.ps1 -Swap      # main = OURS, secondary = ORIGINAL
```

The script decides by content, copies safely and **verifies what it leaves on
disk**. If the table does not say what you expected, record nothing.

## Golden rule: one capture, one thing

The mistake of _30 cost three hypotheses because one capture mixed two questions. If a
capture has to answer more than one, it is badly posed.

Always start with the **original at the front** (`deploy.ps1` without `-Swap`): that is
the reference. You only move to `-Swap` once you know what you want to compare.

## Matrix

Tick what you record. `main` = original unless stated otherwise.

### By game

| # | game | binaries | what is compared | status |
|--|--|--|--|--|
| 1 | AoE2DE | original vs ours | full match against the AI, no observer | [ ] |
| 2 | AoE2DE | original vs ours | **observer**: the client asks to observe and receives the replay | [ ] |
| 3 | AoE3DE | original vs ours | full match | [ ] |
| 4 | AoE2DE | **ours at the front** | the failure that stays open (buildings of _27) | [ ] |

### By functionality

| # | functionality | how it is provoked | what is looked at | status |
|--|--|--|--|--|
| 5 | observation | second instance, "spectator" | `0x19`, whether `0x1E` arrives, and the pages | [x] 2026-09-29_012528, ours at the front. Uncovered the frame store: SESSION-JOURNAL.md 24 |
| 6 | reconnection | kill and reopen the match halfway | the `secret` and that it keeps the slot (6884fa3) | [ ] |
| 7 | forced desync | alter one byte of the stream | the `0xFB` notice and that reconciliation consumes it (14c1eb2) | [ ] |
| 8 | `0x14` echo | long match | 2 of 26 in _26, never isolated | [ ] |
| 9 | match from the lobby | create the match from the lobby | whether the client ever sends `RequestJoin` | [ ] |

## After recording

```
python scripts\export_vectors.py .\captures --out src_go\internal\server\testdata\observer_vectors.json
```

Compare the table with what was expected **before** calling anything good. If `expected_0x19`
comes out as `None` it means the secondary could not be read: record it and investigate, do
not assume.

## What is already known, so that it is not repeated

Extracted from the history before it was deleted, and kept here and in
`src_go/internal/server/testdata/observer_vectors.json`:

| capture | side | `0x19` | `0x1E` | pages |
|--|--|--|--|--|
| _26 | original | **0** | **yes** | **7343** |
| _28 | ours | 3 | no | 280 |
| _29 | ours | 3 | no | 379 |
| _30 | ours | 3 | no | 390 |

The only `0` is the only one followed by `0x1E` and the only one with thousands of pages. A
`3` makes the client never reach activation and then say there is no data from the
start. The golden test `TestStartObserveMatchesCapturedClientRequest` protects this
with _26's real bytes.

---

# RESULTS, CAPTURE BY CAPTURE

This is **the reference of what has been verified**. Each row is a capture
committed in `src_go/internal/sim/testdata/`, with a generated test that replays it.
What **does not** appear here has not been verified.

`MUST-MATCH` is the count of divergences against the original in the three
verdicts the client depends on: the byte of each `0x19`, whether `0x1E` arrives,
and the observer's replay stream with its order and its starting point. **The number
of frames is NOT must-match**: the `0x0A`s are built per simulation tick, so
a compressed replay groups differently without being wrong. "Informative" is counted and
printed with both sides' frames side by side, but does not fail.

## The five real committed captures

`clock` is what the test **costs**, measured: the span of the log plus the settle. The
settle is **not** a constant and has not been since 2026-10-05. It has two stages:
`settleFor` computes the base settle from the observer step's own period,
`interval+2` ticks, with a 400 ms floor and a 3 s cap; `settleForTruth` can then widen
that base to reach the end of the whole second in which the recording stops, with its
own extra cap (`internal/sim/replay.go`). At the defaults the observer-step base is
**1.25 s**, which is why each clock below includes a second and a quarter before any
truth-based widening. The `informative` column is the one from the **latest run's
report** (`go test ./internal/sim/ -count=1 -v`, **772.6 s** for the whole package on
2026-10-05, MUST-MATCH **0** in all seven); it is not a constant: it grows
with the number of informative rules the harness has **and with how much of the run the
settle covers**, so a higher number than the one here is not by itself a change of the
server. **Re-measured by the 45 on the same date and unchanged, which is the point of the
column**: 12, 28, 19, 19 and 18, and MUST-MATCH 0 in all five.

| capture | clock | side at the front | transport | observer | MUST-MATCH | informative | verdict |
|---|---|---|---|---|---|---|---|
| `2026-09-28_205830_314309565d6e` | 90 s | **original** | WebSocket | no | **0** | 12 | green |
| `2026-09-29_012528_03cb67c7cf19` | 200 s | **ours** | WebSocket | **yes**, two attempts | **0** | 28 | green |
| `2026-09-30_010531_a553ef8fd63c` | 250 s | **ours** | WebSocket | **yes** | **0** | 19 | green |
| `2026-09-30_183120_c4e7c284ff1c` | 90 s | **ours** | WebSocket | **yes** | **0** | 19 | green |
| `2026-09-30_201122_c4e7c284ff1c` | 48 s | **ours** | **bsPort** | no | **0** | 18 | green |

The five clocks come out of the log: 89 s, 199 s, 249 s, 89 s and 47 s between the first and the
last stamp, plus the settle. The windows the reports print on 2026-10-05 are
`t=1m29s..1m30.25s`, `3m19s..3m20.25s`, `4m9s..4m10.25s`, `1m29s..1m30.25s` and `47s..48.25s`.
The generated tests run the same.

All five carry `dirty=false` and the commit they were recorded with in the folder name
and in its `metadata.json`. The first is the only `format_version: 1` (not
trimmed); the other four are `format_version: 2` with their trim record, and the
report prints it in full.

The whole package coming out green does not mean there are no tests with Must-MATCH > 0:
`TestTruthComesFromMetadataNotFromTheFileName` mounts a capture with `status = 3` and without
`0x1E` **on purpose**, to check that the truth comes from `metadata.json` and not from the
file name, so there Must-MATCH 4 is the expected result and the test asks for it.

## The two synthetic fixtures

They are not captures: they are crafted ground truth, and that is why they are in the short loop.

| fixture | what it is for | clock | MUST-MATCH | informative |
|---|---|---|---|---|
| `synthetic` | the path with no capture: handshake, one peer, the observer stream | ~2,25 s | 0 | 19 |
| `synthetic-rollover` | **the SAME** capture with `-pageSize 128` and a single block: the cycled store | ~2,25 s | 0 | 12 |

`synthetic-rollover` is the one that found the `3` byte: until 2026-10-01 we
sent it as `4` ("observers deactivated", `sub_1400E3B50:403`), and none of
the five real committed captures had a store that had cycled a block, so
there was nothing to tell them apart. The real `3` is written by `sub_1400E3B50:93` and `:113`.

All four values of the `0x19` byte are in that same function, and they are four
different paths: `3` at `:93`/`:113` (the store already cycled, which does **not** register the
observer), `0` at `:170`/`:219` (the path that **does** register it, `:237-254`), `2` at
`:295`/`:344` (the observer list full) and `4` at `:403`/`:452` (observers
deactivated for the session). The `0` is the only one that is an acceptance and not a
rejection, and it is the only one followed by `0x1E`.

## What each capture contributed beyond that

### `2026-09-29_012528`: the frame store

The second attempt to observe arrives at 3m13s with its `0x19` and its `0x1E`. The
observer stream verdict, by connection:

| | original | ours |
|---|---|---|
| conn 49798 (the one that served the most) | 1.706 subs and 1.070 entries in **1.346** acks | 1.728 and 1.077 in 1.352 |
| conn 49823 | 771 subs and 523 entries in **800** acks | 245 and 322 in **600** |

Pages: **1.359 and 603** against **1.353 and 803** of the original. And what this capture
uncovered is in `SESSION-JOURNAL.md §24`: the store is a number of **BLOCKS**, and the acceptance
byte of the `0x19` is **0** and not 3.

### `2026-09-30_201122`: the first of bsPort

The 48 s trim is justified with its reason: bsPort needs **both**
connections (the second client announces itself at 15 s and the original answers it at
16 s) and the first second with acks, because the original's first `0x09` upload and its first
`0x0A`s land at 47 s.

- For the observer: **7.204 messages against 7.208** of the original, MUST-MATCH 0.
- What **does not** add up, with numbers: the `[pending-peers]` rule gives
  `ours: [] en 1 ack, [1 2] en 7` against `original: [] en 1 ack, [1] en 7`. It is the
  extra second peer, and it is the most expensive finding still open
  (`UNINTENTIONAL-DIVERGENCES.md` point 2).
- The **31 extra `LobbyRelay`** and the **25** of the untrimmed version, and the
  `fe 00 02 01 02 00 00` against `fe 00 01 01 00 00`.
- The report says, with numbers, that the sustained rates are the same (9/s on all
  three sides) and that the seconds that differ are not the first one: **"no cause is
  claimed for it"**.

### `2026-09-30_010531` and `2026-09-30_183120`: the observer's handshake

The two exist so that the observer's handshake is compared in the short loop, and
that is why they are the two that are most alike: 4m10s and 1m30s, both looking at the same
moment. The 1m30s one is the one that **took the 15-minute full suite down to 90 seconds of
clock**, because the 13m40s that go away are the same `0x0A` per second of the same
match.

## A capture that is on disk and is **not** imported

`captures/2026-09-29_101831_a553ef8fd63c`: 24m44s of clock, 92.6 MB on disk, and **1497 s
and 259 MB measured** of replay cost. It does not fit with the other four without going over
the budget. Its report says **0 must-match and 2 extra messages in 24 minutes** against
the original, so what would be lost by trimming it is **tail of the same match**.

If you want to import it:

```
trimsim -trim captures/2026-09-29_101831_a553ef8fd63c -out <dst> -keep 1m -why "..." -verify
```

and **only if `-verify` accepts**. A trim that has not been tested is not imported.

## What no capture can ask

- **Three or more peers.** There is none, and there are at least three differences that are
  only visible with three (the pending list, reconciliation, and the distribution of
  `LobbyRelay`).
- **A reconnected peer.** None of the five has one.
- **A forced desync.**
- **The end of the match with the client closing.**

## The ceiling of the instrumentation, not of the captures

These are the things the harness **cannot ask**, whatever is recorded. They were in
[`UNINTENTIONAL-DIVERGENCES.md`](UNINTENTIONAL-DIVERGENCES.md) 12 until 2026-10-05, which
was their own mistake: a limit of what a tool can measure is not a difference a client, a
peer or the replay host can tell on the wire, so it is not a divergence of anything. They
belong with the matrix, because each one says which row would have to be recorded to answer
it.

- **The transport is decided by the bytes, and with the whole stream.** The first two bytes
  are not enough, because an RLink message of 33.024..33.279 or 33.536..34.047 bytes starts
  `0x81` or `0x82`, byte for byte a WebSocket FIN+binary header. Each candidate is walked
  entirely in both directions and the one that reads without running into a byte that cannot
  start a frame wins; the HTTP upgrade, when it is there, overrides everything else.
  `SESSION-JOURNAL.md §27`, `internal/sim/transport_test.go`.
- **A log whose rows all decode and out of which not one message comes out is rejected**
  (`internal/sim.loadLog`). Before, only a log whose rows could not be decoded was rejected,
  and that was the hole through which an unreadable capture passed as a clean capture and a
  green test on top of two empties. `SESSION-JOURNAL.md §27`.
- **A stream the harness did not collect whole is not compared**, and it says with the
  reason. Before, it was compared. `SESSION-JOURNAL.md §40.4`,
  `internal/sim/truncatedstream_test.go`.
- **The number of replay frames is NOT must-match**; the order and the starting point are.
  An inverted pair crashes the client. `AGENTS.md`, "The two levels of verdict".
- **A settle shorter than the observer step's period cannot be compared**, and the harness
  says so rather than reporting the truncation: the floor is `interval+2` ticks derived from
  the server's own interval, and a configuration whose period exceeds the 3 s ceiling is a
  gap the ceiling wins and `observerStepPeriod` reports. `SESSION-JOURNAL.md §44.4`,
  `internal/sim/settle_test.go`.
- **What is needed and does not exist**: a match with **three or more peers** (no capture has
  three, and there are at least three differences that are only visible with three), a match
  with a **forced desync** (row 7), a match with the **observer** on bsPort, and a real
  **reconnection** (row 6). The first three are rows of the matrix above; each one is also a
  row of [`PENDING.md`](PENDING.md) 1 with what it would settle.