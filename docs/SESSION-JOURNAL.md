# Session Journal — the history, in chronological order

> **WHAT THIS FILE IS, AND WHAT IT IS NOT.** This is **history**: the record of the work
> rounds (the 45 is from 2026-10-05), with their measurements, their falsifications and their decisions in the place where they were
> made. It is where **the evidence** lives: every figure cited anywhere else was measured
> here, and can be read again with its context and with the machine load it had.
>
> **The current design is NOT here.** It lives in living files, and this journal references
> them instead of replacing them:
>
> | | what it is |
> |---|---|
> | **The rules and the front door** | [`AGENTS.md`](AGENTS.md) |
> | **Divergences we chose** ("we chose this", with reason and evidence) | [`INTENTIONAL-DIVERGENCES.md`](INTENTIONAL-DIVERGENCES.md) |
> | **Divergences we did NOT choose** (open bugs, gaps and `UNKNOWN`s) | [`UNINTENTIONAL-DIVERGENCES.md`](UNINTENTIONAL-DIVERGENCES.md) |
> | **The two closed questions** (closed by the user, not reopened) | [`CLOSED-QUESTIONS.md`](CLOSED-QUESTIONS.md) |
> | **What is missing** (by who can resolve each thing) | [`PENDING.md`](PENDING.md) |
> | **Performance** (CPU, RAM, allocations, goroutines, methodology) | [`PERFORMANCE.md`](PERFORMANCE.md) |
> | **Captures**: procedure and results one by one | [`CAPTURES.md`](CAPTURES.md) |
>
> **Why it still exists.** Because a measurement that lives only in a design file is
> lost the moment the design changes, and because a new session needs to be able to find
> *why* something was done, not only *what* was done. **No measurement is ever deleted**: when a
> section became superseded, what was done was **replace it with a pointer** that says where
> it is now and why it moved (**§41** lists the moves one by one). The original text
> is in the git history forever.
>
> **How to read it.** By section, not top to bottom. The index below maps
> **section → topic → living file**, and the list at the end maps **section → one line**.
> Once you know the number, you read that section and that is it.
>
> **The rule that matters here more than anywhere else**: everything asserted about the
> original's behaviour comes from `src/sub_*.cpp` or from a capture. No "it should
> work".

---

## INDEX

Sections **1 to 19** are the history of August and September 2026 and are not touched.
The **20 to 50** are the most recent rounds, and they are the ones that carry the current design in their
content. This index is the only thing you have to read to know where each thing lives.

> **`SESSION-JOURNAL.md §N` means this file.** The living documents and the older rounds of this
> journal cite sections of the journal by the shorthand `SESSION-JOURNAL.md §N`, which is the name
> this file had before §43 renamed it to `SESSION-JOURNAL.md`. **It is a citation token,
> not a different document and not a dead link**: `SESSION-JOURNAL.md §24` and `SESSION-JOURNAL.md
> §24` are the same section. The convention is stated where it is used
> ([`INTENTIONAL-DIVERGENCES.md`](INTENTIONAL-DIVERGENCES.md) and
> [`UNINTENTIONAL-DIVERGENCES.md`](UNINTENTIONAL-DIVERGENCES.md), each in its header note).
> It was left in place on purpose in §43: it is a *number* that survives a rename, and
> rewriting ~100 citations across ten documents in the same commit that renamed the file
> would have put the rename and a bulk edit in one diff. **What is measured**: 95 distinct
> `§N` numbers across the root markdown, of which **91 resolve** once a reference is
> resolved against the document that makes it and not always against this one (§45.1).

### TWO QUESTIONS CLOSED BY THE USER. THEY DO NOT REOPEN.

They stay closed, and **they no longer live here**: they are in
[`CLOSED-QUESTIONS.md`](CLOSED-QUESTIONS.md), with their evidence and with what stays
open inside each one. The summary, because the summary is what you have to keep in your
head:

- **The BattleServer is DETERMINISTIC.** It is the Age of Empires BattleServer that ships,
  so the original's output is reproducible. It stays closed **by authority**, not by
  measurement: it is not measured, no run that would contradict it is looked for. What
  does keep being done is comparing against the recording, which is the available copy of
  that output. The measurement §25 asked for is **voided**.
- **The observer's excessive zoom is NOT a BattleServer bug.** Confirmed from
  outside. Section **25** measured it and closed what can be closed: the observer's
  stream is indistinguishable from the original's, byte for byte. **It is not investigated
  again.** What stays open in 25 (ack grouping, `LobbyConnect` with the same bytes, a
  ping where the other answers with a pong) is informative, and **the `LobbyConnect` one is
  not the zoom**: it is a transport divergence and is still alive, in
  [`UNINTENTIONAL-DIVERGENCES.md`](UNINTENTIONAL-DIVERGENCES.md) 4. The "ping" was **never a
  ping** — it was two bytes of the `101` response's second TCP record — and it was closed as
  a false claim on 2026-10-05 (**§44.3**).

### WHERE EVERYTHING IS

The index goes **section → topic → living file**. The right-hand column is where the current
state is read; the section is where **the evidence** is, and the why of how it got there.
When a topic moved to a living file, its row says so.

#### By subject

| subject | sections | the current state lives in |
|---|---|---|
| The hard rules, the identity of the binaries, the capture layout, the deployment | — | [`AGENTS.md`](AGENTS.md) |
| **Divergences we chose**, with their reason | §16.1, §16.2, §16.2b, §16.4, §16.5, §25, §31, §32.4, §32.5, §33, §36.3, §38.3, §39.7, §40.4 | [`INTENTIONAL-DIVERGENCES.md`](INTENTIONAL-DIVERGENCES.md) |
| **Open bugs and unexplained gaps** | §16.2, §16.4, §25, §27, §29, §30, §32.5, §32.8, §33, §37, §40.4 | [`UNINTENTIONAL-DIVERGENCES.md`](UNINTENTIONAL-DIVERGENCES.md) |
| **The two closed questions** | §25, and the index's own closure | [`CLOSED-QUESTIONS.md`](CLOSED-QUESTIONS.md) |
| **What is missing, and who can resolve it** | §9, §16.6, §21.4, §26, §27, §28, §29, §30, §32.8 | [`PENDING.md`](PENDING.md) |
| **Performance: CPU, RAM, allocations, goroutines** | §34, §35, §36.5, §36.6, §37, §38, §39, §40.5 | [`PERFORMANCE.md`](PERFORMANCE.md) |
| **Captures: procedure, matrix and results** | §4 to §15, §21, §26 | [`CAPTURES.md`](CAPTURES.md) |
| **The names of the documents**, and why they are the ones in English | **§43** | this journal, and `git log --follow` |

#### By section

| What you are looking for | Where |
|---|---|
| What has to be done, and with which rules | `AGENTS.md` |
| The _5 crash and its cause | §4, and inside it `### Cause` / `### Fix` / `### Regression test` (there is **no** `## 5.`, and §45.1 says why the index does not claim one) |
| The format of the 0x0A ack, byte by byte | §3 and its subsection "Ack EndOfGameData"; §33, "where each byte comes from" |
| WebSocket framing, and the 101 | §3 ("WebSocket framing"), §23, `internal/sim/upgrade.go`, `webSocketPort.md` |
| bsPort's RLink framing, and the two transports | §27, `bsPort.md`, `AGENTS.md` |
| **Where the byte of the 0x19 (StartObserve) is written down** | **§22** (the inverted condition), **§24** ("The `0x19` condition was inverted, and the byte did not say so"), **§25** ("Correction of the 22"), **§28** (the 3 byte never reached the wire) |
| Why a `0` works and a `3` does not | §22, §24 |
| What the frame store keeps | §24 ("What the store keeps", "The store is a number of BLOCKS"), §19.1 |
| How a replay is served to an observer | §24, §40.4, `internal/session/session.go` (`StartObserve`, `updateObserversLocked`) |
| The observer's window and the `session+2184` | §24 ("The rest that came out of it") |
| The simulation harness: what it decides and why | §23, and `internal/sim/sim.go` (the package header) |
| The three must-matches and the two levels | §23 ("The three must-matches, and why only three") |
| The settle, and why the harness reads too much | §25 (end), `internal/sim/settle_test.go` |
| **The 11 extra messages, and the per-second histogram** | §23 ("THE REPORT SAYS WHEN"), **§28** |
| Why extras are NOT located by aligning the two streams | §23 ("Why extras are NOT located by aligning") |
| Trimming a capture, and how it is proven | §26, `internal/sim/trim.go` / `trimcheck.go`, `cmd/trimsim`, `AGENTS.md` |
| The identity of the binaries, and why the file name means nothing | §22 ("Capture identity"), `cmd/proxy/identity.go` |
| How it is decided which file is the original | §23, `internal/sim/sim.go` (`Load`) |
| `config.ParseFlags` and the four `os.Exit` | §23 |
| The "compared two empties" bug | §24 ("The harness had a hole here"), §25 ("What really was wrong"), **§28** (the cycled-store fixture) |
| The short loop against the full suite | **§28**, `internal/sim/short.go`, `internal/sim/mode_test.go` |
| **When the original creates a session peer** | **§29**, `internal/sim/peers.go`, `internal/sim/relaypeer.go` |
| The bsPort peer divergence (31 extra `LobbyRelay`) | **§29**, and now `UNINTENTIONAL-DIVERGENCES.md` 2 |
| **The reconnection by `secret`, and that it closes the replaced connection** | §20, **§30** |
| **The observer's lagged flag, and the target of its stream** | **§30** |
| **The body of the `0x1E` and the two time formulas** | **§30** |
| **The field-by-field audit of the session object and the connection object** | **§32** (the whole table, and the three bugs that came out of it) |
| **What a LEAVESIMULATION (0x1F) does, and why it does not remove the peer** | **§32** (`sub_1400E47D0`, `sub_1400E52D0`) |
| **The `0x0A` pending list: what it is, and why there is NO cut** | **§33** (all of it field by field; implemented). §32 read it backwards |
| The 40 bytes of the `session+2192` queue record, and who creates and deletes it | **§33**, `sub_1400E5670:208-219`, `sub_1400E52D0:57-96` |
| The cap of 17 chunks per tick, and why it is not implemented | **§33**, and now `INTENTIONAL-DIVERGENCES.md` 4.1 |
| **Closing a kicked observer, and why closing it from the tick is a self-deadlock** | **§36.1** |
| **The eleven fields 32.5 left unexamined, one by one** | **§32.8** (the tables) |
| **The five that left the divergence register on 2026-10-05, and why `conn+88` stayed** | **§44.1**, and now `PERFORMANCE.md` 5 ("The counters that reach nothing") and `PENDING.md` 2.1/2.3. **One `UNKNOWN` left of the eleven**: `conn+88`, still in `UNINTENTIONAL-DIVERGENCES.md` 10.2 |
| **The sync-message counter (`conn+104`) and the reconnect's echo** | **§36.2** |
| The two status bytes of `E2590:483`, and why the high one is dead | **§36.3** |
| The audit of the parent's root documentation, file by file | **§36.4** |
| **The cost per CLIENT: the store's slope, and the two caps of the upload queue** | **§37**, and now `PERFORMANCE.md` 5 |
| **What a tick COSTS: ns per tick, the per-stage split, and the memory** | **§38**, and now `PERFORMANCE.md` 3 and 4 |
| **Where a tick's CPU goes, with FUNCTION names** | **§39.1**, and now `PERFORMANCE.md` 4 |
| **Goroutines, and whether there is real contention: 8 events in 3 s** | **§39.4** and §35, and now `PERFORMANCE.md` 6 |
| How to read a pprof profile in Go 1.27 | **§39.5**, `internal/profiletest/`, `go1.27.md` |
| **The `goroutineleak` profile: which API it uses, its two traps, and the five paths** | **§35**, and now `PERFORMANCE.md` 6 |
| **The memory and the cost of a long match, and the caps that were chosen** | **§34**, and now `PERFORMANCE.md` 5 |
| What the long-match test does NOT measure | **§34** ("WHAT THIS FILE DOES NOT MEASURE") |
| **The RAM sampler, and the three ways that report 0 MB without warning** | **§40.5**, §36.6, §39.8, and now `PERFORMANCE.md` 2 |
| **THE OBSERVER STEP IS A PHASE**, and the original does not serve a fresh activation in the same tick | **§40.4**, and **§44.4** (answered from the binary, with the bound measured) |
| **The harness's settle was shorter than the observer step's period**, and that was the 1-in-60 flake | **§44.4**, `internal/sim/replay.go` (`settleFor`) |
| **The "WebSocket ping 0x65 against the pong 0x6B" was never a ping** | **§44.3**. Closed as a false claim; the real difference, a missing `Server` header, is fixed and compared |
| **A top-level `0x01` creates nothing**: the dispatcher has no case for it | **§44.2**, `internal/server/handlers_test.go` |
| **The five counters nothing reads are not divergences** | **§44.1**, and now `PERFORMANCE.md` 5 |
| **A session's peers are an array of 255 slots, and the order does not reach the wire** | **§40.2**, `internal/session/peerorder_test.go` |
| The last place that picked a port and released it, and how it is checked now | **§40.3**, `cmd/proxy/process_manager_test.go` |
| A stream the harness did not collect whole is not compared | **§40.4** |
| Which reconnection states resend frames | **§30** (`sub_1400EB5F0:237-249`) |
| **How an observer is paired with its entry: the peerIdentifier of the `0x1C`** | **§31**, `internal/sim/observeridentity_test.go` |
| The observer's zoom | **§25**. Closed: it is not a BattleServer thing. Now `CLOSED-QUESTIONS.md` |
| On bsPort you cannot observe | §27 ("On bsPort you cannot observe") |
| The reconnection by `secret` | §20 |
| The observable flag and the `0x1D` | §20, §19.6 |
| The tick's order, which was our bug | §18.4 |
| The reconnection, the lobby and the replay host | §18.3 |
| The multi-sub tail | §13 |
| The 0x13 relays and the socket that is not erased | §14 |
| The P2 that cannot join | §15 |
| Design decisions, with their reason | §16.1, and now `INTENTIONAL-DIVERGENCES.md` |
| The user's directives | §10 |
| What is pending | §9, §21.4, §26, §27, §28, §29, and now `PENDING.md` |
| The capture plan and its matrix | `CAPTURES.md` |
| **The names of the documentation files**, and why the ones that are technical keep theirs | **§43**, and `git log --follow` |

### THE SECTIONS, ONE LINE EACH

1. Objective. · 2. Architecture (original, reimpl, proxy). · 3. Protocol, RLink messages, and the ack
0x0A byte by byte. · 4. Proxy captures: `_3` to `_10` and the _5 crash, **and inside it the
crash's own `### Cause` / `### Fix` / `### Regression test`** — there is **no 5**, and §45.1
says why the index does not claim one. · 6. Other differences that did not cause the crash. · 7. Build,
deploy and validation. · 8. August's test campaign: coverage and bugs. · 9. August's pending items.
· 10. The user's directives. · 11. Analysis scripts. · 12. Existing documentation. · 13. Capture
`_11`: the multi-sub tail. · 14. Capture `_12`: relays 0x13 to the target, and the socket that is not erased.
· 15. Capture `_13`: P2 cannot join. · 16. Design decisions, the replay body to the host, known
divergences and what is not touched. · 17. Review of plan v3. · 18. Review of plan v4.
· 19. Round 5: the 0x0A inner queue, the sync timeout, StartObserve 4 against 2, the echo
budget, `0x1D`/`0x1E`. · 20. Round 6: reconnect by `secret`, and the observable flag. · 21. Captures
`_26` and `_27`: our server as primary. · 22. StartObserve: the condition was inverted; and the
correction that the acceptance byte is 0. · 23. The simulation harness, done; the three
must-matches; the report that says WHEN. · 24. The second observe attempt: 70 against 803, and the
frame store. · 25. The observer's zoom: measured and outside the stream; the RAM budget after the test
that brought the machine down. · 26. Trimming a capture: the data is shortened, the clock is not; `CheckTrim`.
· 27. bsPort: the transport is decided by the bytes, and the TLS is the port's. · **28. The short loop, the
four acks of second 26, and the 3 byte that never reached the wire.** · **29. When the original
creates a session peer: it is the ADDPEER, and a REQUESTJOIN is needed first.** · **30. The two
unexecuted paths: the reconnection (it closes the replaced connection, and only state 2 resends
frames) and the lagged observer (the flag is the compared hash, not "I sent a hash", and its target
is not the frame count).**
§ **31. The observer list is ordered by the peerIdentifier of the 0x1C, and not by the connection; what happens with two observers that share it, with a 0, and with a repeated 0x1C.** § **32. The field-by-field audit of the session object and the connection object.** The whole table, the three bugs that came out of it (the 0x1F does not remove the peer, `session+2142` had no writer, the 0x22 marker did not look at the client version), the harness's `indexOf`, and the verified-different that stays unimplemented. § **33. The `0x0A` pending list: what it is, where each byte comes from, and why there is no cut.** What §32 read wrong (the "cut" is the bottom of the loop), the 40 bytes of the `session+2192` record, the body of the `0x0A` field by field, six differences fixed and three left with their reason. § **34. The cost of a long match: two hours of match in one second, the memory curve, the allocations per tick, and the four caps that were chosen with those numbers.** That there was nothing to fix, and the three falsifications that prove it. § **35. Go 1.27's goroutine leak profile: which API it uses and its two traps (`Count()` runs one refresh behind, and you have to force the collection), why each cycle runs inside a closure (the detector is by reachability), and the five paths that have goroutines.** The baseline, the cap, the two falsifications, and that there was nothing to fix either. § **36. Closing a kicked observer (which is a self-deadlock if it is closed from inside the lock, and no longer is), the eleven fields 32.5 left unexamined with their verdict one by one, the `conn+104` counter modelled all the way to the reconnect's echo, the high byte of `E2590:483` that is dead in the original, and the audit of the root documentation.** Four falsifications, all with the defect in and taken out. § **37. The other cost axis: the CLIENTS. The sweep of 1 to 32 peers, the slope against `blockSize * f`, the two caps of the upload queue and their sum with no ceiling, the observer that is not a peer, and the cost of a live connection.** No line of production touched: both formulas hold at 0.99x, and the queue aggregate is exactly the product. What comes out is that the tick spends 9.3 objects per peer (short-lived garbage, with its account written down), and five falsifications with each cap's sensitivity written next to it. § **38. What a tick COSTS, and the end of the 9.3 objects per peer.** The axis that was missing (ns per tick at 1, 2, 4, 8, 16 and 32 peers at 125 ms, the per-stage split measured with the same fixture and one stage removed each time, the heap and the transient), the extrapolation to a 2-hour match with the arithmetic shown, and the pooling that took the 9.3 objects per peer and tick to zero (and the ack body that is NOT pooled, with its reason). Also the `TestSim_synthetic` flake: it was not a clock assertion, it was that the harness picked a port, released it and had the server bind it afterwards. § **39. The real PROFILE: where a tick's CPU and garbage go with FUNCTION names, the per-stage split with the clock in place, the three `AckState` maps as a 256 array (-42% tick), where there is NOTHING left to optimise in memory, and the goroutines with the contention (8 events in 3 s, and the server's profile running, which is the half the tick's does not see).** Also the two assertions that had to be fixed because the work made them thinner: a profile's sample count cannot depend on what the machine delivers, and the per-stage split's tolerance cannot be a fraction of a reference that gets cheaper. § **40. The stash, the peers table, the last port, and the observer step.** The stash that was left was a revert, `Session.Peers` becomes a table of 255 slots with every iteration in join order (and the justification the 39.7 left unproven), the last place that picked a port and released it now checks it, and the observer step stays an **8-tick phase** with its measured flake. § **41. The journal split by subject: one living file per topic, and the journal as history.** What moved, what was created and what is kept on purpose. § **42. The divergence criterion tightened to the black box.** Fourteen points that leave the divergence register (three with a false claim), the citations that did not say what they said, and what has been verified line by line. § **43. The documentation files get their names in English, because the prose in them is: ten renames and every reference with them.** § **44. The criterion applied to every entry, and what it moved.** The register walked entry by entry: the "ping" that was never a ping (§44.3), the observer step answered from the binary and with it the two harness flakes (§44.4), and the top-level `0x01` whose handler is gone (§44.2). § **45. The names, the dangling `§5`, and the `conn+88` premise.** §45.1 corrects the index instead of creating a `§5`, and re-measures the cross-references document by document (39 of 39 bare `§N` now resolve, against the 38 of 39 the 43 measured). §45.2 settles `conn+88`: the premise that both writers sat under `if (v15 != 0)` was a misreading of which `if` an `else if` binds to, **both writers are on a first connect**, and there is no third writer in the 7,450 files of `src/`. §45.3 is the six stale literals in `src_go` and the seventh in a comment. §45.4 is the suite over budget: measured, written up in [`PERFORMANCE.md`](PERFORMANCE.md) 9, and **nothing changed**, because which tests run is not this round's decision. § **46. The English sweep finished: the two names §43 missed, the Spanish left in `src_go`, and a committed checker that proves the claim instead of asserting it.** §46.1 is the two renames with every reference; §46.2 is the 246 flagged lines of `src_go` translated, the rest of the tree swept, and the measurement that no code moved; §46.3 is `scripts/check_language.py`, the tool that says so, run against the pre-change tree to show it finds what was removed and against 3,322 lines of English to show it finds nothing there. **47. The Go source stops citing the decompiler.** The review that asked for it (`DERIVATIVE-WORK-RISK.md`) is kept as history and marked as such. 47.1 measures what it called "many comments": **689 references across 48 of 144 files, and 343 of them in `_test.go`**, which the review had not listed. 47.2 is the four artifact counts going to zero with **not one byte of behaviour moving**, and what was deliberately kept. 47.3 is the transformation itself, with the three cases where deleting a citation would have left a sentence with no predicate. 47.4 finds the six copied log strings by searching every 12+ character literal against the decompilation. 47.5 is [`PROVENANCE.md`](PROVENANCE.md): the per-path list of what may not be redistributed, and the statement that this is **not** a clean-room implementation, which is itself **on that list**: an internal record of how the code was built is free and a public admission of it is not. 47.6 is [`BEHAVIOUR-NOTES.md`](BEHAVIOUR-NOTES.md), those findings rewritten as behaviour with no recovered symbol in them. 47.7 names the cost: 55 files, about 2,000 lines of comment diff, and `git blame` now pointing at a rewording round. **48. Performance, measured against the original instead of only against us.** 48.1 is `benchpeer`, a black-box tool that does not know which of the two binaries it is holding; 48.2 is the result, where the one real cost is ours (**idle commit 46,8 MB against 8,9 MB, 5,3x**, and it is the Go runtime's arena reservation and not our data structures) and the per-connection row answers a question section 5 could not; 48.3 is the two rows that are not numbers; 48.4 is the sentence that has to travel with the whole thing, which is that this compares **processes** and not algorithms; 48.5 is three Win32 API facts measured because they are not knowable by reading, two of which fail by returning **zeroes**; and 48.6 is why the tool repeats at all. **50. Two decisions, and one divergence closed without a capture.** 50.1 is the timing criterion, decided by the user, which moved the WebSocket upgrade latency out of the register and was written into the criterion itself rather than left implied by where one entry sits; 50.2 is a pending item its own evidence had already refuted; **50.3 is the upload gate**, a wire-visible divergence closed with no capture, where the failing test turned out to have been asserting a cadence the protocol cannot produce; 50.4 says what is left and who owns it, which is: nothing in the open register is closable without a run. **51. `-maxPagesToStorePerSession` keeps the original's lack of a cap.** The flag is the original's under the same name (`main_1400d1310.cpp:617`, and ours registers it as an alias), so the only question left was whether to put a ceiling on it; the decision is no, because a cap would have been a divergence and not a guard rail, and no line of production changed. **52. The byte range of the replay upload, closed by a negative measurement**: the user ran the original with a listener and `-uploadVia=lobby` and a match played to the end, and the connection arrived but no request followed; `battle-server.txt` agrees, so there is no upload to be byte-identical to, the entry moves to the intentional file, and `TestReplayUploadRequestShapeAndPayload` pins the request against what the binary establishes and fixes our stray `Content-Type`.

---

## 1. Objective

Reimplement **BattleServer.exe** (AoE2DE version 2.58.0, original in C++/RLink) in Go so that it is
**byte-for-byte compatible**, so that the game works the same as with the original and **does not
crash**. Every change is validated with captures from a proxy that duplicates the game's traffic to
our server (T1) and the original (T2).

---

## 2. Architecture

### Original (`E:\SteamLibrary\steamapps\common\AoE2DE\BattleServer\BattleServerOrig.exe`)
- x86-64 Windows PE, MSVC 2019 (v142), custom **RLink** network library.
- ~7,389 functions decompiled in `C:\Users\luska\BattleServer.exe_ai_dataset\src\` (IDA).
- Dispatcher: `src\sub_1400DE050_1400de050.cpp` — cases: `9u` (GameData) l.268, `0x13` (LobbyRelay) l.544, `0x14`, `0x17` (GameDataChecksum) l.727; check `LOBYTE(v95[1]) == 0xFE` l.311.
- Typical ports (original): bsPort **14001**, OOB **14002**, webSocket **14003**.

### Reimpl (Go) — `C:\Users\luska\BattleServer.exe_ai_dataset\src_go`
- Module: `github.com/luskaner/ageLANServer/battleserver`, `go 1.26`, dep `github.com/lxzan/gws v1.8.0`.
- Entry point: `cmd\battleserver\main.go`. Flags: `-relaybroadcastPort`, `-bsPort`, `-outofbandport`, `-websocketport`, `-muninLog`, `-sslCert/-sslKey`.
- Files in `internal\server`:
  - `server.go` — dispatch (`dispatchMessage` l.523), `handleLobbyConnect`, `handleLobbyPing`, `sendEndOfGameDataTicks` (session-level 0x0A tick: it walks `GetSessions()`, calls `PopAckTick()` and sends the same body to every peer via `SendMessage`), `handleGameData`, `handleLobbyGameDataAck`/`handleLobbyChecksumAck` (they queue into the client's session).
  - `ws_listener.go` — WebSocket listener with **gws** (Upgrader l.114). NOTE: the fix `ParallelEnabled: false`.
  - `ws_client.go` — `Process` (drains `incoming`), `enqueue` (FIFO), `lobbyPingTick`, `parseWsMessage`. **It no longer has ack state** (moved to the session).
  - `client.go` — bsPort client (same change: no ack state of its own).
- The 0x0A ack state lives in `internal\session\ack.go` (type `AckState`, session-wide; see §3) and in `internal\session\session.go` (`Session.Ack`, `QueueGameData`, `QueueChecksumEntries`, `PopAckTick`, `PeerIDForConn` l.140).
  - `ping_server.go` — OOB TCP (magic 0x58, TLS 0x4C, echo 0x14) — **does not use gws**.
  - `bs_listener.go` — bsPort. `relay_broadcaster.go` — UDP broadcast. `lobby_client.go` — lobby 8443.
  - Tests: `client_test.go`, `server_test.go`, `integration_test.go` (includes `TestGameDataAckOrder`), `listeners_test.go`, `testcerts_test.go`, `handlers_test.go`, `ws_client_test.go`, `ws_listener_test.go`, `ping_server_test.go` (see §8).

### Validation Manager / Proxy
- Proxy (Go): `src_go\cmd\proxy\main.go`. Duplicates the game's traffic to **T1** (reimpl) and **T2** (original). **Only T1 answers the game**; T2's responses are logged and drained.
- Config via env vars (`loadConfig`, cmd\proxy\main.go): `EXE1_PATH` (T1), `EXE2_PATH` (T2), `EXE1_ARGS`, `EXE2_ARGS` (do not touch per directive), `PROXY_HOST`/`PROXY_PORT`/`PROXY_WS_PORT`/`PROXY_OOB_PORT`, `RELAY_BROADCAST_PORT`, `TARGET1_HOST`/`TARGET1_PORT`/`TARGET1_WS_PORT`/`TARGET1_OOB_PORT`, `TARGET2_*`, `MAIN_LOG_PATH`/`SECONDARY_LOG_PATH`/`LOG_FILE_PATH`, `MITM_TLS`/`MITM_CERT`/`MITM_KEY`/`MITM_KEY_LOG`.
- CSV written **progressively** (goroutine + flush per row): a hard kill loses at most the in-memory queue, not the whole capture. Framing reassembled (`[len u16 BE][type][body]` on bs/oob; WS frames on the ws port) so that segmented TCP does not produce partial rows. Columns: `timestamp,src_host,src_port,dest_host,dest_port,data_b64`.
- Typical ports of a capture: original 14001/14002/14003; reimpl 15001/15002/15003.
- Logs in `F:\ageLANServer\build\battle-server-manager\`:
  - `main_log_N.csv` = game↔T1 traffic: `time,src_ip,src_port,dst_ip,dst_port,data_b64`.
  - `secondary_log_N.csv` = T2's responses to the game: `time,port,data_b64` (the orig writes header+payload in 2 rows/writes).
  - `proxy_N.log` = proxy start-up/process.

---

## 3. Protocol — RLink messages (over WebSocket and bsPort)

| Type | Name | Notes |
|------|--------|-------|
| 0x01 | RequestJoin | creates a session |
| 0x09 | GameData | chunks of replay data. If `body[0]==0xFE` it starts the 0x0A acks. |
| 0x0A | EndOfGameData | **ack tick** (~125ms). See the structure below. |
| 0x0E | SeekResponse | observer |
| 0x0F | Peer | |
| 0x10 | Proxy | |
| 0x11 | Ping | |
| 0x13 | LobbyRelay | forwarded (`SendMessage`), ~50 per session |
| 0x14 | LobbyPing | ping 9B: `[nonce u24][00 00 00 00]`; echoed |
| 0x15 | LobbyConnect | 32B; echoed; reconfigures the session |
| 0x17 | GameDataChecksum | flood at the end; 0x0A ack with entries |
| 0x18 | StartObserve | |
| 0x1C | StartObserveRequest Vx | same handler as 0x18 |
| 0x1A | ActivateObserver | |
| 0x1B | StopObserve | |
| 0x1D | SetObservableParameters | 17B; nobody echoes it |
| 0x47 | GET (HTTP) | WS handshake |
| 0x88 | WS Close | closing frame |

### Ack EndOfGameData (0x0A) — `internal/session/ack.go` (verified byte-exact against `_6`)
The ack is **session-level**: 1 body per tick (~125ms), identical for both players
(verified in `_6`: destinations 38596 and 49899 receive the same bodies, sessionID = sub type).
The body always starts with `[0xFE]`; the subs carry the **player's slot** (1-based peer ID:
P1=01, P2=02) and the players byte is the number of players in the session.

```
flush      = [FE][subCount] + subs + tail
sub        = [slot u8][len u16 LE][payload]
tail 1 sub (2 players) = [01][Y][len u16][02][N][00] + 8N    (Y = the OTHER player's slot)
tail multi-sub          = [00][len u16][02][N][00] + 8N
           with no entries pending: len = 0 and no block (→ `01 Y 00 00` / `00 00 00`)
standalone = [FE][00][players][01][02][len u16][02][N][00] + 8N
heartbeat  = [FE][00][players][01][02][00 00]
```
- Real chunks from `_6` (WS, body after the `0A`): heartbeat `fe000201020000` (7B, ×3124);
  flush 2-subs `fe 02 [01][1f 02][543B zlib] [02][02 00][00 30] 00 00 00`; flush 1-sub
  `fe0101000001020000` and `fe0102000001010000`; with entries `... 01 02 33 00 02 06 00 + 48B`.
- The entries are acked ONE tick after they arrive (only as a flush trailer or standalone
  with the previous tick's entries); the tick of arrival emits a heartbeat.
- The body is sent through each transport's `SendMessage` path (WS: raw `[0A][body]`;
  bsPort: `[u16 len][0A][body]`).
- The heartbeats only start when data flows (in `_4`, with no match, the orig emitted
  not a single 0x0A) → `AckState.PopTick` returns nil until the first ack.

### WebSocket framing (key finding, doc `webSocketPort.md`)
- **Game → server**: frames **masked with a ZERO mask**, header+mask and payload in separate writes:
  - `82 89 00 00 00 00` + `14 <nonce> 00 00 00 00` = WS ping.
  - `82 FE 04 2A 00 00 00 00` + `13 01...` = LobbyRelay (ext-len).
  - `09 FE 64 02 01 78 01 ED...` = GameData (raw payload with no mask applied).
- **Server → game**: frames **unmasked**:
  - `82 09` + `14 <nonce> 00 00 00 00` = ping echo.
  - `82 7E 01 9E` + `0A FE 01 01 64 02 01 78 01 ED...` = EndOfGameData ack.
- 101 handshake: our 129B against orig 156B; same `Sec-WebSocket-Accept`; they differ in header order and `Server: WebSocket++/0.8.2` → **not a cause of the crash**.

---

## 4. Proxy captures — results

### `_3` (~01:26 UTC, 204s) — hybrid ws_listener pre-01:48
- The post-upgrade traffic is WS frames (the framing finding). Compared in `PROXY-CAPTURE-VALIDATION.md`.

### `_4` (~01:53 UTC, 294.2s) — gws ws_listener post-01:48 (build deployed 04:43)
- 934 reimpl acks against 933 orig → **byte-equivalent** (the reimpl "merges" 2 of the orig's ticks).
- Relays 202/202 identical; ping echoes 71/71.
- The game died at 294.2s **with both servers** (it was not our bug). Pre-run 345s+.
- Documented in `PROXY-CAPTURE-VALIDATION.md`.

### `_5` (18:32–18:37 UTC, ~190s) — **the crash**
- The game disconnected at t+189.9s "as soon as the match started" **only with our server**.
- The orig kept sending 7B heartbeats until the proxy was killed (18:37:46:690, mid-line).
- **Byte-level divergence (the only one of the whole session)**: the final ack t+189886 (count 6) had
  its last two subs INVERTED.

### `_6` (15/08, ~3 min) — **first 2-player capture, 0 observers** → session-level ack
- P1 (38596) sent 932 rows and P2 (49899) 496; both send chunks `09 FE 02 00 00 30`
  (GameData FE of 2B); P1's largest frame = 1136B `13 02 01 0F...` (LobbyRelay).
- **Finding**: the 0x0A acks belong to the session, not per client — both players receive
  identical bodies (145 distinct bodies GT / 173 ours; all start with `FE`).
- **The orig's format decoded byte-exact** (see §3): 7B heartbeat with the players byte and
  `[02]` before the len (`fe000201020000` against our 6B `fe0001010000`); flush with
  subCount, subs per player (`[01][1f 02][zlib]...`), and a tail with the complementary slot
  (`01 02 00 00`/`01 01 00 00` for 1-sub flushes; `00 00 00` for 2-subs; `01 02 33 00
  02 06 00 + 48B` with entries).
- The format had NEVER been verified against the orig: `_4` had 0 players and the orig emitted
  not a single 0x0A (the earlier "byte-equivalent _4" was about other messages).

### Corrected reimpl (15/08 04:05, session-level ack)
- New `internal/session/ack.go` (`AckState` + byte-exact format); `Session` exposes
  `QueueGameData/QueueChecksumEntries/PopAckTick/PeerIDForConn`; `server.go` queues
  from `handleLobbyGameDataAck`/`handleLobbyChecksumAck` and the tick emits 1 body broadcast
  to all the peers. `end_of_game_data.go` and the per-client state removed.
- Byte-exact tests with `_6`'s real bodies in `internal/session/ack_test.go`
  (heartbeat, standalone N=1, 556B 2-sub flush, 1-sub flushes, flush with entries,
  session-wide merge). `TestHandleLobbyGameDataAckEdges`,
  `TestHandleLobbyChecksumAckEdges`, `TestSendEndOfGameDataTicks`,
  `TestBsProtocolGameDataAcks`, `TestWsClientEndOfGameDataTick` redone for the new path.
- `go test ./...` green, `go vet` clean. Build:
  `C:\Users\luska\BattleServer.exe_ai_dataset\build\BattleServerReimpl.exe` (10.105.856 B).
- **Pending validation**: a re-capture of 2 players with this build.

### Finding `_7` (15/08 19:20): the session is created in the LOBBYCONNECT, not in RequestJoin
- **The game NEVER sends RequestJoin (0x01)** in this flow. The game's WS frames use a
  **zero** mask (`82 A0`/`82 89` + mask `00000000`), so the payloads are readable: in all of
  `main_log_7` there is no `01...` frame at all, and the reimpl's log (unmasked, the same log
  format as the orig's) shows no `type=0x01` in any leg. The orig's "ProcessPeerMessage:
  fromPeerID=0 REQUESTJOIN" is its session-setup path, logged by the LobbyConnect
  (one per player, 18:36:56:360 and 18:40:43:357).
- The orig creates/joins the session in the LobbyConnect (by the message's sessionID): its first 0x0A
  ack = 18:41:08 (~25s after player 2's LobbyConnect) and 3855 acks until 18:45
  (~16/s ≈ 2 per tick × 2 players). The orig's real types in `_7`: 0x0A=3855, 0x13=171,
  0x14=92, 0x15=3, 0x48=2, 0x21=2. **The orig did NOT emit AddPeer (0x03) in `_7`** (the earlier
  "0x03=2" was a misparse of the long `82 7E 03 A3` = relays of 931B with extended length).
- Our code only created the session in `handleRequestJoin` (server.go) → `FindSessionByPeer`
  was always nil → the session-gated paths (`handleLobbyGameDataAck`/`handleLobbyChecksumAck`/
  tick) discarded everything → 0 acks in `_7`. In `_6` the old build acked because it was per-client
  (it did not need a session).
- **Fix (build 15/08 19:38)**: `session/manager.go` → new `GetOrCreateSession(id)` (creates by
  id, falls back to the counter if id<=0); `server.go`'s `handleLobbyConnect` now creates/joins the session
  by the LobbyConnect's sessionID, adds the client as a peer (parity log "REQUESTJOIN",
  one peer per client), `SetReconnectable` and echo (as before). `handleRequestJoin` stays
  as a fallback. The `TestHandleLobbyConnectSetsReconnectable` test was redone for the new contract
  (one session per sessionID + a single peer). `go vet`/`go test ./...` green.
- Build: `C:\Users\luska\BattleServer.exe_ai_dataset\build\BattleServerReimpl.exe` (10.147.328 B)
  — deployed directly to `E:\SteamLibrary\steamapps\common\AoE2DE\BattleServer\BattleServerReimpl.exe`.
- **Pending**: a re-capture of 2 players; compare the 0x0A bodies (main_log_N vs secondary_log_N)
  and the volume (~16/s × player). Note: the CSV timestamps are in **microseconds** (/1000 for
  DateTimeOffset).
- Note on `_7`'s data path: in `_7` the match data goes INSIDE the 0x13 relays (09 FD) and over
  bsPort (a single raw 09 FE at 18:44:45); in `_6` it went as raw 09 FE over WS. If after the fix the
  volume does not add up, consider feeding the AckState with the relayed content.

### Finding `_8` (15/08 22:03, 2 players): 4 divergences fixed → build 22:44
- Prior validation: 4363 reimpl acks against 4368 orig, ~99% byte-identical with the
  session-per-LobbyConnect fix. 4 mismatches remained, all with an explanation and a fix:
  1. **Dedupe of checksum entries per page** (the orig acks 159 pages, we had 428=2×214):
     both players upload byte-identical 0x17 streams (214 pages each, p1==p2 exactly, no
     dupes within a stream). The orig dedupes **per page, first-arrival-wins**. The 55 missing
     pages (32,33,56,57,65..76,83..97,119,120,129,150..153,158..162,167..169,176,177,
     184,189,197,201..205) have unique CRCs (except 119=crc 1963941 duplicate) → **it is NOT dedupe
     by CRC**; the exact mechanism could not be reverse-engineered (an internal race in the orig: 159 vs 214
     in `_8`, a residual that cannot be replicated). Fix: `AckState.seenPages map[uint32]struct{}` in
     `ack.go`'s `QueueEntries`.
  2. **Multi-sub order in mixed flushes**: the orig merges the per-peer queues in slot-ID
     order (always ascending slot: (1,2),(1,2,1),(1,2,2); never (2,1)); we
     queued by arrival ((2,1)×7). The proxy interleaves from concurrent goroutines per
     connection → non-deterministic per target. Fix: `sort.SliceStable` by slot in `PopTick`
     (it preserves the arrival order within the same slot; it matches 13/14 of the orig's cases).
  3. **Peer not removed when the connection closes**: after P2's disconnect (22:02:4x) the orig
     kept acking P1's 1-player heartbeats for ~80s (players=1); we were still at
     players=2. Fix: `RemoveClient` now does `FindSessionByPeer` → `HandleLeaveSimulation`
     + `RemovePeer`. NOTE: the first version of the fix deadlocked (`RemovePeer` locks `s.mu`
     with `RemoveClient` already holding it → a non-reentrant mutex); fixed by capturing the peerID
     BEFORE `HandleLeaveSimulation` and calling sessionManager/RemovePeer OUTSIDE the lock.
  4. **GameData 0x09 relay removed**: the orig answers GameData 0x09 with 0 acks (logs
     only); we were forwarding 90 relays. Fix: `OnSyncGameData` removed from `handleGameData`.
- 0x17 format (MsgGameDataChecksum) confirmed: `[N u8][N×8 entries]` + an optional trailer
  `[09 fe][len u16 LE][payload]`; entry = `[crc u32 LE][page u32 LE]`. The trailer's `09 FE` is
  queued as a sub exactly like a normal 09.
- Accepted residual divergence: batching per tick phase (the orig groups subs from different
  ticks; the reimpl does not) — identical content per player (2291/2235 B), only the
  number of standalone differs (10 vs 13). Not replicable with our tick-driven architecture.
- Tests adjusted to the dedupe contract (different pages): `TestHandleLobbyChecksumAckEdges`
  and `TestBsProtocolGameDataAcks`. `go vet` + `go test ./...` green (the `RemoveClient`
  deadlock hung the suite; fixed in server.go). Build deployed 22:44:34:
  `E:\SteamLibrary\steamapps\common\AoE2DE\BattleServer\BattleServerReimpl.exe` (10.148.864 B).
- **Pending**: a re-capture of 2 players (`_9`) and compare: sub order sorted, players=1 at the
  end, 0 acks 0x09, checksum entries 214 (not 428), and count the residual divergences.

### Finding `_9` (15/08 23:12, 2 players, build 22:44): 1 real divergence fixed → build 00:09
- Capture: P1=192.168.1.129 (ws 44031, oob 44032, bs 11458), P2=192.168.1.11 (ws 49905, oob
  49906, bs 49844). A single session: 101s at t≈407-420 (ws+oob per player), LobbyConnect 0x15
  at 407.7/419.1, acks from 452.2 to 701.9/702.0. No disconnect until the END of the match.
- **Count**: orig 3543 acks (44031: 1840, 49905: 1703) vs ours 3862 (1999+1863) = **+319
  (~+160/player)**. CAUSE: the orig emits ~7.37 acks/s/player (mean 135.8ms, peaks of
  137-138ms ≈ 10%, 152 gaps >200ms), NOT 8/s; we are exactly 8/s (125ms). It is the jitter of
  the orig's 1.3s animation loop (see `_5`), NOT replicable → **residual accepted** (the orig
  "loses" ~8% of its ticks).
- **Pairing by timestamp** (destination 44031): 1840 pairs, 1408 identical, 432 different; (49905):
  1703 pairs, 1403 identical, 300 different. The differences are of GROUPING (tick phase): the
  orig flushes 1 sub per tick when P1's/P2's chunks land in different ticks (in `_8`
  they landed in the same tick and the orig DID merge); we join both if they fall in the
  same 125ms window. Per-sub content identical.
- **REAL BUG found — multi-sub flush tail with the same slot**: the orig uses the complementary
  tail of the OTHER slot when ALL the subs are of the same slot, even with 2+ subs:
  `(1,)` and `(1,1)` → tail `01 02 00 00`; `(2,)` and `(2,2)` → `01 01 00 00`; only MIXED slots
  `(1,2),(1,2,1),(1,2,2)` → tail `00 00 00`. Evidence: orig `(1,1)→01020000` x10, `(2,2)→
  01010000` x4 in `_9`; `(1,1)` x4 and `(2,2)` x2 in `_8`. Our `PopTick` only did it with
  `len(subs)==1` → we emitted `(1,1)→000000` x12 and `(2,2)→000000` x2. **Fix**: in `ack.go`'s
  `PopTick`, the condition "all the subs are the same slot" (first slot == the rest) instead of `len==1`.
  New test: `TestAckFlushSameSlotTails` in `ack_test.go`.
- **Session end validated (fix 3 of `_8`)**: P2 closed the WS (frame 88 80) at 685.4; the orig dropped
  to players=1 at 685.5 and kept acking P1 with heartbeats `fe000101020000` until 702.4
  (P1's close, end of capture). Ours: IDENTICAL (128 acks in [686-703] both, last
  1-player flushes `fe010102000030000000` byte-equal). **players=1 at the end ✓, peer removal
  ✓, and in [686-703] orig==ours acks==128 → identical count in 1-player.**
- **0x09 relay ✓**: neither server sends 0x09 to the game (0 relays both).
- **Checksums 0x17**: 75 frames (both players), 209 distinct pages uploaded, trailer `09 fe`
  in all of them. Entries acked: orig 316 (160 distinct) vs ours 338 (169 distinct) → dedupe per page
  works (was 428); the orig discards ~25% of pages because of its internal race (the `_8` residual).
  NOTE: the "pages acked that were not uploaded" (10240, 42752, 20480…) = a misparse of f9c by mixing
  the OOB ports (14002) into the parsed stream; not real.
- **Sub order ✓**: our 2 mixed flushes (1,2) come out slot-ascending; orig (1,2) x12 with the
  same order.
- **0x13 relays**: game 171 → the orig forwards 155, we forward 168 (+13). The 6 relays before P2's
  entry (t<418.7) are NOT forwarded by either (the orig discards them, we buffer/discard them too) —
  only 3 vs 16 discards differ (smaller, because the orig discards relays at the end of the match 702.4).
- **0x14 pings**: orig 17 vs ours 69 (+52 self-pings of 10s — pending #3 lobbyPingTick).
  0x15: 2=2; 0x0F: orig 3 vs ours 2; WS ping echo 0x82: 167=167.
- Build 00:09:51 deployed `E:\SteamLibrary\steamapps\common\AoE2DE\BattleServer\
  BattleServerReimpl.exe` (10.149.888 B). `go vet` + `go test ./...` green.
- **Pending**: a re-capture to validate: (a) the multi-sub tail with the same slot, (b) the +319
   count residual (orig ~7.37/s vs 8/s — accepted, not replicable), (c) 0x14 self-pings.

---

### Finding `_10` (16/08, 2 players, current build): 4 discrepancies fixed
- Validation of the current build (tests green) against capture _10. 4 mismatches remained, all with a fix:
  1. **Same-slot multi-sub tail in a flush**: the orig uses the complementary tail (`01 02` for slot 1, `01 01` for slot 2) when **all** the subs are the same slot, regardless of the current number of players. Our `ack.go` had the condition `playerCount == 2` → it did not fire at the end of the match (1 player). **Fix**: `ack.go:121` — the `playerCount == 2` check removed.
  2. **Self-pings 0x14 (lobbyPingTick)**: the reimpl sent its own ping every 10s (62 vs 19 orig). The orig only echoes incoming pings. **Fix**: `lobbyPingTick` removed from `sendEndOfGameDataTicks` (`server.go:682-684`).
  3. **Tests updated**: they expected a flat tail `[00 00 00]` on single-sub; corrected to the complementary one (`01 02 00 00` / `01 01 00 00`) in `client_test.go` and `handlers_test.go`.
  4. **AddPeer (0x03) in RequestJoin**: already implemented in `handleRequestJoin` (line 738); it is not sent in LobbyConnect.
- **Residuals**:
  - Extra flushes: 52 vs 42 (the orig's batch-tick-phase is not replicable).
  - Mixed-slot flushes (1,2): 0 vs 2 orig (the data lands in the same tick).
  - Same-slot tail at session end: 2 mismatches both (the edge case of 1 disconnected player).
- Tests: `go test ./...` green, `go vet` clean.
---

### Evidence
- The game sent (in order): `F8 0C` @t+189861 and `F8 0E` @t+189862 (two 96B, 1ms apart).
- The **original's** ack at t+189942: sub2=`...F8 0C...`, sub3=`...F8 0E...` → correct order.
- The **reimpl's** ack at t+189886: sub5=`...F8 0E...`, sub6=`...F8 0C...` → **REVERSED**.
- The client consumes the ack as replay data in order → crash on validating the inverted chunk.

### Cause
`ws_listener.go:115` used `ParallelEnabled: true` in `gws.NewUpgrader`. gws dispatches each
message through a worker pool (`readQueue.Go(msg, c.dispatch)`, `reader.go:158`, "If ParallelEnabled
is enabled, OnMessage is called in parallel" `types.go:96-97`). Two back-to-back messages are
processed in non-deterministic order → `enqueue` queued them inverted (and although there is a mutex,
the **order** of the concurrent appends is not guaranteed).

### Fix
`ws_listener.go:115` → `ParallelEnabled: false`. `OnMessage` is invoked synchronously in the read
loop; the queue → `Process` (FIFO) → `dispatchMessage` → `queueDataSub` → `popTick` path is
all synchronous and preserves the arrival order.

### Regression test
`TestGameDataAckOrder` in `integration_test.go`: sends two different GameData messages back to back and
verifies that the ack's subs come out in the order sent.
- With the bug (`ParallelEnabled: true`): **3/20 runs fail** with `ack sub order = "BA"`.
- With the fix: it always passes. `go test ./...` → green.

### Rebuilt binaries (14/08 20:19, with the fix)
- `C:\Users\luska\BattleServer.exe_ai_dataset\build\BattleServerReimpl.exe` (10.104.832 bytes)
- `C:\Users\luska\BattleServer.exe_ai_dataset\src_go\battleserver.exe` (the same)
- Previously deployed (04:43) at `E:\SteamLibrary\steamapps\common\AoE2DE\BattleServer\BattleServerReimpl.exe`.

---

## 6. Other reimpl vs original differences (not causes of the crash)

- **Probe (probe-phase)**: the proxy forwarded the probe to both targets. The orig accepted it as WS
  (`WebSocketManager::SockAccept` 18:34:33:433) but its 1.3s animation loop was too
  slow and the game closed the probe at t+200ms → the orig **never answered**. Our Go answered
  instantly (101 + 12 ping echoes `82 09 14 <u24> 05 00 00 00 00`). Wire: 575 of our
  responses against 562 of the orig's = 12 pings + 1 extra x101. **It is not the cause.**
- **Match pings**: 17/17, the same counters/order. Ours: 0x0A×505, 0x0F×1, 0x13×36, 0x14×29,
  0x15×1, 0x48×3 vs orig: 0x14×17, 0x48×2, the rest the same.
- **101 handshake**: 129B vs 156B (see §3).
- `munin2` (18:59:47) = a session AFTER `_5`; it showed `ws_connection_errors.value 1.000000`
  in that new session (it does not apply to the `_5` analysis).

---

## 7. Build / deploy / validation (workflow)

```powershell
# from C:\Users\luska\BattleServer.exe_ai_dataset\src_go
go test ./... -count=1                                  # tests (includes TestGameDataAckOrder)
go build -o "..\build\BattleServerReimpl.exe" ./cmd/battleserver
go build -o "battleserver.exe" ./cmd/battleserver        # the one the proxy used in _5
```
- The proxy launches T2 via `EXE2_PATH` (in `_5` it pointed at `src_go\battleserver.exe`).
- To validate the fix: relaunch the proxy with EXE2_PATH at either of the two rebuilt exes,
  start a match and check that there is no crash at ~190s and that the final ack keeps the
  order `[F8 0C][F8 0E]`.

---

## 8. Test campaign (14–15/08/2026) — coverage ~94%

### Final coverage per package (15/08/2026, go tool cover)

| Package | Coverage | Comment |
|---------|-----------|------------|
| cmd/battleserver | 94.2% | main 60.9% (blocks 60-62 missing: Run error, and SetModule's blocked lines) |
| internal/logger | 100% | ✓ |
| internal/metrics | 100% | ✓ |
| internal/net | 99.2% | `SetWriteBuffer` (UDP) without a possible error; `writeLoop` network blocks |
| internal/protocol | 100% | ✓ |
| internal/replay | 92.9% | `uploader.go` lines 145-146 (timeout trigger), 187-189 (incomplete response) |
| internal/server | 91.6% | multiple error-path blocks (port conflicts, TLS/certs, empty WS messages) |
| internal/session | 97.6% | `updateObserversLocked` 82.4% (dead code paths: Lagged+v5=false, delta<0, target≤0) |
| **Global** | **94.0%** | The 100% goal partially reached; dead blocks accepted |

#### Unreachable/defensive blocks (accepted):
- `session.go:313-314` (v5 always true → the `!v5` path never runs)
- `session.go:318-319` (target ≤ 0 only with PageCount=0, impossible after AddPage)
- `session.go:326-327` (delta<0 or replay nil → impossible conditions in normal flow)
- `session.go:331-333` (empty GetPages → requires inconsistent state)
- `cmd/main.go:60-62` (Run error → requires a ctx.Cancel during Run, not a broad test)

Full suite: `go test -count=1 ./cmd/... ./internal/...` → green (3 consecutive runs without
flakes). `go build ./...` and `go vet ./...` clean. Note: `go test ./...` with `-coverprofile`
fails because the root package `.` has no Go files (use `./cmd/... ./internal/...`, and an
ABSOLUTE path for the profile from pwsh).

### Real bugs found and fixed during the tests
1. **`internal/net/udp.go`'s `broadcast()`**: panic with a nil `buildMsg` (l.70, nil deref) → a
   guard `if b.buildMsg == nil { return }` was added.
2. **`internal/protocol/messages.go`**: uint16 overflow in `Deserialize` (Message) and
   `GameDataMessage.Deserialize` — with a declared length of 0xFFFD–0xFFFF, `3+length` wrapped and
   panicked `slice bounds out of range [3:2]` (a malicious frame could bring the server down). Fixed
   with `int` arithmetic: `len(data)-3 < int(length)` / `len(data) < 3+int(length)`.

### New tests (files)
- `cmd/battleserver/main_test.go` — a subprocess re-running the test binary via
  `os.Executable()` (in `go test` on Windows, `os.Args[0]` is only the basename) for the `os.Exit` paths;
  the `main-*` cases call `main()` **without** a prior `parseFlags()` (a double parse → panic
  "flag redefined" masked by `defer os.Exit(0)`). Covers the defaults (BsPort 27012, OOB 27212,
  ObserverMax 512, LogLevel "debug", Region "global", SimulationPeriod 0, UploadVia "").
- `internal/logger/logger_test.go` (extended) — `TestLogUnknownLevel` needs
  `SetGlobalLevel(Level(99))` (99 > Debug is filtered in `shouldLog`).
- `internal/net/net_test.go` (rewritten, merging the originals) — TCPConn via `net.Pipe` (watch out:
  `net.Pipe` is synchronous; the writer goes in a goroutine or it deadlocks), a real UDP broadcast with
  `DialUDP`+`ListenUDP`.
- `internal/protocol/messages_test.go` — deserialisation edge cases (after the uint16 fix).
- `internal/replay/uploader_test.go` (extended) — `TestStartAndPushReplayData` uses a real TCP
  listener and checks `POST /game/replay/setReplayData?matchid=4242&sequenceid=0`;
  `TestOnAddReplayDataQueueFull` checks `len(sendQueue)==cap` (not draining the queue in the test).
- `internal/session/manager_test.go` — `HandleJoinRequest` broadcasts in the **new** session
  (which only contains whoever is joining) → the joiner never receives anything.
- `internal/server/handlers_test.go` — a mock `mockMsgClient` (ClientConn+messageClient);
  `TestSendEndOfGameDataTicks` requires `srv.AddClient(client)` (the tick walks `srv.clients`);
  for StartObserve you have to set `srv.config.ObserverMax` (0 by default → silent rejection);
  `LobbyConnect` with `reconnectState=2` needs a body of ≥21 bytes (17 does not parse).
- `internal/server/ws_client_test.go` — real WS connections via `dialWS`+`serverWsClient`
  (polling `srv.clients`); **key discovery about gws v1.8.0: `Upgrader.Upgrade` does NOT call
  `OnOpen` — it is invoked from `socket.ReadLoop()` (goroutine)** → the tests that depend on the
  WsClient's registration have to wait. `TestWsClientEnqueueOverflow` and `ProcessMaxPackets`
  use listener-only (without `mainLoop`, which drains the queue every 1ms and breaks determinism).
  `lobbyPingTick`: pass `now` in the future (11s) because `lastPing` is initialised to `time.Now()`.
- `internal/server/ping_server_test.go` — the RFC 6455 accept-key vector
  (`dGhlIHNhbXBsZSBub25jZQ==` → `s3pPLMBiTxaQ9kYGzzhZRbK+xOo=`), `wsFrame` (3 length
  classes), a complete `serveWS` via `net.Pipe` (101+accept, 0x14 echo masked/unmasked,
  frames 126/127, ping→pong 0x8A, oversize >1MB → return; watch out: a real length >1MB,
  `0x1000xx`, not `0x1100`, which is 4352 bytes). The "0x00 ^ 0x02" overflow in Go
  literals is a constant, not a runtime XOR.
- `internal/server/ws_listener_test.go` — a standalone listener (without a server): a Broadcast to 2
  clients checked with `waitForFrame`, `CloseConnection`, `GetLocalAddr`, `UnderlyingConn`,
  OnPing/OnPong via real connections, OnOpen/OnClose with a retry loop (OnOpen↔conns race).
- `internal/server/client_test.go` (extended) — SendMessage 65535 → error, Process timeout
  (lastSeen -6min → Close), SetObserver no-op, Send after Close, `readMessage` with `readStubConn`
  (a stub net.Conn with buffer + EOF; `net.Pipe` blocks on writes pending ReadFull),
  ReadLoop cancel, Send timeout with a full queue (≥900ms).
- `internal/server/handlers_test.go` — `TestMuninLoop` (MuninReportDelay=20ms, tick + cancel),
  the Client deadlines.

### Test notes
- PowerShell: use an ABSOLUTE path for `-coverprofile` (the relative one gets lost); `go tool cover`
  has to be run with the workdir in `src_go` (otherwise it resolves the module wrongly and reports a fake 0.0%).
- `frameType`/`waitForFrame`/`dialWS`/`wireFrame`/`wsTestHandler` (helpers in
  `integration_test.go` l.~560) reused by the ws_client/ws_listener/ping tests.
- Private types (e.g. the uploader's `sendQueue`, the WsClient's `incoming`) are checked
  from inside the package (tests in `package server`).

---

## 9. Pending (August 2026) — MOVED

> **SUPERSEDED by `PENDING.md` (2026-10-04).** This list is from August and almost everything that
> was here is either resolved or no longer applies. What **stays alive** from this section:
>
> - Point 0 (a capture with an observer) **is done**: there are five committed captures with
>   an observer. See `CAPTURES.md`.
> - Point 3 (`lobbyPingTick`) was resolved on 16/08 in `_10`: removed from
>   `sendEndOfGameDataTicks`. The methods stay in the interface and are not called.
> - Point 6 (the `_9` residual, +319 acks) was **closed** in §28: they are not ours, they come
>   from the replay and from the capture's clock.
> - Point 4 (the probe) is the reason for `battle-server.txt`: the original **does not**
>   answer it, and its 1.3 s animation loop is written down there. It is accepted.
> - Point 2 (the `_6` count deltas) **has not been looked at again**, because `_6` is a
>   `main_log_*.csv` log that was deleted. `_6`'s findings are in §4 and in
>   `CAPTURES.md`; the log format is not.
> - Point 5 (coverage) is from August and the current coverage is in `SESSION-JOURNAL.md §8`.
>
> **Where the live list is**: `PENDING.md`.

---

## 10. The user's directives

- **Do not touch** the proxy's `EXE1_ARGS`/`EXE2_ARGS`.
- The proxy only hands the game T1's responses (T2's are drained).
- Do not iterate indefinitely if there is no progress ("If you do not see that you are able to, then after a certain number
  of attempts do not continue indefinitely").
- Keep all the relevant memory in markdown (this document).
- **Every protocol bug we find requires a dedicated regression test** (with the
  exact bytes of the capture when possible) before closing the iteration, so that there
  are no regressions (directive 16/08/2026; first example: `TestLobbyRelayInitCrossRelay13`,
  commit 392bd8c).
- **Document everything that is not touched** (directive 26/09/2026): every known divergence,
  every unimplemented part and every deliberate decision go to §16 with its reason. An item that is
  not written down is an item that is lost in the next sprint.
- **The Go project must work without cgo** (directive 26/09/2026): no `import "C"` and no
  dependencies that require it. Verify with `CGO_ENABLED=0 go build/vet/test ./...`. The
  race detector (`-race`) does require cgo, so it is only a verification
  tool, never a build requirement.

---

## 11. Analysis scripts — MOVED

> **SUPERSEDED (2026-10-04).** All the scripts on this list lived in
> `C:\Users\luska\AppData\Local\Temp\opencode\`, which **is not the repository**, and all of them
> read `main_log_*.csv`, the capture layout the proxy stopped writing. There is no
> live consumer of any of them, and their input logs (`F:\ageLANServer\...`) have not existed
> on this machine for a long time.
>
> The five CSVs that are cited (`main_log_5.csv` .. `main_log_9.csv`) **are deleted**: the
> capture history was deleted whole (294 MB) and what was kept is in §4 to §15
> of this journal, in `CAPTURES.md` and in `src_go/internal/server/testdata/observer_vectors.json`.
>
> **What replaced them, and why**:
>
> | what it did | what does it now |
> |---|---|
> | the ack comparators, the `0x17` formatting, the sub order, the checksum entries | the `internal/sim` harness, which does the same **and also compares byte for byte against the original** |
> | the per-type and per-port counts, the tick intervals, the timestamp pairing | the harness's report, with §23's per-second histogram |
> | `clients.ps1`, which reproduced the 2-player session | `internal/sim` with the synthetic fixture, and `internal/sim/peersweep_test.go` for the sweep |
> | the small-frame scanner by structure, and the racy parser of the original | `cmd/proxy/packets.go` and §27's transport rule |
>
> **What survives from the capture analysis**, and is still live: `scripts/compare_observer.py`
> (as a **module**, not as a CSV CLI), `scripts/export_vectors.py` (the flow
> of `CAPTURES.md`), `scripts/csvlib.py` (imported by the two above and by `fidelity.py`) and
> `scripts/fidelity.py`, which is a parent's change and is not touched.

---

## 12. Existing documentation (index) — MOVED

> **SUPERSEDED (2026-10-04).** This inventory of 15 markdowns is from August and §36.4
> audited it whole, file by file, with its verdict. The **current** inventory is in
> the table "THE DOCUMENTATION OF THE ROOT" in `AGENTS.md`, which is the one that says what is current and
> what is superseded, and that table has been corrected again with this split.
>
> The detail of the audit —which is what it says about each file and why— stays here, in the
> **§36.4**. The four files marked there as superseded
> (`uploadVia-mysql.md`, `uploadVia-odbc.md`, `PROXY-CAPTURE-VALIDATION.md`,
> `WEBSOCKET-BSPORT-COMPARISON.md`) **were not deleted**: all four are linked from
> `README.md`, which is the user's work and is not touched, and from `webSocketPort.md` and the
> three lobby files. What was done is what §36.4 recommended as the second option:
> **a line at the top of each one** saying it is superseded, why, and what replaces it. Deleting them
> would have left broken links in files that are not mine to fix.
>
> **Superseded on 2026-10-05.** The reason above is gone: the user lifted the protection on
> `README.md` and it was rewritten, so "linked from the file that is not touched" no longer
> holds. Each of the four was then decided on its own and **three were deleted** —
> `uploadVia-mysql.md`, `uploadVia-odbc.md` and `PROXY-CAPTURE-VALIDATION.md` — with every
> link to them removed. `WEBSOCKET-BSPORT-COMPARISON.md` is **kept**, because its conclusion
> is superseded but its detail is not and no other document carries it. The reasoning and the
> verification per file are in `AGENTS.md`, "Superseded: three deleted, one kept". **The §36.4
> text above is left as it was written**: it records what was decided then, not what is true
> now.

---

## 13. Capture _11 (16/08/2026) - validation of the multi-sub tail fix (commit c4f59cb)

### Context
- Logs: `F:\ageLANServer\build\battle-server-manager\main_log_11.csv` (617116 B),
  `secondary_log_11.csv` (314226 B), `proxy_11.log`. Start 02:40:03.
  Public ports: bs=44249 ws=44248 oob=44247. T1=14001/14002/14003 (orig),
  T2=15001/15002/15003 (reimpl). The same binaries as `_9` (build 00:09 with the tail fix).
- Capture proxy: `F:\battleServerLogProxy\main_read_2.go`, byte-identical to
  `main_read.go.bak` except for a final newline (24261 vs 24262 B) -> the format did NOT change.
- 2 players confirmed (no observers): P1=127.0.0.1:48759, P2=192.168.1.11:49876.

### The orig's writing (root cause of the "resyncs")
- The orig writes every WS message in TWO writes (header `82 [len]` and payload
  separately) with both orders (`8208`+payload or payload+`8208`) and the writes
  additionally interleave BETWEEN messages (2 concurrent writers), e.g.
  `[relayPayload][flushHdr][relayHdr][flushPayload]`. A full-stream parser
  is not reliable for the orig; a small-frame scanner by known
  structure (header window +/-4 B) was used, only for validation.
- Consequence: `_9`'s "165 frames 0x82" and the "0x82 echo 167=167" were
  parser artifacts (heartbeats `0a fe 00 02 01 02` split), NOT real pings.
- The reimpl writes every message in ONE single write (1989+1857 rows = 1989+1857 frames),
  parses with 0 resyncs (types {10:3550, 15:2, 19:172, 20:120, 21:2} per connection clean).

### Verdict: multi-sub tail fix VALIDATED
- 12 B ORIG flushes (both connections): `0afe01010200003001020000` (tail `01 02 00 00`)
  and `0afe01020200003001010000` (tail `01 01 00 00`) alternating -> complementary tails,
  identical to the reimpl's (the same pair of payloads alternating; additionally the variants
  `0afe01010200004c01020000` / `0afe01020200004c01010000` and the dual flush
  `0afe0201020000300202000030000000` match byte for byte).
- 10 B flushes `0afe0101000001020000` / `0afe0102000001010000` also identical.
- Base heartbeat ack identical: `0afe000201020000` (body `fe 00 02 01 02 00 00`).
- 0x15 echo: the same 32 B format (`15` + session data; P1 verified with header
  `82 20`; P2's appears without a visible header due to the interleaving, content identical).
- End of session: the reimpl ends with hb players=1, flush `0afe0102000001010000`,
  state hb `0afe01020600005100000000000000` and a 1-player flush
  `0afe010102000030000000` (tail `00 00 00`, the same as `_9`); the orig closes
  (frame `88 02 03 e8` from the client) after a `01 01 00 00` flush + heartbeats.
  The reimpl's closing sequence was already validated in `_9`.

### Pending divergences (no change)
- 0x14 self-pings: reimpl 120 vs orig ~75-14 (the 10 s lobbyPingTick cadence,
  pending #3, decided to keep).
- 0x13 relay counts (164 vs 172) are not comparable: the orig's count is not reliable
  because of the interleaving.
- The ack cadence residual (+159) was already accepted in `_9` (7.37/s vs 8/s).

### Scripts _11 (C:\Users\luska\AppData\Local\Temp\opencode\)
- `f11.py` overview (scan), `f11b.py` samples, `f11c.py` strict (crash),
  `f11d.py` strict mixed (garbage), `f11e.py` per connection + handshake skip
  (reimpl clean, orig fails), `f11f.py` racy header/payload-order parser (fix
  wait-when-short: the orig needs to wait for data, not resync, because the sequence
  is always valid; even so the orig is not parseable in-stream because of the interleaving),
  `f11g.py` traces, `f11h.py` small-frame scanner by structure,
  `f11i.py` strict reimpl + orig scanner with positions.

---

## 14. Capture _12 (16/08/2026) - relays 0x13 to the target + heartbeat players=2 on close (commit e068680)

### Context
- Logs: `F:\ageLANServer\build\battle-server-manager\main_log_12.csv` (404102 B),
  `secondary_log_12.csv` (331204 B). The same binaries as `_11` (build 00:09).
  T1=14003 (orig) + T2=15003 (reimpl); the proxy starts T1 with EXE1_ARGS `--bsPort 14001
  --outofbandport 14002 --websocketport 14003` and T2 with `--bsPort 15001 --outofbandport
  15002 --websocketport 15003`.
- 2 players: P1=172.24.112.1 (ws 18365, oob 18366, bs 2687), P2=192.168.1.11 (ws 49882,
  oob 49883, bs 49819). A full session with a match and a close.
- **Row timestamps = NANOSECONDS** (numbers like 5818490→5840115 that were used
  as µs are ns; `/1e6` = ms). Verified: a 1ms gap between rows of the same connection in
  ORIG (different writes), a ~5-30 µs gap in OURS (the same write).

### Finding 1: the 0x0F messages are SENT BY THE CLIENT (no bug in handleLobbyRelayInit)
- The 0x0F is NOT an echo from the server: P2 sent a standalone 0x0F of 321 B at ~38 s; P1
  sent the SAME content (0x0F 321 B) inside its relay#1 `13 01 01 0f 04 00 ...`
  (577 B). The proxy injects the init relays with the `04 00...` prefix + the peer entry
  (which explains the timestamp offset and the different payload).
- Both servers echo the 0x0F to the sender: ORIG to P2 = 32 B truncated (cosmetic garbage
  from its double-write), OURS = 321 B intact. 0x0F count = 2 on both servers (1 to
  each player; the "3" was the `0f 04 00` inside the init relay).
- Conclusion: no change in `handleLobbyRelayInit` (init→MsgPeer to the sender is correct).
- The `0x1F`s detected in ORIG streams = parser artifacts (corrupt streams), not
  real messages.

### Finding 2 (REAL BUG): relays 0x13 are routed to the TARGET, not to the sender

- Evidence: P1 sent 971 relays with target=2; ORIG delivered them to P2 (indirectly: P2 received
  749 relays with target=1 from P1; it did not forward them). OURS echoed them to the sender.
- Format (confirmed by disassembling another AI, LobbyRelayMessage__vmethod_20):
  `13 [toPeerID][fromPeerID][0x0f][payload]` — byte@1=to, byte@2=from. ORIG delivers to
  `session.peers[toPeerID]` (a direct SendMessage; if the slot is nil, it sends nothing).
  Init relays go with to==from (13 01 01, 13 02 02, 13 00 00) and the orig detects them by
  content → MsgPeer to the sender (not a bug).
- **Fix** (`server.go`'s `dispatchMessage` case MsgLobbyRelay): if `!client.isRelayed() &&
  len(body) >= 3 && body[0] == body[1]` → `handleLobbyRelayInit`; otherwise the new
  `routeLobbyRelay`: `GetPeer(body[0])`, type-assert to `ClientConn`, `SendMessage
  (MsgLobbyRelay, body)`, drop if nil (which covers `13 00 00` and dead peers; `WsClient.Send`
  fails silently on a closed conn → routing to a dead peer is safe).

### Finding 3 (REAL BUG): the session peer is NOT removed when the socket closes
- Heartbeat window at close (ns): P1 closed the WS at ~1786845818301 (with no 0x1F), P2 continued
  ~21.6 s more. OURS→P2: 170× players=1 (`0afe000101020000`) + 4× players=2; ORIG→P2:
  174× players=2 (`0afe000201020000`). The orig only removes the peer with an explicit 0x1F
  (LeaveSimulation); closing the socket does not remove it (reconciles `_8`: the leaver DID send
  0x1F → players=1, and `_9`: a clean leave → players=1).
- **Fix** (`RemoveClient` reworked): it keeps the peer in the session (only removing it from the
  routing map `s.peers`), and ends the session when no live peer is left
  (`sessionHasLivePeerLocked`). `HandleLeaveSimulation` (0x1F) still removes the peer
  explicitly.

### Clean OURS inventory (strict parse per connection)
- →P1: {0x15:1, 0x0F:1, 0x13:975, 0x14:27, 0x0A:5240}; →P2: {0x15:1, 0x0F:1, 0x13:758,
  0x14:27, 0x0A:5417}.

### Note on the ping side (NOT the focus; pending explanation)
- The 4478 side (ping) carries masked 0x14 pings (`82 89 00 00 00 00 14 ...`); OURS
  answered 52× (15002→18366) vs ORIG 104× (14002→18366); and 13 vs 27 for 2687/49819:
  exactly half, with no explanation yet.

### Tests and build
- New: `TestRemoveClientKeepsSessionPeer`, `TestRouteLobbyRelay` (handlers_test.go);
  updated `TestDispatchMessageLobbyRelayInit`/`TestDispatchMessageLobbyRelayShortBody`
  (they register the mock as peer 1 so that the routing delivers to the client itself).
- `go vet` clean, `go test ./...` green (all packages). Commit `e068680`.
- Build: `C:\Users\luska\BattleServer.exe_ai_dataset\src_go\BattleServerReimpl.exe`
  (10.152.960 B, SHA256 B6247022...) — deployed on BOTH:
  `E:\SteamLibrary\steamapps\common\AoE2DE\BattleServer\BattleServerOrig.exe` and
  `BattleServerReimpl.exe` (the proxy's two processes = reimpl now).

### Pending _12
- A functional capture with both processes = reimpl (the logs should come out identical);
  verify: (a) relays 0x13 to the target, (b) players=2 in the heartbeats after a disconnect with no
  0x1F, (c) general parity of types/counts.

### Scripts _12 (C:\Users\luska\AppData\Local\Temp\opencode\)
- `f12h.py` overview (types per connection), `f12i.py` strict (`0f` transitions), `f12j.py`
  relay/echo/0x0F parity, `f12k.py` 0x1F scanner, `f12l.py` closing heartbeat window
  (players=1/2 per server).

---

## 15. Capture _13 (16/08/2026) - regression: P2 cannot join the lobby (commit 82ab1c8)

### Context
- Logs: `F:\ageLANServer\build\battle-server-manager\main_log_13.csv` (36528 B),
  `secondary_log_13.csv` (15672 B), `proxy_13.log`. Start 17:37:36.
- **ROLE SWAP**: T1 (main_log, 14001/14002/14003) = OUR build
  (`BattleServerOrig.exe` = reimpl B6247022); T2 (secondary_log, 15001/15002/15003) =
  THE REAL ORIGINAL (the user restored `BattleServerReimpl.exe` = pristine 3.216.736 B,
  SHA256 CF4DD19D...). main=ours, secondary=real.
- Public ports: bs=32592, ws=32591, oob=32590. P1=172.29.80.1 (ws 53816, oob 53817),
  P2=192.168.1.11 (ws 49880/49888, oob 49881/49889). P1 was the first to connect over WS and
  sent 7 `13 01 01` relays (~every 2s) + pings; P2 connected 2 times (49880 and 49888) and
  withdrew without joining.
- **Symptom**: P2 cannot join the lobby (retries `13 00 00` 20B every ~2s).

### Root cause (the ONLY functional divergence of the capture)
- The init relay (`13 00 00` / `13 01 01`, to==from, with the `0f 04...` entry) → the ORIGINAL
  delivers the MsgPeer (0x0F) to ALL the session's peers, the sender included:
  P2's init (327B, `82 FE 01 47` + `13 00 00 0f 04 00 00 00 00 00 00 ff ff ff ff
  00 00 ff ff ff ff 78 01 ...`) produced on the REAL one: `82 7E 01 44` + `0f 04 00...` (324B)
  to P2 **and** to P1 at the same time (4938516055). That is the mechanism by which the host (P1)
  discovers that a new player is joining.
- WE only echoed the MsgPeer to the sender → P1 never saw P2 → the join hung
  (P2 retrying `13 00 00` 20B relays that the REAL also discards — GetPeer(0)
  = nil → drop, exactly like us).
- Regression introduced by e068680 (the init handler was not touched, but _12 had no
  case with P2 in the session: its per-connection 0x0F counts could not distinguish
  "echo to the sender" from "broadcast to a single peer").

### Byte-exact verification (the rest of the flow = identical)
- 0x15 echo: both `82 20` + identical 32B (P1 `...8B 89 D1 3D D2 A7 A9 45...`, P2
  `...69 C3 1D 1A E0 F7 4C 6B...`; shared sessionID `AF 8F 94 61 00 00 18`).
- P1's init (586B) → MsgPeer 583B = relay - `13 01 01`, IDENTICAL on both.
- P1's relays #2-6 (`13 01 01 0f 04/00...`) → verbatim relay echo (581/929/939/955/1042B)
  on both, with correct `82 7E [len]` headers (0x0245=581 etc.).
- P1's relay #7 (1042B, 4979553053): the REAL did not answer (P1 closed the WS 11ms later —
  the closing window, not relevant); we did echo it.
- OOB/WS pings: identical echoes on both (_12's note about the 4478 ping side no longer
  applies: here the OOB pings are `82 89`+`14 ...` with echo `82 09` on both sides).
- P2 received NOTHING more from the REAL than from us (only the 0x15 echo + MsgPeer + pings).

### Fix (commit 82ab1c8)
- `server.go`'s `handleLobbyRelayInit`: it now sends `SendMessage(MsgPeer, body[3:])` to
  ALL the session's peers (`FindSessionByPeer` → `GetPeers()` → type-assert
  ClientConn); with no session it falls back to the sender. The `isRelayed()` gating (only the
  first relay of each connection is an init) stays the same and matches the REAL (the init of
  P2's 2nd connection was also broadcast on the REAL).
- `TestDispatchMessageLobbyRelayInit` extended: a 2nd mock as peer 2 → it checks that the
  MsgPeer reaches BOTH peers.
- `go vet` clean, `go test ./...` green (9 packages).
- Build: `src_go\BattleServerReimpl.exe` (10.152.960 B, SHA256 D5364570...) deployed
  ONLY on `BattleServerOrig.exe` (T1 = ours); `BattleServerReimpl.exe` (T2) = the
  real original, untouched.

### Pending _13
- Re-capture of 2 players with this build: P2 must be able to join the lobby (P1 receives P2's
  MsgPeer), cross relays `13 01 02`/`13 02 01` to the target, heartbeat players=2 and a full
  match; the main vs secondary logs should come out byte-equivalent.

---

## 16. Design decisions and open questions (2026-09-26)

> **PARTIALLY MOVED (2026-10-04).** This section used to be both the divergence register
> and the journal of why each decision was taken. The divergence register **has moved out**
> into two living files, and §16.3 (the bugs found and fixed in the
> revision) **stays here**, because it is history with its commit.
>
> | subsection | where it is now |
> |---|---|
> | §16.1 Decisions taken (with a reason) | [`INTENTIONAL-DIVERGENCES.md`](INTENTIONAL-DIVERGENCES.md), and what is not a divergence in §16.3 |
> | §16.2 What body is uploaded to the replay host | the doubt in [`UNINTENTIONAL-DIVERGENCES.md`](UNINTENTIONAL-DIVERGENCES.md) **1**; the "we do not change it" in [`INTENTIONAL-DIVERGENCES.md`](INTENTIONAL-DIVERGENCES.md) **4.2** |
> | §16.2b The uploader from `0x13` | [`INTENTIONAL-DIVERGENCES.md`](INTENTIONAL-DIVERGENCES.md) **4.2** |
> | §16.3 Bugs found and fixed | **it stays here.** It is history with a commit |
> | §16.4 Known unresolved divergences | [`INTENTIONAL-DIVERGENCES.md`](INTENTIONAL-DIVERGENCES.md) and [`UNINTENTIONAL-DIVERGENCES.md`](UNINTENTIONAL-DIVERGENCES.md) |
> | §16.5 What has not been touched, and why | [`INTENTIONAL-DIVERGENCES.md`](INTENTIONAL-DIVERGENCES.md) |
> | §16.6 The next capture: 1 observer | **done.** See `CAPTURES.md` and `PENDING.md` 1.1 |

**What stays here without moving**, and why: §16.3 (the seven types the original
dispatcher does not handle, the `0x0F` with an invented layout, the dead framings, the duplicated type
table in the proxy, the `matchid=0`, the `DropoutPeer` re-read as host-only, the flag with
the original's name) and §16.4 only in the part that **is evidence from the binary** and not
state: the addresses, the string literals and the blocks of code. The **verdict** of
each row now lives in the two divergence files, with their separation rule —"if there is
a reason to accept the difference, it is intentional; if not, it is a bug"—.

**State after the plan review** (`E:\Descargas\RESPONSE_TO_REIMPLEMENTATION_PLAN.md`) and
the P0/P1 code cleanup. Commits in `src_go`: `acc332d` (ack queue cap), `0b1b064`
(lobby registration), `a58187b` (framing in the transports), `11ae493` (dead framing
removed), `c74f413` (beacon docs), `1addadf` (replay matchid).

### 16.3 Bugs found and fixed in the revision

*( this subsection was not moved: it is the history of what was wrong and which commit
fixed it )*

### 16.1 Decisions taken (with a reason) — MOVED TO `INTENTIONAL-DIVERGENCES.md`

> The twelve points go to [`INTENTIONAL-DIVERGENCES.md`](INTENTIONAL-DIVERGENCES.md).
> The one that is consulted most is **point 1** (framing is decided by the transport and the observer stream's
> `[u16 BE length]` prefix is **not** removed without a capture that contradicts it),
> which is 1.4 and 2.1 over there.
>
> The points that are **not** divergences, because we do the same as the original, stayed
> here: the per-slot ack queue cap (300 entries / 512 KiB, the same as
> `sub_1400E2E00:53-54`), the `LobbyConnect` echo only if the parse succeeds
> (`sub_1400DE050:138`), the lobby registration with the shared serializer, the peers table indexed by a u64, and the `RemovePeer`/`RequestJoinDenied` with a 30 s close.

### 16.2 What body is uploaded to the replay host (mostly resolved) — MOVED

> What is **established** by reading the binary is in
> [`UNINTENTIONAL-DIVERGENCES.md`](UNINTENTIONAL-DIVERGENCES.md) **1**, together with
> the part that is still unresolved and the reason why it is **not** changed. The separation is
> that of the general criterion: what is known goes in the unintentional file because it is a
> gap, and the "we do not change it" goes in the intentional one because it is a decision. The two
> files are mutually exclusive and the same fact does not appear in both.

### 16.2b The uploader is fed from 0x10 and from 0x13, and we only use 0x10 — MOVED

> [`INTENTIONAL-DIVERGENCES.md`](INTENTIONAL-DIVERGENCES.md) **4.2**, with the two
> dispatcher calls and the reason for not doing both at once.

### 16.4 Known unresolved divergences — MOVED TO THE TWO DIVERGENCE FILES

> **Intentional** (there is a reason to accept them):
> [`INTENTIONAL-DIVERGENCES.md`](INTENTIONAL-DIVERGENCES.md) — the `DropoutPeer` gate
> assumed to hold, the endianness of the unsent serializers, `maxreplay` with no
> consumer, the `titleID` accepted even when it does not match, the `0x08` to the observer that is not
> sent, the RequestJoin session-setup, the relay without a "host only" output, the observer stream's
> prefix, and the log that is not added when the beacon stays off.
>
> **Unintentional** (or there is no reason, or it is an `UNKNOWN`):
> [`UNINTENTIONAL-DIVERGENCES.md`](UNINTENTIONAL-DIVERGENCES.md) — the replay's byte
> range, `conn+88`, `session+2140`, the unexamined counters, the reconnection without
> validation, and the pairing of observers with a shared `peerIdentifier`.
>
> **What stays here** is the part of this table that is not state but **evidence**: the
> addresses, the binary literals and the cited blocks. Not the table itself, because its
> "why it is still open" column is exactly the column that has been rewritten in the two
> files, and the same finding in three places is a finding that gets updated in two.

### 16.5 What has not been touched, and why — MOVED

> [`INTENTIONAL-DIVERGENCES.md`](INTENTIONAL-DIVERGENCES.md), which is where the
> register of "what is left out and why" lives. The two items that are **not** divergences but
> gaps —**AoE3** and the **integration into the ageLANServer repo**— are in
> [`PENDING.md`](PENDING.md), because they are not a behaviour difference but
> work that has not been done.

### 16.6 The next capture: 1 observer (never done) — DONE

> **This capture was made.** Five committed captures have an observer, and the observer's
> stream is compared **byte for byte** in the harness. See `CAPTURES.md`.
>
> That runbook's checklist (`scripts/OBSERVER_CAPTURE.md`) was **deleted** on 2026-10-04:
> it described the `main_log_*.csv` layout and the `MAIN_LOG_PATH` /
> `SECONDARY_LOG_PATH` variables, which the proxy **deliberately ignores**. What replaces it is
> `CAPTURES.md` plus the harness.
>
> The comparator it named, `scripts/compare_observer.py`, was **not deleted**: its CSV CLI
> is dead but the **module** is used by `scripts/export_vectors.py`, which is the step
> documented in `CAPTURES.md`. And the observer stream does not need a Python comparator:
> `internal/sim` compares it, and that is what makes it must-match.
---

## 17. Review of plan v3 (2026-09-27)

The plan `REIMPLEMENTATION_PLAN (2).md` is the third one, and it is the best of the three: it retracts four
claims from its earlier revisions, three of them coinciding with corrections already made
here. The complete answer, with the verification of each point, is in
`E:\Descargas\RESPONSE_TO_REIMPLEMENTATION_PLAN_V3.md`.

**What the plan gets right and needed no action:** the observer list
(`0x14, 0x18, 0x1a, 0x1b, 0x1c`, mask `0x1d100000`), the unhandled inner `8` with the verdict `2`
dead, not answering `0x11`, the write-only host ID, the upload contract with zlib level 9 + base64,
the two gates before the jump table, and the AddPeer's second opaque blob (which we already read, called
`Extra`). The OOB as WSS with an `0x14` echo **we already were**: we have TLS, a WebSocket upgrade and the
`0x14` check on both paths, with a raw fallback that is a superset.

**The plan's mistake that corrects us:** the `ProcessPeerMessage` verdict has a live `0`,
and `0` means skip that destination. The consumer states it explicitly
(`sub_1400DF920:150-153`, `v13 = false`) and `v13` governs the send (`:181`). In `2bf35af` we had
removed the whole verdict type; the reason is detailed in §16.4.

**Implemented in this round:**

| Item | Commit | What changed |
|---|---|---|
| Observer framing: it is WebSocket, not a stream with a `[u16 len]` prefix | `b1331b9` | `WsClient.Send`/`SendMessage` without an observer branch; the `loadgen` client and the test helper read WebSocket frames |
| Sync timeout: reaps the peer that stops sending | `7573bd8` | `lastSync` per slot, a 60 s deadline with a LobbyConnect override, and the immediate scheduler `ScheduleCloseIfNone` |
| Session key as the pair `(titleID, sessionID)` | `8e6b4f7` | `SessionKey`, with `titleID` as a signed `i16` |
| Checksum reconciliation between peers | `2a6645f` | A vector per peer, positional merge, up to ten rounds per tick. **The order was wrong, fixed in `b59ef37`; see §18** |


**Documented, unimplemented, and why:**

- **The `0` (consume) verdict is not implemented, and that is deliberate.** "Consumed" is a **pair** of
  behaviours: the original consumes a RequestJoin *and* re-retransmits a stripped `0x0F` to every
  connection of the session and to every pending joiner (`0x1400e6200`). We implement neither
  of the two halves, so today we forward, which is incorrect but not destructive. Adding only the
  "skip the destination" would deliver that message to nobody, **worse than today**. The per-type
  distribution (`2`/`5`/`9-with-`0xFE`` to `0`, the rest to `1`) is **confirmed in §18 from the producer**, so it is
  no longer only the plan's reading; what is still unimplemented is the semantics of "consume", not
  the table.
- **The desync notice `0xFB` was not implemented; `b59ef37` implements it** as a
  GameDataMessage with body `FB 00 00`. The real layout turned out to be the minimum, not the plan's; see §18.3.

- **The `0x14` echo limit before joining is not implemented.** The field exists and its value is
  confirmed: the peer constructor sets it to **100** (`sub_1400D4770:45`,
  `*(DWORD *)(a1 + 64) = 100`, over a 0xA8-byte object). What does **not** appear in the decompiled
  set is the place that consumes it: the echo's call site (`sub_1400DE050`, which only does
  `sub_1400D49A0(a3, Block)`) and `sub_1400D49A0` never touch `+64` at any point. Without seeing the rule
  (whether it decrements, when it stops) implementing it would be inventing it. It is the same pattern that made
  us distrust the `titleID`: a bare offset with a value, and no path that reads it.
- **The uploader fed from `0x13` is not implemented.** It is still blocked by the same doubt as
  §16.2: whether the byte range that is uploaded includes the framing prefix or not. The plan adds a
  third source (the final page when the last peer leaves, `0x1400e74c0`), which is noted.

**The peer constructor as a second way of confirming.** `sub_1400D4770` (the constructor, called
right after `operator new(0xA8)`) zeroes `+56, +68, +72, +80, +88, +90, +96, +104, +112, +152,
+156` and sets `+64` to **100**. That independently confirms, from the initialiser and not from
use, the offset identifications that were made throughout the revision:

| Offset | Field | How it was confirmed |
|---|---|---|
| `+0x40` (64) | echo budget, 100 | constructor |
| `+0x44` (68) | `clientMessageVersion` | constructor to 0 + `sub_1400EB5F0:470` writes it from `LobbyConnect+0x58` |
| `+0x48` (72) | timeout override | constructor to 0 + `sub_1400D4BD0` reads it as a deadline |
| `+0x58` (88) | "already joined" | constructor to 0 + the gate at `sub_1400DE050:122` |
| `+0x60` (96) | last sync | constructor to 0 + `sub_1400D4BD0` compares it |
| `+0x70` (112) | close deadline | constructor to 0 + the two schedulers write it |
| `+0x98` (152) | close state | constructor to 0 + both schedulers read it |
| `+0x9c` (156) | observer flag | constructor to 0 + `sub_1400DE050:248` gates the mask with it |

**Pending risk warning.** Checksum reconciliation changes the content of the
`0x0A` ack's block, which is the most verified message we have (13 captures byte for byte). None of those
captures included the checksum-block case, so the change **cannot be validated locally**.
It is the part of this batch that would justify a capture before being taken as good.

## 18. Review of plan v4 (2026-09-27)

`REIMPLEMENTATION_PLAN (4).md` is the fourth revision. The answer, with the verification of each point,
is in `E:\Descargas\RESPONSE_TO_REIMPLEMENTATION_PLAN_V4.md`. This section records what
actually changed.

The plan arrives with a methodological observation about our previous round, and it is right:
**we argued from the consumer when the binary itself refutes that argument.** We claimed that
the `0` of the verdict was reachable because `sub_1400DF920` branches on it. But that same consumer
also branches on the `2`, which the producer never generates, so branches on impossible values do
exist here and the inference was not worth anything. The table is still correct, but the proof now comes from the
producer.

### 18.1 Verification from the producer (`sub_1400DAAF0`)

Reading the actual assignments of the return register:

| Nested type | Verdict | Evidence in the producer |
|---|---|---|
| `2` `RequestJoin` | `0` | falls into `LABEL_128` with `v21 = 0` (set at 162, never reassigned) |
| `5` `RequestJoinDenied` | `0` | same path |
| `3` `IntegrityCheck` | `1` | `goto LABEL_103` |
| `6` `AddPeer` | `1` | `goto LABEL_103` |
| `7` `RemovePeer` | `1` | `goto LABEL_103` |
| `9` `GameData` | `0` only if it was queued as `0xFE` **and** there is a session; otherwise `1` | `*(_BYTE *)(v19+72) != 0xFE` -> `1`; with a session the path ends at `break` with `v21 = 0`, without a session it falls into `LABEL_103` |
| `4`, `8`, and anything else | `1` | the outer `switch` only lists `2,3,5,6,7,9`; the rest falls into `default:` -> `LABEL_115` -> `1` |
| precheck failure | `1` | `sub_1400EF390(a2) == 0 -> return 1` |

**The verdict `2` is unreachable, and it is now proven from the producer.** The only
`v17 = 2` (line 153) is guarded by `v63 == 8`, but the outer `switch` has no `case 8`, so
control never reaches it. The decompiler kept the branch because it is in the source's `if`, not
because it is reachable. This confirms the finding in §17 and closes the doubt left open there.

One row **could not** be confirmed: "no nested message, or scratch-buffer failure -> `1`". We verified the
precheck and the unknown-type `default`, and both give `1`, but the deserialiser's failure return
`sub_1400ED400` is not visible at this call site (if it failed, `v19` would be dereferenced with no
guard, so the most likely thing is that it throws).

### 18.2 `LobbyConnect` does not create sessions; `RequestJoin` does

`sub_1400EA680` is `OnPeerRequestJoin` with no ambiguity (it throws
`"OnPeerRequestJoin - Server fatal error session == NULL"`): it rejects if `peer+8 != 0`, it requires that if the
peer already has a session it matches `{msg+104, msg+112}`, and if not, it reserves `0x938` bytes, builds it with
`sub_1400E1DE0` and inserts it.

`sub_1400EB5F0` (the connection handler) **creates no session at all**: it looks it up, and
additionally `sub_1400E5CB0(session, msg+72)` locates the peer **by `secret`**; with mode `msg+80 != 0` it does
reconnection, swapping the socket (`peer+24` <-> `conn+24`) and resending stored frames, with the
exact log strings (`"OnReplacePeerSocket, sock delivered titleID=..."`,
`"...the first frame requested has already been deleted"`, `"...that is an invalid range"`). With mode
`0` it does an **echo of the incoming packet** as the ack, whether there is a session or not. The echo-by-echo ack **we already did**.

**The tree's comparator sorts by SIGNED `i16`** (`sub_1400EA680:70-81`, `v10[16] < v9` with both
operands `__int16`). So the plan's point that "signedness is not needed because the lookup
is exact equality" is incorrect: it does not affect the lookup, but it does affect the **iteration order**, which is
observable. `SessionKey.TitleID int16` is the faithful choice and it stays. `0xFFFF` sorts as `-1`, that is
before title `0`.

**Consequence in the code (`8b42e34`).** In `handleStartObserve` we defended that a `0x18` request
"cannot match by construction" because "a real titleID is never `0xFFFF`". That is false: a client
that does not set `RL_TITLEID` sends `-1` in its `LobbyConnect`, its session is indexed by `(-1, sessionID)` and the
legacy `0x18` request —which also leaves `-1`— **does** find it. Our own WS client sends
`0`, and that is why the test passed. The *logic* was correct (lookup by pair, `titleID = int16(-1)`); what
was wrong was the premise and the test. Both were corrected, with a new case for a client without
`RL_TITLEID`.

### 18.3 The `0xFB` notice goes out empty, and the replay uploader uploads nothing

**The uploader never serialises.** `sub_14010F4F0` is `ReplayUploadManager::OnReceivePacket`
(source `D:\a\1\s\cohoreplaytool\Src\ReplayUploadManager.cpp`). Inside the scope
`"OnReceivePacket - PackReplayData"` it only reads `{titleID, sessionID}`, formats
`"failed PackReplayData replay data titleID=... sessionID=... seq N dropped"`, sets the flag with
`sub_1400E7610(session, 1)` and leaves. **There is no serialisation or upload call anywhere in the whole function.**
Therefore D.7 is moot and our `internal/replay/uploader.go` is **new functionality, not
parity**; §16.2 ("mostly resolved") was misleading and has been relabelled.

**The `0xFB` does not carry a payload either.** The plan takes it as byte-exact
`[len u16][0x02][count u16][entries]`. I cannot support that. `sub_1400E4940` builds a list of 12 bytes
per peer (`u8 peerID` + a pair of 8-byte values, lines 131-146) and **never passes it to the serialiser** (it is
freed at 316-326). What it does send is a `GameDataMessage` (type `9`, `v44 = 9` at 159) with `0xFB`
as its first byte (`v49[0] = -5` at 160), a `u16` length (222-225) and the blob if it is different from
zero (229-237). And the blob's buffer is initialised to `{0, 0}` at 157 and **never written**: nothing between
there and the read at 222 gives `&v47`. The length is always 0 and the `if` at 229 is never taken.

**The original emits exactly `FB 00 00`.** The plan's `0x02` marker, the `u16` count and its entries (4)
exist only in that dead list. Our previous `appendTail` (`[u16 len][0x02][count u8][0]`)
was a third variant that matches neither of the two; it has been withdrawn from the notice's path.
It is the same pattern as the uploader: a feature whose payload assembly was compiled out,
the signalling left intact.

**The `0x0A`'s inner queue is still unverified.** `sub_1400E2590:346-353` adds `2 + blobLen` bytes
when `version > 1`, with `blobLen` from the `MultiGameDataMessage` serialiser's own buffer. The
`[u16 len]` prefix is confirmed by the size accounting, but the filling of that buffer lives in
`sub_1400E34F0`, which **we have not read**. The current block is kept because it works against captures,
but it is marked as ours and not as parity.

### 18.4 The tick's order: it was our bug (`b59ef37`)

`sub_1400E2590` is the per-session tick, and this is its order: `sub_1400E4EC0()` (pending observers) ->
`StartObserveResponse` queued (type `0x19`, code `4`, with `"Observer session not available"`) ->
one-time start-up and reservation of the frame store (`sub_1400D4E10`, capacity
`ceilf(max(peers,8) * float@+2168)`) -> build and send the `MultiGameDataMessage` (type `0x0A`) to each
peer (362-373) -> **`sub_1400E4940` up to 10 times (374-381)** -> tick counter.

`2a6645f` reconciled **before** building the frame, so it sent the merged entries in the
same tick. Wrong: what reconciliation merges can only reach the peers in a later tick.
Besides, the return code was inverted: the original **stops** when reconciliation says "there is
nothing left to do" (`1`) and **counts and continues** when there is a desync (`2`); we stopped on the
discrepancy. Now `ReconcileChecksums` returns `0`/`1`/`2` with those meanings, the desync is
broadcast as a GameDataMessage `FB 00 00` and counted in `Session.DesyncCount` (`session+2144`).

And one detail that `b59ef37` also had wrong, corrected in `14c1eb2`: **the discrepancy also
consumes.** The deletion loop of `sub_1400E4940:342-355` is reached by both exits, the agreement
one with `goto LABEL_70` at 104 and the discrepancy one by falling through. The original pops one entry from each peer
per round whether they match or not; `b59ef37` returned the desync without consuming, and its comment
affirmed exactly the opposite. Consuming on a mismatch too is what stops a stuck peer from blocking the
queue, and it explains why the caller can bound it to ten rounds.

Two end-to-end tests asserted the wrong order on the same wrong premise
(`TestBsProtocolGameDataAcks` and `TestHandleLobbyChecksumAckEdges`); all three scenarios shift by one
tick, which is exactly the expected effect of the fix.

### 18.5 What is documented and not done

- **The `LobbyConnect`/`RequestJoin` divergence is not corrected, on purpose.** Our captures are of the
  **relay** connection, where creating the session on connect is plausibly that
  component's correct role; the main server creates it in `RequestJoin`. Moving creation to `RequestJoin`
  would break the only flow we have seen work end to end, on the strength of a
  claim about another component.
- **"Real clients send `RequestJoin` on every host and join"** clashes with our capture `_7`
  ("the game never sends a join"). Both can be true if the client only sends it in the
  lobby `HostAsync`/`JoinAsync` flows, which our captures did not exercise. They are different components and flows,
  so they are not treated as mutual refutations and **the forwarding is not changed** for this.
- **The `0x14` echo budget** (100, confirmed by the constructor) still has no clear consumption site.
  Plan (4) offers a direction (`sub_1400DE050` and the decrement at `0x1400dec92`); it remains
  to be verified before coding.
- **Full reconnection** (lookup by `secret`, socket swap, frame resend) is
  confirmed in the binary and absent here. It is new work, not a correction.

## 19. Round 5: closing what §18 left pending (2026-09-27)

Following what §18 marked as unverified. Everything here comes from reading the binary, not from
the captures.

### 19.1 The `0x0A`'s inner queue was our guess, and no longer is

`§18.3` said the block `[u16 len][02][N][00][entries]` was unverified because `sub_1400E34F0`
filled it and we had not read it. Read. The chain is:

- `sub_1400E34F0:351-360` builds an `RLink::CheckSumMessage` and adds it to the body as
  `[u16 len][blob]`, adding `2 + blobLen` to the frame's size when `version > 1`
  (`sub_1400E2590:346-353`). Confirmed.
- The blob is serialised in `sub_14007D550` -> `sub_14007EAA0`, and the latter does
  `v6 = ((char *)a2[1] - (char *)*a2) >> 3` and `*(WORD *)(v5 + v7) = v6`: that is a **`u16` holding the number
  of 8-byte elements**. Each element is written with `sub_14007FFA0`, which writes 4 bytes, that is
  **two `u32` LE per entry**.

The real layout: `[0x02][u16 count LE][count * 8 bytes]`, with the `[u16 len]` prefix.

**Our `[0x02][count u8][0x00]` is the same byte sequence for every `count < 256`**, because
`u8 N` followed by `0x00` is exactly `u16 N` in LE. The format was already correct; the only thing that was
wrong was the note marking it as unverified. It is replaced by the real layout and a
regression that reads it back as a `u16`. The only possible divergence is `count >= 256`, unreachable with the
cap of ten rounds per tick.

Incidentally the sub is confirmed: `v24 = payloadLen + 3` and the record's first byte is the peerID
(`sub_1400E34F0:149, 187`).

### 19.2 The sync timeout was more aggressive than the original's

`sub_1400E34F0:92` and `sub_1400D4BD0`. The check is made **only on the first round** of the tick
(`a4 != 0`; the following calls pass `0` and skip it), and only for a peer that meets the three
things: it exists with `peer+0x5A != 0` (it takes part in the already-started match, line 89), its **data queue
is empty** (line 92) and `now - peer+0x60 > override` with 60000 ms by default.

We only checked the deadline and for all peers, which reaped in two cases the
original respects: a peer with accumulated backlog, and a peer of a session that never started. Now both
conditions are respected. Commit `ef5bff1`.

### 19.3 StartObserve: state `4` is a different rejection from `2`

The cascade of `sub_1400E3B50:78-403` has one constant byte per branch:

| State | Constant | Condition |
|---|---|---|
| `3` | line 93 | `maxObservers > 0`, `session+2332 != 0`, observer list under the cap, recording and with a frame |
| `0` | line 170 | the list has room but it fails the replay gate |
| `2` | line 295 | the observer list is already at the cap |
| `4` | line 403 | `maxObservers <= 0` or `session+2332` clear, that is observers deactivated |

We mapped both the cap and "not reconnectable" to `2`, so we sent `2` where the
original sends `4`. Separated in `1715888`. The `4` is also emitted by the tick (`sub_1400E2590:116`,
`v81 = 4`, with the log `"Observer session not available"`).

**Plan (4) is wrong at both ends**: it says `0` is the accepted one (it is `3`) and does not mention
the `4`. Its table also has neither `1` nor `5`, which appear nowhere.

### 19.4 The `0x14` echo budget: closed

The point §17 left open ("the place that consumes it does not appear") is resolved. `sub_1400DE050:
695-726`:

```c
case 0x14:
    log("LOBBYPING from=%d")
    if ( peer+0x28 != 0 ) sub_1400E61B0(session, peer, packet);   // echo always
    else if ( (v60 = peer+0x40) > 0 ) { peer+0x40 = v60 - 1; echo(); }
```

They are **two branches**, not one condition with two parts as the plan said. The budget is `peer+0x40`,
which the constructor sets to 100, it is only consulted without a session, and on reaching 0 the ping is left unanswered without error or log. `sub_1400E61B0` echoes without a counter. Implemented in `f2e956e`.

### 19.5 The "compression of the nested one with flag 0x04" is a wrong reading

The plan claims that clients compress the nested PeerMessage with the `0x04` flag and that it has to be
inflated to classify. The flag comes out of `sub_1400DAAF0:75`, `(*(_BYTE *)(a2 + 64) & 4) != 0`, and
`a2` is the peer. But **`peer+0x40` is the echo budget**, the same field we just
closed in 19.4. That is, `& 4` is the **bit 2 of a counter**, which toggles as it decrements from 100:

```
100 = 0b01100100 -> bit2 = 1
 99 = 0b01100011 -> bit2 = 0
 95 = 0b01011111 -> bit2 = 1
```

And `sub_1400EF1F0`, which extracts the type, does exactly `*a2 = *(u8 *)buffer[0]`: it reads byte 0 with no
inflating of anything.

**No inflate is needed, and we do not implement it.** Reproducing the artifact (routing the payload through
`sub_1400ED010` one ping in five, because of a bit of a counter) would be absurd. It is the same
principle that made us distrust the `titleID`: a field read in a way that does not match its
meaning. Besides, our 13 captures never showed a byte 0 that was not the type, so the
path is not exercised in the observable flows.

### 19.6 `0x1D` was not a no-op, and `0x1E` was already right

While reviewing plan §3.10's layout table (the one we had never looked at) a real
gap appeared. `sub_1400EC5C0`, reached from `sub_1400DE050:865-897`, does four things with
`SetObservableParamatersMessage` and only while the peer has a session and that session **has not
started recording yet** (`session+2141 == 0`):

| Field | Value |
|---|---|
| `session+2344` (lag pages) | third `u32` * (1000 / `session+2172`) |
| `session+2332` (observable) | (first `u32` != 0) |
| `session+2336` | the `u64`, the credential a `StartObserveRequest` carries as a password hash |
| `session+2160` | `maxObservers <= 0 ? 60000/ticksPerSec + 1 : -1` |

Our handler only logged. The plan says players send `0x1d`, so it is exercised.
The impact on observing is real: a first `u32` of zero leaves the observable flag at zero and
subsequent `StartObserve` answers `4`. Implemented in `fc8278a`. The reader of `session+2160` is not
identified, so the value is recorded and nobody reads it; the `u64` is stored and exposed
but **not** compared against the `StartObserveRequest`'s, because there is no proof that the original
compares them.

`0x1E` did turn out to be right: the response is sent from `Session.ActivateObserver` and not from the
server's handler (which is why the handler appears to do nothing), and it matches `sub_1400E43E0` in
all four things: `[result i8][u64]`, result `0`, `-(pageCount / ticksPerSec)` without lag and
`(lagPages - pageCount) / ticksPerSec` with lag. The plan table's *lagged* branch also
adds up: the `u32` of `0x1D` are **seconds**, and the original multiplies them by ticks per
second to get pages.

And one detail from 19.1 closed outright: the checksum block's counter was written as
`[0x02][u8 N][0x00]`, identical to the real layout `[0x02][u16 N]` while `N < 256` but truncated above it.
Now it is a real `u16`, with the clamp to `0xFFFF` that the original serialiser does
(`sub_14007EAA0:23-27`). `4fc6fcb`. The bytes do not change for any reachable count.

## 20. Round 6: reconnecting, and the observable flag (2026-09-27)

### 20.1 The connect does not write the observable flag (`977c6c2`)

`session+2332` has **exactly two writers** in the binary:

| writer | value |
|---|---|
| `sub_1400E1DE0` (session constructor) | `*(BYTE *)(a1 + 2332) = 0` |
| `sub_1400EC5C0` (the `0x1D` handler) | the message's first `u32` != 0 |

`sub_1400EB5F0`, the connect handler, **does not mention the field**. And `sub_1400EBE60` only **reads** it
(line 66, the same bookkeeping check that `E9E90` and `E2590` do): my first search
counted it as a writer because of a regex that was too lax (`\+ 2332\)\s*=`, which also matches
the read after the `==`).

`handleLobbyConnect` set it from the connect's `reconnectState`, which was a wrong
attribution and made **two places fight over the same field** from `fc8278a` onwards. The
`reconnectState` selects the reconnect branch of the connect handler; it is not the observable
flag. Removing that write is neutral: if `0x1D` arrives it is the one that writes it, and if it does not arrive the value stays
at the constructor's.

**The constructor's initial value (0) is deliberately not reproduced.** `NewSession` starts with the
flag active because **it is not confirmed that the game sends `0x1D`**, and starting deactivated
would reject with state `4` every observer in any setup that does not send it. It is the highest
risk left in the tree, and the reason why `fc8278a` cannot be validated here.

### 20.2 Reconnecting by `secret` (`6884fa3`)

It was not a missing feature: it was a **defect**. `handleLobbyConnect` decided with
`sess.PeerIDForConn(client) == 0`, and for a new connection that is always true, because the
`secret` was never looked at. A reconnecting client got:

- a **second** peer slot;
- the old slot pointing at a dead connection, until the sync
  timeout (60 s) reaped it, that is a ghost player in the session;
- `nextPeerID` growing without limit;
- the `0x0A` pending list naming two slots for one player;
- and no stored frames, so it came back with no replay state.

`sub_1400EB5F0` does something else: `sub_1400E5CB0(session, secret)` locates the peer and, if it
finds it, **swaps the socket** (`peer+0x18` <-> `conn+0x18`, lines 137-139) keeping the slot,
key and every per-peer field. Implemented with `Peer.Secret`, `SetPeerSecret`, `PeerBySecret`
(a zero secret is neither stored nor looked up, so that two peers without a secret do not match),
`ReconnectPeer` (which removes the reverse index of the replaced connection, or `PeerIDForConn` would keep
resolving the dead one) and `FramesFrom` for the resend.

**What is inference and not proof**: the resume point. The range that is resent is
`[dataReceivedCount, last-1]`, and `dataReceivedCount` is a field that `LobbyConnectInfo` **already
parsed** at that offset. That that field is the frame the client already has is not demonstrated,
only that it occupies the position the original reads. If the range were a different one, the resend would be
misaligned; the slot fix is correct either way.

An unknown secret falls into the new-connection path, which is what the original does after
`"Attempt to reconnect to a non-existent sessionID"` (`sub_1400EB5F0:506`).

When the retained window no longer covers what was asked for, `FramesFrom` trims to the start of the window
instead of serving nothing, which is what `UpdateObservers` already did. The original logs one of four
different failures (`"...the first frame requested has already been deleted"`, `"...the last frame does
not exist"`, `"...that is an invalid range"`, `"...an unknown reason"`) and continues.

### 20.3 What is still not done, and why
- **Session creation in `RequestJoin` instead of `LobbyConnect`**: confirmed divergence, not
  changed (high risk, doubtful benefit, different components). See §18.5.
- **Consuming + re-rebroadcasting `RequestJoin`** (`sub_1400E6200`): the plan takes it as live, our
  captures say the game sends no join. With no capture that settles the conflict, the forwarding is
  not touched.
- **Reconnection by `secret`**: new work, not a parity correction.
- **`0x0A`'s inner queue with `count >= 256`**: resolved in `4fc6fcb`, no longer a divergence.
- **The reader of `session+2160`**: unidentified. The value is recorded but not used.
- **`0x1D`'s `u64` against `StartObserveRequest`'s `passwordHash`**: not compared because
  there is no proof that the original compares them. A candidate for the future if the read appears.



## 21. Captures _26 and _27: our server as primary (2026-09-27)

The first time the client acted on **our** output: the binary
names were swapped, so process 1 (Target 1, the one the client sees) became
BattleServerReimpl.exe and the original was left in the shadow. That inverts the CSV labels:
**main_log_NN.csv = ours**, secondary_log_NN.csv = the original.

The names are fixed by constant in cmd/proxy/config.go (exe1Name/exe2Name), so
swapping means renaming the files. The proxy has **no** way to choose a target: if
Target 1 does not connect, handleConnection returns and does not try T2.

### 21.1 _26: one player against the AI and an observer. Observing works

- 0x19 with state **3** (accepted), then 0x1E with state 0, pages, and 0x1B on exit. The
  codes of 1715888 and 977c6c2 are correct: the risk that fc8278a broke
  observing **is refuted**.
- The client **does send 0x1D**: 2 times by the player and 1 by the third connection. That is,
  fc8278a is on a live path, and 977c6c2 (the connect not writing the observable flag) is
  the right thing, because 0x1D is what writes it.
- The **player's** flow was nearly perfect: 6966 0x0A on both sides, and the sub-payloads
  **11809 out of 11809 identical**, zero divergent. The checksum queues matched in size
  (2, 13, 21 and 29 bytes), which independently validates the reconciliation's entry count.
- 0x13 LobbyRelayMsg: **765 of 766 identical**.

**The bug that did show up, and it was real**: the player's heartbeat carried the pending list
fe 00 02 01 02 00 00 (peers 1 and 2) where the original carries fe 00 01 01 00 00 (peer 1).
**3621 of 6966 bodies** carried the wrong list, because handleLobbyConnect adds a session peer for every
connection that connects and the observer was never taken out. The original treats the
observer as a passive receiver of the same 0x0A the player sees: 6966 of its 7343 bodies are
byte-identical to the player's, and it never sends 0x09 or 0x22 in that flow.
Fixed in 9858ab0.

**Correction to §19.5:** there it was claimed that the 0x04 compression flag was a wrong reading of
bit 2 of the echo budget. It is **false**: the capture shows LobbyRelayMessage with 0x04
and zlib 78 01 inside, in flight. Plan (4) was right. The practical conclusion (that a
server does not need to inflate, because internal/protocol/peer.go documents that the block is stored
and forwarded as it is) still holds, but the reason given was wrong.

### 21.2 _27: two quick matches, our server as primary

Symptom: in the first match, without the user doing anything, **several of the
player's own buildings were deleted on entry**. The second match did not have it.

What was measured, ours (primary) against the original (shadow) with the same client input:

| | client 54437 (match 1) | client 48284 (match 2) |
|---|---|---|
| 0x0A bodies | 570 vs 571 | 298 vs 298 |
| byte-identical | **563** | **290** |
| pending list | [1] on both | [1] on both |

Besides: zero 0x09 (replay pages) sent to the player, and the two matches use different sessionIDs
(1968136880 and 78064792), so there is no inherited state. During the 13 minutes of
inactivity between connecting and playing, **both servers did the same thing**: only LobbyPing
echoes, no 0x0A.

**The cause was not found.** The client only saw our output, but it was practically the same
that the original would have produced with the same input, so it cannot be attributed without a
control: **running the original as primary with the same setup**.

What IS a real and systematic divergence: the client sends 74 relays 0x13, the
original delivers 72 and we deliver 73. The 72 common ones are identical and in the same order, and our extra one
is **always the last of the match** (1049 bytes, the same content in both matches). That is,
**we discard one fewer**: the original has already removed the destination peer and drops it; we still
have it. It is a teardown bug, but it happens at the end of the match, so it does not explain a symptom
at the beginning. **It is not corrected yet**: it changes the session's proxy forwarding, which is what must not be
touched without the prior control.

Hypotheses discarded by measurement: the rule that "the checksum blob is only emitted in a round
without subs" (from sub_1400E34F0, `v9 == 0`) is **false**: the original also emits them in bodies
with subs, and our distribution is nearly identical (89/361/17/103 vs 88/365/17/101). It is not
session reuse nor replay-page leakage either.

### 21.3 Missing instrumentation (5d306c0)

secondary_log_NN.csv only recorded server -> client for Target 2, so **everything
our server received was invisible**, and there was no way to distinguish a discarded
message from one never seen. That is what blocked the diagnosis of _26's 0x14 echo.
It now also records **client -> T2**, in the same logger, only if Target 2 is
connected.

The next capture will have both directions of Target 2 and will answer several of the
open questions by itself.

### 21.4 What is missing and only a capture answers

- **The control**: the original as primary, the same match. Without it, _27's symptom cannot be
  attributed.
- **Savegame and replay** from a match with the symptom: so far there is only the server stream, and
  the simulation state is needed to confirm or discard it.
- **A disconnect halfway through a match**, which validates 6884fa3 (the reconnect), today
  completely unverified.
- **A forced desync**, which validates the 0xFB notice and that reconciliation consumes on a mismatch
  (14c1eb2).
- **0x14 echo**: in _26 the observer got 2 of 26 and the original 26 of 26. Suspicion of
  f2e956e, **not isolated**. With the client -> T2 direction now recorded it is possible to see whether T2
  received all 26.
- **A match created from the lobby** (not by a direct connection): it is the only thing that can settle whether
  the client ever sends RequestJoin, the only architectural question that has not been touched.
## 22. StartObserve: the condition was inverted (2026-09-28)

### What was broken

The client could not be observed with our server as primary. Captures _28 and _29
confirm it without ambiguity: they asked to observe and got `StartObserveResponse = 0x03` with
our binary, while the original answered `0x00` at the same instant.

The cause is at `sub_1400E3B50:83-85`, which requires three things to serve a request:

1. `session+2141 != 0` -> the session is recording.
2. the frame store has content (`store+48 != 0`).
3. `**(int **)(store+80) > 0` -> the **first retained frame** has already advanced.

Our code had the third one backwards:

```go
// before: wrong
if !s.Recording || s.replay == nil || s.replay.Base() == 0 || s.replay.Base() > 0 {
```

That is, we rejected the observer **exactly when the store had already evicted frames**,
which is exactly when the original accepts it. We accepted while the store had never expired, that is,
with a replay that did not yet cover the start of the match. Hence the symptom on the
client: the early data was not there.

### The knock-on effect of the unlimited store

Fixing the condition on its own would have broken everything, and that was the real risk. `AddPage` created the
store **with no limit** when `maxPagesPerMatch` was 0, which is the default. With an unlimited
store, `Replay.Base()` **never advances**: it stays at 0 forever. Inverting the gate
would then have meant rejecting observers **permanently**.

The original does not have that problem because its frame store **is always finite**:
`sub_1400E2590:235-249` reserves `ceil(peers * f(session+2168))` and never grows. Since the store
is never infinite, the first retained frame advances and the gate opens normally.

### The two changes

1. `StartObserve` rejects while `Base() == 0` and accepts when `Base() != 0`.
2. `AddPage` uses a finite store also in the default case. An explicit `maxPagesPerMatch`
   wins when it is given; with 0 a derived value is used.

The default **64 is measured, not deduced**, and that is how it is written in the code.
`f(session+2168)` is still not identified, but two captures bound it: the original
was still at `firstFrame == 0` at 36 pages (capture _29) and already had it `> 0` at 1818
(capture _26), that is capacity in `[36, 1818)` with 8 peers. 64 falls inside. For the exact value you have to read `sub_1400D4E10` and
the frame store object.

Accepted side effect: `-maxPagesPerMatch 0` no **longer** means "no limit", because
an infinite store is exactly what breaks observation. The flag's help text was updated to
"derived default" in `internal/config/config.go`. The option's name does not change.

### Tests

The suite was dragging the old rule along, so it had to be migrated whole instead of patching the
first failure that appeared:

- The tests that expect an **accepted observer** now fix `maxPagesPerMatch` at a small value and
  add exactly enough pages for the store to have expired before `StartObserve`. Each one
  carries the value its own assertion needs: `1`, `2`, `8`, `10` or `250`.
- `TestObserverStartObserveDeniedAfterEviction` and `TestObserverNotStalledByEviction` had
  **inverted** expectations (accept without wrap, reject with wrap). They were rewritten: now
  they reject before expiry and accept afterwards.
- `observerTestServer` in `internal/server/handlers_test.go` mounts the server with
  `MaxPagesPerMatch: 1`, so two `AddPage`s are enough to reach the observation state.
- `TestWebSocketObserverFlow` moves to 4 pages with cap 3: the store expires and the retained window
  still has the 3 pages the observer expects to receive.

`internal/session` and `internal/server` green, including `-race`.

### What is still open

- **The exact value of the capacity**, which is still an estimate inside an interval.
  It does not change the behaviour (the gate opens as soon as the store expires, which is what the
  client asked for), but if you want the exact figure: `sub_1400D4E10` and whoever writes
  `session+2168`.
- **_27's building symptom**: unattributed, pending the control with the original as
  primary.
- Reconnecting, forced desync and the `0xFB` notice: unverified end to end.

### Added: the exact value of the capacity (closes point 2 of the 22)

The chain is **complete and verified**, not inferred:

```
main:346        v43 = 1.0                                   <- default
main:617        f = (float)atoi(-maxPagesToStorePerSession) * 0.125
main:818        sub_1400D2F40(..., LODWORD(v43), ...)        <- bit pattern
D2F40:73        sub_1400E9DC0(..., a6, ...)                 a6 = LODWORD(v43)
E9DC0:29        *(DWORD *)(settings + 28) = a5              28 = 7*4, that is a1[7]
EA680:109       sub_1400E1DE0(..., a1[7], ...)
E1DE0:42        *(DWORD *)(session + 2168) = a6
E2590:243       v21 = ceilf( (float)peers * *((float *)a1 + 542) )   542*4 = 2168
E2590:249       sub_1400D4E10(store, v23, ...)               v23 = ceil(peers * f)
```

That is `size = ceil(peers * f)` with `peers` minimum 8, and **`f = 1.0` by default**. The float
travels as a bit pattern and is read back as a float, so the default survives intact.

#### The earlier 64 was wrong, and so was the reason for the error

An interval `[36, 1818)` was estimated from the fact that the original kept rejecting at 36
pages. **That reasoning was invalid**: the original's `status 0` covers all THREE
conditions of `sub_1400E3B50:83-85`, and the rejection was attributed to `firstFrame == 0` with no way to
tell it apart from "not recording" (`session+2141`) or "empty store" (`store+48`). An interval was built
on an unverified assumption. With `f = 1.0` the real capacity is **8**, and
the first retained frame advances on the ninth page.

That 64 gave the same capacity (8) was a coincidence: `8 * 64 * 0.125 = 8`. The number was
right for the wrong reason, which is worse than simply being wrong.

#### The four changes

1. **Default 8, not 64.** It is now expressed as the original's default (`f = 1.0`) instead of
   as an odd treeshot number.
2. **`math.Ceil` instead of truncating with `int()`**, to match `ceilf` and not only when
   the product comes out whole.
3. **Capacity and window unified.** The original passes **one single size** to the store
   (`E2590:249`), and there is where the eviction window is: there are not two numbers. `NewReplay(size, size)`.
   With the previous formula they were two different values for >8 peers; now they are the same, as
   in the original. Since `peers` has a minimum of 8, for every real match (2 players) both
   paths coincide exactly, so the change does not alter practical behaviour.
4. **The flag's help text and the logs** updated: the flag scales by `peers/8` and `0` is the original's
   default, not "no limit".

#### Tests

`TestSessionAddPageUnlimited` encoded the old "0 = unlimited", so it was rewritten as
`TestSessionAddPageDefaultIsFinite`: it checks that with the flag unset the store retains 8 of
10 pages and that `Base()` has advanced to 2. It is a better test, because it documents the real
behaviour and not a withdrawn assumption.

The observer tests' explicit caps (`1`, `2`, `3`, `8`, `10`, `250`) still give
exactly the same sizes, because with `peers = 8` the `0.125` factor cancels out. No
eviction test moved.

Verified with build, vet, the full suite and `-race`.

### Correction of the 22: the acceptance byte is 0, not 3

Section 22 was **wrong from beginning to end** and is rewritten here. Three
hypotheses in a row that turned out to be false, for reading the meaning of a jump and
building on top of it instead of using the bytes:

1. "The third condition of sub_1400E3B50:85 was inverted". False.
2. "`store+80` is the first retained frame and you have to accept once it has advanced". False.
3. "The interval [36, 1818) bounds the capacity". Baseless: the original's `status 0`
   covers all THREE conditions, so a rejection at 36 pages does not
   prove the store had not expired.

#### The method that was missing: ground truth from two captures

Captures _26 and _30 have the binaries in **OPPOSITE** positions, and the proxy log's
`Server::Start` marker (present in our build) confirms it in the two setups separately:

|  | _26 (21/08, AoE3DE) | _30 (28/09, AoE2DE) |
|--|--|--|
| Process 1 -> Target 1 -> `main` | **without** `Server::Start` | **with** `Server::Start` |
| Process 2 -> Target 2 -> `secondary` | **with** | **without** |

That is: in _26 `main` was the ORIGINAL, and in _30 `main` was us. Comparing
_26 against _30 as if they were the same binary invalidates every table that was
made before.

Recomputed with the correct identity:

| capture | the client's `0x1C` | `0x19` | `0x1E` | pages | result |
|--|--|--|--|--|--|
| _26 | `c850a26b000018008bce7129c5c0428c0000000000000000` | **00** | **00** | 7343 | IT WORKS |
| _30 | `d96a4b6300001800070171bfc60c37e30000000000000000` | **03** | **none** | 390 | IT FAILS |

**The original answers 0 and still serves the observer.** The client sends
ActivateObserver, receives `0x1E=00` and 7343 pages flow, including the initialisation
burst (18 KB + 18 KB + 23 KB + 16 KB) which does not exist in _30.

We answered `3`, the client **never reached sending the `0x1E`** (it does not
appear anywhere in the capture) and from there the "there is no data from the start of the
match". The `3` is the direct cause.

#### The function's real shape

The `0` is at once **"there is no replay" and "accepted"**: the original puts the same byte
on the wire in both cases, and the only thing that distinguishes them is whether the client ended
up in the observer list. That is why the acceptance arm is decided with
`HasObserver(client)` and not with the status value. Before it was decided with the
status, and on returning 0 the `case 3` stopped running: the peer stayed in the
session (`PeerCount = 2`) and ActivateObserver timed out. The suite exposed that on
changing the value, not a capture.

The arms 2 (list full) and 4 (observers deactivated) are passed without being touched:
they are the original's other two constants, `sub_1400E3B50:295` and `:403`.

Commit `4158335` is undone entirely.

#### The golden test, and why it was the missing part

All of this investigation was 20-minute capture cycles. With the client's
input and the original's output already extracted, it is a test:

`internal/server/golden_observer_test.go` pushes the two captures' **real** bytes through the
real `dispatchMessage` and checks

1. that the answer is `0x19 = 0` (the original's byte),
2. that the client ends up registered as an observer and in the list,
3. that the following `ActivateObserver` gets `0x1E` and the observer starts being
   served, which is exactly where _30 died.

It has been **verified to be falsifiable**: injecting a `3` into the acceptance arm
makes the test fail with the right message, and restoring it passes. A test that cannot
be broken is not worth having as a net.

Acceptance is tested by state (`client.observer`, `HasObserver`), not by the
byte, precisely because the byte is 0 in both cases.

### Capture identity: decided by content, not by file name

The file name turned out to be useless and had already failed in both directions:
a `BattleServerOrig.exe` contained **our** build on 2026-09-28, and on
2026-08-21 it contained the original. That is the reason for the three false hypotheses about the
StartObserve byte: we were comparing without knowing which was which.

#### The fingerprint

Identity is decided by reading the binary:

- **Ours**: a Go binary, by the markers the toolchain always embeds
  (`Go build ID:`, `runtime.main`, `go1.`) and which the MSVC original does not have.
  Checked on the real executables: the 11.2 MB one has them, the 3.2 MB one does not.
- **Original**: MSVC markers.
- **Unknown**: neither one nor the other, and nothing is concluded.

Discarded as a fingerprint: `Server::Start`. **It does not appear in the bytes of either**
executable, it is generated at runtime, so it serves neither to identify
a loose binary nor retrospectively.

What did work for the history is the start-up banner in `proxy_NN.log`,
confirmed in two opposite setups. For new captures it is no longer needed: the proxy
identifies itself before launching anything.

#### Automatic from now on

`cmd/proxy/identity.go`, invoked from `main` **before launching the processes**:
it writes `roles.json` next to the capture with each side's identity, sha256 and size,
and leaves an unambiguous line in the log. If a side comes out `unknown` or if the
two sides turn out to be identical, the manifest carries a warning, because that
capture is not usable for comparing and must not look decided.

#### The previous 27 captures

`scripts/audit_captures.py` reconstructs them from the `proxy_NN.log` banner.
All 27 are decidable, **so nothing has to be deleted**, and the 27
`roles_NN.json` are written.

The script **independently reproduces** the two points that were deduced
by hand, which validates the method: `_26` -> ours=`secondary_log_26.csv`, and `_30` ->
ours=`main_log_30.csv`.

And it shows why the confusion kept repeating: the swap alternated many
times over the history.

| ours in... | captures |
|--|--|
| `secondary_log_NN.csv` | _6 to _12, _15 to _20, _22 to _26 |
| `main_log_NN.csv` | _13, _14, _21, _27 to _30 |

### scripts/deploy.ps1: deployment with revert

The two slots the proxy starts are fixed (`cmd/proxy/config.go`):
`BattleServerOrig.exe` -> Target 1 -> `main_log_NN.csv`, and
`BattleServerReimpl.exe` -> Target 2 -> `secondary_log_NN.csv`. The script decides
which binary goes in each one **by content**, with the same markers as
`cmd/proxy/identity.go`, and then **reads back what is on disk** to
verify it. A wrong deployment is visible in the table, not two hours later.

Default arrangement: **main = ORIGINAL, secondary = OURS**. With `-Swap` it is
inverted, which is the comparison capture mode (like _27 to _30).

| parameter | effect |
|--|--|
| (none) | builds `cmd/battleserver` and `cmd/proxy`, copies safely, places and verifies |
| `-Swap` | main = ours, secondary = original |
| `-Revert` | restores the newest timestamped copy and exits |
| `-SkipBuild` | uses the binaries already built in `build/deploy` |
| `-WhatIf` | dry run |
| `-Title AoE3DE`, `-GameDir` | a different game or path |

Details that cost blood:

- **`Get-FileHash` inherits `-WhatIf`** through the provider, so in a dry run
  there was no hash and it blew up. The hash is computed in pure .NET.
- **The antivirus blocks a freshly copied `.exe`** for one or two seconds. Without
  a retry, the revert failed with "in use by another process" with nothing
  running. `Copy-FileWithRetry` retries 6 times and the final error says it is
  almost certainly the antivirus.
- **The build directory is created with `[System.IO.Directory]::CreateDirectory`**
  and not with `New-Item`, so that a `-WhatIf` does not block compiling: otherwise the
  dry run cannot show what it would place.
- `Show-State` is defined **before** being used, not at the end: in PowerShell a
  function only exists from its definition onwards.

End-to-end verified cycle: default deployment, `-Swap`, `-Revert`,
and the final state left is the default one (main = original).

### COMMIT.txt: which code produced each capture

A capture that does not say where it came from cannot be compared with another, and the commit is
the first thing you are asked for when two captures disagree. `BuildCommit` and `BuildDirty` are
injected via `-ldflags` from `scripts/deploy.ps1`, and at proxy start-up `COMMIT.txt` is written
next to the CSVs with the commit, whether the tree was dirty, the time and the
sha256 of the two binaries with their identity. Plain text on purpose: it has to be
readable with `cat` on any machine three years from now, without this tool.

`COMMIT.txt` and `roles.json` stay together, which are the two things needed to
know **what ran** and **who was who**.

Checked by reading the deployed `.exe`'s bytes, not just the script's output.

### What was missing of the folder reorganisation

Requested and **not done yet**: one folder per capture with the date, and inside
`main.log` / `secondary.log` / `proxy.log` instead of loose `main_log_NN.csv` /
`secondary_log_NN.csv` / `proxy_NN.log` in one directory.

The blocker is that **the `NN` number and the names come from the manager**, not from the proxy:
`cmd/proxy/config.go` only reads `MAIN_LOG_PATH` and `SECONDARY_LOG_PATH` from the environment and,
if they are not set, uses `main_log.csv`. The manager is a compiled binary and there is no source
to change, so the proxy would have to **reject** those names and write into
`captures/<date>_<commit>/` on its own, saying that it is ignoring what it was
told. That breaks the convention the manager expects and it has to be decided
on purpose, not in passing.

Besides, for captures to become tests an exporter `captura -> testdata` is missing that extracts vectors
like the ones in `golden_observer_test.go`. There is reusable code in `scripts/compare_observer.py` (which already knows how to find the
observer and decode the framing), so it is not starting from scratch.

Both things are done together and with a clean context, in a dedicated session.

### The folder reorganisation, done

The layout is **hardcoded** and is **relative to the working directory**:

```
captures/2026-09-28_034816_65de70a/
    main.log            T1, client <-> primary
    secondary.log       T2, client <-> secondary
    proxy.log           the proxy
    metadata.json       commit, dirty, time, and who each side was
    main.meta.json      one sidecar per log, with the SAME base name
    secondary.meta.json
    proxy.meta.json
    COMMIT.txt          the same provenance, in plain text
    roles.json
```

The names **no longer carry a number**. Before they were loose `main_log_NN.csv` and
`secondary_log_NN.csv` in one directory, with the `NN` put there by the
manager, and the two binaries swapped from one capture to the next. That was
exactly the mechanism that produced the three false conclusions about the StartObserve byte:
it was not just the identification, it was that the names
meant nothing and nobody recorded it.

The proxy **deliberately ignores `MAIN_LOG_PATH` and `SECONDARY_LOG_PATH`**, and it says
so in the log. The layout is the contract, not a preference.

#### The metadata with the same name

One sidecar per log, `main.log` -> `main.meta.json`. The property fixed in the
test is the important one: **the file name does not imply the side**. `main.log`
can hold our traffic or the original's, and what says which is the metadata's
`identity`, not the name. The test deliberately inverts the sides so that a failure here shows up.

#### Tested for real

By starting the real proxy, not just with tests: it creates the folder, the nine
files, and the `metadata.json` with the correct identity on each side
(`main.log` = original, `secondary.log` = ours, which is the default
arrangement of the deployment).

### The exporter: each capture becomes a test

`scripts/export_vectors.py` walks the captures, in both layouts, and exports the
observer's real exchange: the bytes the client sent in `0x18`/`0x1C`, the
byte the server answered with in `0x19`, whether `0x1E` came back, and how many replay pages
the observer received.

What comes out, with nobody interpreting it, over the 30 captures of the history:

| capture | side | `0x19` | `0x1E` | pages |
|--|--|--|--|--|
| _26 | **original** | **`0`** | **yes** | **7343** |
| _28 | ours | `3` | no | 280 |
| _29 | ours | `3` | no | 379 |
| _30 | ours | `3` | no | 390 |
| _14 | ours | (none) | no | 4809 |
| _4 | ours | (none) | no | 934 |
| _15 | original | (none) | no | 2 |

**A perfect correlation across five independent captures**: the only `0x19=0` is also
the only one followed by a `0x1E`, and the only one with thousands of pages. The
conclusion that was drawn by hand from _30 alone is corroborated by the complete
history, without depending on reading the binary.

#### What is missing

`expected_0x19` comes out as `None`: reading the **original's** answer is missing in the secondary
log of the flat captures, which is exactly the field that would allow an assertion. The
`find_observer` of `compare_observer.py` scans with `strict=True` and finds nothing
there, because the original splits every message into two writes; there is a tolerant detector
in the exporter but it **does not manage to chain in the secondary**. With what there is, the
vectors are exported and visible, but `assertable` is 0 and **no test is generated from them yet**.

The vectors live in `src_go/internal/server/testdata/observer_vectors.json` and
`golden_observer_test.go` already covers the _26 case with real bytes, so the regression
that matters is protected even though the exporter does not close.

A Go detail that came out: `find_observer` returns a **string** and the rows' ports are **int**, so
comparing them without converting never matches.

### Discarding the capture history, and why that was right

The previous 30 captures (**113 files, 294 MB**) have been deleted. Reason: their
layout is what caused the confusion, not their content. Loose `main_log_NN.csv` and
`secondary_log_NN.csv` in one directory, with the `NN` put there by the manager and
the binaries swapped from one capture to the next. Four of them end up
contradicting each other about which was which, and from that came the three false
hypotheses about the StartObserve byte.

Kept **before** deleting, which is the only thing that cannot be recovered:

| what | where |
|--|--|
| _26's real bytes | `internal/server/golden_observer_test.go`, hardcoded |
| the table of 7 vectors | `internal/server/testdata/observer_vectors.json` (commit `56c1309`) |
| the 0 vs 3 correlation | `CAPTURES.md` and this section |
| the code | the repo |

That is: **the conclusion and the proof stay, the loose logs go.** A
294 MB history in the old format is not evidence, it is noise that invites
going back to guessing.

#### Out: the migrator

`scripts/audit_captures.py` only existed to reconstruct the identity of those
old captures from the `proxy_NN.log` banner. Without them it makes no
sense, so it is deleted. `export_vectors.py` also loses the flat-layout
support and keeps only `./captures/<date>_<commit>/`.

#### From now on: one capture, one thing

`CAPTURES.md` has the matrix. The rule that comes out of all this: **a capture that
has to answer more than one question is badly posed**, which is exactly
what happened in _30. And you always start with the original at the front, because that is
the reference; ours at the front only once you know you want to compare.

### AGENTS.md: the memory moves into the repository

The history of this session lived in the conversation and was lost. `AGENTS.md` at
the root collects what is needed to continue: the hard rules, the capture layout,
the deployment, and **the concrete pending item of the simulation harness** with the design
decisions already taken, so that they are not argued again.

There too is the lesson of this session: commit before moving files, because the
whole harness was lost by moving it without committing.

## 23. The simulation harness: done (2026-09-28)

The PENDING that `AGENTS.md` opened with ("none of this has been started") is closed.
`src_go/internal/sim/` replays a capture against our server and compares it
with the original, and `src_go/cmd/gensim/` turns any capture into a test.

```
captures/2026-09-28_040213_56c1309/   what the proxy writes
        |  cmd/gensim
        v
internal/sim/testdata/<capture>/      the capture, copied
internal/sim/sim_<capture>_test.go    generated, one function
```

### The flow, and why in that order

1. **Concatenate first, then frame.** The proxy logs TCP writes, not
   messages: a row is a frame already reassembled *if it could be*, and a raw
   `flush` when it could not (`cmd/proxy/framer.go:60`). Chaining row by row
   (which is what `scripts/compare_observer.py` did in its CSV CLI) loses the message
   when the original splits the header and the payload into two writes, which is exactly
   what `scripts/OBSERVER_CAPTURE.md` documented —runbook deleted on 2026-10-04,
   see §41.4—. Here the bytes of each direction are concatenated
   and the result is framed, once.

2. **The direction is decided by the source port, and it can only be so.** The proxy
   logs the client's upload against **its own** listen port
   (`cmd/proxy/tcp_proxy.go:265`) and the download against the child's real port
   (`:315`). The child's port is the only one of the two that is in the recorded
   args, so a row is server->client when its `src_port` is one of
   them. With the proxy's port it could not be done: both sides of the comparison
   share it.

3. **`msg_type` is not read.** It is the label the proxy itself puts
   (`cmd/proxy/packets.go:33` on top of `internal.ParseMessage`): `0x82` on a WebSocket
   frame, but the real message type on a bsPort frame, and `0x47` on the
   handshake. The source is `data_b64`.

4. **The client is replayed through a real socket**, not through
   `dispatchMessage`: the transports differ (length prefix on bsPort,
   WebSocket frames on webSocketPort, and the observer flag that changes how a client's socket is
   processed), and a comparison that skipped them would
   not be comparing what breaks. The WebSocket side is masked when writing
   because the harness reads from the log the payload **already unmasked**: the proxy
   `unmaskWS`s in place and leaves the mask and the key in the row
   (`cmd/proxy/framer.go:53`).

### `config.ParseFlags` and the four `os.Exit`

`ParseFlags` reads `flag.CommandLine` and `os.Args`, which are process-wide, and calls
`os.Exit` on four paths: unknown flag or bad value (2), `-help` (0), and the two usage
checks (1). **A `FlagSet` with `ContinueOnError` contains none of the four**,
because the exits are inside the function, not inside the
`FlagSet`. A flag with a semicolon proves it: the test binary printed the
whole usage and then died without a single `--- FAIL`.

That is why parsing goes in three steps, and that is why they are in this order:

1. A call with an **empty** command line: it only registers the flags, it
   cannot exit, and it leaves the complete set of names visible.
2. Validation over that same `FlagSet`: `ParseFlags` with `os.Args` without
   arguments and then `flag.CommandLine.Parse(args)`, which does return the error.
   The values of `-uploadVia` and `-simulationPeriod` are also read there, because the
   final call is the one that exits with `exit(1)` if they are not valid.
3. Only then the final call, with `ExitOnError` and arguments that are already
   known to parse.

A flag the parser does not know (one of the original's own) is dropped **and
reported**: our default for that knob is the closest available
reproduction, and saying so is more useful than refusing to compare the capture.

### Which of the two command lines is reproduced

With **our side's args**: the question is whether the server, run as it
was run, answers what the original answered. The original's args are printed
alongside and **their difference with ours is reported** before any
verdict, because a divergence caused by a different `-pageSize` is not a
fidelity finding. If the two sides were not launched with the same thing, the report
says so on its first line.

### What does change, and why it is declared

Every deliberate change is a `Deviation` with its reason, and what is not on that
list has to be identical to what was recorded. Today they are: the listen ports (the capture's
belong to the run that produced it), the out-of-band (it is another TLS service and
no battle message goes through it), the lobby and the replay host (they do not exist in
a test), the certificate if the file is not here, `muninLog`, and the log level at
`warning` (the captured log is not part of the comparison and it would bury the
report; the level does not change one byte of the wire).

### The three must-matches, and why only three

The byte of each `0x19`, whether `0x1E` arrives and with which status, and the **order** of the
replay pages. The order and not the number: an inverted pair crashes the client
(_5, `TestGameDataAckOrder`), whereas how many fit in a tick depends on
where the capture was cut and where the replay window ends, which are not
evidence of anything. Both directions of a missing `0x19` are must-match: the
silence in front of a byte the client expects is exactly what _30 cost, and a rule that only looks at "we answered
more" would have swallowed it.

Everything else is informative and is printed with both frames side by side. The reason is
the usual one: the original has oddities of its own (2 of 26 `0x14` in _26,
`CAPTURES.md` row 8) and a test that fails because of them ends up disabled wholesale.

### A single masked field

`0x1E` body[1:9], the estimated start time: Unix seconds plus an offset
that comes out of the page count (`internal/session/session.go`,
`Session.ActivateObserver`). It depends on when this process started, not on the
protocol. The list goes in every report so that it can be argued about. The masking
is observable: the synthetic test puts `0x1122334455667788` there and the clock puts what
it puts, and without the masking a `body-mismatch` of `0x1E` comes out.

### Verification without a real capture

None of the history's captures are left, so `sim_test.go` builds one in the
exact format the proxy writes (`internal.ParseMessage` included, so that
`msg_type` is what it would be) and `TestWriteSyntheticFixture` leaves it committed in
`internal/sim/testdata/synthetic` with its `ORIGEN.txt` saying it was written by
hand. It encodes _26's shape: `LobbyConnect`, three replay pages with
`-maxPagesToStorePerSession 2` (which is what makes the store expire and therefore what
opens the observer's gate, `sub_1400E2590:242-243` and
`sub_1400E3B50:83-85`), `0x1C` -> `0x19=00` -> `0x1A` -> `0x1E=00` -> the retained
window in order -> `0x22`.

**And it is falsifiable, which is the only thing that makes a golden test useful**: by changing in
`testdata/synthetic/main.log` the original's `0x19` byte from `00` to `03`,
`TestSim_synthetic` fails with a `start-observe-status` must-match and both
frames side by side (`original: 0x19 ... 03` / `ours: 0x19 ... 00`). Restoring it
passes. Besides there are tests that invert each rule separately, and one that swaps
the `writer`s in `metadata.json` to show that the truth comes from there and not from the
file name.

### Two generators, and why only one is left

`cmd/gensim` is in Go, inside the module, with its `-check` mode (it writes nothing and
fails if a generated test does not correspond to what the captures would produce: a
capture cannot be added without its test, nor can a test be edited by hand). A
draft in Python did the same and produced a different file, which made the Go
one's `-check` fail against the other's: two generators for the same
file is a broken invariant, and the one that can run in continuous
integration stays.

### Two sessions writing the same package

Halfway through the implementation, files appeared in `internal/sim/` that were not from this
session (`capture.go`, `config.go`, `confighelper_test.go`), and `cmd/gensim` with an
in-process child-process-parser approach. It ended up merged: `cmd/gensim` stays
(better than the draft, because of the `-check`), and the parser stayed **in process**, with
the three-step validation from above instead of a subprocess. A subprocess
would have worked too, and it would have been simpler to reason about; it is discarded
because it forces the configuration to be serialised to JSON and the test binary
re-executed for every capture, and because the real problem (the `os.Exit`s are inside the
function) is solved without them.

## THE REPORT SAYS WHEN, AND THAT DECIDES THE 11-MESSAGE QUESTION

The open question was whether the 11 extra messages we sent in the first
real capture (10 0x0A and 1 0x13 LobbyRelay out of 547) were an artifact of the
harness's settle or a tick cadence that really emits more. A total does not
answer that: both look the same. Now the clock answers it, and it comes out that **it is not
only the settle**.

### How it was measured

- The replay reader stamps every read with its arrival moment and carries it
  through to Msg.At (a scan over []chunk instead of a buffer). Before, the time
  was lost on concatenation.
- The two sides are put on **one single axis**, counted from the first message the
  client sent: Capture.origin, which is the same instant from which
  Replay() schedules that entry. Timeline() uses that same definition, so that
  there is only one.
- The two clocks do not have the same precision and the report says so: the proxy
  stamps whole seconds (`cmd/proxy/packets.go:47`), the replay has a real clock. It does
  not blur the question because **the start of the window falls on a second
  boundary** (it is a subtraction of two integer stamps): a message stamped with that
  exact second is at or after it, whatever the instant inside the second.
- Every finding carries AtOurs and AtTheirs; the connections table carries
  ours=555 (last t=1m29.38s) original=547 (last t=1m29s). The new
  [message-tail] rule splits the two streams by the window and says which of the two
  things it is, **computed from the timestamps** (judgeTail), not written by hand.
  It is informative and not must-match, and nothing was promoted: if the difference
  were from the settle it would be a choice of the harness, not of the protocol.

### What comes out on the real capture (2026-09-28_205830)

`
window [t=1m29s, t=1m29.4s]: 4 of ours at or after it opened and 503 before it,
                            against 1 of the original's and 499
per second, original|ours, only the seconds that differ: 26s: 3|7 89s: 1|4
by payload: len=1052: we sent it 1 time(s) at t=1m29s, the original 0
4 of the 8 ... were emitted BEFORE the settle window opened ... THE EXTRAS ARE
INTERLEAVED THROUGH THE MATCH, not an artifact of the settle
`

Breakdown, and the arithmetic closes exactly: 550 before against 546 of its own are **+4
before the window**; 5 after against 1 of its own are **+4 after**; 4+4 = 8.

- **+4 0x0A in second 26** (7 of ours, 3 of its). The first of the whole
  match is at t=26.13s. This is NOT the settle: it is the acks starting up.
- **+3 0x0A in second 89** and **+1 0x13 in second 89** (the 0x13 of
  1052 bytes that the client sends at 89s and the original never relayed, because the
  client disconnected just after). That one IS an edge of the capture.

### A harness failure that comes up along the way


ReplayConn received the **end** of the window in the parameter called span and
added the settle to it again: it read twice the settle that settleFor chooses (800ms
instead of 400ms). A report over a window the replay did not read is a
report about nothing, so the parameter is now called `end`. That is why the 11
extra messages are 8: three were from the non-existent 400ms second.

### Why extras are NOT located by aligning the two streams

It was tried. The 0x0A are hundreds of equal bytes in a row, the resync fires at
almost every step and declares an extra at each one: **382 "placed" extras for 7
real ones**. A resynchronisation there is a guess, and a guess with a timestamp
on top is worse than saying nothing because it looks like a measurement. In its place
**payloads** are counted, which is independent of the order, and when even that is not enough
(the real acks carry bytes that change between runs: 102 more, 95 fewer) the
report says so and counts per second, which is evidence and not
attribution. The 0x13 is located by payload, because there it is.

### Falsifiable

TestJudgeTailVerdictFlipsWithOneTimestamp: the same two streams, the same
totals, the same window, and **one** message moved 100ms across the boundary. The
conclusion has to change, and change in the direction the movement implies.
TestJudgeTailArithmeticCloses checks that the partition gives exactly the
extras in all four cases (interleaved, all in the window, we are shorter,
one side empty). TestReportSaysWhenEveryDivergenceHappened does the whole path
over the synthetic capture and **requires both verdicts in one report**: the
player's connection has to come out interleaved (the harness acks the pages at
t=0 and the window does not open until t=1s) and the observer's has to come out as a
possible artifact (everything the observer answers arrives after its last input).
No must-match rule was touched.
## 24. The second observe attempt: 70 against 803 (2026-09-29)

The capture `2026-09-29_012528_03cb67c7cf19` enters as a test (row 5 of `CAPTURES.md`,
with our binary at the front, which is the arrangement that was missing) and brings two
observer connections because the game squeezes observe twice. The squeeze
works on both: `0x19 = 0` and then `0x1E`. The failure that remains is the connections
table, and it is a single line:

    conn 1944   ws  compared  sent=2485  ours=1875  original=1873   the player
    conn 49798  ws  compared  sent=8     ours=1582  original=1353   the first observer
    conn 49823  ws  compared  sent=4     ours=70    original=803    the second

The four suspects of the brief (the initial `Progress`, the observer list,
the rolled window and the dedupe by identity) are all four
wrong. The reason is in one word: **what is stored in the store**.

### What the store keeps, measured on the original

The original's two observer connections received 1346 and 800 messages, and
**each one is byte-identical to a message of the player's `0x0A`**, in strictly increasing
order relative to it: ack #0 (the first of the match, t=22s)
up to #1345 and #799. Zero `0x09` and zero `0x22`. And the binary says the same in
the same direction:

```
sub_1400E2590:354-357   (*(store + 32))(store, &msg, size)    <- append to the store
sub_1400E2590:362-373   for (peer) sub_1400D49A0(peer, &msg)   <- and THEN to the peers
sub_1400E4EC0:79        sub_1400D5560(store, progress, progress+delta-1, out)
```

The store is fed by **the session tick with the `0x0A` it has just built**,
not by what the players upload. We were storing the `0x09` GameData body
(`handleGameData` -> `AddPage`) and serving it to the observer as `0x09`; besides we
joined it with a `0x22` end-of-match that the original never sends, and with
a second copy of the live ack (`sendAckToObservers`). Three flows where the
original has one.

### The store is a number of BLOCKS, not of frames

```
E2590:242-243   v21 = ceilf(peers * *(float*)(session+2168))   with peers minimum 8
E2590:255       sub_1400D4E10(store, v21, session+2152, session+2160)
D4E10:18-19     store[1] = session+2152    store[2] = v21
D4E10:29-30     store[3] = session+2160 != 0 ? session+2160 : -1
vmethod_20:76   v5 = a3 + 57            what each frame costs its block
vmethod_20:82   if (v5 + block.used <= store[1]) -> same block
vmethod_20:118  if (blocks created >= store[2]) -> recycle the oldest one
```

`session+2152` is the `-pageSize` (`E9DC0:22` -> `EA680:109` -> `E1DE0:41`), 65536
by default. And `f` is the `-maxPagesToStorePerSession` with **f = 1.0 by default**
(`main_1400d1310:346` initialises it, `:617` divides it by 8), that is **8 blocks
of 64 KiB: half a megabyte of ack**.

The capture's match is **259651 bytes of ack in 1570 frames** (measured on
the bytes, not estimated), that is four blocks. The original's store **never**
got to discarding its first frame. We, with a ring of 8 FRAMES, discarded
at the ninth, and an observer arriving late had only eight frames to
chase: that is the 70.

`session+2160`, the field that had been commented for months as "reader not
identified", is the store's cap in FRAMES, and the game's `0x1D` writes it:
`-1` as soon as `observerMax > 0`, which is always the case. In a real match
it bounds nothing.

### The `0x19` condition was inverted, and the byte did not say so

The **3** arm (`sub_1400E3B50:87-160`) builds the response and returns: it **does not
touch the observer list**. The **0** arm (`sub_1400E3B50:162-254`)
builds the response, sets `client+156` and **adds the 16-byte entry** with
`progress 0`, `active 0` and the lagged flag.

That is, the 0 is the arm that REGISTERS and the 3 is the one that does not, the opposite of how it was
written. The wire byte does not distinguish them because it is 0 in both cases: what
distinguishes them is whether the client ended up in the list, which is exactly why the server
decides with `HasObserver`. This capture's 0 fits entirely: its store had not
discarded anything, hence `**(store+80) == 0`, hence it went in through the 0 and ended up
registered. And so does _30's 3: the store had expired, it answered 3, it did not
register it and the client never sent the `0x1A`.

The "0 works, 3 fails" correlation of section 22 was correct; its explanation
was not. The 0 is not "there is no replay", it is **registered**, and the 3 is not "accepted", it is "the
store already lost the start of the match". The two arrangements are a single one:
fixing the store without fixing the gate would have made us never accept an
observer in a normal match, which is the knock-on effect the 22 already stumbled on two
times.

### The rest that came out of it

- The end marker (`0x22`) is only sent with the `v5` condition of `E4EC0:97`
  (`session+2142`, which nothing in this server ever sets, or a session with no peers) and
  with a client version >= 4. With one player in the session it is not sent, and the
  capture confirms the original sent not one to its two observers. We
  were sending 86 and 4.
- `UpdateObservers`'s step is `session+2184` in TICKS. The original
  initialises it to **8** (`main_1400d1310:299`) and only converts the milliseconds flag
  **when it is on the command line** (`main_1400d1310:735-738`).
  Reading the default of 8 as ms gave 0 ticks, which is the original's own infinite loop
  (`E4EC0:199` advances the index by the interval), and the guard that saved
  it turned into 1: eight visits per second where the original does one. The
  capture measures it: the original serves 200 frames per second, in bursts.

### The harness had a hole right here

The page-order rule (`replay-page-order`) read the `0x09`, which is the
PLAYER's stream. The original sends an observer not a single `0x09`, so the
rule compared 0 against our 31 and passed. **0 must-match with the second
observer served 70 messages against 803** is exactly what a
rule that looks at the wrong type produces: it compared two empties.

Now it reads the `0x0A`, which is what the store keeps, and compares by **frame
identity** against each side's own player's `0x0A`. So it compares two
things that are comparable: at what point of the match the observer starts, and in what
order the frames arrive. Both are must-match; the number is still
informative, because how much fits in a tick says nothing about fidelity. And the
body scan now reports **one per type** with the count of the others: with
two streams of 1352 and 1346 the positional comparison produced 273 findings that
were only a consequence of the offset.

### Result

```
conn 1944   player       ours=1875  original=1873   no change
conn 49798  1st observ.  ours=1359  original=1353   was 1582
conn 49823  2nd observ.  ours= 603  original= 803   was   70

[replay-page-order] conn 49823: match frames served: ours [0..599] of 1571 |
                                             original [0..799] of 1570
```

Both windows start at frame 0, which is what was not known before. There are 0
must-match across the three captures. The second observer now gets a window
comparable to the original's: the same starting point, the same order, 200 frames per
visit, and the difference of 600 against 800 is one of the four visits the
observer lived through before the client sent the `0x1B`.

### Falsified three times

With each defect reintroduced by hand, the new test fails with that
defect's failure and not with another:

| defect | what the test says |
|--|--|
| store of 8 frames | `first attempt: expected status 0, got 3` (the whole old bug) |
| observer served as `0x09` | `the late observer got 0 frames on the first visit` |
| initial `Progress` from the queue | `the late observer got 0 frames on the first visit` |

### Tests

- `TestObserverStartObserveTwoConnections`: two connections, the second 171 s late and
  with the first already present; it is served from frame 0, 200 per visit, in order
  and as `0x0A`.
- `TestSessionRecordFrameDefaultHoldsAWholeMatch`: the match's 259651 bytes
  fit in the default store and `Base()` does not move.
- `TestObserverStartObserveArms`: the four arms of the cascade, including that the
  3 **does not** register and that the 0 registers even when the store is empty.
- `TestHandleGameDataDoesNotFeedTheFrameStore` and `TestWebSocketObserverFlow`: the
  `0x09` does not feed the store, and the observer gets `0x0A`s identical to the player's and nothing else.
- `TestObserverNoEndOfGameDataMarkerWhilePlayersRemain`: with one player in the
  session there is no end marker.

The synthetic capture at `testdata/synthetic` was regenerated with the new shape (the
observer receives the session's `0x0A`, not the uploads' `0x09`), and its
`ORIGEN.txt` says where each thing comes from.

---

## 25. The observer's zoom: measured, and not in the stream (2026-09-30)

The capture `2026-09-30_010531_a553ef8fd63c` enters as a test (row 2 of `CAPTURES.md`,
ours at the front, with an observer) and brings the symptom that was left: the observer
**joins and sees the match**, and the only thing the user complains about is the verbatim
**"the observer has a lot of zoom"**. Everything else is fine.

The brief's four lines were measured on both logs, and there is **no
difference at all**. The observer's stream is indistinguishable from the original's:

| | 0x19 | LobbyConnect / 0x1E | subs | entries | frames | visits |
|--|--|--|--|--|--|--|
| ours (`main.log`) | 0 | identical | 12312 | 4948 | 5292 | 200,200,200 and then 9/s |
| original (`secondary.log`) | 0 | identical | 12313 | 4948 | 5363 | 200,200,200 and then 9/s |

- The three cells that are not the observer's `0x0A` (`0x15` LobbyConnect, `0x19`
  and `0x1E` = `00 8a 47 bc 6a 00 00 00 00`) are **byte-identical**, WS
  binary frame with the stripped `[type][body]` payload included, and with no RLink length prefix on either side.
- The **12312 game sub-messages** come out in the **same order** on both
  sides, and so do the **4948 checksum entries**, over the whole sequence.
  Decoding the `0x0A` per the format in `internal/session/ack.go`: the
  25,856 acks of the four flows parse to the last byte, none with
  leftover bytes.
- The observer's window **starts at sub 0 of the match** on both, and
  the observer's stream is a **byte-identical prefix** of the `0x0A` of
  its own side's player. The player gets 8 acks/s and the observer 9
  (200 on the first three visits = the delay, and then 1 extra per
  second), on both sides.
- **There is no setup burst**: the original sends the observer NOTHING other than
  `0x15`, `0x19`, `0x1E`, `0x0A` and `0x14`. The message-type histograms are
  the same, and in the raw log not even a `0x54`, `0x41`, `0x00`, `0x25` nor an
  `IntegrityCheck` appears in the whole capture. The old history's hypothesis
  is **discarded by measurement**, not by oversight.

That is: **the zoom is not in the observer's stream**, and the evidence **does not
identify the cause**. Nothing in the server is touched for this.

### What really was wrong: the harness could not answer that question

`replay-page-order` identifies each frame by its **whole body** against the
player's `0x0A`. As soon as the two sides group the acks differently it stops
finding anything, and `indexOf` gives up with `false`. In this capture the report
said, verbatim:

    [replay-page-order] conn 50569 #0: the replay stream could not be located in the match's own ack stream
        match frames served: ours [0..5363] of 5369 | original none

`original none` with 5363 frames served. It is the **"compared two empties" of
the 24** all over again, and it meant that the only question of this capture went unanswered
exactly when the capture is the one with the longest recorded match.

The new rule `replay-content` (`internal/sim/replaycontent.go`) reads the stream
by what it **carries** instead of by which match frames it is: it decodes each
`0x0A`, pulls out the ordered sub-message sequence and the checksum-entry
sequence, and compares them without looking at how each side groups. Must-match are the starting
point, the sub sequence and the entry sequence; how many each side carries is still
informative, as is the number of frames. The decoder is **strict on
purpose**: a body whose lengths do not add up, or that carries something after the
block, is counted and named instead of being read as far as it parses, because a
sub sequence that looks real and is not is worse than none.

With the real capture:

    [replay-content] conn 50569 #0: the observer's replay carries the same game data on both
        sides: 12313 sub-messages and 4948 checksum entries, in the same order, from the same
        sub-message of the match. The acks are grouped differently (5364 frames against 5363),
        which is what RuleReplayPageOrder cannot see and the client does not read
        ack frames: ours 5364 | original 5363
        game-data subs served: ours 12313 | original 12313
        checksum entries served: ours 4948 | original 4948
        window starts at: ours sub 0 of 12313 | original sub 0 of 12312

0 must-match, 18 informative (17 before; the new one is one of them).

### The synthetic capture had the 0x0A written wrong

`fixtureAcks` wrote them by hand with the sub's length in **big-endian** and without
the pending byte, that is in a shape no ack has ever had. It did not
matter while the only rule that read them compared whole bodies between
themselves; `decodeAck` rejects them, and a decoder that does not read the fixture cannot be
exercised with it. Fixed and `testdata/synthetic` regenerated. **The new
rule found them:** with the length wrongly set to `04 00` instead of `02 00`
a must-match came out on the synthetic one, which is the shape it has to give.

### Falsified

The six tests of `replaycontent_test.go` fail with the rule set to `return`
and pass with it: a dropped sub, an inverted pair, a window that starts
late (_30's symptom), a different checksum entry, a different grouping that is
**not** a divergence, and an ack the decoder does not read (which is
reported as not comparable, not as a difference).

### The only thing left different, and not touched

- **Grouping**: 3447 acks with subs against 3470, and 3230 checksum blocks
  against 3118, for the same 12312 subs and the same 4948 entries. It is
  informative by definition: how much fits in a tick says nothing about
  fidelity (AGENTS.md).
- **Our `LobbyConnect` carries the same 8+4 bytes on the player's connection
  and on the observer's; the original varies them per connection.** For the
  observer our bytes are the same as its, that is, there is
  nothing to fix there. On the player's they do diverge, and the player works fine.
- We send a **WebSocket ping** (`0x65`) where the original answers with a
  **pong** (`0x6B`), once per connection, and there is no pong from the client in the log.

### Closed

The user confirms it: **the observer's zoom is not a BattleServer thing**, so
this section stays as a measurement, not as pending work. The control match with the original at the front is not made, and the lobby is
not chased for this.
### RAM and time budget for the tests, after a test that brought the machine down

An unverified harness test reached **20 GB of RAM** and had to be stopped by hand.
What was measured, not assumed:

| state | peak RAM |
|--|--|
| `internal/sim` committed, over the 1m29s capture | **49 MB** |
| `internal/sim` committed, over the synthetic capture | **33 MB** |
| unverified work (discarded) | **20,000 MB** |

That is: **the harness we use does not have the problem**. The trimming work that caused it
was left unverified and thrown away whole (`git clean`), not committed. The exact line that
grows could not be located: `matchEvents` generates about 2,600 small
events, which cannot be 20 GB, so the growth is downstream, in the
`Render()` or in `compare.go`'s diff. The diff **resynchronises at every step**, which is
exactly what makes it O(n^2), so it is a real risk of the committed code and not only
of the scratch. It remains to be measured with a big capture, not assumed.

That is why `AGENTS.md` now carries an explicit budget: **10 minutes and ~500 MB**,
no synthetic data without a cap in the code itself, and measure the working set instead
of assuming it. A test that goes past a few hundred MB is a failure even if it passes.

---

## 26. Trimming a capture: the data is shortened, the clock is not (2026-09-30)

The previous section's budget (10 minutes, ~500 MB) was the right one and
**measured**, and the time could not be touched: every capture is replayed in real time
on purpose, because the server emits `0x0A` per tick and not per input, and speeding it up
measures the harness and not the server (the 176-against-500 of `Options.ReplaySpeed`). What was
missing was the other half of the problem: **the data**.

### The arithmetic that motivated the trim, measured

With the test binary's working set watched every 200 ms, and **with nothing else
running on the machine** (it matters: a first pass of these measurements added
the Go compiler and the linker to the test's consumption, and the linker eats more than
200 MB on the first build after a change, so the figures from that pass
are not the test's):

| capture | before | after |
|--|--|--|
| `synthetic` | 5.9 s / 11.6 MB | 5.7 s / 10.9 MB (not trimmed) |
| `2026-09-28_205830` (1m29s) | 93.9 s / 16.2 MB | 93.7 s / 16.0 MB (not trimmed) |
| `2026-09-29_012528` (3m38s) | 222.8 s / 22.8 MB | **203.8 s / 24.3 MB** |
| `2026-09-30_010531` (14m00s) | **845.5 s / 54.3 MB** | **254.0 s / 17.7 MB** |
| all of `internal/sim` | 1168 s for the four captures alone | **559.9 s** the whole package |
| **`go test ./...`** | did not fit | **565.0 s / 114 MB peak, entirely green** |

The diff's O(n^2) **does not reproduce**, and this closes it with a number: the heaviest
test of the committed harness used **54.3 MB** before being trimmed and **24.3 MB**
after. The 20 GB of the discarded work was not this code, and the ~500 MB
ceiling of the budget is a factor of four above what the suite really uses.
What **was** unsustainable was the sum of the clocks: 1168 s of captures plus
the 40 s of the rest of the tree.

The 24m44s one is 1497.6 s measured, with the old sampler (which also summed the
`go` process and whatever was around, so its 259.5 MB of RAM is a ceiling, not
the test's figure). The caveat is left in rather than mixing two methods in one
table.

### What a trim is, and why it can be proven

The truth of a capture is a **recording**: the original's stream is on
disk, byte for byte. Taking its tail off leaves a **prefix** of what was recorded, and that
can be checked without believing anything: the harness compares against the recording, not against
itself. So a trim is a checkable statement of the form

> what remains is a prefix of what there was, and the verdict is the same.

`internal/sim/CheckTrim` (with `trimcheck.go`) checks it by replaying **both**
captures with **the same rules** and deciding by rule name. What it decides:

- **the verdict**: the same must-matches on the same connections, in both
  directions. Losing a finding is as serious as inventing one: a trim that swallows
  a divergence leaves the suite green over a server that is still wrong.
- **the original's client->server stream, as a prefix**: a trim that changes what
  the game said replays a different match.
- **the original's server->client stream, as a prefix**: that is the trim itself.
- **the observer's sub sequence and checksum-entry sequence, and the point
  of the match where its window starts**, as prefixes. A suffix does not move the
  head of a sequence, and the starting point is must-match (_30's symptom).
- **the coverage**: every connection compared before has to still be compared, and an
  observer that was served frames before has to keep receiving them **on
  both sides**.

The last one is what makes the failure of section 24 impossible. A trim deep
enough to leave an empty stream makes `replay-page-order` and `replay-content`
**say nothing**, and the report paints a green over two empty things. That is not a
detail: it is the reason the harness can be trimmed without lying.

The digests are **one per message** (8 bytes, `fnv64` over `[type][body]`), not one
per stream, and that is where the trick is: a tail trim leaves the digests equal in the
prefix, and one that removes a message **from the middle** leaves the same totals and changes
every digest after it. A whole-stream hash would not serve, because a tail
trim would change it too and the proof could not tell them apart.

Comparing by digests instead of by messages is also what makes this cheap in
RAM: `Report.Measures` is 8 bytes per message, and two 24-minute captures in
memory at once would be exactly the path to the 20 GB of the discarded work.

### The tests are at two levels, because each one sees something the other does not

`trim_test.go`, without sockets: it fakes the shapes a replay would hide. An empty
stream (in all four variants, including the one only visible by the content: ours
keeps the handshake and runs out of replay, both totals are still different
from zero and only the content detects it), a moved window, a changed sub or checksum
entry, a swallowed finding, an invented finding, a connection the
trim took away, a connection the
trim invented, and a trim that makes the server say
more than it had said. And the control that a capture compared
**with itself** passes, which otherwise the checker would be measuring its own
bookkeeping.

`trimreplay_test.go`, with sockets: a three-second capture with _26's
shape, its trim to two, and both replayed. There a trim that slackened for the wrong reason looks
**the same** as one that slackened for the right reason, and only two real runs
tell them apart. And the negative: a trim to 500 ms takes the observer away and
`CheckTrim` has to reject it.

The hardest evidence does **not** come from the harness: **git** gives it. The four trimmed logs
differ from the originals in **zero added lines and 47,929 deleted**
(`git diff --numstat` over the four files: 0 and 452, 0 and 450, 0 and 23479, 0 and
23548). That is, they are a **literal prefix of the recording**, with nothing rewritten, reordered
or re-encoded. What remains is a prefix of what there was, and that is not a
promise of the trimming code: it is what the diff shows.

### What is trimmed and what is not, with the evidence for each

| capture | trimmed to | why there | verdict |
|--|--|--|--|
| `2026-09-29_012528` | **3m20s** (of 3m38s) | the second observe attempt arrives at 3m13s with its `0x19` and its `0x1E`, and the two observer connections go quiet from 3m14s | `-verify` **ACCEPTED**: the two observers with **exactly** the same messages as untrimmed (1353 and 803), and the window at sub 0 on both sides |
| `2026-09-30_010531` | **4m10s** (of 14m00s) | the observer registers at 4m02s and the `0x1E` arrives at 4m03s: before that there is no handshake to compare | `-verify` **ACCEPTED**: window at sub 0 on both sides, the original's 229 subs and 309 checksum entries as a prefix of its 12313 and 4948, and ours (227 and 303) too |

And what is **not** touched, with the reason:

- **The head.** The player's `0x15 LobbyConnect` is what creates the session, and in
  the 14m00s one the first 2m30s are the heartbeats before the first `0x09`: without
  them the store would fill at a different moment and the `0x19` would come out with a different value, which
  is exactly the byte being measured.
- **The handshake and the control responses.** `0x19` and `0x1E` are inside the
  window by construction, and `CheckTrim` checks that their sequence is still the
  same.
- **The redundant frames.** The idea of collapsing repeated acks (the identical
  heartbeats in a row are most of the log) is **discarded**: that is not trimming
  the recording, it is **changing the truth**, and the truth is what the comparison is against. There
  is no way to prove it with this harness, because the harness compares against the recorded
  truth. It would lower the repository size, and size was not the problem: the
  tail trim already takes away 97% of the bytes of the 14m00s capture.
- **The 1m29s capture.** It is the cheapest (93.9 s) and the only one with the **original at the
  front**; trimming it means losing 19 s for a third of its match.
- **The 24m44s capture.** See below.

### The 24m44s capture: the decision, and why

It is the one that does not fit, and the budget is what it is. Trimmed to one minute it is 60 s
more, and the sum of everything else is already at the minute: **it does not go in**. Besides, a
one-minute trim of that capture is a single-player match, which is exactly the
shape the 1m29s one already covers. The only thing it adds is 24 minutes of match
byte for byte against the original with **2 extra messages in total**, and that does not fit in
10 minutes alongside everything else.

It is not thrown away: it is in `captures/2026-09-29_101831_a553ef8fd63c/`, with its 97 MB, and it
is imported with `trimsim -trim ... -keep 1m -why "..." -verify` only if `-verify` accepts.
It is a capture that cannot be recorded again, and deciding by budget that it does not
fit is a decision that is said out loud, not one taken in silence by deleting it.

### The RAM budget, closed

No test goes past **54.3 MB** before the trim and **24.3 MB** after, and
all of `go test ./...` with every package in parallel peaks at **114 MB**: a
factor of four below the ~500 MB ceiling, and a factor of over a hundred below
the 20 GB of the discarded work.

The trims are written **row by row**: the code copies the row as it is, with its
`data_b64` undecoded, so a 48 MB trim does not need 48 MB of memory
nor a byte of decoding. And there is no new O(n^2) path: the digests are
one per message and `subIndex` (the only O(n) per lookup) is called once per
connection, not per message.

In bytes, the logs: the 14m00s capture's log goes from **14.06 MB to 0.45 MB** per
log (3.2% of what it was) and the 3m38s one from 2.45 MB to 2.25 MB (91.7%, which is the
proportion of clock taken off it). The whole `testdata` drops from **32.1 MB to
5.6 MB**.

### The signing key, and why the data commit is half-done

`git commit -S` fails with `gpg: firma fallida: Vida maxima` (error 0x0500003E) with the
configured signing key, `B859D4BE7C4AAB9F8118A3DFBA3D5801A6CFE711`, which **is not
expired** (`sec ed25519 2025-01-16`, with no expiry date, and the last three
commits on the branch are signed with it and verify correctly). Pinentry is launched
(`PINENTRY_LAUNCHED ... qt`) and that is where it fails; with `--pinentry-mode loopback` and an
empty passphrase the error is a different one (`No se ha dado ninguna frase contraseña`), that is, the
key has a passphrase and it has to be typed. **The code commit did go out** (at
06:47, with the pinentry answered); the data one did not, and it is not made unsigned nor with
`--no-gpg-sign`. It is left in the index and said, instead of leaving an unsigned commit
or a half-done tree without warning.
---

## 27. bsPort: the transport is decided by the bytes, and the TLS is the port's (2026-09-30)

### The capture

`2026-09-30_201122_c4e7c284ff1c` is the **first bsPort one**: a real LAN
match ("use local LAN server"), `primary.writer=ours`, `secondary.writer=original`, commit
`c4e7c28`, `dirty=false`. The game spoke over `bsPort` (14001/15001) and not over
`webSocketPort`, and there are **two** client connections (19850 and 50316), both
over bsPort. The traffic runs from 20:27:20 to 20:30:46: **3m26s**, out of a 16-minute
run because the game took 16 minutes to connect.

The CSV's shape is the one that had never been seen: `msg_type` is the **internal type**
of the message and not `0x82`.

| | rows | 0x0A | 0x13 | 0x09 | 0x14 | 0x17 | 0x15 | 0x0F | `0x82` |
|--|--|--|--|--|--|--|--|--|----|
| `main.log` (ours) | 2928 | 2533 | 250 | 64 | 40 | 34 | 4 | 3 | **0** |
| `secondary.log` (truth) | 1591 | 1257 | 189 | 64 | 40 | 34 | 4 | 3 | **0** |

### The brief assumed a failure that does not exist

**The reader has never filtered on `msg_type == 0x82`.** The column is not read, that is
written in the code since the harness's first commit (`internal/sim/sim.go`,
`loadLog`) and the comment there says so. Measured on the new capture:
`trimsim -inspect` reads it whole, with both connections and `rlink-len` on both. A
reader that filtered on `0x82` would have returned the capture empty, and that was not
the state of the tree. It is written down so that the next session does not measure it again.

What really was wrong, and produces **the same symptom** (a clean report over two
empty captures), are two things.

#### 1. The transport was decided with the first two bytes

They are not enough, and it is not a matter of taste. An RLink message starts with its length in
big endian, so **one of 33,024..33,279 or 33,536..34,047 bytes starts `0x81` or
`0x82`**, which is byte for byte a WebSocket FIN+binary header. Believing the first byte
frames the whole capture as WebSocket frames and produces a mess that additionally
"frames", that is, the failure does not even look like a failure.

Now each candidate is walked **whole in both directions** with the same step
as the `scan` that comes next, and the one that reads without running into a byte that cannot
start a frame wins. The HTTP upgrade, when it is there, overrides everything else: that
a connection was upgraded is a fact of the capture and not an inference. When both
framings read, the one that reads more is taken and the connection **says** so in its
notes, because that is not something the reader resolves.

`internal/sim/transport_test.go` has the exact case (a message that declares
`0x8200`), and **it fails with the two-byte decision**: `an RLink stream has to be read
as RLink, got "ws"`. The TLS-record case also fails (which before came out as
`rlink-len` without framing anything) and the two-bytes-that-reach-a-message case too.

#### 2. A log that decoded whole and framed nothing came out clean

`loadLog` rejected a log only if **no** row could be decoded. A log
whose rows all decode and out of which not one message comes came back with its
connections and a note ("not the battle protocol"), and the report said the capture
had nothing to compare: a green test over a blank report. Now it is
rejected, and the error names the transport and the two shapes the harness has to
read.

And the test that had to tie the two halves together **passed right over that**:
`cmd/proxy/crosscheck_test.go` wrote eight arbitrary bytes with the production
logger and checked that the reader had found a connection. With the new
rejection **that test fails**, which is the falsification, and it has been rewritten for the two
shapes: the framer and the real logger write the capture (WebSocket frames with
handshake, masked the way the game masks them, and RLink messages on bsPort) and
`sim.Load` has to return **the same messages, with their types and their bodies**.

One detail that appeared while writing it and that is not decoration: the fixture has to
carry the mask set. The framer removes it in place (`framer.go:53`), that is, what
the CSV carries is the frame with the key and the payload **in the clear**, and the reader does
well not to unmask again (`internal/sim.wsPayload`). With the payload in the clear
in the fixture the framer masks it once more, the capture carries garbage, and it still
frames: the test passed and checked nothing.

### The bug the capture uncovered when replayed: the TLS is the PORT's

With the reader fixed, the capture replayed and gave **`ours=0` on both
connections** against 24 and 2 messages of the original, with 0 must-match and a clean
report. It was not the reader: the harness decided the TLS **once for the whole replay**,
and the certificate pair was in force because the files in
`F:\ageLANServer\build\server\resources\certificates\` **do exist on this machine**.

Our `bsPort` **is not a TLS listener**: `internal/server/server.go:195` builds
the bs one without the pair and `:199` gives it to the WebSocket one. That is, the harness was
marking TLS on a port that speaks in the clear, the server read the ClientHello as if it
were an RLink length, and answered nothing. No earlier capture would have shown this because
**all the committed ones went over `webSocketPort`**, where TLS is indeed the
right thing: the assumption was accidentally correct and nobody had put it to the test.

Now the TLS is decided **by port** (`internal/sim/replay.go`), and when the pair is in
force and there is some connection that is not WebSocket a `Deviation` comes out saying so
with its reason. The reason is written in place, because a server that answers
nothing on the bsPort looks **exactly the same** in the connections table.

With that the comparison is real:

```
conn 19850  rlink-len  compared  sent=44  ours=53  original=58   first disagreement at index 48
conn 50316  rlink-len  compared  sent=43  ours=39  original=5
MUST-MATCH  : 0        INFORMATIVE : 15
```

And the 15 informative ones are real differences: we send 31 LobbyRelay to the second
client that the original sends none to, and on the player's connection the original
sent 8 acks where we sent 3. None of them is must-match (ack counts are
informative by AGENTS.md), and none of them turned the crop into something else.

### The two trims, and why these

Neither of them is the one you would want by taste: both are asked for by the budget, and
both were **verified** with `CheckTrim` before being imported.

| capture | before | after | why there | `-verify` |
|--|--|--|--|----|
| `2026-09-30_183120` | 15m10s | **1m30s** | the observer announces itself and gets `0x19` and `0x1E` at 1m16s, and 1m30s leaves that burst served; the 13m40s that go away are the same `0x0A` per second | **ACCEPTED**: window at sub 0 on both sides, and the original's 431 subs and 257 checksum entries are a prefix of ours |
| `2026-09-30_201122` | 3m26s | **48s** | bsPort needs **both** connections (the second client at 15s, the original answers it at 16s) and the first second with acks: the first `0x09` upload and the first `0x0A`s land at **47s** | **ACCEPTED**: both connections compared on both, the truth's streams as a prefix |

A 16s trim of the bsPort capture is **rejected** by `CheckTrim`, and rightly so: at
16s the original had already answered the second client and the trim left it
without truth, that is, a connection compared before stopped being compared. The 17s does not either:
the window is semi-open and the 16s rows fall outside it. That is why 48s and not 45.

The trimmed files are a **literal prefix** of the recording, with nothing rewritten:
`git diff --numstat` over the four 183120 logs gives **0 added lines and
13,434 deleted** per log. And the whole 183120 capture **is not lost**: it is in
commit `aa471d5`, so `git checkout aa471d5 --
internal/sim/testdata/2026-09-30_183120_c4e7c284ff1c` brings it back.

### The budget, measured again

With the test binary's working set watched every 200 ms and nothing else running:

| | before | after |
|--|--|--|
| `TestSim_2026-09-30_183120` | 910.7 s | **~91 s** |
| `TestSim_2026-09-30_201122` (new) | - | 47.5 s |
| `internal/sim` | 1471.6 s | **698.2 s** |
| **`go test ./... -count=1 -timeout 25m`** | 1477.7 s / 60.8 MB | **705.9 s / 70.0 MB** |
| **`go test -race sim+server+session -timeout 25m`** | did not fit (the package alone, untrimmed, asked ~1900 s) | **718.6 s / 217.8 MB** |

Both verification commands pass, with **over 13 minutes of margin** against
their 25-minute timeout. Before, the whole `go test` was **22 s** from breaking, and the `-race`
did not fit.

The **10 minutes of the budget are not reached with five captures**, and it is not a
decision that can be solved by trimming more: the three old captures are already
trimmed right down to their observer handshake (3m20s, 4m10s and 1m29s), which is
exactly what `CheckTrim` does not let through, and the 1m29s one is the only one with the original at the
front. The measured floor is **~700 s**. What does change is that the ceiling is no longer
within a hair, and that the peak RAM is 70 MB (217.8 MB with `-race`), a factor of
seven below the ~500 MB of the budget.

Incidentally, the `-race` factor **fell from 2.1 to 1.02**: the detector's cost rode
behind the per-second work of a long capture, and with the captures trimmed there is
far less to instrument per second of clock.

### What is left from this session

- The **25 extra LobbyRelay** we send to the second client, and the **5 fewer acks** in the
  player's connection. They are informative and have not been touched: no
  capture separates them from the tick cadence without another trim of the clock.
- The LAN's `bsPort` is the only transport with two client connections at once,
  and it is the only case where we send messages the original does not send. Neither of the
  two things was in the `CAPTURES.md` matrix.
- Row 1 of the matrix (AoE2DE, full match, **original at the front**) is still
  pending, and now it is easier: the harness already reads both transports.

### On bsPort you cannot observe, and that explains the extra peer

The user's fact, and the one that was missing: **on bsPort you cannot observe**. The
original stays at a single peer because there is no observer possible, not because it counts
differently. That **discards** the hypothesis we were going to test, which was that the
peer's removal lives in `handleStartObserve` and bsPort does not go through there: with no
observer, there is nothing to remove.

Therefore the bsPort capture's divergence (31 extra LobbyRelay, extra acks,
`fe 00 02 01 02 00 00` against `fe 00 01 01 00 00`) **is not about observation**, but about why
our server adds a second peer on a transport where the original does not
do it. It is a different question and it stays open.

Methodological consequence that has to be written down:

- The bsPort capture **is not a reference for observer behaviour**.
  It would serve for the opposite, which is exactly the point.
- For the observer, the reference is WebSocket, and there we already know it matches:
  7204 messages against 7208 of the original, MUST-MATCH 0.
- A `0x0A` with a two-peer pending list on a transport with no observer is
  a clue, not a conclusion. Rule 1 still rules: nothing is touched without evidence.

#### What is still needed, in order of value

1. **A match with three or more clients on bsPort.** With two connections it is impossible to
   distinguish "one peer per connection" from "one peer in total". With three, if the original stays
   at 1 or goes up to 2, the question answers itself.
2. **The destination bytes of each `LobbyRelay`**, not just the count: the distribution needs
   a non-null slot in `session + 24 + 8*toPeerID` (`sub_1400E6160`), and with the count
   alone you do not know which peer each one points at.
3. **`0x0F`**, which `SESSION-JOURNAL.md` has been asking for since before: if it creates a peer, seeing it
   brings the creation path.

---

## 28. The short loop, the 4 acks of second 26, and the byte of the 3 (2026-10-01)

Four things. The first is a real fidelity bug and a new synthetic
fixture found it, which is exactly what they exist for. The other two close or narrow
questions that were open, and the fourth is about a harness mechanism.

### The `0x19 = 3` never reached the wire. IT WAS OURS.

The 3 arm of `sub_1400E3B50`'s cascade writes 3 into the response and returns
**without touching the observer list**:

    sub_1400E3B50:93      v58 = 3;
    sub_1400E3B50:113     LOBYTE(v69) = v58;
    sub_1400E3B50:87-160  the whole block: builds the response and leaves

And `handleStartObserve` mapped that state to `default`, which sent **4**, which is the
other rejection: `v68 = 4` at `sub_1400E3B50:403`, observers deactivated for the
session. That is, the client received "observing is deactivated" instead of "the store
already lost the beginning of the match", which are not the same failure and do not lead to the
same place in the code to look.

Why none of the committed captures had seen it: because **none has a
store that has cycled a block**. In the five the original store lost nothing and
answered 0. The 3 was a number measured in _30, historical and deleted, and an arm of
a state machine with unit tests. The response path, the one that puts the byte on
the wire, had no coverage.

`testdata/synthetic-rollover`, the second synthetic fixture, found it: the SAME
capture as `testdata/synthetic` with two different things in the recorded command
line, which is the only thing that changes:

    -maxPagesToStorePerSession 1    a single block in the store
    -pageSize 128                   of 128 bytes in that block

With 128 bytes the second ack does not fit in the block (a frame costs its size plus 57,
`WWBS::ReplayPacketAllocator__vmethod_20:76`) and the store recycles its own
(`vmethod_20:118`), that is, by the time the observer asks at t=1s the store no longer
has the first frame of the match. Both directions come out of the same body:

| `-pageSize` | store | truth | ours | verdict |
|--|--|--|--|--|
| 128 | cycled | `0x19=3`, no `0x1E`, no window | `0x19=3` | 0 must-match |
| 65536 (the original's own default) | intact | `0x19=3` | `0x19=0` | **must-match** `start-observe-status` |

The bottom row is the falsification: the same fixture with a different flag fails in
the other direction, which is what separates it from a golden test.

And incidentally the other warning from the 24 comes out: with both observer streams empty, the
report says verbatim "neither side served this observer a single replay frame", that is,
the green is not on top of two empty things.

### The 4 acks of second 26: THEY ARE NOT OURS

The question was "26s: original 3 | ours 7". **The capture is not the one the
brief said**: it is `2026-09-28_205830_314309565d6e`, not the `010531`. The 205830 is the
important one because it has the **original at the front**, that is, it recorded in the SAME run
the original binary and ours, against the same client and the same wall clock. The
010531 cannot answer this question or any other of this kind: both
recordings are of our own binary.

Reading the two logs of the 205830 (not the replay):

    second 26:            3 from the original, 2 from ours    (the replay said 7)
    seconds 27..88:       8 and 8, both, without a single drift in 62 seconds
    totals:               500 and 500

The same in the other four, which is what makes it not a coincidence:

| capture | first second of acks | what separates the most in one second | totals (orig \| ours) |
|--|--|--|--|
| `2026-09-28_205830` | 26s: **3 \| 2** | 8 \| 8 over the 62 following ones | 500 \| 500 |
| `2026-09-29_012528` | 22s: 8 \| 7 | 8 \| 7 and 8 \| 9 (one tick of jitter) | 356 \| 355 |
| `2026-09-30_010531` | 169s: 8 \| 8 | **none** | 584 \| 584 |
| `2026-09-30_183120` | 13s: 6 \| 6 | **none** | 508 \| 508 |

The sustained rate is 8 per second on both sides of all four. The surplus is
from the replay.

**The mechanism**, which is not a hypothesis but a property of the code checked
over the capture: the proxy stamps whole seconds (`cmd/proxy/packets.go:47`) and
`timeline()` schedules each message at its exact stamp, that is, a message written at
any point in second N is fed at N.000, **up to one second early**. With one
ack per 125 ms tick that puts acks inside the second in which the stream starts and
in none after, which is exactly the measured shape. The phase moved; the cadence
did not. `TestTheReplayFeedsEveryMessageAtItsWholeSecondStamp` checks it over the
capture: all 206 client entries are scheduled on a whole second.

What does stay open, and is said: the binary the 205830 recorded as "ours" is
the one from commit `3143095`, thirteen commits behind this one. The argument does not depend on
that (the phase pulse comes from the harness, not the binary), but the claim "what today's
build would do" has not been measured and is not asserted.

#### The rule that writes it down forever: `tick-cadence`

`internal/sim/cadence.go` compares **three** series per second instead of two: the
original, this build's replay, and what the capture itself recorded of the other
binary. It is the only third series that exists, and it is needed because the client was the same and
the clock was the same.

The conclusion comes out of the numbers and only says what the numbers support:

- equal sustained rates on all three -> "**THE TICK CADENCE IS THE SAME ON
  ALL THREE SIDES**", because a cadence is a rate and there is none to look for;
- and it only calls it a phase when the differing second **is** the second in which
  the stream starts. A second that differs later is neither cadence nor phase, and it
  says so: "no cause is claimed for it";
- when the rates do not separate and the per-second gaps point at nothing, it
  concludes nothing and says so.

That it comes out on its own in the reports is the useful part. In `2026-09-28_205830` the rule says, verbatim:

    the sustained per-second rate is 8 on all three sides, over 62, 63 and 62 of their
    seconds respectively, so the TICK CADENCE IS THE SAME ON EVERY SIDE. The second
    that differs IS the second the ack stream starts in (t=26s), and what differs there
    is the PHASE.

#### And a cross-check that had never run

`crossCheck` walked `cap.Conns`, and it was called **before** `cap.Conns` had been
filled. That is, `otherUp` was never stored and the `input-agreement` rule had not been able to
fire once in the harness's whole life. `Load` now fills the list before crossing, and `crossCheck`
also stores what the other side **answered**, which is the third series above. Both things are
visible in the five reports.

### The short loop and the full one

`go test ./...` was **762 s** and the captures cannot be trimmed further: `CheckTrim` already
leaves them right at their observer's handshake, and a suffix does not move the must-match
facts. A capture's cost is its recorded clock, because the replay goes in real
time on purpose (`Options.ReplaySpeed`: the server emits `0x0A` per tick, not per
input).

The cut is by **recorded clock**, not by capture name, with the budget
as a constant in the code (`internal/sim/short.go`, `ShortModeBudget = 20s`): the
synthetic one records ~1 s and the shortest real capture 47 s, that is, the short loop is
"the capture that needs no capture, plus everything that is not a capture". A list of
names goes stale the day a capture is added, and a stale list either drops a
capture out of the short loop by accident or keeps one in forever.

Measured, with the whole tree and the working set of **all** the test binaries added up
every 200 ms, with nothing else running:

| | before | after |
|--|--|--|
| `go test -short ./... -count=1` (short loop) | did not exist | **40.6 s / 77.0 MB** |
| `go test ./... -count=1 -timeout 25m` (full) | 762 s | **766.9 s / 115.5 MB** |
| `go test -race sim+server+session -timeout 25m` | 718.6 s / 217.8 MB | **772.4 s / 220.5 MB** |

The RAM is the sum of the binaries running at the same time, which is what "peak
of the suite" means; the package that commands it is `internal/sim` with 54 of the 115.5 MB. Before
it was only measured on `internal/sim`'s binary (62.8 MB) and that is why the figure comes out higher
now: they are two different measurements and the suite's is the one that counts. The 115.5 MB matches
the 114 MB that the 26 already measured with the same method.

The difference against the earlier 762 s is 5 s: the second synthetic fixture (2 s) and the
settle test, which before ate 48 s of the bsPort capture and now is skipped in the
short loop. The ~767 s floor is the same one from the 27: the captures cannot be
trimmed further.

What stops a green short loop from being mistaken for a green whole suite:

- `TestMain` prints the MODE before starting, and every long capture is skipped with a
  `Skipf` that carries its recorded clock and the command that replays it;
- `TestFastPathIsNotTheFullSuite` **fails** if the short loop leaves out the same thing the
  full loop does (which is how a budget goes stale without anyone noticing), and
  checks that the classification does not depend on the mode: a list that moved with the
  mode would be a rule choosing what to **run** instead of what to **skip**, which is how a
  capture falls out of the short loop without anyone choosing it;
- `TestGeneratedTestForEveryCapture` ties each `testdata` capture to its generated
  test, **in both modes**, and the synthetic one also has to be in the short
  loop: it is the only coverage that needs no capture.

Honest caveat, measured: a `go test` of a package that passes **prints nothing** of what
the test binary wrote (checked on this toolchain: a test that prints through
stdout, through `t.Log` and through stderr, and `go test` without `-v` shows none of the three).
So the mode cannot go inside the `ok <pkg> <s>` line. What does survive without
`-v` is the number on the `ok` line (41 s against 768 s), and with `-v` or `-json` it is all there.

### New coverage that needs no capture

- **the store really cycled**, end to end (fixture) and at the byte level
  (`TestHandleStartObserveRolledStoreAnswersThree`, which also fails if the store did NOT
  cycle, that is, the arm under test is the one the name says).
- **one ack per tick, exact**, and no duplication with two peers
  (`internal/session/cadence_test.go`): no ack before there is data, one per tick
  from the first, this tick's sub goes in THIS ack, and the heartbeat of a two-peer
  session is the `02 01 02` that _6 measured. No cap in the code: 64 ticks, which
  are more seconds than the four captures sustain at 8/s.
- **the TLS by port as a pure function** with its four cases (`internal/sim/tls_test.go`),
  which is the bug that cost a whole capture, plus the case bound to the real bsPort
  capture.
- **the per-second histogram of the three series** over the real capture, which is this
  section's measurement and costs no replay.

Everything falsified by hand before committing:

| defect reintroduced | what the test says |
|--|--|
| the 3 arm back to `default` (4 on the wire) | `must be answered with status 3 ..., got 00021904` |
| the tick with no data produces no ack | `tick 1 produced no ack: the original emits one per tick once game data has started` |
| the cycled-store fixture with the original's default | `must-match start-observe-status`, `0x19 ... 03` against `0x19 ... 00` |

### What stays open

- The 205830's binary is thirteen commits behind. The conclusion (the surplus is from the
  replay) does not depend on it; "what today's build would do" has not been measured.
- The **31 extra LobbyRelay** and the extra acks of bsPort (§27) stay open, and
  now that capture's report also says, with numbers, that its sustained rates
  are the same (9/s on all three sides) and that the differing seconds are not
  the first one: "no cause is claimed for it". It is still a different question.
- The measurement that the 25 asked for (the zoom with the original at the front) is **voided** by the
  user: the BattleServer is deterministic and the zoom is not this server's thing.
- Row 1 of `CAPTURES.md` (AoE2DE, full match, original at the front) is still
  pending. With the short loop it no longer costs 700 s per change, so the new
  capture goes in on its own.

---

## 29. When the original creates a session peer (2026-10-01)

The question the 27 left open: the bsPort divergence (31 extra `LobbyRelay`,
`fe 00 02 01 02 00 00` against `fe 00 01 01 00 00`) is not about observation
— the user said you cannot observe on bsPort — but about **why
we add a second peer and the original does not**. This section answers it
as far as the evidence can, and says explicitly where it stops.

### The trigger is ADDPEER, and it needs a REQUESTJOIN first

From the binary, without inferring from names:

- **`sub_1400E5670`** is the only function that writes a peer into the session's array
  of slots: line 100, `*(_QWORD *)(v10 + 24) = v9[5]`, with
  `v10 = a1 + 8LL * a3`, that is `session + 24 + 8*peerID`. It is the same array that
  `sub_1400E6160:18` reads to deliver a relay. Its only incoming call:
  **`sub_1400EAE60`**, reached only from the ADDPEER case of
  `sub_1400DAAF0:298`.
- Before creating anything, `sub_1400E5670:75-85` looks for the peer's **global id** in the
  session's map at `session+2080` and **returns without creating** if it is not there.
- That map is written by **one single** function, `sub_1400E5B00`, reached only from
  **`sub_1400EA680`**, the REQUESTJOIN case of the same switch (`sub_1400DAAF0:193`).

That is: **a peer exists when a REQUESTJOIN registered its global id in the session
and an ADDPEER then claimed a slot for it.** The map is per session, not per connection,
and the two messages can arrive on different connections: the capture does that.

- **`sub_1400EB5F0`, the `0x15` handler, adds no peer on any of its
  paths.** Its 520 lines read: it looks the session up by (titleID, sessionID), in the
  reconnect branch it looks for a peer **by secret** (`sub_1400E5CB0`), and in the new
  branch it writes the profile, the secret, the version and the timeout into the connection.
  It never touches `session+24`.

### The client does not send either of the two as 0x0F

Measured over the **seven** of `testdata`: **0** top-level `0x0F`, **13**
inside relays. The `0x13` carries a whole `0x0F` inside, and the original dispatches it
recursively (`sub_1400DE050:567`, `v40 = sub_1400DE050(a1, &v111, a3)`
on the relay's sub-message). The first ones come **zlib compressed**
(`flagBits 0x04`, `sub_1400DAAF0:75`), so without inflating them the internal
type is `0x78` and there is no peer message to see at all. The 13 compressed ones are
exactly the ones carrying REQUESTJOIN and the first ADDPEER of each connection.

### The 0x0A names the session's peers, and the original always names one

`sub_1400E34F0:80-91` walks a 40-byte record per peer, takes the id from
offset 32 and looks the peer up in `session + 24 + 8*peerID`; those without a slot are
skipped and their id never reaches the wire. The list is an enumeration of the peers that
survived that pass.

Read from the committed captures, per connection:

| capture | connections | the original names |
|--|--|--|
| `2026-09-28_205830` | 1 | `[1]` |
| `2026-09-29_012528` | **3** | `[1]` on all three |
| `2026-09-30_010531` | 2 | `[1]` on both |
| `2026-09-30_183120` | 2 | `[1]` on both |
| `2026-09-30_201122` (bsPort) | 2 | `[1]` on the first, **nothing** on the second |

The 012528 one **does have three client connections**, which is what the 27 asked for,
but two of them are observers: they do not distinguish "one peer per connection" from "one
peer in total", because in both cases they are one.

### The bsPort capture: the client asked for the second peer and did not get it

In `2026-09-30_201122` the client sends **both halves**:

- client **19850** sends `REQUESTJOIN` with peerKey `0x56b4d65de6f6084f` and
  immediately `ADDPEER peerID=1 globalID=0x56b4d65de6f6084f isHost=true`, and at
  15 s `ADDPEER peerID=2 globalID=0x373fdcafe5ad6934`;
- client **50316** sends `REQUESTJOIN` with peerKey `0x373fdcafe5ad6934` — the
  peer 2's global id — and **never** an `ADDPEER` of its own.

The session's map receives both ids and there is an ADDPEER claiming slot 2.
**The original did not create peer 2**: its pending list says `[1]`, and it did not
forward any of the 31 relays we sent to the second connection
(`sub_1400E6160:19` does not deliver to a null slot). The client **was not wrong**:
peer 2 existed in its own account and in the host's, which announces it.

### What is NOT known, and why the server is not touched

**Which of the two messages the original rejected is not known**, and it cannot be known
with this capture:

- The two messages of peer 2 go in **different connections** (the REQUESTJOIN on
  50316, the ADDPEER on 19850), the proxy stamps **whole seconds**
  (`cmd/proxy/packets.go:47`), and each connection is read on its own socket. The order in
  which the original **processed** them is not in the capture.
- If the ADDPEER was processed before the REQUESTJOIN wrote the map,
  `sub_1400E5670:75-85` returns without creating and the refusal is the code's. If it was
  processed after, the refusal is **another** one, and it is not identified.

Adding the peer in the `0x15` (where we are), removing it from there and putting it in the
`ADDPEER`, would be a large change on top of a question missing the fact that
decides it. And a change that swaps one visible absence for another is not a fix
(`AGENTS.md`, the three must-matches). **It is not touched.** What is done is that the
question can be answered without 20 minutes of capture.

### What has been done

**The measurement, in the harness** (`internal/sim/peers.go`, `relaypeer.go`), because it
costs not one second of game time and turns the paragraph into an account:

- The rule **`[pending-peers]`** reads each side's pending list and
  reports when they name **different sets** of peers. It started that way, and
  it now says why in the finding's own lines, with the addresses.
- It compares **sets** and not **counts**, and this was measured: in
  `2026-09-29` both sides name `{1}` and only how many acks carried each
  list changes, which is grouping and not peers. A rule that compared counts
  would have reported a peer divergence in a capture where there is none, and it did it on **all three**
  connections. The test
  `TestTheRuleStaysQuietWhenOnlyTheCountsDiffer` is that falsification.
- The **13** peer messages can be read from the captures, and only if the
  compressed ones are inflated. `TestTheFirstPeerMessagesAreCompressed` fails if the
  un-inflated byte ever reached being a peer type, that is, if the flag stopped being
  informative.
- `TestTheOriginalNamesOnlyPeerOne` **pins** the original's measurement across the five
  captures. If a future capture shows the original naming a second peer,
  that test fails instead of passing unnoticed.

Falsified by hand before committing:

| defect reintroduced | what the test says |
|--|--|
| the empty pending list goes back to being `[]` (peerRead with a value receiver, an `append` that does not land) | the five captures of `TestTheOriginalNamesOnlyPeerOne` read `[]` |
| the rule goes back to comparing counts instead of sets | `TestTheRuleStaysQuietWhenOnlyTheCountsDiffer` |
| the rule is marked must-match | `TestTheRuleFailsWithTheDivergenceReintroduced` |

### The measurement that is missing, and it is a single one

**A LAN match with three or more real clients on bsPort**, with the game session
started (so that the 0x0A exist) and **no observer**. With two
connections, "one peer per connection" and "one peer in total" are indistinguishable; with three,
if the original stays at 1 or goes up to 2, the question answers itself. If
additionally the third client's `ADDPEER` arrives **over the same connection** as its
`REQUESTJOIN`, the order is observable and it tells which of the two messages was
rejected.

Until then: the **mechanics** are established (`sub_1400E5670`, `sub_1400E5B00`,
`sub_1400EA680`, `sub_1400EB5F0`) and the **divergence** is measured. What is not
established is the condition under which the original refuses.

### The budget, measured with the new code

With the working set of the test binaries added up every 200 ms, in `src_go`:

| | before | with the rule |
|--|--|--|
| `go test -short ./... -count=1` | 40.6 s / 77.0 MB | **40.7 s** |
| `go test ./... -count=1 -timeout 25m` | 766.9 s / 115.5 MB | **767.9 s / 117.2 MB** |
| `go test -race server+session` | - | green (37.0 s) |

The **10 minutes of the budget are not reached** with five captures, same as
before (§26, §28): the floor is the captures, and the new rule does not lengthen them. The
117.2 MB are two test binaries at once, and the package that commands them is still
`internal/sim`.

---

## 30. The two unexecuted paths: the reconnection and the lagged observer

The two gaps that were left unexecuted under test, verified against the binary.
Both **had code and neither had ever been run**, and both came out
with differences. Commit `0206f0b`.

### The reconnection: the original closes the replaced connection and does not resend outside state 2

What was already verified (§20.2) is confirmed and completed. From the binary:

- The session is looked up by (titleID, sessionID) and the peer **by secret**:
  `sub_1400EB5F0:128` calls `sub_1400E5CB0(session, msg+72)`, which walks the
  map by the session key (`session+2080`) comparing the **peer's secret**
  (`peer+80`), not the map's key.
- That secret is written on the **connection** (`sub_1400EB5F0:468` and `:489`,
  `*(u64 *)(a2 + 80) = *(u64 *)(a3 + 72)`) and the session's "peer" **is the connection
  itself**: `sub_1400E5B00(session, key, conn)` inserts the connection as the map's value
  (`sub_1400EA680:259`) and `sub_1400E5670:100` copies that value into
  `session + 24 + 8*peerID`. There are not two objects.
- The socket swap is `sub_1400EB5F0:137-139`, and **the peer keeps the new socket**.
  That is why all of that branch's sends go to the peer
  (`:248`, `:409`) and not to the connection.

What was not:

| | before | now | address |
|--|--|--|--|
| the replaced connection | was forgotten | it is closed (`sub_1400EB5F0:220`) | `sub_1400D4860(conn, reason, 0)`: deadline = now; `sub_1400D4930` takes it in the socket loop (`sub_1400D8E10:753`). The object is left with the **old** socket, so closing it is the same teardown |
| reconnect state 1 | resent `[0, PageCount-1]` | resends nothing | `sub_1400EB5F0:237-249`: if `msg+80 == 1` it hands the connect to the peer and leaves |
| states 3 onwards | resent `[0, PageCount-1]` | resend nothing | those messages do not carry `dataReceivedCount`: the parser puts **-1** (`LobbyConnectMessage__vmethod_20:59`) and the range `[-1, PageCount-1]` is rejected by the check as invalid (`sub_1400D5520:12-13`) |

The last point is half of the second one: the Go parser left the field at 0
(Go's zero value), and a 0 is a servable range. With both as they were,
state 1 and state 3 resent the whole window.

**What is NOT reproduced, and why**: the original rewrites the `dataReceivedCount`
of the connect it returns with `peer+104` (`sub_1400EB5F0:341`), which is the count of
accepted sync messages from that peer (`sub_1400E2E00:202`). This
server keeps no such count and **no committed capture has a reconnection**
(measured: the `0x15`s of the seven of `testdata` all carry `reconnectState = 0`),
that is, there is nothing to check the value against.

**A deliberate difference that is left**: a zero secret is not looked up. The original
would compare `peer+80 == 0` and could find the first peer without a secret; here
`PeerBySecret(0)` returns nil. No capture brings a zero secret, so there
is no evidence in either direction.

### The lagged observer: the flag was exactly the opposite, and the target was not the frame count

`Session.StartObserve` deciding `lagged = passwordHash != 0` was **the rule
inverted**. The original's is `sub_1400E3B50:241-243`:

```
v27 = *(u64 *)(session + 2336);                       // the u64 published with 0x1D
if (v27 == 0 || (v28 = *(u64 *)(a3 + 72) == v27, entry.lagged = 0, !v28))
    entry.lagged = 1;
```

That is: **lagged unless it presents exactly the published u64**. Without a published
hash, lagged; with a different hash, lagged; only the matching one clears it.
This **closes the point the 20.3 left open** ("the 0x1D's u64 is not compared against the
passwordHash because there is no proof that the original compares
them"): if they are compared, and the u64 is `session+2336`.

Measured over the seven captures of `testdata`:

| capture | 0x1D | 0x1C (passwordHash) | lagged in the original |
|--|--|--|--|
| `2026-09-29_012528`, `2026-09-30_010531`, `2026-09-30_183120` | hash `0x21616c6661aa7a42`, lag **0** | 0 | **yes** |
| `2026-09-28_205830` | hash `0x21616c6661aa7a42`, lag **0** | (no observer) | - |
| `synthetic`, `synthetic-rollover` | (there is no 0x1D) | `0x61605f5e5d5c5b5a` | **yes** |

The **five** observers of the four real captures are lagged in the original
and none of them was here. And the target was not the frame count either: the existing
comment read `sub_1400E4EC0:65` as "the message queue is empty", and what
it says is `session+2142 != 0 || the session has no peers` (`v5`). A lagged observer's target
with the match alive is `session+2216 - session+2344`, that is, it
**never sees the most recent LagPages frames**.

**Why none of this showed in any capture, and why the five stay green**: the `0x1D`'s third u32 is the
**lag in seconds** and the game sends **0** in all four. With lag 0 the two arms agree:

- lagged target = `PageCount - 0` = `PageCount` = the normal target;
- the lagged `0x1E` time = `(0 - PageCount)/ticks` = `-(PageCount/ticks)` = the
  normal one.

The two failures were masking each other, and the field that distinguishes them is
exactly the one the harness **masks** (`0x1E`, the estimated start time). That is
why a test with lag > 0 is needed, and why the captures cannot be the ones that
demonstrate it.

### The `0x1E`: the body and the two formulas

`sub_1400E43E0` whole, which is 50 lines:

```
:39-49   walks the observer list (session+2272/2280) and stops at the entry
         whose connection matches; sets entry+12 = 1
:47      if none matches, returns WITHOUT SENDING ANYTHING
:56      v6 = 0x3E8 / *(u32 *)(session + 2172)        (ticks per second)
:57-61   lagged:  v8 = (session+2344 - session+2216) / v6
         normal:  v8 = -(session+2216 / v6)
:69      v27 = _time64(0) + v8
:54      v23 = 30  (the wire type, 0x1E)
:118-142 writes 1 byte (v26 = 0) and then 8 bytes of v27
```

Body: `[status u8][time i64 LE]`, 9 bytes. The status is **always 0** because the
only other exit is the `return` with no response. The **direction** of the offset is not
"the other sign": they are two different subtractions, and with a lag smaller than the match the
lagged observer receives a time in the **past**, not in the future. This was already
right in the code; what was missing was the formula written with its direction
and a test that pins it.

One detail that comes out of reading it: the entry lookup compares
`*entry.conexion` with `*a2`, that is the **u64 at connection offset 0**, which
`sub_1400E3B50:240` has just written with the request's `peerIdentifier`. The
original pairs by that identifier, not by the connection's address
(`sub_1400E43E0:43` and `sub_1400E46C0:25`, the same comparator). Here it pairs
by connection identity. It is left as a difference: the four captures have a
different identifier per observer, and there is none where two observers
share one.

### Falsified by hand, each defect with its own failure

| defect reintroduced | what the test says |
|--|--|
| `lagged = passwordHash != 0` | `entry.Lagged = false, want true` and `time = now-12s, want now+17s`: the two observers of the same test swap roles |
| target always `PageCount` | `lagged observer got 100 frames, want 60 (PageCount-LagPages)` |
| `(PageCount - LagPages)` in the `0x1E` | `lagged observer time = now-17s, want now+17s` |
| state 1 resends from 0 | `state 1 resent 3 frames, want 0` |
| the parser leaves the count at 0 | `reconnect state 3 resent 3 frames, want 0` |
| the replaced connection is not closed | `the replaced connection must be closed by the reconnect` |

The `0x1E` with the sign flipped **does not** break the target and the broken target **does not**
break the `0x1E`: they are two different facts with two different tests, which is what was needed so that they do not
mask each other.

### What cannot be established

- The value the original rewrites into the `dataReceivedCount` of the returned
  connect (`peer+104`, the peer's accepted sync-message count). Not
  modelled, and without a reconnection capture there is no way of knowing whether the client cares.
- Whether the original would also close the peer's socket, or only the replaced
  connection's: the branch has only one call to `sub_1400D4860` and it is on the
  connection, so from the binary it is only that one. Without a capture it cannot be ruled
  out that the client would notice the close.
- The `peerIdentifier` pairing condition when two observers share it:
  it happens in no capture and there is no way of provoking it here.

### The budget, measured with the new code

With the working set of the test binaries added up every 200 ms, in `src_go`:

| | 29 (without these changes) | 30 |
|--|--|--|
| `go test -short ./... -count=1` | 40.7 s / 77.0 MB | **43.1 s / 117.6 MB** |
| `go test ./... -count=1 -timeout 25m` | 767.9 s / 117.2 MB | **768.2 s / 115.5 MB** |
| `go test -race server+session` | green (37.0 s) | green (38.8 s) |

The full suite comes out the same: the five captures replay green **without a single byte
changed**, which is the proof that the captures' lag 0 makes the two
corrections invisible there. The short loop goes up ~2.4 s and the peak goes up, but
that peak is a **sum over concurrent binaries** and on this machine the same
code with no changes measured 63.6 MB and 86.9 MB in two runs: the noise band is
bigger than the difference. Measured package by package, which is comparable:
`internal/sim` 25.1 -> 27.3 MB, `internal/server` 20.8 MB, `internal/session`
4.6 MB. The new tests synthesise nothing (60 frames and 100 frames from a
43-byte body, with a cap in the test itself) and add no capture.

---



## 31. The observer list is ordered by the peerIdentifier, and that is not the connection



The 30 closed this as a difference that "is left": "the entry lookup compares

`*entry.conexion` with `*a2`, that is the u64 at connection offset 0, which

`sub_1400E3B50:240` has just written with the request's peerIdentifier. The

original pairs by that identifier, not by the connection's address

(`sub_1400E43E0:43` and `sub_1400E46C0:25`, the same comparator). Here it pairs

by connection identity." It was written down as a gap because the four captures

have a different identifier per observer. From the binary it comes out whole.

Commit `471340e`.



### What field it is, and where it is



The field is the **third u64 of the StartObserveRequest Vx, offset 9 of the body**:

`sub_1400DDAB0:37-39` (`*(u64*)(a1+64) = v10`) inside an 8-byte read at

position 17, over a body of exactly 25 bytes. The original itself names it

in its log, `sub_1400DE050:792-798`:

`"STARTOBSERVEREQUEST_Vx titleID=%d sessionToObserve=%lld peerIdentifier=%llu passwordHash=%llu"`,

with `v95[0]` = request+64 and `v95[1]` = request+72. The legacy 24-byte form does not
carry it and the original writes **0** on purpose (`sub_1400DDAB0:62`), with the
titleID at `0xFFFF` (`:63`).



The connection stores it at its **offset 0**, and in the whole binary that offset has

two writers and two readers, no more:




| | |

|--|--|

| `sub_1400D4770:20` | `*(_QWORD *)a1 = 0;` the connection's constructor: it starts at 0 |

| `sub_1400E3B50:240` | `*(_QWORD *)a2 = *(_QWORD *)(a3 + 64);` **the only writer** |

| `sub_1400E43E0:43` | `while ( **(_QWORD **)v4 != *a2 )` the 0x1A's lookup |

| `sub_1400E46C0:25` | `if ( *(_QWORD *)*v4 == v6 ) break;` the 0x1B's lookup |



`v4` is the 16-byte entry, and its 16 bytes are `{conn*, u32 progress, u8

active, u8 lagged}` (`sub_1400E3B50:237-239`): **it does not carry a copy of the

identifier**, so both lookups read it LIVE through the pointer.

That is not an implementation detail, it is what makes "the entry follows the
connection" a property and not a coincidence.



The comparator is an equality over a u64, the **first** matching entry wins, and the
lookup is inside the connection's own session: both callers take the session
from the connection's offset +40 and do nothing if it is
null (`sub_1400EC580:14-16` and `sub_1400EC5A0:15-17`). There is no peerID, no
address, no hash and no second key in either function.



### What that decides, which is what no capture could ask



- **The same identifier twice.** Two connections that share it are the
  same observer as far as the list is concerned. A `0x1A` from the second arms the first's
  entry and the `0x1E` goes to the second (`sub_1400E43E0:49` arms, `:150`
  answers `a2`): two different facts, and that is why both are checked. A
  `0x1B` from the second erases the first's entry and closes the second
  (`sub_1400E46C0:31-33` closes the one it was given, not the entry's).

- **An identifier 0 is an identity.** Offset 0 starts at 0 for every
  connection, and the legacy `0x18` forces it to 0. With an entry at 0, any
  connection of the session that has not sent a Vx `0x1C` also resolves it, and a
  `0x1A` or `0x1B` from a player is blocked by nothing: the observer's type mask
  (`sub_1400DE050:248`, `0x1D100000`) goes in one direction only.

- **A second `0x1C` on the same connection.** `sub_1400E3B50` does not scan the
  list before adding (`:245-254`): it overwrites offset 0 and adds a second
  entry. The "a repeated StartObserve is idempotent" guard this code
  had swallowed the new identifier and never recorded it, so
  it is removed too, and its written justification ("the comparison is over the
  connection's identity, not over anything the client sends") was exactly what
  this change undoes.

- **The 0x19 byte is the arm's.** It is a fact of the cascade, not a
  question about the list afterwards: with a connection that already has an entry and a
  second request that lands on a rejection arm, `HasObserver` said "accepted" and
  the client received a 0 where the original would put a 2 (`sub_1400E3B50:295`).

  `StartObserve` now returns the record separately from the status.

- **A reconnection does not reach an observer, not even in the original.** The path by
  secret (`sub_1400EB5F0:128`) looks in the session's peer table, and an
  observer is not in it: peers are created by REQUESTJOIN + ADDPEER
  (`sub_1400E5B00` and `sub_1400E5670`, §29) and an observer sends neither of
  the two. When its socket goes, the entry stays (nothing removes it) and the game
  opens a new connection with its own `0x1C`.



### What is NOT changed, and why

When a connection dies **here its entry is removed** and in the original it is not. Measured:

the `session+2272/2280/2288` list is touched by six functions in all of `src/`, the
session's constructor (`E1DE0:58-60`), the destructor (`E2050:186-198`), the
append (`E3B50:245-254`), the two lookups (`E43E0:39`, `E46C0:17`), the reader
(`E4EC0:56`) and an emptiness check (`EA460:41-43`). **None removes an
entry when its connection dies**, nor on reconnect. Here it is removed, and it is documented
as a difference: a dead entry would keep being written to forever and
would take a slot from the cap the game needs. The server crashing by
writing to a freed object is not something that can be reproduced or measured.



### Eleven tests, and the wire offset

The `session` ones are the five cases of the brief plus the second `0x1C` and the replaced
connection; the `server` ones pin the **body's offset 8** (which is the wire's 9:
`parseWsMessage` leaves the version as `body[0]`), the two observers over the
wire, the legacy `0x18` as identity 0, the slot's release on close, and that
the byte is the arm's. All with data on both sides: two entries, two
`0x1E`, two progresses.



Falsified by hand, each defect with its own failure:




| defect reintroduced | what the test says |

|--|--|

| the lookup goes back to connection identity | `the first entry should be the armed one` and `the armed entry's connection got 0 frames, want 200` |

| the removal goes back to connection identity | `the wrong entry survived: the first match (A's) is the one the original removes` |

| the identifier is read from offset 0 | `connection identifier = 0x1, want 0xfeedfacecafebeef` |

| the idempotent guard comes back | `ObserverCount = 1, want 2` |

| the record is deduced with `HasObserver` | `the repeated request got 00021900, want status 2` |



### The budget, measured with the new code



| | 30 | 31 |

|--|--|--

| `go test -short ./... -count=1` | 43.1 s / 117.6 MB | **43.9 s / 114.1 MB** |

| `go test ./... -count=1 -timeout 25m` | 768.2 s / 115.5 MB | **769.8 s / 115.5 MB** |

| `go test -race server+session` | green (38.8 s) | green (36.5 s / 1.5 s) |

| `go run ./cmd/gensim -check internal/sim/testdata` | green | green |



The peak is the sum of **this** repository's test binaries, every 200 ms:
the filter is a whitelist of names, because `%TEMP%\go-build` is shared with
other projects on the machine and path alone is not enough (another project's `router.test`
lives there too and this time it does slip into the sum). The five captures
stay green **without a single byte changed**, which is the proof that the identifier
moved nothing of what the captures compare: each observer carries its own and
there is only one per connection.



**A flake that was seen twice and has not come back.** During this work
`TestSim_c2026_09_30_183120_c4e7c284ff1c` failed twice with the same signature: the
observer's first window landed on frame 503 instead of 0, and the
`[replay-page-order]` must-match counts it as an inverted pair. With only one
`0x1C` in the whole capture (measured) and only one entry added, that cannot be
this change. The mechanism the report leaves visible is the **store's
window**: if the block is recycled before the observer's first visit, the
jump to `Base()` that was already a documented difference puts it at 503, and the
harness's `indexOf`, when it finds no candidate greater than the previous one, falls to
`cands[0]` (`internal/sim/compare.go:970-972`) and the inverted pair comes out of that. It has
not reproduced in 30 later runs (3 full suites, 24 of the
`sim` package with `-count`, 3 with a synthetic load of 6 and 10 processes) nor in 13 of the base
before the change, with 10 load processes. **It stays open and has not been touched**:
neither of the two branches of that code is from this work.

**The second branch was ours and no longer is**: it is the `cands[0]` of `internal/sim/compare.go`,
and it is closed in §32.6 with two tests. The first one (the store's window at 503) is still
one of our documented differences, not a harness failure.

---

## 32. Field-by-field audit: the session object and the connection object

### WHY THIS ROUND

All the recent bugs were the same mistake: reading a field backwards, or reading it as
the name suggests instead of as it is written. The ones already fixed are in
§22, §24, §28, §30 and §31. That pattern says the payoff is in **systematically
enumerating the fields the original writes and confronting ours one by one**,
not in looking for more symptoms. That is what this section is.

The source is the binary: `sub_1400E1DE0` (the session's constructor, 130 lines) and
`sub_1400D4770` (the connection's, 47 lines), plus their writers. The meaning of
each field comes from **where it is written and where it is read**; nothing is inferred from a name.
When the decompilation does not resolve it, the row says **UNKNOWN** and moves on: that is a
useful finding, not a gap in the work.

Verdict per row: **matches**, **differs**, **we do not read it**, **UNKNOWN**.

### THE SESSION OBJECT

| dir | constructor (E1DE0) | what it is, with the addresses | ours | verdict |
|--|--|--|--|--|
| +0,+8,+16 | `= 0` (:20-22) | `vector<peer*>` of the match. The tick walks it (E2590:220), E52D0:169 erases it, and `*a1 == a1[1]` is "nobody is left" (E4EC0:65) | `Peers map[byte]*Peer` + `peerOrder []byte` | **matches**: the map is membership and `peerOrder` is AddPeer's order, which is the original vector's order (E5670:88-97) and where the pending list and the sub order come from (§33) |
| +24..+2071 | `memset(a1+24, 0, 0x800)` (:73) | table of 256 pointers, `peerID -> peer`. Written by `sub_1400E5670:100`, erased by E52D0:56 | `Peers[byte]` indexed by id | matches on the key |
| **+2072** | `*(WORD*)(a1+2072) = 0` (:23) | **the FIRST peer's message version**, cached when creating the store (`E2590:233-234`) and read when building the `0x0A` (`E2590:301`): `v31 = 6 or 4` depending on `v68 > 1`, and `E2590:346` puts the checksum block in only if `v68 > 1` | **does not exist** | **differs, we do not read it**. Measured: real clients announce **5** (`2026-09-30_201122`, LobbyConnect `body[17:19] = 05 00`), that is, the block always goes out and it is not visible. See 32.4 |
| +2080,+2088 | `= 0` (:25-26) | map `globalID -> peer` (`sub_1400E83E0`) | `Peer.Key` | matches on the key, not the map |
| +2096,+2104 | `*(_OWORD *)(a1 + 2096) = *a2` (:33) | `(titleID u16, sessionID u64)`, the sessions map's key | `SessionKey` | matches |
| +2112 | `*(BYTE *)(a1 + 2112) = 0` (:34) | **the host's peerID**, the byte the log prints as `hostPeerID=%d` (`sub_1400DFE60:153`). It is written **only** by `DFE60:156`, from `E5670:188` (the AddPeer with its host flag) and `DAAF0:243` (IntegrityCheck with `*(BYTE*)(msg+80) == 0`). It is read by `DF920:143` and `DE050:621`, and **both reads are behind the dispatcher's return value 2, which is dead code**: `DAAF0:141` requires `v63 == 8` inside a switch whose labels are 2,3,5,6,7,9 | `Session.hostID` | **matches**. The 32's name was right and the condition is no longer an UNKNOWN: the field decides a delivery ("host only") that the original does not have. See 32.8 |
| +2120,+2128 | `= 0` (:35-36) | a `shared_ptr` to the session's replay-metrics object. Single writer `E7500:23-25`, which only calls `sub_14010F4F0`, the callback **`"OnReceivePacket - PopReplayInfo"`** (`10F4F0:107`); readers `E74C0` and `E2050:202`. `E7500:33-35` writes the pointee's `session+2104` (the sessionID) at +0x10000 and the counter at +65548 | nothing | **we do not read it**, and it is not a gap: it is **metrics**, with nothing on the game path. See 32.8 |
| +2136 | `= 0` (:37) | the replay-info publication counter: `++` at `E7500:32`, read at `:34` and copied into the pointee | nothing | **we do not read it**, metrics |
| +2140 | `*(WORD *)(a1 + 2140) = 0` (:38) | one byte with its own getter and setter (`E7600` reads, `E7610` writes) and **a single pair of calls**, both inside `sub_14010F4F0`: `if (sub_1400E7600(session) == 0 && ...)` at the start of the callback and the setter at the end. That is an "the replay-info was already published" latch | nothing | **we do not read it**, metrics. Which metric it is is still **UNKNOWN**, but its use no longer is |
| +2141 | `= 0` (:39) | "the frame store exists". Set by E2590:230 and **cleared** by E47D0:91 | `Recording` | matches, and **it is now also turned off** (32.2) |
| +2142 | `= 0` (:40) | "the match has ended". Only writer: `*(_WORD *)((char *)a1 + 2141) = 256` at E47D0:91 and E52D0:189, which clears 2141 and sets this | `GameStarted` | **never written before**; 32.2 |
| +2144 | `= 0` (:40) | desync count, `++` E2590:380 | `DesyncCount` | matches |
| +2152 | `= a7` (:41) | the store's block size | `pageSize` | matches |
| +2160 | `= 0xEA60 / ticks + 1` (:71) | the store's cap in frames; `sub_1400EC5C0:33/37` rewrites it | `observableWindow` | matches |
| +2168 | `= a6` (:42) | the block-count scale, `ceilf(peers * this)` at E2590:241 | `maxPagesPerMatch * 0.125` | matches |
| +2172 | `= a3 / 0x3E8` (:43) | **the simulation period in milliseconds**. `1000 / this` are the ticks per second (E43E0:56, EC5C0:25) | `simulationTicksPerSec()` | matches: `main:597` leaves the flag in microseconds (`v42 = ms * 1000`) and E1DE0 divides it by 1000 |
| +2176 | `= 0` (:44) | tick counter, `++` E9E90:153 | nothing | **we do not read it**, metrics |
| +2184 | `= a4` (:45) | the observer update interval, in ticks | `ObserverUpdateInterval` | matches |
| +2188 | `= a5 * ticks / 100` (:72) | write-only | nothing | **we do not read it** (already noted: it is `settings+24`) |
| +2192,+2200,+2208 | `= 0` (:47-49) | 40 bytes per peer: the game-data queue. See §33: one record per peer in **join order**, and its 40 bytes are the chunk queue (`vector<vector<u8>>`), a byte counter and the `peerID` | `AckState.subs/queued` + `Session.peerOrder` | matches on the content, and now also on the order and on the record's death at disconnect (§33) |
| +2216 | `= 0` (:50) | frame count, `++` E2590:382 | `PageCount` | matches |
| +2224 | `= 0` (:51) | `+= size + 56` at E2590:361 | nothing | **we do not read it**, metrics |
| +2232 | `= 0` (:52) | reconnections, `++` EB5F0:136 and EBE60:64 | nothing | **we do not read it**, metrics |
| +2240 | `= 0` (:53) | "the most that has been resynchronised": E3260:34, E9E90:125, EBE60:90 | nothing | **we do not read it**, metrics |
| +2248..+2264 | `= 0` (:55-57) | 32 bytes per peer: the checksum entries | `AckState.perPeer` | matches |
| +2272,+2280,+2288 | `= 0` (:58-60) | the observer list, 16 bytes per entry | `observers` | matches |
| +2296 | `= 0` (:61) | `UpdateObservers`'s step counter | `observerTick` | matches |
| +2304,+2312,+2320 | `= 0` (:62-64) | the pending checksum block, serialised at E34F0:333-359 | `AckState.merged` | matches |
| +2328 | `= a8` (:65) | the observer cap (`observerMax`) | `MaxObservers` | matches |
| +2332 | `= 0` (:66) | "observable". Only writer: `sub_1400EC5C0:27`, which sets `(msg+48 != 0)`, that is 0 or 1 | `Reconnectable` | matches as long as the writer sets 0 or 1. Note: E2590:96 reads it as `== 1` and E3B50:79 as `!= 0` |
| +2336 | `= 0` (:67) | the u64 players publish with `0x1D` | `observableHash` | matches |
| +2344 | `= 0` (:68) | the lag in pages | `LagPages` | matches |
| +2352 | `= 0` (:69) | the frame store | `replay` | matches |

### THE CONNECTION OBJECT (which in the original is at once the peer)

| dir | constructor (D4770) | what it is, with the addresses | ours | verdict |
|--|--|--|--|--|
| +0 | `= 0` (:20) | the `peerIdentifier`, written **only** by E3B50:240 and compared at E43E0:43 and E46C0:25 | `peerIdentifier` | matches (§31) |
| +8 | `= 0` (:21) | the `peerID` | `Peer.ID` | matches |
| +16 | `*(_QWORD *)(a1 + 16) = -1` (:22) | **a u64, not a 32-bit profileID**: the constructor puts -1 with all eight positions set. It is written by `EB5F0:467` and `:488`, `= *(_QWORD *)(msg + 64)`, and the name of that field comes from the binary itself: `EB5F0:316`'s log prints `*(a3 + 64)` with the label **profileID** | nothing | **we do not read it**. The 32 named it profileID by assumption; now it is demonstrated, and it makes no difference because **there is no reader of conn+16 at all**. See 32.8 |
| +24 | `= 0` (:23) | the socket, the first argument that `D49A0:21` and `D8E10:199` pass to the manager of +32 | `Peer.Conn` | matches in function (see +32) |
| +32 | `= a2` (:14) | **the connection's socket manager**, and if it has readings: `D49A0:20-21` (sending one packet) and `D4A70:38` (sending several, the observer's path) take `conn+32`, pull slot 120 out of its vtable and call it with `conn+24` as the first argument; `D8E10:195-199` does the same through slot 136 on the read side | `PeerConnection` (the {manager, handle} pair in a single object) | **matches in function, differs in shape**. The 32 said "no reading found" and there are two: sending a message finds it |
| +40 | `= 0` (:24) | the peer's session | `Manager.peerSession` | matches |
| +64 | `= 100` (:45) | the LOBBYPING echo budget, DE050:710-713 | `pingEchoBudget` | matches |
| **+68** | `= 0` (:34) | the client's message version. Written **only** by the LobbyConnect handler (EB5F0:470 and :491), read at E4EC0:97 (`>= 4u`) and at DAAF0:141 | before: stored on the peer and read by nobody | **we do not read it**: 32.3 |
| +72 | `= 0` (:35) | the timeout override, with the gate `manager+129` (EB5F0:471) | `TimeoutOverride` | matches |
| +80 | `= 0` (:36) | the `secret` | `Peer.Secret` | matches |
| +88 | `*(WORD *) = 0` (:37) | **one byte, and it has readings.** It is written by `EB5F0:469` and `:490` with the literal 1, in the two branches nested under "the secret already found a peer". It is read by `DE050:136`, which is the **dispatcher's gate** (`if (conn+88 == 0) { if (the type != 21) return 0; }`, that is, a connection without a LobbyConnect could only send a LobbyConnect) and by `EB1E0:33`, which uses it as a slot index: `session[24 + 8*conn+88]` | nothing | **UNKNOWN** which connections come to be 1, and **differs**: we have neither the byte nor the gate. See 32.8 |
| **+89** | `*(BYTE *) = 0` (:37) | "this peer has uploaded a `0xFE` at least once": set by E2E00:42 and read **only** by E2590:226 | nothing | **differs, we do not read it**, and it is a branch: 32.4 |
| **+90** | `*(BYTE *) = 0` (:38) | "it is in the started match". Set by E2590:232 (for all, on store creation), cleared by E47D0:31, and read by E2E00:35, E34F0:89, E4940:76/85/130, E47D0:80, E52D0:178 | `Peer.Participating` | **we did not have it before**; 32.2 |
| +96 | `= 0` (:39) | the time of the last sync, E2E00:43 | `AckState.lastSync` | matches |
| +104 | `= 0` (:40) | **the peer's accepted sync messages**: `++` at `E2E00:202` (one per `0xFE` message, including those dropped for a full queue, and without counting if the peer does not exist or if `E2E00:35`'s gate rejects it). Its only reader is `EB5F0:341`, which puts it into the returned LobbyConnect's `dataReceivedCount` | `Peer.syncMessagesAccepted` + `Session.NoteSyncMessageFrom` | **matches, modelled** (commit `a6eb4d6`), with the limitation written in the code: no committed capture has a reconnection |
| +112 | `= 0` (:41) | the last activity, D8E10:164 | `lastSeen` | matches |
| +120 | `= 0` and `*(BYTE *)(a1 + 120) = 0` (:16-17) | **the close-reason string**, and the 32's row that said "+136,+144" describes this SAME object's queue. The constructor sets `+120` (the 16 bytes of the SSO buffer), `+136` (the size, to 0) and `+144` (the capacity, to 15): it is a 32-byte `std::string` based at +120. Both schedulers write there: `D4860:24` and `D48D0:22`, `sub_140081470((void **)(a1 + 120), reason, len)` | `Peer.closeReason` | **matches in role**, with the correction of the +136,+144 row |
| +136,+144 | `= 0` / `= 15` (:18-19) | **the queue of the +120 string**, not a second string: they are its size and its capacity. A second string at +136 would overlap the one at +120 | (the 32 associated them with `closeReason`) | **correction**: the 32 read them as an object of their own |
| +152 | `*(DWORD *) = 0` (:40) | **the close state word**, not "it has pending data": `D4860:16,20` sets it to 1 (schedule with a deadline) and `D48D0:14,18` to 2 (schedule immediately), and it is read by `D4930:12` (`!= 0 && now >= conn+112`, "the close is due") and by `D4970:12` (`== 2`). The 32 pointed it at `D8E10:166`, and `D8E10:164-176` is the observer's skip branch (`conn+156`, `conn+160`, `manager+604`) | `Peer.closeAt` with its "already scheduled" | **matches in behaviour**: the two states 1/2 live in the scheduler that is used, not in a field |
| +156 | `= 0` (:43) | "it is an observer", set by E3B50:244 | `observer` | matches |
| +160 | `= 0` (:44) | the observer's socket-skip counter | `observerSkip` | matches |

### THE FIELDS THAT DECIDE A BRANCH, ONE BY ONE

This is where an inversion hides, so the original's condition and
ours are side by side.

| where | the original's condition | ours | verdict |
|--|--|--|--|
| `E2590:96` | `session+2332 == 1 && (a1[285] - a1[284]) >> 4 != 0` | `s.Reconnectable && len(s.observers) > 0` | matches (the only writer sets 0 or 1) |
| `E2590:226` | `peer+89 == 0` or `session+2142 != 0` -> `goto LABEL_126` (neither built nor sent: no `0x0A`) | we have no `peer+89`; `GameStarted` **yes** | **differs halfway**: the `2142` term is there, the `peer+89` one is not (32.4) |
| `E2590:233` | `session+2072 == 0` -> `session+2072 = first_peer->68` | does not exist | **differs**, unimplemented |
| `E2590:303` and `:346` | `v31 = v68 > 1 ? 6 : 4`, and the checksum block only if `v68 > 1` | the block goes out whenever there are entries | **differs**, unimplemented |
| `E4EC0:65` | `session+2142 != 0` or `*a1 == a1[1]` | `s.GameStarted \|\| len(s.Peers) == 0` | matches |
| `E4EC0:97` | `v5 != 0 && progress == v6 && *(u16*)(entry->conn + 68) >= 4u` | the version was missing | **fixed**, 32.3 |
| `E4EC0:199,207` | `v2 = *(u32*)(a1+2184) + v36`, and `v2 + 1 > 2184` -> 0 | `interval = max(interval, 1)`, same wrap | differs on purpose: with interval 0 the original **hangs** in its `do while` |
| `E2E00:35` | `peer != 0 && (session+2141 == 0 \|\| peer+90 != 0)` | no gate | **differs**, 32.4 |
| `E3B50:79` | `maxObservers > 0 && session+2332 != 0` | `MaxObservers <= 0 \|\| !Reconnectable -> 4` | matches (inverse) |
| `E3B50:83-85` | `2141 != 0 && store->48 != 0 && *store->80 > 0` -> 3 | `Recording && replay != nil && Total() > 0 && Base() > 0` | matches |
| `E43E0:56-65` | `1000/2172`, and `lagged` chooses between `(2344 - 2216)/v6` and `-(2216/v6)` | the same | matches |
| `EC5C0:22` | `if (session+2141 == 0)` | `if s.Recording { return false }` | matches |
| `EB5F0:471` | `if (manager + 129) *(u32*)(conn+72) = msg+92` | `if s.config.AllowTimeoutOverride` | matches |
| `DE050:621` | `v54 == 0` or `v54 == 1 && v112 == session+2112` | the LobbyRelay relay is delivered to everyone, or only to the host if the dispatcher returned 2 | **matches**: no reachable path produces the 2, so it always ends at "everyone". See 32.8 |

### 32.1 THE BUG: A LEAVESIMULATION DOES NOT REMOVE THE PEER

`0x1F` is `sub_1400EC5E0`, which is three lines, and its only call is
`sub_1400E47D0(session, conn->peerID)`. Those 95 lines do four things and **not one
of them removes the peer**:

```
sub_1400E47D0:28-31   peer = session[24 + 8*peerID]; if there is none, do nothing
sub_1400E47D0:31      peer+90 = 0
sub_1400E47D0:32-53   empties its game-data queue (session+2192); the entry stays
sub_1400E47D0:54-72   empties its checksum entries (session+2248); the entry stays
sub_1400E47D0:74-93   if the session was recording and no peer with +90 is left,
                      *(WORD*)(session+2141) = 256
```

The only function that removes a peer is `sub_1400E52D0`, and its only caller is
`sub_1400EA460:40`, which is **the disconnect** and which also deletes the session once it is
empty (`sub_1400EA460:41-43`).

Ours did exactly the opposite, and it did it from the message's name:
`Manager.HandleLeaveSimulation` removed the peer from the session and deleted the session if it
was left empty. `Server.RemoveClient`'s comment even said that "the original
removes a session peer only with an explicit LeaveSimulation", which is reading by
the name again. With that, a `0x1F` removed the peer from the others'
`0x0A` pending list, and killed the whole session if it was the last one.

None of the five committed captures has a `0x1F`: measured over
`2026-09-30_201122`, which is the only one with a reliable type column (0x15 x4, 0x13 x144,
0x14 x20, 0x0F x3, 0x09 x2, and nothing else), no replay would have seen it. The
decompilation resolves it without ambiguity.

Fixed: `Manager.HandleLeaveSimulation` calls `Session.LeaveSimulation`, which removes
nothing. With the defect back in place:

```
--- FAIL: TestManagerHandleLeaveSimulationKeepsThePeer (0.00s)
    manager_test.go:457: PeerCount = 1, want 2: sub_1400E47D0 does not remove the peer
    manager_test.go:460: peer 1 should still be a member of the session after a LeaveSimulation
--- FAIL: TestSessionManagerHandleLeaveSimulation (0.00s)
    session_test.go:817: SessionCount = 0, want 1: a LeaveSimulation does not end the session
    session_test.go:820: PeerCount = 0, want 1
FAIL	github.com/luskaner/ageLANServer/battleserver/internal/session	0.481s
```

### 32.2 THE BUG: `session+2142` HAD NO WRITER

`*(_WORD *)((char *)a1 + 2141) = 256` is a single 16-bit store across two contiguous
bytes, and 256 is `0x0100`: in little endian it **clears `session+2141` and sets
`session+2142`**. The only two writers of `session+2142` in the whole binary are
that line at `sub_1400E47D0:91` and the same one at `sub_1400E52D0:189`, that is the two ways
of ending a match: the last peer that was in the game leaving, or a peer being
removed. `session+2142` is "the match has ended", not "the game has started" as our
field's name said.

What that byte is read by:

- `sub_1400E2590:226` reads it **before** deciding whether to build the `0x0A`, so with
  `2142` at one the tick goes to `LABEL_126` and **sends not one more `0x0A`**. We
  went on sending acks forever.
- `sub_1400E4EC0:65` reads it as the `v5` term of the observer's target, so
  an observer starts receiving the live end of the store and the end marker.

Fixed in all four parts needed for it to be coherent: `Peer.Participating`
(the `peer+90`), the flip to `GameStarted` in `Session.LeaveSimulation`, the `PopAckTick` that
goes silent when the match has ended, and `peer+90` for everyone at the moment the store is
created (`RecordFrame`, which is where E2590:232 puts it). With the flip removed:

```
--- FAIL: TestLeaveSimulationEndsTheRecordingWhenTheLastPeerLeaves (0.00s)
    leavesimulation_test.go:60: session+2141 should be clear once the match is over (sub_1400E47D0:91)
    leavesimulation_test.go:63: session+2142 should be set: it is the other byte of the same store
    leavesimulation_test.go:75: PopAckTick after the match ended returned 6 bytes / [], want nothing
--- FAIL: TestLeaveSimulationKeepsServingObservers (0.00s)
    leavesimulation_test.go:191: the match should be over after the last peer leaves
FAIL	github.com/luskaner/ageLANServer/battleserver/internal/session	0.502s
```

### 32.3 THE BUG: THE 0x22 MARKER DID NOT LOOK AT THE CLIENT VERSION

`sub_1400E4EC0:97` is `if ( v5 != 0 && entry->progress == v6 && *(WORD *)(*entry + 68LL) >= 4u )`.
They are three conditions, and the third one we did not have: **the message version of the observer's own
connection**. Offset 68 is written in one single place, the LobbyConnect
handler (`sub_1400EB5F0:470` and `:491`), so an observer that does not send a
LobbyConnect has it at 0 and receives no marker. The two observers of
`2026-09-29_012528` connect and go straight to the `0x1C`, and the original sent them
no `0x22` at all: measured, and now it fits the condition instead of luck.

The other arm, the `len(s.Peers) == 0` of our `v5`, **was** reachable without this: a
single-client session where that same client becomes an observer is left with no peers,
and there we sent a `0x22` the original does not send. So the fix was mandatory
before touching `session+2142`, not an extra.

The version lives on the connection (`PeerConnection.ClientMessageVersion`), not on the peer,
because it is read from the observer's connection and because a connection that did send a
LobbyConnect keeps it even if it later becomes an observer (`E3B50:237-254` does not touch
offset 68).

### 32.4 VERIFIED-DIFFERENT, AND DELIBERATELY NOT IMPLEMENTED

Four rows of the table are demonstrated differences that are **not** touched this round.
The reason is written because the next session has to decide with the same
criterion and not try again blindly.

1. **`session+2072`**: the first peer's version, and therefore whether the `0x0A` carries the
   checksum block. The original caches it from the **first element of the vector**, and
   our `Peers` is a map with no insertion order, so "the first" does not exist
   yet. And given what was measured (all real clients announce 5) the block comes out the same
   on both sides, that is, fixing it would not change one byte of any capture.
   What would have to be done first is an insertion order, and that touches `AddPeer`,
   `ReconnectPeer` and `RemovePeer`.

2. **`peer+89` and the gate at `E2590:226`**. The original **does not create the store nor send
   the first `0x0A` until ALL the peers have uploaded at least once**, because the loop
   at E2590:226 exits as soon as it finds one that has not. We start with the
   first one. With a single peer it is not noticeable, and the captures that exist do not distinguish it:
   the bsPort one, the only one with two clients, has **one** peer in the original (§29) and
   **two** in ours, so the gate would stop the acks of a session the
   original keeps serving. Implementing it without settling the peer divergence first
   would make it worse.

3. **`peer+90` at `E2E00:35`**: `if (v4 != 0 && (session+2141 == 0 || peer+90 != 0))`.
   Translated: once the store exists, a peer **that joins later** does not have the
   mark (E2590:230-232 is on the path that is no longer walked) and its uploads are
   silently discarded. We accept them. Same reason as 2, and it is the
   "a player joins a match in progress" case, which the game does not allow and which no
   capture has.

4. **The `0x0A`'s pending list is not the list of peers.** THIS CLAIM
   WAS HALF RIGHT AND §33 CORRECTS IT. What follows is what §32.4 said before reading
   `sub_1400E34F0` whole, and it is kept so the mistake is visible: it walks the vector of
   `session+2192` queues, not the peers vector; for each entry whose peer exists **and
   has `peer+90`**, if the queue is empty it puts its `peerID` in the list. Two of the three
   differences listed there are real and are fixed (the order is by join, not by slot; and the cut by `peer+90` **does not exist**, because the jump
   the audit read as the end of the list is the bottom of the loop). The third, "a peer
   that joins and does not upload", is also real, and for the same reason. The harness's
   `[pending-peers]` rule still compares **sets**, not counts nor order.

### 32.5 WHAT IS STILL UNEXAMINED

> **Updated by the 36 (2026-10-03).** The nine session fields and six connection
> fields listed here **are no longer UNKNOWN**: each one has a verdict in the
> two tables above and the detail is in **32.8**. Of this list only `+2176`, `+2224`, `+2232` and `+2240` of the session remain unexamined, and they are
> metrics counters read only by `sub_1400E9E90` (the munin) and `sub_1400EBE60` (the state
> dump), and they were already classified. The two points below are also
> resolved: the first is the 36's fix and the second turns out to be a byte
> **dead** in the original.

From the session object: `+2112`, `+2120`, `+2128`, `+2136`, `+2140`, `+2176`, `+2224`, `+2232`,
`+2240`. The last five are metrics counters read only by `sub_1400E9E90` (the
munin) and `sub_1400EBE60` (the state dump); the first four were **UNKNOWN**, not
"we are not interested": they were structures with writers and no reader located.

From the connection object: `+16` (profileID), `+32`, `+88`, `+104` (already known: the reconnect's
LobbyConnect echo), `+120`, `+152`. `+32`, `+88` and `+120` were **UNKNOWN**.

Two more differences that were found, and which the 36 leaves as they are:

- ~~The original's tick kicks observers with `sub_1400D4860(*v6, "Observer session not
  available", 0)`, that is with **deadline 0**, which the socket loop executes by throwing it
  away, and it **does not clear the list** (`E2590:184`). We cleared the list and did not
  close the connection.~~ **Resolved in the 36**: the close exists and now we do it too, but **after**
  releasing the lock, which is where the self-deadlock was. The list we still empty, which is still
  the deliberate difference (the entry goes the same way when the connection closes, `E52D0:227-242`).
  See 36.1.
- `E2590:483` returns two status bytes (`v85`), which `E9E90:70-71` reads. We do not
  return anything. **Resolved in the 36**, and the high byte turns out to be **dead in the
  original**: see 36.3.

### 32.6 THE HARNESS'S `indexOf`: THE FALLBACK THAT FABRICATED THE INVERTED PAIR

Besides the table, the open flake from the end of section 31 was located and
fixed here. `internal/sim/compare.go` fell to `cands[0]` when no candidate was
greater than the previous one. `cands[0]` is the **earliest** occurrence of that body in the
match, and with a repeating body (the heartbeat is five hundred in a row) it is the opposite
end of a long list; besides the guess **became the cursor**, so `prev`
went backwards and everything after it was aligned against a position no
frame of the window had reached. An unplaceable frame misaligned the rest of
the window, and from that came the phantom inverted pair when the store's block
was recycled before the observer's first visit.

Now the frame is reported at the **closest** candidate that does not advance, and the cursor
stays where it was. The must-match rule is not weakened: a frame that cannot advance
still does not go up, and with the defect in place the second test still sees the real inverted pair.
With the defect back in place:

```
=== RUN   TestIndexOfDoesNotWalkTheCursorBackwards
    matchindex_test.go:66: unplaceable frame #7 (A) reported at match frame 0, want the nearest candidate 3 (cands[0] is the earliest occurrence in the match and invents a reversed pair); full: [0 1 2 3 4 5 6 0 1]
    matchindex_test.go:73: frame #8 (G) placed at match frame 1, want 7: one unplaceable frame must not move the cursor backwards and misalign the rest of the window; full: [0 1 2 3 4 5 6 0 1]
--- FAIL: TestIndexOfDoesNotWalkTheCursorBackwards (0.00s)
=== RUN   TestIndexOfStillReportsAGenuineReversal
--- PASS: TestIndexOfStillReportsAGenuineReversal (0.00s)
FAIL	github.com/luskaner/ageLANServer/battleserver/internal/sim	1.104s
```

### 32.7 WHAT WAS MEASURED

With the working set of this repo's test binaries added up every 200 ms, in
`src_go`, and with nothing else running:

| | before (previous method) | after (this measurement) |
|--|--|--|
| `go test -short ./... -count=1` (short loop) | 40.6 s / 77.0 MB | **41.6 s / 112.7 MB** |
| `go test ./... -count=1 -timeout 25m` (full) | 766.9 s / 115.5 MB | **766.9 s / 43.1 MB** |
| `go test -race sim+server+session -timeout 25m` | 772.4 s / 220.5 MB | **779.8 s (sim) + 35.8 s (server) / 125.0 MB** |

The two RAM figures **are not comparable**: the previous round's sampler counted
processes this one does not, so the right-hand column's figure is the sum of the
`.test` binaries alive at the instant of the peak, with the `*go-build*` filter. What
is comparable is the **time**, and it is the same: the full suite stays at 766.9 s,
exactly the 766.9 s the previous round left, and the short loop at 41.6 s against
40.6 s. The short loop's 112.7 MB are `server.test` and `sim.test` at once.

The new tests are microseconds, synthesise nothing and add no memory.

### 32.8 THE ELEVEN FIELDS 32.5 LEFT UNEXAMINED, one by one

All of them are resolved by reading the binary, and **none needs a new capture**. The
table above already carries each one's verdict; here is why, which is what
is needed so that they are not read again.

From the session object:

| field | what it is, with the addresses | verdict |
|--|--|--|
| `+2112` | the **host's peerID**. The binary names it: `sub_1400DFE60:153` prints it as `hostPeerID=%d` and :156 writes it. One writer, two calls (`E5670:188` with the AddPeer's host flag, `DAAF0:243` with IntegrityCheck's `migrateState` 0) | **matches** our `Session.hostID`, and the detail that was missing is that **it decides nothing**: its two readers (`DF920:143` and `DE050:621`) compare against the dispatcher's return value 2, and that 2 is only set by `DAAF0:141-154`, which requires `v63 == 8` (DROPOUT) inside a switch whose labels are **2, 3, 5, 6, 7 and 9**. That is, dead code, and that is why the original has no "host only" delivery |
| `+2120`,`+2128` | a `shared_ptr` to the session's replay-metrics object. Only writer `E7500:23-25`, only caller `sub_14010F4F0`, which is the callback whose literal is **`"OnReceivePacket - PopReplayInfo"`**; `E7500:33-35` writes the sessionID (`session+2104`) and the counter into the pointee | **we do not read it**, and it is not a gap: it is **metrics** |
| `+2136` | that callback's publication counter (`++` at `E7500:32`) | **we do not read it**, metrics |
| `+2140` | one byte with its own getter and setter (`E7600`, `E7610`) and **a single pair of calls**, both in `sub_14010F4F0`: `if (E7600(session) == 0 && ...)` on entry and `E7610(session, ...)` on exit. An "already published" latch | **we do not read it**, metrics. Which metric it is is still **UNKNOWN**; its use no longer is |

From the connection object:

| field | what it is, with the addresses | verdict |
|--|--|--|
| `+16` | a **u64** at -1, written by the LobbyConnect with `msg+64` (`EB5F0:467`, `:488`). The name of `msg+64` is printed by the binary itself: `EB5F0:316`'s log calls it **profileID**. **No reader of `conn+16` at all** | **we do not read it**. The 32 named it profileID by assumption and was right; the correction is that it is a u64 and that it does not matter because it is not read |
| `+32` | the connection's **socket manager**. If it has readings: `D49A0:20-21` (sending a message) and `D4A70:38` (sending several, the observer's path) take it, pull slot 120 out of its vtable and call it with `conn+24` as the first argument; `D8E10:195-199` does the same through slot 136 on the read side | **matches in function** with our `PeerConnection`, which is the {manager, handle} pair in a single object |
| `+88` | one byte that the LobbyConnect sets to 1 (`EB5F0:469`, `:490`), read by the **dispatcher's gate** (`DE050:136`: if it is 0, the only class of message handled is the LobbyConnect) and by `EB1E0:33`, which uses it as a slot index (`session[24 + 8*conn+88]`) | **UNKNOWN** on which connections it comes to be 1, and **differs**: we have neither the byte nor the gate. The only two writers are in the branches nested under "the secret already found a peer" (`if (v15 != 0)`, `EB5F0:131`), and that **cannot be** true for a first connect, which the captures give as served; either the decompilation nests it wrongly there or there is a writer that does not appear. It is written down and not implemented: a "LobbyConnect only" gate would reject the observers of `2026-09-29_012528`, which go straight to the `0x1C` |
| `+104` | the peer's **accepted sync messages**: `++` at `E2E00:202`, one per `0xFE` message, including those dropped for a full queue, and without counting when the peer does not exist or `E2E00:35`'s gate rejects it. Its only reader is `EB5F0:341` | **matches, modelled** (commit `a6eb4d6`). See 36.2 |
| `+120` | **the close-reason string**. The constructor sets `+120` (the 16 bytes of the SSO buffer), `+136` (size 0) and `+144` (capacity 15): a 32-byte `std::string` **based at +120**. Both schedulers write there (`D4860:24`, `D48D0:22`) | **matches in role** with `Peer.closeReason`. **Correction of the 32**: its `+136,+144` row ("the disconnection reason string") describes this same object's queue, not a second string, which would also overlap |
| `+152` | the **close state word**: `D4860:16,20` sets it to 1 and `D48D0:14,18` to 2; it is read by `D4930:12` (`!= 0 && now >= conn+112`) and by `D4970:12` (`== 2`) | **matches in behaviour**. The 32 read it as "it has pending data" citing `D8E10:166`, and `D8E10:164-176` is the **observer's skip** branch (`conn+156`, `conn+160`, `manager+604`); our `incoming`'s name came from there. The states 1/2 live in the scheduler that is used, not in a field |

**"We do not read it" is a finding and not a gap**, and of the eleven only two
genuine UNKNOWNs are left: `+2140` (which metric the latch is) and `+88` (which connections set
it to 1). Both are written above with what is needed to close them.

---

## 33. The `0x0A` pending list: what it is, where each byte comes from, and why there is no cut

### WHY THIS ROUND

§32 left the pending list as the most serious candidate remaining, and left it
badly read: it said the walk **stops** at the first entry whose peer is missing or not
participating. That is not what the code does, and the difference is not a nuance: a
"cut" and a "skip" produce different lists as soon as there is a non-participating peer in the
middle. §32's lesson (that all the recent bugs were reading a field or a condition
backwards) applied to the audit itself, so this section reads
`sub_1400E34F0` whole and what it calls, and then field by field against ours.

### WHAT THE LIST IS

The walk is `sub_1400E34F0(session, msg, sizeCounter, checkTimeouts)`, and it is called by
`Session::Tick` (`sub_1400E2590:314`) **once with `checkTimeouts=1` and up to sixteen
more times with `checkTimeouts=0`** (`:315-320`, while the previous call returns something
other than zero). Translated line by line:

```
:80-81   v10 = session->queue.begin;  if ((end - begin)/40 == 0) goto LABEL_69
:88      peer = session->peerTable[record->peerID]        (session + 24 + 8*peerID)
:89-90   if (peer == 0 || peer->+90 == 0) goto LABEL_10    -> SKIPS the entry
:91-92   if (record->begin == record->end) {              <- "queue empty"
:94-98       if (!checkTimeouts) goto LABEL_10
:100          if (sub_1400D4BD0(peer, now))  sub_1400D48D0(peer, "Sync data timeout expired")
:106-118      else  msg->pending.push(record->peerID); ++*sizeCounter
:120          goto LABEL_10
          }
:124-231  else: copies the HEAD of the queue into the `0x0A`'s sub, erases it (erases the
          first one), and adds payload+3 to the counter
:289-300 sawData = sawData || "I have drained something"
:320-323  LABEL_10: v9 = sawData;  LABEL_11: ++i; if (i >= count) break     <- the only exit
:329-400 if (!sawData) -> serialises the session+2304 block and empties it
```

The list is then: **the peers that are participating and whose game-data queue is
empty at the moment the tick starts**, in the order the records are in, and
minus those past their sync deadline (which are closed in that same
branch). "Queue empty" is literally `record->begin == record->end` (`:91-92`): the
record has a queue of chunks and has none.

### WHY THERE IS NO CUT

Because `LABEL_10` is not the end of the list: it is `v9 = v52; goto LABEL_11`, and `LABEL_11`
is `++v12; v11 += 40; if (v12 >= count) break` (`:320-326`), that is **the bottom of the loop**.
The only exit is the counter, and an entry that fails the `:89` test is skipped and the
walk continues. The audit read the skip as a `break`.

There is a second thing that is read wrong if you do not look at the caller: the list is written **only
on the pass that checks deadlines**. The sixteen following passes go through `:94-98` and leave
before `:106`, so they add nothing. That is why the list is "the state at the moment the
tick starts" and not "the state at the end".

### THE RECORD'S 40 BYTES, AND WHERE THEY COME FROM

```
+0,+8,+16   std::vector<std::vector<u8>>: the peer's game-data queue. One chunk per
            each `0xFE` uploaded (E2E00:190-199), each chunk with the [len u16][body] of the
            message as it was. Cap: 0x12C chunks or 0x80000 bytes (E2E00:53-54), which is our
            `maxAckQueueEntries` / `maxAckQueueBytes`.
+24         byte counter: each upload adds to it (E2E00:199) and the drained
            chunk takes from it (E34F0:236), and it is the second term of the byte cap. It is the only
            consumer of that field: nobody else reads it.
+32         the peerID (E2E00:46 finds it by walking, E34F0:88/106/146 read it)
```

One record per peer **when it joins** (`sub_1400E5670:208-219`) and **it is deleted when it
disconnects** (`sub_1400E52D0:57-96`, with `sub_1400E9B70` destroying whatever chunks
were left). The vector is therefore in **join order**. LeaveSimulation
(`sub_1400E47D0:46-52`) empties it but **leaves it in its place**, with its `peerID` and its
position: that is why a `0x1F` puts a non-participating record IN THE MIDDLE of the vector, which is
the case that separates "cut" from "skip".

### EACH FIELD OF THE `fe 00 NN MM ...` BODY, FROM ITS ORIGIN

From the serialiser, `RLink::SessionMessage::MultiGameDataMessage__vmethod_18` (`1400e92b0`),
which is the only place that writes it:

```
body   = [FE][subCount] + subs + [K] + pending + [blockLen u16 LE] + block
sub      = [peerID u8] (:90) [len u16 LE] (:98) [payload] (:105)
block    = [len u16 LE][02][N u16 LE] + 8N bytes, and N is (end-begin)>>3 of session+2304
```

- `fe` is byte 48 of the message, which `Session::Tick` fills with -2 (`E2590:300`) and the
  serialiser writes at `:60`. It is not one of our "flush markers": it is the first byte
  of the body, and it is -2 in every `0x0A`.
- `NN` is `(end-begin)>>5` of the subs vector (`:64`), that is, how many chunks were
  drained this tick.
- `MM` and the bytes that follow are the pending vector: its length at `:116` and its
  bytes at `:140`. The bytes are `record->peerID` as they are, from the record that names the
  peer.
- The final block is only written if the client's message version is **greater than 1**
  (`session+2072`, `:146`), and the block is serialised by the drainer itself from
  `session+2304..2312` (`E34F0:333-360`), which empties it when serving. Empty it is `00 00`.

That is, **`fe 00 01 01 00 00` is "no sub, one pending, peer 1, empty block"**:
a one-peer session that is participating and uploaded nothing this tick. And
`fe 00 02 01 02 00 00` is the same with two peers. The bsPort capture's
difference is therefore **two peers where the original has one** (§29), not a different list
rule: that capture's `[pending-peers]` still says exactly the same thing as
before this round (`original: [] en 1 ack, [1] en 7` against
`ours: [] en 1 ack, [1 2] en 7`).

One detail of the block that comes out of reading the call loop and not the drainer: the block is
added **only if the pass drained nothing** (`E34F0:329`), but the pass that adds it is
always the last one of the loop at `E2590:315-320`, and that one returns zero because there is nothing
left. That is, it is always added, unless the sixteen passes run out. It is not a
"if there was no flush, there is no block": it is an "if the pass cap was exhausted, there is no block",
and in the normal case our behaviour (add the block and empty `session+2304`) is the
original's.

### FIELD BY FIELD AGAINST OURS, AND WHAT CHANGED

| what | the original | ours before | verdict |
|--|--|--|--|
| the list's order | the records vector's order = join order (E5670:208-219) | ascending slot order | **differs, fixed** |
| non-participating peer | does not go out (E34F0:89) | it did | **differs, fixed** |
| peer past its deadline | does not go out, and is closed (E34F0:100-103) | it did | **differs, fixed** |
| the match's newest `0x0A` | the `peer+90` marks are already set (E2590:230-232 before :314) | it was built before marking them | **differs, fixed** |
| a disconnected peer's data | dies with it (E52D0:57-96 + sub_1400E9B70) | it kept being drained | **differs, fixed** |
| the subs' order | that of the same walk, and within each record that of arrival | slot order | **differs, fixed** |
| reconciliation's "the first" | the first record with `peer+90` (E4940:85-86) | the lowest slot, without a `peer+90` filter | **differs, fixed** |
| the queue cap's bytes | only the payload (E2E00:54) | `1 + payload` per sub | differs, **not touched**: 512 KiB against 512 KiB + 300 bytes, and the cap is a drop |
| chunks per tick | at most 17 (one per pass, E2590:315-320) | all of them | differs, **not touched**, see below |

The six "fixed" ones are the same code in three places: `Session.peerOrder` (the order),
`Session.queueEntries` (the three filters, including the deadline, which now shares the
computation with `SyncTimedOutPeers` instead of recomputing it), `Session.ensureRecording` (which
is called **before** building the body, not when the frame is added), and
`Session.RemovePeer` (which empties both of the peer's queues).

### WHAT IS LEFT UNDONE, AND WHY

- **The cap of 17 chunks per tick.** The drainer takes **one** chunk per call (it copies the
  queue's head and erases the first one, `E34F0:124-231`), and the tick calls it up to
  sixteen more times. That is, a tick carries at most 17 chunks and the rest stay
  queued for the next one. It is the most clearly demonstrated of everything remaining, and it is not
  implemented because it changes **how many `0x0A`s go out per second**, which is exactly
  what `[tick-cadence]` and `[message-count]` have as informative and what we already
  know we group differently (§28: four extra acks in second 26). Besides, the
  frame store keeps the whole `0x0A`, so changing the grouping changes the
  replay's content, and no capture separates "17 chunks" from "all of them": with
  a client that uploads one chunk per tick, which is what the five captures are, both
  behaviours give the same bytes.
- **`session+2072`** (§32.4 point 1): now the session knows the join order, so
  "the first peer" is available, but with all real clients announcing 5
  (§32.4) it does not change one byte of any capture, and the **moment** at which the version is
  read still is not established: the peer's LobbyConnect can arrive after
  the store exists. That is the reason it is not done "because now it can be".
- **Reconciliation compares with more peers than the original compares with.**
  `sub_1400E4940:114-116` compares the first record with participating against **the
  next record in the vector**, which need not be participating; ours compares
  against all of them that have participating. With two peers they agree; with three they only
  differ when the peers do not agree with each other, and no capture has three peers.

### WHAT IS STILL NOT KNOWN

Nothing new in this area. What was **UNKNOWN** and still is, is that of §32.5, which has
not been touched: `session+2112`, `+2120/+2128/+2136`, `+2140`, and `conn+16`, `+32`, `+88`,
`+120`.

What has been closed in this area: the record's byte `+24` (the byte counter)
**is read by nobody** other than `E2E00:54`'s cap; the `+8` and `+16` are the queue vector's `end` and
`cap`, not a counter; and the "first upload order" that
§32.4 said is no order at all: the record is created when the peer joins, not when it uploads.

### WHAT WAS MEASURED

With the working set of this repo's test binaries added up every 200 ms, in `src_go`,
and with nothing else running (the same method as §32.7):

| | before | this measurement |
|--|--|--|
| `go test -short ./... -count=1` (short loop) | 41.6 s / 112.7 MB | **43.2 s / 111.5 MB** |
| `go test ./... -count=1 -timeout 25m` (full) | 766.9 s | **767.7 s / 114.3 MB** |
| `go test -race sim+server+session -timeout 25m` | 779.8 s (sim) + 35.8 s (server) / 125.0 MB | **780.0 s (sim 771.0) + 35.9 s (server) / 322.3 MB** |

The time is the same: the full suite stays at its measured floor of ~767 s and the short
loop at 43 s, and the new tests are microseconds. **The `-race` RAM is NOT
comparable with §32.7's** and it has to be said: 322.3 MB against 125.0 MB measured by that
round's sampler. That sampler sums the working set of the `*.test.exe` under
`go-build*`; measured by it, a single `-race` binary already takes 73.5 MB
(`internal/session` alone), that is, the peak is the race detector of `sim.test` and
not this round's tree, which in `internal/sim` only touches comments. The useful
comparison is the time, and it has not changed.

What the captures say after the change, with none of them moving: the five
are still at **MUST-MATCH 0**, and the bsPort one still reports `[pending-peers]` with the
same numbers as before (`ours: [] en 1 ack, [1 2] en 7` against
`original: [] en 1 ack, [1] en 7`). That nothing changed is what was expected: in the five
captures the ids are laid out in join order (so slot order and record order
coincide), there is no `0x1F` at all (so no non-participating record), and no
peer passes its deadline (so the deadline filter removes nobody). The three cases
that §32.4's rules no longer distinguish are exactly the ones no capture has, which
is what was said when they were written and is still true: **this round fixes the
behaviour, not a measured discrepancy**.
---

## 34. The cost of a long match: two hours in one second (2026-10-02)

### WHY THIS ROUND

Everything that had been timed was the harness, which is slow ON PURPOSE because
it replays the captures in real time. The longest recording is 24 minutes
(`2026-09-29_101831`, the one §26 decided not to import), so "a LONG
match stays bounded" had never been asked of anything. With the requirement that long
matches work, the question is ours: **does the memory grow, and does the
per-tick cost grow**.

The fixture does not sleep. `Session.PopAckTick` and `Session.Tick`
are driven directly, in the order the server's main loop runs them
(`server.go:374-376`), so **57600 ticks of match (2 hours at 125 ms per
tick, the `simulationPeriod` that every capture records) cost 0.3 s of wall
clock**. The captures' horizon is not trimmed: it just stops being paid.

`src_go/internal/session/longmatch_test.go`, four tests, all in the short
loop:

| test | what it asserts |
|--|--|
| `TestLongMatchMemoryAndCostAreBounded` | the memory plateau against its baseline, the curve's slope, the per-tick allocation cap, and that the per-tick cost does not grow |
| `TestFrameStoreRecyclesWithoutGrowing` | the same with a store that recycles 28000 times instead of 40, plus the ring's structural invariants |
| `TestObserverServeCostDoesNotGrowWithTheMatchLength` | the cost of ONE observer visit, first half against second half of the horizon |
| `TestObserverChurnDoesNotGrowTheList` | 1000 observers that come and go, against an identical control that saw none |

### THE FIXTURE, CAPPED IN THE CODE

Every cap is a named constant, in the file, with its reason. **The direct
lesson of the test that reached 20 GB (§25)**: a synthetic horizon that lives in
someone's head is a horizon the next one duplicates without anyone noticing.

| cap | value | why that one |
|--|--|--|
| horizon | 2 h = 57600 ticks | the longest capture is 24 min |
| players | 2 | the shape of _6's two peers |
| observers | 2 | the two that `2026-09-29_012528` brings |
| `-pageSize` | 65536 | the original's own default (`main_1400d1310:347`): 8 blocks, 512 KiB |
| `-maxPagesToStorePerSession` | 0 | the original's default, f = 1.0 |
| `session+2184` | 8 ticks | the original's default: one visit every nine ticks |
| `0xFE` payload | 150 B | §24 measured 259651 B of acks in 1570 frames |
| checksum records per upload | 4 | reconciliation drains 1 per peer and round, and at most 10 rounds per tick (`E2590:374-381`): four fits, more than 10 would fill the queue **in the original too** |
| observers in the recycling test | they join at tick 2 | with `-pageSize 1024` the store cycles every 16 frames, so at 600 ticks it would already answer 3 |
| churn | 1000 observers | it has to fit in the store's first window (about 1400 frames), because outside it `StartObserve` answers 3 and does not touch the list |

### WHAT WAS MEASURED

With the LIVE heap (`HeapAlloc` **after a forced GC**, which is the only thing
that says whether something is still reachable), 57 samples spread over the 57600 ticks:

```
long match: 57600 ticks = 7200 s of match at 125 ms/tick, 225 buckets of 256 ticks, 57 memory samples
live heap after a forced GC: baseline 799 KiB, plateau 820 KiB (min 761 KiB), drift 0.26 B/tick
allocs/tick: first quarter 30.01, last quarter 30.01 (ceiling 48)
tick time: first quarter 5928 ns (84.98ms), last quarter 6412 ns (91.92ms), fastest of 2 repetitions
recycling store: the window holds 16 frames, so 57584 of 57600 frames went out through
                 dropOldestBlockLocked over 57600 ticks; array capacity at most 16
live heap: baseline 377 KiB, plateau 337 KiB, drift 0.04 B/tick
first half: 28800 ticks in 173ms, 6268 visits, 57597 frames served (27678 ns and 9.2 frames a visit)
second half: 28800 ticks in 181ms, 6400 visits, 57600 frames served (28281 ns and 9.0 frames a visit)
churn: 1000 observers joined and left over 1000 ticks; live heap 328 KiB with the churn
        against 328 KiB without it, so 0 KiB left behind by 1000 connections
```

What those numbers say, one by one:

- **The memory plateaus.** Baseline 799 KiB, plateau 820 KiB: **21 KiB of
  growth over two hours**, and the least-squares slope over the
  last 53,248 ticks is **0.10 to 0.30 B/tick** (0.04 in the fixture that
  recycles 28000 times). The 512 KiB the store occupies is the plateau; the rest is the
  test binary.
- **The allocations per tick are flat.** 30.01 in the first quarter and 30.01 in
  the last: **zero drift**, and the number is exactly reproducible. With
  `-race`, 31.00 and 30.98. It is all from the ack path: the pending list, the
  flushed map, the body, the subs slice, the reflective `sort.SliceStable` of the
  flush with mixed slots, and the eight `make([]byte, 8)` of `QueueEntries`.
- **The store recycles and does not copy.** In the realistic fixture the store goes round
  about 40 times; with `-pageSize 1024`, **57584 of 57600 frames go out through
  `dropOldestBlockLocked`** and the frames array stays at capacity 16 with a
  16-frame window. `cap(r.frames)` is the assertion that says so: recycling by copying into a
  bigger array breaks it immediately.
- **An observer visit costs the same in minute 120 as in minute 1.**
  27678 ns and 9.2 frames per visit in the first half; 28281 ns and 9.0 in the
  second. The **9 frames per visit** are what the store imposes (one visit every nine ticks, one
  frame per tick), and §24's 200-frame burst is the first visits of an observer that
  arrives late.
- **The observer list does not grow.** 1000 that come and go leave **0 KiB**
  above an identical control that saw none. And the residue **does not scale
  with the number**: measured 41, 3, 3 and 3 KiB for 125, 250, 500 and 1000
  observers. The 41 in the first case is the allocator warming up, and that is why the
  test keeps the minimum of three runs.

### THE CAPS, AND WHERE EACH ONE COMES FROM

All of them are in the code, next to the measured number that chose them, and all
are checked in both modes (normal and `-race`).

| cap | value | measured | margin | what it catches |
|--|--|--|--|--|
| heap growth | 4 MiB | 21 KiB | 190x | any leak that reserves something per tick (4 MiB / 57600 = 72 B) |
| curve slope | 8 B/tick | 0.30 B/tick | 26x | a one-slice-header-per-tick leak (48 B) already fails it |
| absolute heap ceiling | 32 MiB | 820 KiB | 40x | a leak that is already in the baseline |
| allocations per tick | 48 | 31 (with `-race`) | 1.55x | a new path that reserves per tick |
| allocation growth | 8 per tick | **-0.03** | — | constant degradation |
| cost per tick | 3x + 5000 ns | <= 1.02 (no load) / 1.56 (with the whole short loop in parallel) | — | an order-of-magnitude degradation |
| cost per visit | 3x + 5000 ns | <= 1.01 | — | the same on the observer's path |

### WHAT THIS FILE DOES NOT MEASURE

Said here so that the next session does not read it as more than it is:

1. **The sync timeout.** The clock inside the session is the wall one
   (`queueEntries` takes a `time.Time` and `syncTimedOut` subtracts it), so two hours in
   two seconds cannot make a peer expire at 60 s.
   The fixture's players upload every tick, and the timeout is not what is being
   probed.
2. **Compile-time wraps and integer widths.** A 2 h horizon is
   57600 frames; `session+2216`, `session+2296` and the ring index are a long way from
   any limit. A test that took those counters to the limit would have to synthesise
   billions of frames, which is exactly the mistake of §25.
3. **The measurement's own clock is wall and is quantised.** Measured on
   this machine: timing with `time.Since` a 700 ns operation gives 0 in
   95% of samples, and **19 of the 55 buckets of the first quarter give
   exactly 0**. That is why nothing here times a single tick or a single visit: each
   quarter is measured with **one single pair of readings** spanning 14336 ticks
   (60 to 120 ms), and it is repeated twice keeping the faster one, because the
   machine's load only adds time. With that the worst late/early ratio
  measured drops from 1.26 to 1.02.

### THE RESULT: THERE WAS NOTHING TO FIX

**Not one line of production was touched.** The server does not leak, does not drift and does not
get more expensive over two hours of match. And it is not a green over two empties: each
test additionally asserts what must really have happened, and fails if not:

- one frame per tick in the store, and `Base() > 0` (that is, that recycling happened)
- one `0x0A` per tick to each player, and both observers served and with no
  growing delay against the store
- `DesyncCount == 0` with 4 checksum records per tick **identical** between the
  two peers, which is what proves reconciliation was really exercised
- in the recycling case: `len(blockAt) <= 8`, `cap(frames) <= 2x` the high-water mark, the
  retained frames below what fits in the byte window, and a late observer with
  **status 3 and not registered** over a store that has let go of 57584 frames

### THE THREE FALSIFICATIONS

A test that has never been seen to fail is not a test. All three were put into the
production code by hand, measured, and removed. `git status` afterwards: clean
apart from the new file.

**1. A leak in the store** (a packet-level slice kept per frame, the typical oversight):

```
=== RUN   TestLongMatchMemoryAndCostAreBounded
    longmatch_test.go:598: the live heap grew from 1631 KiB to 22003 KiB over 57600 ticks, a growth of
        20371 KiB; the budget is 4096 KiB. A heap that only goes up has the shape of a leak, and it took
        362 bytes a tick to do it
    longmatch_test.go:604: the live heap drifts at 387.23 bytes per tick over the last 57600 ticks
        (from 2023 KiB to 22003 KiB); the budget is 8 B/tick.
--- FAIL: TestLongMatchMemoryAndCostAreBounded (0.99s)
=== RUN   TestFrameStoreRecyclesWithoutGrowing
    longmatch_test.go:746: the live heap grew from 66351 KiB to 88219 KiB over 57584 recycled frames;
        the budget is 4096 KiB. Recycling that copies, or that lets the buffers behind the frames grow,
        looks exactly like this
    longmatch_test.go:751: the live heap drifts at 379.93 bytes per tick over 57584 recycled frames;
        the budget is 8 B/tick
--- FAIL: TestFrameStoreRecyclesWithoutGrowing (0.61s)
```

Both memory caps bite, and the slope one bites 48 times before
the maximum one, which is exactly what it is for.

**2. Recycling by copying into a bigger array** (`dropOldestBlockLocked` stops
moving within the array and builds one of `cap+64`):

```
    longmatch_test.go:604: the live heap drifts at 13.02 bytes per tick over the last 57600 ticks
        (from 820 KiB to 1496 KiB); the budget is 8 B/tick.
    longmatch_test.go:630: the average tick costs 9829 ns in the last quarter against 6669 ns in the
        first; the budget is 9670 ns (3x plus a 5000 ns floor).
    longmatch_test.go:648: the store's frame array has capacity 23808 for a high-water mark of 1296
        retained frames; the recycling memmoves inside the array and never grows it, so a capacity
        above twice that is the array growing
    longmatch_test.go:754: the store's frame array has capacity 1842704 for a high-water mark of 16
        retained frames
--- FAIL: TestLongMatchMemoryAndCostAreBounded (1.43s)
--- FAIL: TestFrameStoreRecyclesWithoutGrowing (104.60s)
```

Three different assertions on the same defect. **And the worst case is measured:**
with this defect the recycling test took **104.6 s**, because the copy is
O(n) per recycle. It is accepted: 105 s and 58 MB in a test that FAILS, against the
20 GB that §25 had to be stopped by hand for.

**3. A cost that grows with the match** (a store walk per visit, the classic "something
gets older"):

```
    longmatch_test.go:630: the average tick costs 118187 ns in the last quarter against 23220 ns in
        the first; the budget is 31186 ns (3x plus a 5000 ns floor).
--- FAIL: TestLongMatchMemoryAndCostAreBounded (12.60s)
    longmatch_test.go:857: a visit to an observer costs 468853 ns in the second half of the match
        against 174296 ns in the first; the budget is 228585 ns (3x plus a 5000 ns floor).
--- FAIL: TestObserverServeCostDoesNotGrowWithTheMatchLength (4.09s)
```

Ratio 5.09. And this is why the clock caps are 3x and not 1.3x: with the whole
short loop running in parallel, the **same** build measured 10083 ns against
15681 ns, a 1.56 that is the machine's load and not the server. What IS exact, and has no
clock in it, is the allocation count: in that same
loaded run they were 30.01 and 30.01. **The sharp per-tick assertion is the
allocation count; the clock caps catch an order-of-magnitude
degradation and do not pretend to more.**

### WHAT WAS MEASURED (suite)

With the working set of **this repo's** test binaries added up every 200 ms
(the name list from §33, so that the sampler does not pick up the `*.test.exe` of
other projects running on the machine), in `src_go`:

| | before (§33) | with the long match |
|--|--|--|
| `go test -short ./... -count=1` (short loop) | 43.2 s / 111.5 MB | **42.3 s / 120.7 MB** |
| `go test ./... -count=1 -timeout 25m` (full) | 767.7 s / 114.3 MB | **768.2 s / 118.8 MB** |
| `go test -race server+session+sim -count=1 -timeout 25m` | 780.0 s / 322.3 MB | **779.1 s / 287.3 MB** |

**The time has not moved from where it was**: the full suite stays at its measured
floor of ~767 s (768.2 s, +0.5) and the short loop at 42 s. RAM goes up 9 MB in the
short loop and 4 MB in the full one, and in `-race` it **drops** 35 MB (this time
§33's measurement was dragging in another project's processes). What does change is
`internal/session` in isolation: **from 1.8 s to 15.3 s and from 75.7 MB to 87.9 MB under
`-race`**, which is what these tests really cost. They do not move the critical path
because the package that marks the clock is `internal/sim` (762.8 s).

### What is left

- Nothing from this round. The question has been answered with numbers and the answer is
  that the server is bounded over two hours of match. **If there is ever a
  match longer than two hours**, the horizon is a constant
  (`longMatchHorizonHours`) and the rest of the caps are in bytes per tick and in
  objects per tick, that is, **they do not have to be recalibrated when it grows**: a leak
  of N bytes per tick looks the same at two hours as at ten.
- The sync timeout (`peer+96`, 60 s) still has no coverage over a long
  horizon, because it is the only clock in the session that is not the tick. It
  can be covered with an injected clock, which is a production change that does not
  pay for itself.
- The constant of 4 checksum records per tick is a fixture decision
  informed by the binary (`E2590:374-381`'s 10 rounds). If the game ever sends more than 10
  records per tick and per peer, the `session+0x8c8` queues grow **in the original too**,
  and that file is not the place
  to fix it.

---

## 35. Go 1.27's goroutine leak profile: how to read it, where it looks, and the five paths that have goroutines (2026-10-03)

### WHY THIS ROUND

§34 measured the heap and the allocations per tick of two hours of match and
found 21 KiB of growth and **30.01 flat objects per tick**. And it was
right: that is the shape of a leak that reserves per tick. What it **cannot**
see is the class of defect this question is about: **a goroutine per connection
blocked on a channel moves neither the heap nor the allocation-per-tick
counter**, because its stack is small and there is nothing per tick that grows.

And this server is not a program without concurrency: it starts real listeners,
accepts real connections, and through the reconnection work (§20, §30)
**swaps a socket and closes a connection**. A `go readLoop`/`go writeLoop` per
connection (`client.go:87-88`) that never lets go is exactly the shape of bug
that §34 could not find.

### WHICH API IS USED, AND HOW IT IS CHECKED THAT IT EXISTS

Go 1.27 (release notes in `go1.27.md`, section "Goroutine leak profile")
brings the new `goroutineleak` profile. **No API was assumed**: it was
read with `go doc`, and the three checks are in the code and run in
the suite:

| where | what it says | how it is checked |
|--|--|--|
| `go doc runtime/pprof.Profile` | the predefined profiles are `goroutine`, **`goroutineleak`**, `allocs`, `heap`, `threadcreate`, `block`, `mutex` | `pprof.Lookup("goroutineleak")` and `pprof.Profiles()` |
| `go doc runtime/pprof.Profile.Count` | `func (p *Profile) Count() int` | it is used, and cross-checked against the text |
| the release notes | it is also in `net/http/pprof` as `/debug/pprof/goroutineleak` | **not used**: reading the profile in the binary is the same thing without standing up an HTTP server inside a test |

That is: **the stdlib does expose the facility programmatically**, so the
assertion is about `runtime/pprof`, not about hand-parsed text nor about the
endpoint. If the endpoint were needed anyway, `net/http/pprof.Handler("goroutineleak")`
is the same profile.

### THE TWO TRAPS, MEASURED

**1. `Profile.Count()` runs one refresh behind.** A forced `runtime.GC()` does NOT
refresh `Count()`; what refreshes it is **writing the profile**. Measured with 1, 2
and 3 planted leaks: `Count()` read right after the `GC()` gave **0, 1 and 2**
while the header said **1, 2 and 3**. That is why the code writes first
(`WriteTo(&buf, 1)`), takes the number from the header
`goroutineleak profile: total N`, and **then** reads `Count()` and **requires that the
two match**. That way the number comes from the reliable place and the cross-check turns
"the profile stopped counting" into a red test instead of a green one.

**2. The collection has to be forced.** The detection is done by the collector, so the
profile describes what the last collection **saw**. With zero forced `GC()`
a parked goroutine can be invisible. `leakProfileForceGC = 1`.

### THE TRAP THAT GOVERNS THE DESIGN: REACHABILITY

The release note says the runtime "may fail to identify leaks caused by blocking on
concurrency primitives **reachable through global variables or the local
variables of runnable goroutines**". Everything this server blocks hangs off the
`Server` object, that is, **a goroutine that leaked while the test held the
`*Server` would NOT be reported**.

Measured in `TestLeakDetectorIsArmed`'s child: a `go` parked on a channel
**reachable** from the test's frame **is not reported** (`total 1`, the other
leak's); as soon as that reference is dropped, the same goroutine **appears**
(`total 2`). That is exactly the documented caveat, checked.

The file's shape comes out of that: **each cycle runs inside its own closure**
(`runServerCycle`), which builds, starts, drives and **stops** the server and does not
return until no reference to it is left. On return, the server's object graph
is unreachable and a goroutine still parked on one of its
channels **is a leak by the runtime's own definition**.

That is why the primary assertion is the **goroutine count** with a cap, a settle
window and a retry (reachability does not matter to it), and the profile is the second
property, the one the collector sees.

### THE FIVE PATHS, WITH THEIR BASELINE AND THEIR CAP

`src_go/internal/server/goroutineleak_test.go`. All of them with a warm-up cycle
before the baseline, because the runtime starts goroutines lazily (the first
`dial`'s resolver, gws's own).

| test | what it brings up and checks | cycles |
|--|--|--|
| `TestServerStartStopReleasesEveryGoroutine` | a whole server (both transports) started and stopped; a real client does the `0x14` echo in each cycle | 3 |
| `TestClientConnectDisconnectReleasesEveryGoroutine` | 4 clients per cycle: `LobbyConnect`, a replay `0xFE`, and **a `0x0A` with dumped subs** (which only comes out of the tick's flush), and they leave | 3 |
| `TestReconnectSwapReleasesEveryGoroutine` | **the reconnection path**: 4 socket swaps with the same secret, alternating state 1 and state 2. Each replaced connection **has to see EOF, and the test does not close it** | 3 |
| `TestObserverJoinLeaveReleasesEveryGoroutine` | an observer over WebSocket: `0x1C` -> `0x19`=0, `0x1A` -> `0x1E`, **a `0x0A` served from the frame store** (that is, the tick's `UpdateObservers` running), and the **three** ways out of the list: `0x1B`, abrupt disconnect, and the tick's kick | 2 |
| `TestSyncTimeoutReapReleasesEveryGoroutine` | the **only clock in the session that is not the tick**: a peer announces 60 ms with `-allowTimeoutOverride`, starts the match with one page and then goes quiet; the tick reaps it, the socket sees EOF **by the server's decision**, and the session goes with it | 3 |

None of them goes past "we saw a few messages": each one additionally asserts
the state that must really have happened.

- **The reconnection** checks from the client's side what the server did:
  the replaced connection sees EOF and the session continues with **one** peer. Without those two
  things the test would pass even if the reconnect branch never ran.
- **The observer** checks all three ways out. The tick's kick
  (`Session.Tick`, `sub_1400E2590:99-218`) is conditioned on the session not
  being recorded nor reconnectable, and `Recording` is turned on as soon as the
  frame store is created: that is why a **second session with no pages**
  is needed (`leakKickSessionID`), which is the price of exercising that arm. And what
  ours does where the original **does not** is written down: `RemoveClient` removes
  the entry, `handleStopObserve` removes it, and the tick empties the list. The entry left
  in `srv.observerSessions` goes with the client, so it waits for it to reach zero instead of
  reading it once (the server finds out about the closed socket from its own read
  loop).
- **The timeout** uses the session's wall clock, so the horizon cannot be
  compressed as in §34: the deadline the client announces is shortened, which is
  what `-allowTimeoutOverride` exists for.

**Where the goroutines really are**, a `grep` of `go ` over the sources without
tests: the main loop and the munin one (`server.go:298`, `:301`), the
`http.Server` and a gws `ReadLoop` per WebSocket connection (`ws_listener.go:162`,
`:176`), the accept loop and the `TCPListener`'s `watchCtx` (`internal/net/tcp.go:72-73`),
the accept handler, and each bsPort connection's read/write loop
(`client.go:87-88`).

**`internal/session` starts not a single goroutine**: `session.go`, `manager.go`,
`ack.go` and `replay.go` have not one `go` statement. That is why its timeout path has no goroutine
that could leak, and why it is covered **through a real server**, which is where the
connection is closed.

It is not the same as `TestMainLoopAdaptiveTickerClientChurn`: that one watches the **registry**
of clients between WebSocket connections, and says not one word about
goroutines.

### THE FALSIFICATION, IN TWO PARTS

**1. In the suite**, `TestLeakDetectorIsArmed`: it plants a goroutine parked on
a channel that exists only as its own local and that nobody will point at, and requires that
**the same measurement** the other five use sees it. It lives in a **child
process**: `assertNoGoroutineLeaks` hangs off `startServer` and
`startBlackBoxServer`, that is, a leak planted in this process would turn the rest of the
package red. The child does the three measurements and prints them, so the
parent's log carries the evidence:

```
=== RUN   TestLeakDetectorIsArmed
    goroutineleak_test.go:708: child process, planted leak, out:
        === RUN   TestLeakDetectorIsArmed
            runtime go1.27.0 lists the "goroutineleak" profile among 7 predefined ones
            a goroutine is parked and no collection was forced: total 1
            after 1 forced GC(s): total 1, and the stacks name leakForever:
                goroutineleak profile: total 1
                1 @ 0x7ff6fd2f7bea 0x7ff6fd2847ee 0x7ff6fd284332 0x7ff6fd621225 0x7ff6fd2fec81
                #	0x7ff6fd621224	...internal/server.leakForever+0x44	...goroutineleak_test.go:814
            a goroutine parked on a channel this frame references: total 1, not reported
            after dropping that reference: total 2, the parked goroutine is now reported
        --- PASS: TestLeakDetectorIsArmed (0.21s)
--- PASS: TestLeakDetectorIsArmed (0.84s)
```

The three lines are the three measurements: that the facility exists, that it sees a leak
that cannot wake up, and that **its blindness is the documented one**.

**2. Against production**, a real per-connection leak in `client.go:Start` (a
`go` parked on a channel nobody points at), the five tests, and the result. All
five fail and the profile **names the exact line**:

```
--- FAIL: TestServerStartStopReleasesEveryGoroutine (4.09s)
    goroutineleak_test.go:460: a started/stopped server left 3 goroutine(s) above the baseline of 3
        (now 6, budget 2).
    goroutineleak_test.go:460: the goroutineleak profile reports 4 leaked goroutine(s):
        goroutineleak profile: total 4
        #	0x7ff7e2db6864	...internal/server.(*Client).Start.func1+0x24
        	C:/Users/luska/BattleServer.exe_ai_dataset/src_go/internal/server/client.go:93

--- FAIL: TestClientConnectDisconnectReleasesEveryGoroutine (4.41s)
    goroutineleak_test.go:499: settled baseline after one warm-up cycle: 10 goroutines, budget 2 above it
    goroutineleak_test.go:499: a client connect/disconnect cycle left 12 goroutine(s) above the
        baseline of 10 (now 22, budget 2).
    goroutineleak_test.go:499: the goroutineleak profile reports 20 leaked goroutine(s):
        goroutineleak profile: total 20
        #	0x7ff7e2db6864	...internal/server.(*Client).Start.func1+0x24
        	.../internal/server/client.go:93

--- FAIL: TestReconnectSwapReleasesEveryGoroutine (4.11s)
    goroutineleak_test.go:553: a reconnect swap left 15 goroutine(s) above the baseline of 27
        (now 42, budget 2).
    goroutineleak_test.go:553: the goroutineleak profile reports 40 leaked goroutine(s):
        goroutineleak profile: total 40
        #	0x7ff7e2db6864	...internal/server.(*Client).Start.func1+0x24
        	.../internal/server/client.go:93

--- FAIL: TestObserverJoinLeaveReleasesEveryGoroutine (4.65s)
    goroutineleak_test.go:629: an observer join/leave cycle left 4 goroutine(s) above the baseline
    of 44 (now 48, budget 2).

--- FAIL: TestSyncTimeoutReapReleasesEveryGoroutine (4.54s)
    goroutineleak_test.go:660: a sync-timeout reap left 3 goroutine(s) above the baseline of 3
        (now 6, budget 2).
    goroutineleak_test.go:660: the goroutineleak profile reports 4 leaked goroutine(s):
        goroutineleak profile: total 4
        #	0x7ff748ff6864	...internal/server.(*Client).Start.func1+0x24
        	.../internal/server/client.go:93
```

Note that the baseline also rises (10, 27, 44): the warm-up cycle's leak counts as
baseline. That is why **the budget has to be a cap per cycle, not a comparison with the
baseline from before it started**. Leak planted, removed: `client.go` goes back to HEAD, `git status` clean of production.

### THE RESULT: THERE WAS NO LEAK

**Not one line of production was touched.** All five paths let go of everything they
have, and the measured baseline is **2 goroutines** (the test binary) with a
cap of **2** above it, in all five: zero growth margin and zero
growth, with 3 cycles of 3 and 4 connections, 4 socket swaps, 3
observers and one reap per timeout cycle.

What **did** come out, and is not a leak:

- `srv.observerSessions` is **not** emptied with the tick's kick: the kick empties
  the session's list (`s.observers = nil`) and the routing map's entry goes
  when the client disconnects. It is written in the test and it waits for zero.
- The `Dropping stalled observer` that appears in the observer path's
  log is the tick's failed-send counter (**smaller and a different thing**, not a
  leak): the observer left abruptly and the store kept sending `0x0A`s to a dead
  socket.

### WHAT WAS MEASURED (suite)

With the working set of **this repo's** test binaries added up every 200 ms
(the name per package, as in §34, so that the sampler does not pick up `*.test.exe` from
other projects on the machine), in `src_go`:

| | before (§34) | with the leak profile |
|--|--|--|
| `go test -short ./... -count=1` (short loop) | 43.2 s / 118.2 MB (trimmed) | **43.9 s / 121.3 MB** (and 43.2 s / 119.6 MB in another run) |
| `go test ./... -count=1 -timeout 25m` (full) | 768.2 s / 118.8 MB | **768.6 s / 115.9 MB** |
| `go test -race server+session+sim -count=1 -timeout 25m` | 779.1 s / 287.3 MB | **780.2 s / 295.5 MB** |
| `internal/server` alone, without `-race` | 15.5 s (in the short loop) | **37.8 s** |
| `internal/server` alone, with `-race` | — | **38.8 s** |

**The time has not moved from where it was.** The full suite stays at its measured
floor of ~767 s (768.6 s, +0.4) and the short loop at 43-44 s. What went up is
`internal/server` in isolation, from 15.5 s to 37.8 s, and **it does not move the critical path**:
the package that marks the clock is `internal/sim` (762.6 s in the full one, 37.0 s in the
short one) and `internal/server` is 18.2 s in the short one, well below. With `-race`
too: 38.8 s against 769.5 s of `internal/sim`.

### What is left

- **`internal/session` has no goroutines to cover.** If any are ever added to it,
  this file is where they fit: the pattern is `runServerCycle` plus
  `assertCycleBudget`, and the `drive` can be as synthetic as needed
  because the assertion does not depend on how the server is driven.
- **The reachability blindness cannot be closed from here.** A leak whose
  primitive stays reachable from a runnable goroutine is not seen by the runtime either.
  That is why the primary assertion is the count with a cap, and why the
  `drive` runs inside a closure: if someone writes a test that keeps the
  `*Server` alive until after asking, the profile will say nothing and the count
  is all that is left.
- **`leakGoroutineSlack = 2` is measured, not chosen by taste**, but it is the most
  fragile figure of the file: it is a count of the whole process, not of this server.
  With the rest of the repo's packages running in parallel (which is how the suite
  runs) the real margin is that of the run, and that is why the settle window
  with a retry is there. If it ever starts failing intermittently, the
  place to look is `settleGoroutineCount`, not the cap.
- **The `0x1F` and the pending list were not touched** (§32, §33) nor either of the
  two closed questions. This round found no divergence with the
  original: it found nothing.

---

## 36. Closing a kicked observer, §32.5's eleven fields, and the documentation audit (2026-10-03)

### WHY THIS ROUND

The 32.5 left eleven fields unexamined and two differences untouched. One of the two
differences was a **self-deadlock**, not a risk: that has priority over everything
else because it hangs the whole server, not because it is the most interesting one. And the eleven
fields were the last work that rule 1 left open without a new capture: all of them
are resolved by reading the binary.

What was **not** done here, and why: no new capture (it is not needed, and
adding one is what keeps the suite in its budget), neither of the two
closed questions, and nothing about the `0x1F` or the pending list, which belong to the 32
and the 33.

### 36.1 CLOSING A KICKED OBSERVER: WHAT THE ORIGINAL DOES

The question was which of the three functions that the 32.5 named is the kick path.
The answer is that **the kick is in the tick**, and that the other two are not
on that path:

| | what it is | what it does on the kick |
|--|--|--|
| `sub_1400E2590:99-218` | the kick arm, inside `Session::Tick` | **is** the kick. For each entry in the observer list it builds the `StartObserveResponse` (type 25 = `0x19`, one-byte body, `v81 = 4`), sends it with `sub_1400D49A0` (`:183`) and schedules the close with `sub_1400D4860(*v6, "Observer session not available", 0)` (**`:184`**) |
| `sub_1400E4EC0` | `UpdateObservers`, the **serving** arm (`:96-97`) | it is not on the kick path. Its only close is `sub_1400D48D0(v10, "Observing peer, but can't update")` (`:194`), which is also a schedule and not a close, and it does not touch the list |
| `sub_1400EC5E0` | `LeaveSimulation` (0x1F) whole, three lines, and its only caller is the dispatcher | it is not on the kick path |

The three close details that matter:

- **The deadline is 0**, that is `conn+112 = now` and `conn+152 = 1` (`sub_1400D4860:16-20`).
  `sub_1400D4930` is the "the close is due" predicate
  (`conn+152 != 0 && now >= conn+112`) and the socket loop asks it at
  `sub_1400D8E10:164`, that is, the socket falls in the **same pass**. It is not a
  deferred close: it is an immediate close written as a schedule, because that is how it is
  in the binary.
- **The list is not touched.** The loop advances `v6 += 2` (`:210`) and its end condition is
  the same pointer it started with, that is, the entry is still there when the
  connection drops. The disconnect removes it (`sub_1400E52D0:227-242`).
- **The reason** goes to the string at `conn+120` (`sub_140081470((void **)(a1 + 120), ...)`),
  which does not reach the wire: our `WriteClose(1000, nil)` does not carry a reason either.

Us, before this round: we sent the `0x19` = 4 and **emptied the list, but
did not close the connection**. The 32.5 wrote that down as the reason not to
touch it, and the reason was real but in the wrong place: the problem was not
closing, it was **closing from inside the lock**.

#### WHY CLOSING FROM THERE IS A SELF-DEADLOCK, NOT A RISK

`Session.Tick` enters `s.mu`. `Close()` is not a leaf:

```
WsClient.Close()      internal/server/ws_client.go:117  -> c.host.RemoveClient(c)   (synchronous)
Client.Close()        internal/server/client.go:352     -> c.host.RemoveClient(c)   (synchronous)
Server.RemoveClient   internal/server/server.go:458     -> sess.RemoveObserver(c)
Session.RemoveObserver                                       -> s.mu.Lock()
```

A `sync.RWMutex` is not reentrant. And the goroutine that runs `Tick()` is the **server's main
loop** (`internal/server/server.go:376`), that is, the first
observer that closes from there does not just delay its own send: **it delays the pages of
all sessions**, and everything else that loop dispatches.

The shape of the fix is the one `closeExpiredPeers` already did right: the tick's body
runs with the lock held and **returns** the connections, and closes them afterwards. The
new body is called `tickLocked` and the close stays in `Tick`. Nothing else of the tick was
touched.

```
85dd779  the kicked observer is closed, and it is closed after releasing the lock
a6eb4d6  the sync-message counter reaches the reconnect's echo
```

#### THE TESTS, AND WHY EACH ONE FAILS IN ONE WAY

| test | what it pins | how it fails with the defect |
|--|--|--|
| `TestObserverKickClosesTheConnection` (session) | the `E2590:184` with a connection whose `Close` **cannot** re-enter | as an **assertion**, never as a hang |
| `TestObserverKickDoesNotReenterTheSessionLock` (session) | the whole cycle with a connection that closes like the real ones (`Close` -> `RemoveObserver`), requiring `Tick` to return within `observerKickReturnBound` | after **2 s**, with the number and the reason in the message |
| `TestObserverKickTickReturnsWithoutObservers` (session) | that a tick with nothing to kick closes no connection | if the close were moved out of the kick arm |
| `TestObserverJoinLeaveReleasesEveryGoroutine` (server, step 3) | that the **connection drops**: the client's read loop has to finish and nothing in the test closes it | `the server did not close this connection within 3s` |

The first way out of the list (the `0x1B`) and the second (the abrupt disconnect) were
already covered by that leak test; what is new is that the **kick** also has to
drop the connection, and that the file does not limit itself to counting goroutines.

Both falsifications, with the defect in and out:

```
=== CLOSING FROM INSIDE THE LOCK ===
=== RUN   TestObserverKickClosesTheConnection
--- PASS: TestObserverKickClosesTheConnection (0.06s)
=== RUN   TestObserverKickDoesNotReenterTheSessionLock
    observerskick_test.go:189: Session.Tick did not return within 2s with 2 observers in the list. The kick closes
    each observer's connection from inside the section that holds s.mu (internal/session/session.go), and Close()
    re-enters the session through RemoveObserver, so the tick is waiting on a lock it already owns. The goroutine
    that runs it is the server's main loop, so this is a whole-server hang, not one observer's send. The connections
    close after the lock is released, the way closeExpiredPeers does it.
--- FAIL: TestObserverKickDoesNotReenterTheSessionLock (2.00s)
=== RUN   TestObserverKickTickReturnsWithoutObservers
--- PASS: TestObserverKickTickReturnsWithoutObservers (0.00s)
FAIL	github.com/luskaner/ageLANServer/battleserver/internal/session	2.529s

=== WITHOUT CLOSING THE CONNECTION (what it was before this round) ===
=== RUN   TestObserverKickClosesTheConnection
    observerskick_test.go:139: the kicked observer's connection is still open: sub_1400E2590:184 closes it, and a client
    left on a session with no data is the state that arm exists to end
--- FAIL: TestObserverKickClosesTheConnection (0.07s)
=== RUN   TestObserverKickDoesNotReenterTheSessionLock
    observerskick_test.go:210: observer 0 had Close called 0 times, want exactly 1: sub_1400E2590:184 schedules the close
    once per observer
    observerskick_test.go:210: observer 1 had Close called 0 times, want exactly 1: sub_1400E2590:184 schedules the close
    once per observer
--- FAIL: TestObserverKickDoesNotReenterTheSessionLock (0.00s)

=== ON THE SERVER, WITHOUT CLOSING ===
    goroutineleak_test.go:642: the kicked observer: the server did not close this connection within 3s
    --- FAIL: TestObserverJoinLeaveReleasesEveryGoroutine (9.72s)
```

That the test does **not** hang is part of the design, not a coincidence: `Tick` runs in
its own goroutine and the wait has a cap with a constant (`observerKickReturnBound`), and
when the cap runs out the test **does not touch the session again** (the tick holds it
and is not going to release it, so reading it would hang the test instead of failing it).

### 36.2 `conn+104`: THE ONLY ONE OF THE ELEVEN THAT WAS MODELLED

The eleven fields' verdicts are in **32.8**. This is the only one that was touched, and
why: it is the only one with an identified writer **and** reader, both interpretable
without assuming anything.

`conn+104` is the number of sync messages the session has taken from that
peer (`sub_1400E2E00:202`), and its only reader in the whole binary is the reconnection:
`sub_1400EB5F0:341` puts that number into the `dataReceivedCount` of the LobbyConnect it
returns, over a **clone** of the client's message (`EB5F0:306`, the `vmethod_4` whose
error line is `"failed to get LobbyConnectMessage from packet"`). Before
we returned the client's body as it was; now the counter exists and the field is
rewritten in state 2.

Both falsifications:

```
=== THE ECHO WITHOUT REWRITING ===
    syncmessages_test.go:123: the echoed connect carries dataReceivedCount = 0, want 3 (sub_1400EB5F0:341 copies the
    peer's offset 104 into that field)
--- FAIL: TestSyncMessageCountReachesTheReconnectEcho (0.08s)

=== THE COUNTER WITHOUT INCREMENTING ===
    syncmessages_test.go:109: accepted sync messages = 0, want 3: sub_1400E2E00:202 counts one per 0xFE upload
    --- FAIL: TestSyncMessageCountReachesTheReconnectEcho (0.07s)
    syncmessages_test.go:192: peer 2 counted 0 accepted sync messages, want 1: its own upload
    --- FAIL: TestSyncMessageCountIgnoresEverythingElse (0.00s)
```

**The limitation, written in the code and not here:** no committed capture contains
a reconnection, so none of this can be checked against the original's bytes. And the
original's serialiser writes that field for **all** states except
2, with `-1` inside (`LobbyConnectMessage__vmethod_20:59`), that is, its
state-0 and state-3+ echoes are four bytes longer than the requests our
parser accepts. Only the shape of state 2 is reproduced, which is the one that round-trips;
the other two still return the request as it was.

The other half of the original's count guard, `session+2141 == 0 || peer+90 != 0`
(`E2E00:35`), is **not** modelled: it is the divergence the 32.4 already records, and the test
does not claim it.

### 36.3 THE TWO STATUS BYTES OF `E2590:483`: ONE IS DEAD

`sub_1400E2590` returns an `__int16` in `*a2` (`:483`) and `sub_1400E9E90:69-71` reads it,
which does both things with the same value:

```
sub_1400E9E90:70    *((_DWORD *)v2 + 10) += HIBYTE(v41);          // accumulates the HIGH byte
sub_1400E9E90:71    if ( (_BYTE)v41 != 0 && session+2328 > 0 && session+2332 == 1 )
```

The low byte is what the 32.5 said, with the exact shape of `:270-283`: "the store
has frames", that is `v27 = the store is not empty ? first element of the store+80 container : -1`,
and `LOBYTE(v85) = that element > 0`.

**The high byte can never be set.** Line 273 loads
`v28 = *((_DWORD *)a1 + 536)` (that is `session+2144`, the
desync counter) and lines 285-286 are:

```
if ( v28 == 0 )
  HIBYTE(v85) = *((_DWORD *)a1 + 536) > 0;
```

If `v28` is zero, the comparison itself is false; and if it is not, the branch is not taken. The
only `++` of that field is at line 380, **after**, so there is not even a race
that would fix it. That is, the original munin's "sessions with desync" counter **never moves**,
and the high byte of `E2590` is 0 with or without a desync.

We return nothing (`Session.Tick` has no return value), so the
divergence is of a field that does not travel: **informative**, and now with the certainty that
the original does not use it for anything. It is left as it was.

### 36.4 WHAT THE DOCUMENTATION AUDIT FOUND

Fifteen markdowns in the parent's repo root, **none of them touched**. The verdict on each
one, and what is **recommended** (nothing is deleted: that decision is the user's):

| file | what it is | verdict | recommendation |
|--|--|--|--|
| `AGENTS.md` | the rules and the index | **current**, it is the front door | keep it; link it to the index below |
| `SESSION-JOURNAL.md` | the journal | **current**; this section is its continuation | keep it |
| `CAPTURES.md` | how to read a capture | **current**, and the only one that describes the current layout | keep it |
| `go1.27.md` | a copy of the Go 1.27 release notes | **current as a reference** (§35 cites it for the leak profile) | keep it; mark it as a third-party copy, not as a project document |
| `bsPort.md` | the original's game port | **still true** as a description of the original (flags, RLink framing, 27012) | keep it; link it from `AGENTS.md` |
| `webSocketPort.md` | the original's WebSocket port | **still true**, and with the expensive fact already corrected (RFC 6455, not Hixie-76, verified in captures) | keep it; link it |
| `MESSAGES.md` | the table of message types | **current**; this journal cites it as the source of the types | keep it; link it |
| `oob.md` | the out-of-band port | **still true** as a description of the original | keep it |
| `LOBBY-PROTOCOL.md` | the heartbeat with the lobby | **still true**; `-lobbyPort` exists and the harness turns it off (§23) | keep it |
| `replays.md` | the replay upload over HTTP POST | **still true** as a description of the original, and it is the only place the POST is | keep it |
| `uploadVia-lobby.md`, `uploadVia-lobby-connection.md`, `uploadVia-lobby-state.md` | the three steps of `-uploadVia=lobby` | **current**: that is the only method the server accepts (`config.go:117-125`) | keep it; **link them to each other**, today they are referenced with loose relative links |
| `uploadVia-mysql.md`, `uploadVia-odbc.md` | the other two `-uploadVia` methods | **SUPERSEDED**: the flags exist but the server **rejects** `mysql` and `odbc` with an error and `os.Exit(1)` (`config.go:119-121`). They document a feature that does not exist | **delete them**, or mark "NOT IMPLEMENTED" on the first line. **Nothing is deleted without the user saying so** |
| `WEBSOCKET-BSPORT-COMPARISON.md` | its conclusion is "the two ports are equivalent" | **SUPERSEDED in its conclusion**: the 27 proved that on bsPort **you cannot observe** and that the TLS is the port's, and the document does not mention the observer in any of its 394 lines | mark the conclusion as superseded with a link to the 27; the rest (packaging, handshake, socket managers) still holds |
| `PROXY-CAPTURE-VALIDATION.md` | how the reimpl is validated with proxy captures | **SUPERSEDED from the root**: it describes a format that no longer exists (`main_log_*.csv` with `data_b64`, T1/T2 with ports fixed per run). The current proxy writes `main.log`/`secondary.log` plus `metadata.json` and `*.meta.json`, and the validation is the `internal/sim` harness | **delete it**, because besides it is the link that `webSocketPort.md:55` and others give as "evidence", and that link leads to a document describing captures that were mostly deleted. If it is kept, it should only be history, with a line at the top saying so |

The three reasons why **nothing** was deleted: rule 5 of `AGENTS.md`
forbids touching others' changes, and a deleted file cannot be discussed; besides these
conclusions ("SUPERSEDED") mean nothing to whoever reads them without context, so what
is needed is a line at the top of each file and not its absence; and the
decision is the user's.

### 36.5 WHAT WAS MEASURED

Full verification of both commits, in `src_go`, with the working set of this repo's
test binaries added up every 200 ms (the `.test.exe` under the temporary
`go-build` directory, which is where `go test` puts them; no other project on the machine counts):

| | before (35) | this round |
|--|--|--|
| `go test -short ./... -count=1` (short loop) | 43.9 s / 121.3 MB | **44.8 s / 122.9 MB** |
| `go test ./... -count=1 -timeout 25m` (full) | 768.6 s / 115.9 MB | **767.9 s / 124.9 MB** (and 769.6 s on a first run) |
| `go test -race server+session+sim -count=1 -timeout 25m` | 780.2 s / 295.5 MB | **796.9 s / 296.2 MB** (`sim` 772.8 + `server` 38.9 + `session` 17.1) |
| `internal/server` alone | 37.8 s | 37.6 s |
| `internal/session` alone | — | 2.9 s (17.1 s with `-race`) |
| `go run ./cmd/gensim -check internal/sim/testdata` | ok | **ok** (0) |

**The time has not moved where it matters.** The short loop is +0.9 s and the full
suite is within the noise (one run -0.7 s and another +1.0 s), and the package that
marks the clock is still `internal/sim` at 762 s. The new tests are
microseconds, synthesise nothing and add no memory: the +9 MB of RAM in the full
suite and the +1.6 MB in the short loop are the two packages running at once
(`sim.test` and `server.test`), not the new tests.

The only number that really moves is the race run: **+16.7 s**, and almost
all of those 16.7 s are `internal/session` under `-race` (17.1 s), which is the package of
the long-match test and which before was not measured separately. It runs **in parallel**
with `sim`, so it does not move the critical path: `sim` is at 772.8 s against the 769.5 s
of the 35, and `server` at 38.9 against 38.8.

### 36.6 WHAT WAS MEASURED, WITH NUMBERS

The figures of 36.5. The full suite's first run came out with the sampler
not filtering properly (0 MB, because it looked for `*.test` instead of `*.test.exe`) and that is why it
was repeated: 769.6 s the first time and 767.9 s with the good filter. Both are green and the
number reported is the second. Noted because a sampler failure is
exactly the class of thing that makes a "0 MB" look like an improvement.

---

## 37. The other cost axis: the CLIENTS (2026-10-03)

### WHY THIS ROUND

The 34 measured the cost of a **long** match: two hours of simulated clock, 57600
ticks, 30.01 flat allocations per tick and 21 KiB of heap growth. That is the
**TIME** axis, with a fixed peer count. The other axis, **how many clients
cost**, had never been asked of anything, and the binary says it grows in two places,
linearly, and that **neither of the two has a global cap**:

1. the frame store is `ceil(peers * f)` blocks of `-pageSize`, with the floor at 8;
2. each peer's upload queue is capped at `0x12C` entries or `0x80000` bytes.

A third thing was not measured either: what a **live connection** costs. The 35 measured the
*residue* of connections (0 KiB after 1000 observer cycles), and the residue is not
the cost.

The question answered here is not "how much do 32 clients use" but **"how much does
ONE cost"**, because the first can only be reported and the second can be **asserted**:
both formulas come from the binary, with addresses, and a formula can be checked.

### 37.1 THE TWO FORMULAS, FROM THE BINARY

Read out of `src/`, not from a summary:

```
sub_1400E2590:221   v17 = (a1[1] - *a1) >> 3			 the length of the peer slot vector
sub_1400E2590:235   if (v17 < 8) v17 = 8			 the floor
sub_1400E2590:241   v21 = ceilf(v20 * *(float *)(a1 + 2168))	f, as a float
sub_1400E2590:255   sub_1400D4E10(store, v23, a1[269], a1[270])
sub_1400D4E10:18    a1[1] = a3	a1[2] = a2	 blockSize = session+2152, maxBlocks = v23
main_1400d1310:346  v43 = 1.0		the default of f
main_1400d1310:347  v93 = 0x10000		the default of -pageSize
main_1400d1310:617  v43 = (float)atoi(v49) * 0.125	-maxPagesToStorePerSession / 8
main_1400d1310:818  sub_1400D2F40(..., LODWORD(v43), v93, ...)	the bits of f go down to the store

sub_1400E2E00:53   (v7 - *v6) / 24 >= 0x12C				300 chunks of 24 bytes
sub_1400E2E00:54   *(v6+24) + (a3+56) - (a3+48) >= 0x80000		512 KiB
```

`session+2168` holds the **float bits** of `v43`: `E1DE0:42` stores it as
`_DWORD` and `E2590:241` reads it as `*(float *)(a1 + 2168)`. With the defaults, that is
`f = 1.0` and `-pageSize = 65536`:

| peers | blocks | window |
|--|--|--|
| 1, 2, 4, 8 | **8** (the floor) | 512 KiB |
| 16 | 16 | 1 MiB |
| 32 | 32 | 2 MiB |

That is, **64 KiB per peer above the floor of 8, and nothing below it**. And `f` comes
from the configuration, so the ceiling is the operator's, not the code's.

### 37.2 THE MEASURED TABLE

`src_go/internal/session/peersweep_test.go`. The fixture does not sleep: it drives
`QueueGameData` + `QueueChecksumEntries` + `PopAckTick` + `Tick` as the server's main loop
(`server.go:374-376`), and the store is measured at the **peak** of its sawtooth: the
store grows until the last block cannot hold one more frame and then loses a whole
block (`Replay.AddFrame` calls `dropOldestBlockLocked` exactly there), so the
peak of one tooth is the whole window and any other moment is a fraction that
depends on where you read. With the peak taken and the **minimum of two passes**, and with the live
heap **after two forced GCs with the fixture alive** (`runtime.KeepAlive` behind the
read: without it, the heap being measured is the test binary's and not the
session's).

```
peers blocks  window  store KiB charged KiB  frames  heap KiB  allocs/t  fill  ticks
    1      8    512K        497.8        511.7    2096      805       17.00  1835  2603
    2      8    512K        493.6        510.0    1296      795       30.00  1135  1903
    4      8    512K        538.1        509.6     736      819       49.00   645  1413
    8      8    512K        508.1        505.7     392      812       86.00   344  1112
   16     16   1024K       1068.2        994.1     400     1382      165.02   376  1144
   32     32   2048K       2202.5       2028.4     416     2484      310.09   404  1172
```

With `-race` the last three columns come out practically the same (allocs/t 17.56 /
31.00 / 51.05 / 90.05 / 172.85 / 326.12, heap 810 / 795 / 819 / 813 / 1383 / 2491,
store byte for byte equal) because the store is exact bytes and the race detector does not move the heap.

**The slope, which is what is asserted** (least squares over the 8, 16
and 32-peer points; the four below are compared against the window, not against the slope):

| slope | measured | formula | ratio |
|--|--|--|--|
| store, **charged** bytes | 65144 B/peer | 65536 | **0.99x** |
| store, **actual** bytes (Go size classes) | 72337 B/peer | 65536 | **1.10x** |
| the process's live heap | 71238 B/peer | 65536 | **1.09x** |
| allocations per tick | 9.30 objects/peer (9.81 with `-race`) | - | - |

The difference between the first and second rows **is not spare memory**: it is the
rounding to size classes. An ack of 193 bytes falls in the 208 class (7.8%) and one of
4936 in the 5376 one (8.9%), and since the frame grows with the number of peers, the actual-bytes
slope comes out 10% above the charged-bytes one. The charged-bytes one is the unit of the
formula (`vmethod_20:76` subtracts `a3 + 57` per frame) and that is why it is the one
with the narrower band.

And below the floor **nothing moves**: the four points at 1, 2, 4 and 8 peers
declare the same 8 blocks and the same 524288 bytes, and their store peaks run from
505456 to 550992 bytes, a difference that is the class rounding and the smaller frame
that fits in the same window.

### 37.3 THE CAPS, AND WHERE EACH ONE COMES FROM

All in the code, next to the measured number that chose them, and checked in both
modes:

| cap | value | measured | margin | what it catches |
|--|--|--|--|--|
| the store's blocks | the `ceil` from 37.1's table | identical | exact | a store from another formula, and it is the sharpest assertion of the file |
| charged-bytes slope | 0.90x .. 1.10x | 0.99x | 10% / 11% | a peer that costs 10% more (or less) than a block |
| actual-bytes slope | 0.85x .. 1.40x | 1.10x | 1.3x | the same defect from the allocator's side |
| store peak against its window | 0.85 .. 1.40 | 0.96 .. 1.08 | 1.3x | half the window, or twice it |
| live-heap slope | 0.75x .. 1.75x | 1.09x | 1.6x | a per-peer cost **outside** the store that hides in the heap |
| absolute live heap | 32 MiB | 2.4 MiB | 13x | a per-peer leak that is already inside the baseline |
| allocations per peer and tick | 12.0 | 9.30 (9.81 `-race`) | 1.22x | **three or more** extra allocations per peer and tick (see 37.8) |
| allocations per tick, absolute | 400 | 310 (326 `-race`) | 1.23x | a leak of one object per tick, which does not fit the per-peer shape |

The "what it catches" column is what makes the margins useful: the charged-bytes
cap, with 10%, is the file's **sharp edge** (a peer ten times more expensive gives 9.94x
and blows up), and the allocation one is the **thick** one, and its sensitivity is written in the
code itself because the margin cannot hide it.

### 37.4 THE UPLOAD QUEUES: THE TWO CAPS AND THEIR SUM

`src_go/internal/session/peerqueue_test.go`. `ack_test.go` already covered the mechanics of the
cap over a one- or two-slot `AckState`; what had never been asked is **how
much a full session of peers that all reach the cap costs**, and whether the cap holds
when the peer keeps uploading.

| case | offered | left | charged | actual |
|--|--|--|--|--|
| entries cap (`0x12C`) | 400 x 64 B | **300** entries | 19500 B | 84984 B |
| ...and then 3000 more | 3000 | **300** (the same) | 19500 | 84984 |
| bytes cap (`0x80000`) | 5 x 128 KiB | **3** entries | 393219 B | 417888 B |
| ...and then 20 more | 20 | **3** (the same) | 393219 | 417888 |

`0x80000 / (1 + 131072) = 3`: the fourth chunk would land at 524292 of 524288 bytes. And the
actual 417888 is 3 objects of 139264 bytes, which is what a 131073-byte chunk costs
in Go: **the bytes cap is holding real memory, not a counter**.

With **32 peers saturated** at the entries cap:

- live heap 2982 KiB against 323 KiB of the identical control with not a single upload: **2655 KiB,
  85114 B per peer**;
- aggregate **exactly** the product: 9600 subs (32 x 300) and 624000 B (32 x 19500);
- the ceiling the two caps imply with no global one: `32 x 0x80000 = 16 MiB`;
- after the tick that drains them (a `0x0A` with the 9600 subs, 624005 bytes), the queues are at
  zero and the heap drops to 948 KiB: **what is returned is the ack the store keeps**.

The 85114 B per peer are 4.4 times the 19500 the queue charges itself, and the reason
is written in the file: each sub is a pool buffer with a **256-byte floor**
(`ack.go`, `subPool.New`) plus 24 of header, that is 280 per sub, and 300 subs are 84000.
The queue's **charged** figure is the unit of the formula; the **actual** one is what the process pays.

### 37.5 THE OBSERVER IS NOT A PEER

`StartObserve` adds to the observer vector at `session+2272`
(`sub_1400E3B50:245-254`) and **writes no peer slot at all**, and the store's block count
is a `ceil` over the length of the peers vector (`sub_1400E2590:221`). Both
are counts of different things, so an observer cannot move the store.

Measured with 64 observers over a 2-player session: the store stays at **8
blocks and 524288 bytes** with none and with 64; the 64 sit down and receive **88768
frames in 256 ticks**; the heap goes up **194 B per observer** (125 with `-race`), and the
cost **per tick** is **0.00 extra allocations** (30.00 against 30.00; 31.04 against
30.89 with `-race`). That last number is the one that squeezes hardest: the observer update
is the only thing in the session that runs per tick **and per observer**
(`sub_1400E4EC0`, from `sub_1400E2590:96`), and it reads the store with `GetFrames`, which
returns a subslice of frames the store already has. With the 8-tick interval
8 of the 64 are visited per tick, so the 1.0-object-per-tick cap for the whole
update holds with margin and **does not scale with the observers**.

### 37.6 THE COST OF A LIVE CONNECTION

`src_go/internal/server/conncost_test.go`. Over `webSocketPort`, which is the port of
all the committed captures and the expensive one (the `bsPort` client reads from its own socket;
this one goes through `net/http` and gws).

| connections | live heap | stacks | goroutines |
|--|--|--|--|
| 0 (control) | 420 KiB | 288 KiB | 6 |
| 8 | 553 KiB | 448 KiB | 22 |
| 32 | 933 KiB | 896 KiB | 70 |

**16 KiB of heap and 34-36 KiB of heap plus stacks per connection, with 2.0 goroutines per
connection** (gws's `ReadLoop`, `ws_listener.go:162`, and `net/http`'s per-connection
goroutine). It is the distinction from the 35 that has to be said: there the residue
was measured and here the cost, and a live connection **is not** free even if it leaves nothing behind.

Of the heap you can say where it comes from: 2 KiB of gws read and write buffers
(`ws_listener.go:127`, `ReadBufferSize: 1024`) and about 8 KiB of `net/http` bufio around the
socket. The **stacks are reported but have no cap below**, and the reason is
in the code: `StackInuse` counts whole 8 KiB chunks, so it moves in steps and
a chunk freed anywhere in the process **lowers** the reading. Measured in three
runs of the same build: the control read 288, 960 and 1504 KiB with the same six
goroutines, which puts the per-connection figure between 7 and 20 KiB. A cap of 128 KiB
(3.2 times the measured one) holds that spread and still blows up with a stack that
leaks.

### 37.7 WHAT DOES NOT ADD UP WITH THE FORMULA

Three things, and **none of them is a defect**. They are put here so that the next session does not
find them again as if they were new.

1. **The tick spends 9.30 objects per peer**, not one. The account, all of it on the ack path
   and everything else: an interface box per `subPool.Put` (passing a `[]byte` to `any` for a
   `sync.Pool` **reserves** 24 bytes, because the pool keeps it), four 8-byte checksum
   records (`QueueEntries` makes one per record, `sub_1400E2E00:190-199`),
   and about three from the **array** of each peer's checksum vector, which reconciliation
   empties to zero capacity every tick (it takes one record per peer and round, and a round runs
   at most ten times per tick, `sub_1400E2590:374-381`; `[1:]` in Go consumes
   capacity, whereas the original's `std::vector`'s `erase` keeps it). It is
   **short-lived garbage**, not a leak: the 34 already measured that the heap plateaus at 30 allocations
   per tick, and here it is the same with a factor of peers. And the original has **nothing** to
   fix: its per-peer `std::vector` (`session+0x8c8`) has the same one-record-per-round
   drain and therefore grows the same way if a peer uploads more than ten records per
   tick. **Production is not touched**: the allocation count does not travel on the wire, and
   changing it would be a performance change with no evidence that the original does it.

2. **One byte per sub in the byte count.** The binary loads `msg+56 - msg+48`, that is
   the payload (`sub_1400E2E00:54` and `:199`); here it charges `1 + len(payload)`, because the
   queued sub is `[slot][len u16][payload]` (`AckState.QueueDataSub`). That is 300 bytes
   per peer saturated at 512 KiB, and it is only visible in the byte cap.

3. **The queue's two caps are independent**, and it was seen by planting the defect: with
   `maxAckQueueEntries` multiplied by 100, one slot ends up with **8065 entries** instead
   of 300, and the 32 peers together with **258080 subs and 71 MB**, because **the bytes
   cap still holds**. That is, the entries cap is the one that shows in the number
   of objects and the bytes one is the one that shows in memory, and neither is the other's
   safety net. That is why both are checked separately.

### 37.8 THE FIVE FALSIFICATIONS

A test that has never been seen to fail is not a test. All five were put into the production
code, measured and removed; `git status` afterwards: clean apart from the three new
files. **Each one with what the cap detects AND what it does not**, because a margin
that is not declared is a margin that cannot be argued about.

**1. A peer that costs ten blocks** (`ensureRecordingLocked`: `len(s.Peers)*10`):

```
--- FAIL: TestPeerCountSweepMatchesTheStoreFormula (3.91s)
    peersweep_test.go:612: 32 peers: the store was built with 320 blocks, and sub_1400E2590:235-241 with f = 1.0 says ceil(max(32,8) * 1.0) = 32. Every other figure in this row is about a store that is not the one the formula describes
    peersweep_test.go:596: 32 peers: the store never had all 32 of its blocks live in 4096 ticks, so every figure for this point is the cost of a store with room to spare rather than the cost of a full one
    peersweep_test.go:691: each peer above the floor of 8 costs 651439 bytes of frame store as the store CHARGES them, 9.940x the 65536 the formula says (...); the ceiling is 72090
    peersweep_test.go:704: each peer above the floor of 8 costs 725157 bytes of frame store, 11.07x the 65536 the formula says (...); the ceiling is 91750
    peersweep_test.go:715: the live heap grows 724173 bytes per peer above the floor of 8 peers, 11.05x the 65536 the store formula accounts for and over the ceiling of 114688
```

Four different assertions on the same defect, and the last one is what acts as the net:
the heap too. And besides that, the tick cap speaks on its own: a store ten times bigger does not
fill up in 4096 ticks, and the test says so without anyone asking. With the defect in place the peak
was 22 MB: this test's net does not need the 25's.

**2. One allocation per peer and per record** (`QueueEntries`):

```
--- FAIL: TestPeerCountSweepMatchesTheStoreFormula (0.93s)
    peersweep_test.go:673: slope over the points at or above the floor of 8 peers: ... allocations 13.30 objects/tick/peer over a budget of 12.0
    peersweep_test.go:722: the tick allocates 13.30 more objects for each extra peer, over a budget of 12.0 (measured 9.30, and 9.81 under -race: one subPool.Put interface box, four checksum records and three backing arrays per peer). Anything above that is a per-client allocation on a path that did not have one
    peersweep_test.go:728: at 32 peers the tick allocates 438.1 objects, over the absolute ceiling of 400
```

**3. One allocation per observer and per served tick** (`updateObserversLocked`):

```
--- FAIL: TestObserversAreNotPeers (1.14s)
    peersweep_test.go:900: 64 observers being served cost 7.68 extra allocations per tick over a budget of 1.0, and the budget does not scale with them. The update reads the store through GetFrames, which returns a subslice of frames the store already holds (Replay.GetFrames), and sends them: an allocation here is a copy of the replay per observer per tick
```

**4. A connection buffer sized for the worst** (`ws_listener.go:127`,
`ReadBufferSize` from 1024 to 65536):

```
--- FAIL: TestLiveWebSocketConnectionCostIsBounded (0.21s)
    conncost_test.go:248: a live webSocketPort connection costs 79.0 KiB of live heap, over the 64 KiB the bound allows. The buffers this transport is known to allocate are 2 KiB of gws read and write buffer (ws_listener.go:127) and about 8 KiB of net/http bufio around the socket; a connection costing several times that is a per-connection buffer sized for something else
```

**5. A cap that is not enforced** (`maxAckQueueEntries * 100`):

```
--- FAIL: TestPeerUploadQueueStopsAtTheOriginalsCap (0.08s)
    peerqueue_test.go:172: one peer offered 30100 chunks holds 8065 of them; the cap is 30000 (0x12C, sub_1400E2E00:53). A queue that keeps taking them is not bounded by anything
--- FAIL: TestSaturatedPeerQueuesScaleWithThePeerCount (2.01s)
    peerqueue_test.go:306: 32 saturated peers hold 258080 subs in all, want 960000 = 32 peers x the 30000 cap. There is no global limit in sub_1400E2E00:53, so this number has to be the product and nothing else
```

### 37.9 WHAT WAS MEASURED, AND WITH WHAT SENSITIVITY

What **each** cap detects, measured by putting the defect in front of it:

| cap | detects | does NOT detect |
|--|--|--|
| the store's blocks | any change to the formula | - |
| charged-bytes slope (0.90x..1.10x) | a peer 10% more expensive | 5% |
| actual-bytes slope (0.85x..1.40x) | a peer three times more expensive, or half | 20% |
| store peak against its window | half the window or twice it | 20% |
| live-heap slope (0.75x..1.75x) | a per-peer cost of 2.7x the formula | 20% |
| allocations per peer and tick (12.0) | **three or more** per peer and tick | **one or two**: 9.30 + 1 = 10.30 and 9.30 + 2 = 11.30 pass |
| allocations per tick, absolute (400) | a leak of over 90 objects per tick at 32 peers | - |
| one peer's queue (exact) | any change in the entry or byte count | - |
| queue aggregate (exact) | any peer that does not reach the cap, or that exceeds it | - |
| observers' allocations per tick (+1.0) | one object per observer served, or per visit | one loose object in 256 ticks |
| heap per observer (1 KiB) | a 1 KiB leak per observer | 400 KiB across the 64 |
| heap per connection (64 KiB) | a 64 KiB buffer per connection | 8 KiB |
| heap plus stacks per connection (128 KiB) | a stack leaked per connection | - |
| goroutines per connection (3.5) | one more goroutine per connection | - |

The allocation-per-peer row is the only one with a hole, and the hole is
**deliberate**: the two figures (9.30 normal and 9.81 with `-race`) differ by 0.5 per peer
because the race detector's shadow is per peer and per tick, and a cap tight
enough to catch one would be finer than that separation and **would fail because of a toolchain change and not because of a
defect**. That is why the file's sharp edge is the charged-bytes slope,
and the allocation one is there to catch an order of magnitude, which is what a
real `Count()`-style change brings.

### 37.10 WHAT WAS MEASURED (suite)

With the working set of this repo's test binaries added up every 400 ms. The
filter is the one §36.6 warns about and it has to be looked at: on Windows `Get-Process` returns the
name **without** extension (the `go test` binary is `session.test.exe` and shows up as
`session.test`), so a filter on `*.test.exe` against `ProcessName` gives **0 MB**, and
a filter on the bare name also picks up the `*.test` of **other projects** on the
machine (this machine had another project's `go test ./... -race` running, and the
filter ended up discarding between 15 and 330 of their samples depending on the run).
Here each candidate binary is filtered by the command of
**its parent**, which has to carry this repo's path, and `go -C <repo>` is used
so that it does.

| | 36.5 (previous code) | this machine, with these tests |
|--|--|--|
| `go test -short ./... -count=1` (short loop) | 44.8 s / 122.9 MB | **45.3 s / 116.8 MB** |
| `go test ./... -count=1 -timeout 25m` (full) | 767.9 s / 124.9 MB | **772.4 s / 115.2 MB** |
| `go test -race server+session+sim -timeout 25m` | 796.9 s / 296.2 MB | **805.2 s / 303.1 MB** |
| `internal/session` alone | 2.9 s (17.1 s `-race`) | 4.1 s / 50.1 MB (**18.8 s** `-race`) |
| `internal/server` alone | 37.6 s | 38.5 s / 39.6 s `-race` |
| the three new tests alone | - | 0.71 + 0.12 + 0.08 s (**5.2 s** `-race`) |
| `go run ./cmd/gensim -check internal/sim/testdata` | ok | **ok** (0) |

**+0.5 s in the short loop, +4.5 s in the full one and +8.3 s in the race run,
with RAM equal to or lower than before** (6, 10 and 7 MB lower, within the sampler's noise). The
time goes almost entirely to `internal/session` and `internal/server`, which
run **in parallel** with `internal/sim`, the package that marks the clock: `sim` is at
776.7 s of the full loop's 772.4, and that is why the +4.5 s of the full suite is
build noise and not tests.

The **same measurement with the machine loaded** (a 2 GB game and another project with its
`-race` on top, which is how it was before this round) gives other figures that are worth
having too, because the delta measured **paired**, with and without these tests and minutes
apart, is what really bounds the cost: **+2.3 s and +29 MB** in the short loop,
**+10.2 s and +17.5 MB** in the `-race` one (with `internal/session` at 29.1 s against 19.5 s
without them). Those +29 MB **are not from these tests**: the three tests alone measure **31.5 MB**
of peak, below the **47.7 MB** that `internal/session` already reached without them, and
`internal/server` measures the same with and without (`conncost_test.go` is six server cycles
with 0, 8 and 32 connections). What raises the short loop's sum is **overlap**: `session` and
`server` now take longer, so their peak coincides with `sim`'s more often. The
package that marks the RAM is still `internal/sim` and nobody touches it.

The **+10.2 s** of the race run is almost all `internal/session` (19.5 s -> 29.1
s with the machine loaded; 17.1 s -> 18.8 s with the machine quiet), and it runs **in
parallel** with `sim`, which does not move: the critical path does not change.

### A flake that is **not** from this round

In one of the short loop's runs `internal/sim` failed with
`TestSim_synthetic`, and in another `TestReportSaysWhenEveryDivergenceHappened`: **once
every 8 to 10 runs** of the package, and it **does not** reproduce in 8 consecutive runs of the test
alone nor in 6 of the whole package. Both are in `internal/sim`, which is byte for
byte the same as in `a6eb4d6`: this round does not touch that package and has not investigated the
flake, which looks like clock under this machine's load. It is said here so that
next time it comes up nobody looks in the peers tests.

### What is left

- **Not one line of production touched.** The two formulas hold: the store's slope at 0.99x of `blockSize * f` over the bytes the store charges itself, and the
  aggregate of the queues exactly equal to the product of the per-peer caps. There was no
  defect to fix, and none was invented.
- **The per-peer allocation cap has a declared hole** (one or two it does not
  catch, and the reason is in the code and in 37.9). If it ever has to catch them, the
  path is **decomposing**: measuring the same sweep without checksum records and subtracting,
  instead of an absolute cap on a slope.
- **The per-peer checksum queue grows with no ceiling** if a peer uploads more than ten records
  per tick, because reconciliation drains at most ten per tick
  (`sub_1400E2590:374-381`). **That is what the original does too**, so it is not a
  defect of ours and it is not fixed in this file (§34 already wrote it down).
- **`-maxPagesToStorePerSession` multiplies the store's ceiling** (`f = flag/8`, with no cap
  anywhere). The sweep uses `f = 1.0` because it is the default; with a larger `f` the
  per-peer slope is `blockSize * f`, and the test would say so by changing only the
  constant `peerSweepStorePerPeer`, which is the only thing that would have to be touched.
- **The connection was measured on `webSocketPort`**, which is the expensive one. The `bsPort` one
  has no equivalent because its client reads from its own socket (`client.go:87-88`) and does not
  go through `net/http`; all you can say is that it is cheaper, not by how much.

---

## 38. What a tick COSTS, and the end of the 9.3 objects per peer (2026-10-03)

### WHY THIS ROUND

The 34 measured the cost of a long match (2 h, 57600 ticks, 21 KiB of heap, 30
allocations per tick) and the 37 the cost per client (the store's blocks, the two
upload-queue caps, and 9.30 objects per peer and tick). Both measured
**how much is reserved** and **how much is withstood**, which are the two exact figures and the ones that do
not need a clock, and that is why they are the ones that were asserted. Nobody had asked
either side **how much it costs**: the nanoseconds of a tick.

And the 37 left an open residue written down with its account: nine objects and three
tenths per peer and tick, "short-lived garbage, not a leak", with the sentence that the
count does not travel on the wire. That was true of the wire and false of the cost: 9.3 objects
per peer in a tick that runs eight times per second is about 250 objects per
second and client of short-lived garbage, and the file itself said so without ever having measured
it. This section measures both things.

Besides, a flake leaves through the door that the 37 noted as "looks like clock" and
that was not about clock (§38.4).

### 38.1 WHAT A TICK COSTS

`src_go/internal/session/tickcost_test.go`, three tests, all in the short loop. The
fixture is the 37's sweep one (`sweepFixture`, same peers, same 150 B of
upload, same four checksum records) and the reference mode is
exactly what drives the 37, so the figures are comparable with its.

The clock discipline is the 34's and it is not optional: this machine's monotonic
clock is quantised, and timing a tick gives 0 most of the time (measured there: 19
zeros out of 55 buckets of a single tick). Every figure is a whole window of 4096 ticks
with ONE pair of readings, repeated 3 times and keeping the fastest window: the
machine's load only adds time. The two counters (allocations and bytes) come
from `MemStats` and are exact.

**Measured**, reference mode (1 upload + 4 checksum records per peer and tick),
at `-simulationPeriod` 125 ms, which is what every capture records:

| peers | blocks | ns/tick | allocs/tick | allocs/peer | B/tick |
|--|--|--|--|--|--|
| 1 | 8 | 2234 | 4.00 | 4.00 | 422 |
| 2 | 8 | 2551 | 6.00 | 3.00 | 777 |
| 4 | 8 | 4519 | 6.00 | 1.50 | 1497 |
| 8 | 8 | 11532 | 6.01 | 0.75 | 2666 |
| 16 | 16 | 21508 | 6.02 | 0.38 | 5534 |
| 32 | 32 | 44947 | 6.05 | 0.19 | 11016 |

- time slope over the points at or above the 8-peer floor:
  **982 ns/peer** (18.5 us with `-race`);
- allocation slope: **0.00 objects/peer/tick** (0.50 with `-race`, which is its
  own shadow). The fixed part of a tick is 4 objects and does not grow with the peers.

With `-race` the same tick at 32 peers measures 585 us, that is, about thirteen times more. That is why
this file's clock caps (2 ms per tick, 30 us per peer) are ORDER-OF-MAGNITUDE caps
and say so: the file's sharp assertion is the count, which is a
counter and not a clock.

### THE EXTRAPOLATION, WITH THE ARITHMETIC

Measured at 4 players: 4198 ns/tick. Ticks of a 2 h match at 8 ticks/s =
`2 x 3600 x 8 = 57600`. CPU = `57600 x 4198 ns = 0.242 s` for the whole match.
Measured at 8 players: 8345 ns/tick -> `57600 x 8345 ns = 0.481 s`.

That is, **a two-hour match costs between 0.24 and 0.48 seconds of CPU**, with a
tick every 125 ms. The calculation is a tautology (it is the per-tick times the number
of ticks) and that is why what matters is where the per-tick comes from, which is the
table above and what follows.

A match's memory is the store's, which is `ceil(max(peers,8) * f)` blocks of
`-pageSize` with `f = 1.0`: **512 KiB with 4 or with 8 players**, and 64 KiB more per
peer above eight. It does not grow with the match's duration: it is a cap in BYTES
(the 34: 21 KiB of growth over two hours, and the 512 KiB are the plateau).

**What is measured and what is extrapolated.** Measured: the whole table, the slopes, the
live heap per point, the bytes and allocations per tick, and the block count.
Extrapolated: the CPU seconds per match, which are the product of the measured row
times 57600. What would make the extrapolation false, in order of how much:

1. **The ack body's size.** The fixture uploads 150 B per peer and tick and produces
   an ack of 193 B with one peer and of 4936 B with 32. A real ack in a
   4-player match can be a few KiB. The store is capped in BYTES, so a bigger body **does not
   cost more heap**: it costs more CPU per tick (copy of the body, memcpy into
   the store, more bytes on the wire). The CPU figure above is a floor, and the floor
   grows with the body.
2. **That all the players upload on EVERY tick.** The other way round here: a real
   match uploads in bursts, so the table is the ceiling and not the average.
3. **That the main loop is the only per-tick cost.** It is for the session
   work; connections have their own cost (the 37.6: 16 KiB of heap and 34-36 KiB of
   heap plus stacks per WebSocket connection, with 2 goroutines).
4. **One session per core.** The main loop is ONE goroutine
   (`internal/server/server.go:366`), so N concurrent matches share the same
   tick: the total CPU is the sum and the wall clock is not.

### 38.2 THE PER-STAGE DISTRIBUTION, AND WHY RECONCILIATION IS CHEAP

Measured with the same fixture and **one stage removed each time**, with the windows
taken in round-robin (one per mode and repetition) so that a drift of the machine does not
fall on one mode and not on the others. A mode with no work has to come out
cheaper and one with work more expensive: if a sign comes out
the other way, the table is noise and the test says so.

8 peers (windows of 2048 ticks, the fastest of 3):

| stage | ns/tick | allocs/tick | delta ns | per peer |
|--|--|--|--|--|
| full tick (1 upload + 4 checksums) | 8422 | 6.01 | - | - |
| 3 fewer reconciliation rounds | 7554 | 6.01 | -868 | -108 ns, 0.00 allocs |
| no uploads (heartbeat ack) | 4158 | 4.00 | -4264 | -533 ns, 0.25 allocs |
| 4 uploads per peer | 19403 | 6.07 | +10981 | +635 ns per extra sub |
| 256 observers served | 10186 | 6.01 | +1764 | 61 ns per visit |

32 peers (the same five modes):

| stage | ns/tick | allocs/tick | delta ns | per peer |
|--|--|--|--|--|
| full tick | 30612 | 6.01 | - | - |
| 3 fewer reconciliation rounds | 25154 | 6.02 | -5458 | -171 ns, 0.00 allocs |
| no uploads | 17586 | 4.00 | -13026 | -407 ns, 0.06 allocs |
| 4 uploads per peer | 65035 | 6.18 | +34423 | +494 ns per extra sub |
| 256 observers | 37442 | 6.01 | +6830 | 236 ns per visit |

**Which part dominates: the draining of each peer's queue and the building of the body.**
One extra sub per peer costs 494-635 ns and is the most expensive thing measurable here; with
eight players it is 4-5 us of the tick's 8.4, and with thirty-two it is 16-20 of the
30.6. Reconciliation is 108-171 ns per peer for three rounds, that is, **about 40-60 ns
per round and peer**, and it is not comparable with the draining.

**Reconciliation is linear in the peers, and it is proven by both
magnitudes, not by reading the code.** `AckState.ReconcileChecksums` compares the front record
of each active peer against that of the FIRST peer and nothing else: a `bytes.Equal`
of 8 bytes per peer and round. If it were a pairwise check, the per-peer term would
grow when the peers are multiplied by four, and here it is 108 ns at 8 peers and 171 ns at
32: **1.6 times, not four**, and the difference is the session map's cache and the
cost of walking the entry list, not a second loop. Besides, its cost in
allocations is 0.00 per peer: with the flat cursor buffer (§38.3) it reserves nothing.

**Serving the observer is the cheapest stage and does not scale with the number of
observers**: 61-236 ns per visit for 29 visits per tick (session+2184 is 8
ticks, so the stride reaches one in nine, `sub_1400E4EC0`), and **0.00
allocations per tick** with 256 observers served. The 37.5 already asserted that with 64
observers and this measures it again with 256.

**The pending list and the flushed map** are the part that CANNOT be separated
by mode difference: they live inside `PopTick` and are only visible when there are subs to
drain. What can be said about them is what they cost before (§38.3) and what they
cost now: zero allocations per tick, both.

### 38.3 WHAT WAS POOLED, WHAT WAS NOT, AND WHY

Measured at the 32-peer point of the sweep, with the same test that produced the 9.30:

| | before | after |
|--|--|--|
| allocations per tick | 310.10 | **6.05** |
| objects per peer and tick | 9.30 | **0.00** |
| bytes per tick | 21869 | **11016** |
| ns per tick | 60190 | **44947** |
| live heap per peer | 71284 B | 71107 B (the store, which is not touched) |

The heap does not move because pooling does not touch the store: what was taking 9.3 objects
per peer was all of the ack path, and a match's memory is the 37.1's.

**What was done**, each one with its reuse cycle proven on the same
path and every tick:

1. **`subPool` becomes `*subBuf` and no longer `[]byte`.** A `[]byte` inside a
   `sync.Pool` **reserves 24 bytes on every `Put`**, because an interface cannot carry
   a three-word header: that was the object per peer and tick. A pointer goes in as it is.
   The property is one line and is written in place: a `*subBuf` is in
   `AckState.subs` or in the pool, never in both, and `PopTick` copies the bytes into the
   body **before** returning it.
2. **The checksum records are a flat buffer per peer with a read cursor**, not
   a `[][]byte`. The problem was not the record but the **array**: removing from the front
   in Go consumes capacity, so the next append had to grow again. With a
   cursor, emptying is `buf = buf[:0]`, which keeps the capacity, which is exactly what
   the original's `std::vector`'s `erase` does (`session+0x8c8`). Besides, the cursor
   puts the comparison in one place (§38.2).
3. **The pending list is a reused field** (`AckState.pending`) instead of a
   `make` per tick: it is copied into the body before returning it and nobody outside the file sees it.
4. **The set of emptied slots is 256 bits on the stack** (four `uint64`), not a
   `make(map[byte]struct{})` per tick. The slot is a byte, so the whole set
   is four words and there is nothing to reserve.
5. **The sort's rank map is a `[256]int16` array** in the `AckState`. The
   ranks are filled from the entries vector, which is already walked.
6. **The per-slot accounting is NOT cleared when it reaches zero.** It cost an allocation
   per peer and tick and bought nothing: a slot is a byte, the whole map is 4 KiB
   for a session that at some point used all 256 slots, and `QueueEmpty` already treats
   `count == 0` and "entry absent" the same.

**What was NOT done, and why**, which is the part that has to be written down:

- **The ack body is not pooled.** It is not caution, it is property: **it leaves this
  file**. `PopAckTick` hands it to the store (which copies it) and to every peer, and its life
  ends in `internal/server`, where each transport builds **its own** frame from
  it. Nobody outside the package says "I do not need it any more", so there is nowhere to return it;
  and a pool you hand to N times (one per peer) and return once is exactly how the same buffer ends up in two messages. It is the defect this project cannot
  afford, so it stays as one allocation per tick and not per client.
- **The `sort.SliceStable` stays.** It still has a closure and a reflective
  swapper: two objects per tick, and only in a flush with mixed slots (with a single
  peer the flush is one slot and there is nothing to sort). Writing a stable
  order by hand by ranks would remove them, at the price of hand-writing the order of the
  bytes the client reads, and two objects per tick are not that price.
- **The queue records' snapshot (`queueEntries`) is not reused.** It would be
  a shared buffer **under a read lock**, which is a race waiting to happen, for
  one object per tick per session and not per client.
- **The send closure is not removed.** `ForEachPeer` takes a function, and both the
  fixture and the server's loop (`handlers.go:341`) build it with the ack
  body inside. It is one object per tick per session.

**The caps are tightened because the figure that justified them no longer exists.** The
allocation slope per peer goes from 12.0 to **3.0** (measured 0.00 without the detector and
0.50 with it), and the absolute per tick from 400 to **24** (measured 6.05); in the long-match
test the per-tick cap goes from 48 to 24 (measured 6.01). With the old caps
the first of the falsifications below **passes**, which is exactly what a cap
with twelve times of slack does.

**The three falsifications**, with the defect in production, measured and removed:

| defect | what it sees |
|--|--|
| an `append` per checksum record again (4 per peer and tick) | `the tick allocates 4.00 more objects for each extra peer, over a budget of 3.0` and `at 32 peers the tick allocates 134.1 objects, over the absolute ceiling of 24`: both caps |
| `subPool` back to `[]byte` (the interface box per `Put`) | the 37's sweep in red, with the previous figure of 9.30 |
| the observer visit copying the frame list instead of `GetFrames` | `31.9 extra allocations per tick` in both files, and the ns cap **does not** see it: they are 292 ns per visit instead of about 100, because what is copied are the headers and not the bodies |

What the third row leaves said is what the 37.5 already asserted and what is now measured
by both routes: **the observer's sharp assertion is the allocation
count, not the clock**.

### 38.4 THE `TestSim_synthetic` FLAKE: IT WAS NOT ABOUT CLOCK

The 37 left it like this: "once every 8 to 10 runs of the package, and it does not reproduce in 8
consecutive runs of the test alone nor in 6 of the whole package. Both are in
`internal/sim`... it looks like clock under this machine's load".

**There is no clock assertion in that test.** What fails is the replay server's start-up, and the
signature is this:

```
--- FAIL: TestSim_synthetic (0.04s)
    sim: start server: ws listener: listen :31681: listen tcp :31681: bind:
    Solo se permite un uso de cada direccion de socket
```

The harness **picked a port, released it and had the server bind it
afterwards** (`freePort` + `prepareConfig`). Between the release and the bind the port belongs to
no one, and on Windows a listener's `bind` is exclusive: whatever takes it in that
window keeps it, `Start` fails and the whole replay dies **before a single message is
compared**.

Measured on this machine, and why the 37 did not say it:

| how it is stressed | result |
|--|--|
| `go test -short ./internal/sim` with 12 load processes | **1 failure in 8 runs** (and 1 in 3 in another series) |
| the test alone, `-count=30`, with the same load | **0 failures** |

That is, it does not reproduce with the test alone, and that is why it looks like clock. The whole
package creates far more connections per minute than the test alone, so the window is
won many more times.

Whether the window is won depends on what is creating sockets at the same time. Measured on a
quiet process on this machine, "ask for a port, release it and bind it" **never
fails**: 65000 consecutive pairs, 0 rejections, and not even with a 20 ms window between
the two calls (which is what a scheduler that has preempted does). The defect
needs competition in the ephemeral port range, which is what the rest of the
package does.

**The fix is not to retry nor to widen a tolerance: it is not to pick the port.**
`Config.EphemeralListenPorts` makes both listeners bind `":0"` and
`Server.BoundPorts` returns what they bound, so the port **is never left
free**. It is not a flag: `-bsPort 0` has to keep meaning that that transport
is off, and that is why it is a `Config` field that no flag writes and that only
`internal/sim` uses. The WebSocket listener had no way to say which port
it had (`GetLocalAddr` returned the configured one, and that is why `internal/server`'s
tests were worked around with `wsPortOf()`); now it has `GetPort()` like
bs's.

The two ports become **one** deviation, `listen ports`, because the harness no longer
picks them: they arrive from the server and are declared together (so
`TestRecordedArgsReachTheServer` expects that line and not two).

**Falsified in both directions.** With `EphemeralListenPorts` unread in `New`,
`TestEphemeralListenPortsBindTwoDistinctPorts` reads 0/0 and fails. And with the old harness
in place, the loaded package stress is what made it appear (§38.6).

**What is left of this flake, untouched:**

- `internal/server` has its own `freePort(t)` for the tests (`load_test.go:19`)
  with the same shape and the same race. The same mechanism fixes it and it has
  not been touched because it is not the reported flake and the change touches ten test places that
  would have to be re-measured one by one.
- In one of the stress runs **another** test of the same package was seen failing,
  `TestCroppedCaptureStillAnswersTheSame`, and that time **not** with a bind (the
  instrumentation that was put in for a while to identify the port thief did not
  fire). That test does TWO `Check`s in a row and compares the trimmed capture with the
  whole one, that is, it has double the start-up exposure **and** a comparison of
  two streams in real time; its cause is not established and no cause is asserted.

### 38.5 WHAT WAS MEASURED (suite)

With the working set of **this repo's** test binaries added up every 400 ms, and
filtering each binary by the command of **its parent** (which has to carry this repo's
path), which is what §36.6 and §37 warn about: on Windows `Get-Process` returns the
names WITHOUT extension, so a filter on `*.test.exe` gives 0 MB and one on the
bare name picks up other projects' `*.test`.

| | 37 (this machine) | with this |
|--|--|--|
| `go test -short ./... -count=1` (short loop) | 45.3 s / 116.8 MB | **43.6 s / 110.8 MB** |
| `go test ./... -count=1 -timeout 25m` (full) | 772.4 s / 115.2 MB | **770.0 s / 115.6 MB** |
| `go test -race server+session+sim -timeout 25m` | 805.2 s / 303.1 MB | **790.1 s / 309.3 MB** |
| `internal/session` alone | 4.1 s / 50.1 MB (**18.8 s** `-race`) | **6.0 s** (**49-77 s** `-race`) |
| `go run ./cmd/gensim -check internal/sim/testdata` | ok | **ok** |

**The short loop's time and the full suite's have not moved** (-1.7 s and
-2.4 s, within the noise) and neither has RAM. What goes up is `internal/session`, which is
where the new file lives: +1.9 s without the detector and from 18.8 to 49-77 s with it, because
the clock figures are measured in windows of 4096 ticks and the detector multiplies by
13 the work of each tick. **It does not move the critical path**, which is `internal/sim`
(764.6 s of the full loop's 770.0): `session` and `server` run in parallel with it.

A trap from the 37 that came back and is worth not forgetting: the sampler's filter
requires the **parent's** command to carry the repo path, and that only happens if `go` is
launched with `go -C <repo>`. Launched from the working directory, the parent is a
`go test -short ./...` with no path and the sampler measures **0 MB across 53 samples** without
warning. The first time this was measured with the change, that is exactly what came out.

### 38.6 WHAT WAS MEASURED (the flake, before and after)

With 12 load processes and `go test -short ./internal/sim -count=1` repeated:

| | before | after |
|--|--|--|
| package runs | 1 failure in 8, and 1 in 3 in another series | **0 failures in 12** |
| `TestSim_synthetic` with `-count=30` | 0 failures | **0 failures** |
| `TestSim_synthetic` with `-count=50` | (not measured) | **0 failures**, 204 s |

And with the defect in place (port picking returned to the harness) the failure appears in
the series of 3 to 8, which is the same window as before. The standalone test's `-count=50`
**is** not proof of anything and is written down so that it does not look like it: without the
whole package next to it, which is what competes for the ports, that test does not fail even with the defect.

---

## 39. The real PROFILE: where a tick's CPU and garbage go, with function names (2026-10-04)

### WHY THIS ROUND

The 34 measured what a long match RESERVES, the 37 what it costs per CLIENT, and the
38 what a tick COSTS. All three use counters, and that is why they are exact and do not
need a clock. What nobody had asked was **where**: the three say HOW MUCH
and the 38.2's per-stage split says WHICH STAGE, and both are categories, not
functions. "An extra sub per peer costs 494 ns" says how much; it does not say where. And a subtraction
between two modes cannot attribute cost to the code that carries it, because the subtraction
eats everything between the two modes.

And the 38.2's per-stage split has a limitation that cannot be removed with more
subtractions: it is measured BY DIFFERENCE, and a stage cheaper than the clock's
resolution comes out as zero or with the wrong sign. With the clock in place instead of by subtraction, each
stage brings its own list of functions.

### 39.1 WHAT WAS MEASURED, AND HOW

Two files, and they are TWO HALVES that do not overlap:

| | `internal/session/tickprofile_test.go` | `internal/server/serverprofile_test.go` |
|---|---|---|
| what it profiles | the tick, with the 37/38 fixture | a real server, with real TCP clients |
| what goes in | PopAckTick, PopTick, each peer's queue, reconciliation, Session.Tick, the observers | all of the above PLUS the socket: accept, read, write, goroutines, scheduler |
| what does NOT go in | **the socket**: `sweepConn` counts and stores nothing | **the tick**: it is invisible, with the arithmetic below |
| window | 1.2 s of wall clock per profile | 3 s of wall clock |

Each profile is **1.2 s of WALL CLOCK**, not a number of ticks, and that is what makes
the parts comparable between modes: a flat part is a part of SECONDS of CPU,
so equal time is equal samples. The sampling rate is raised to 1000 Hz
(`SetCPUProfileRate` before `StartCPUProfile`; the runtime prints "cannot set cpu
profile rate until previous profile has finished" when `StartCPUProfile` tries to
set its 100 and is refused, and that message is NOT a failure: what decides is the number of
samples, which is read from the file). At 100 Hz that window would carry 120 samples.

With what was gained, over the 38.2 fixture, 32 peers, full tick
(1 upload + 4 checksum records per peer and tick):

| function | flat | cumulative |
|---|---|---|
| `runtime.mapaccess2` | **11.02%** | **36.41%** |
| `internal/runtime/maps.memHashAES` | **10.18%** | 12.17% |
| `maps.ctrlGroup.matchH2` | 4.72% | |
| `maps.(*groupReference).key` | 3.46% | |
| `runtime.memequal8` | 3.25% | |
| `(*AckState).QueueEmpty` | 2.52% | 7.35% |
| `runtime.memmove` | 2.52% | |
| `(*AckState).QueueEntries` | 2.20% | 6.09% |
| `(*AckState).ReconcileChecksums` | 2.20% | **24.13%** |
| `(*AckState).PopTick` | 0.94% | 9.76% |
| `(*AckState).QueueDataSub` | 3.20% | 8.83% |

The **per-stage split with the clock in place** (cumulative over the whole window), which
is what the 38.2 could not give:

| stage | % of the window | where it is inside |
|---|---|---|
| `Session.PopAckTick` | **48.5%** | `queueEntries` 25.8% + `ReconcileChecksums` 24.1% |
| `Session.Tick` | 18.1% | `closeSyncTimedOutPeers` 15.5%, almost all `syncTimedOut` |
| `Session.QueueGameData` | 9.4% | `QueueDataSub` |
| `AckState.QueueEmpty` | 7.4% | |

**Where the map is, which is the question that comes out of here**: `ReconcileChecksums`
spends 172 ms of its 230 in one `mapaccess1` (that is 75%), and it is two lookups per round
(ack.go:400 and :421). `QueueEmpty` 33 of 70, `QueueDataSub` 19 of 90, `releaseSub` 19,
`QueueEntries` 18, `LastSync` 21 of one `mapaccess2`. Summed: **282 ms of 953 = 29.6%
of the tick**. The other 82 ms (8.6%) are `Session.Peers`.

And one thing that is **NOT** what it looks like: the `lock2` + `semasleep` +
`lockWithRank` block (15% of the window, which looks like a very expensive mutex) is **the runtime's own
lock**. `semasleep` comes 100% from `lock2`, and `lock2` 100% from `lockWithRank`, and their
callers are `gcMarkTermination` and `findRunnable`. The session's lock is
`sync.(*Mutex).Lock` at **0.16%**. That is: **the session mutex's cost is NOT the
problem**, and half a window of the "synchronisation cost" you see in a profile of this
code is the collector and the scheduler.

**The tick does not show up in the server's profile, and the arithmetic is the reason**:
`-simulationPeriod 125` is 8 ticks per second, and at 4 peers a tick measures ~3.2 us, so
a 3 s window carries `3 x 8 x 3.2 us = ~77 us` of tick in 3 s of clock, that is
**0.0026%**. At 1000 Hz that is 0.08 expected samples. Measured: the server's window
carried **16 to 54 samples in total** (the whole process, **0.53% of a core**), with the
tick appearing once. One sample is not a measurement, so the tick's attribution is the one of
the `session` file and the server's is everything else's, and they are
reported side by side and not mixed.

### 39.2 MEMORY: WHERE THE SHORT-LIVED AND THE LONG-LIVED ARE

With `runtime.MemProfileRate = 1` (**not** the default 512 KiB, and the reason is that without
this the profile LIES: at 512 KiB the window's 48.75 MB come out ranked in TWO
places and the 4 to 6 objects per tick do not appear, because 6 KiB against a
512 KiB interval is a 1% probability. That is, the profile was saying "nothing is left" for the
wrong reason: it could not see what was left).

Short (4096 ticks, everything allocated), 32 peers:

| site | objects | per tick | bytes | % of the bytes |
|---|---|---|---|---|
| `(*AckState).PopTick` | 12,306 | **3.00** | 21,600 KiB | 47.8% |
| `(*Session).queueEntries` | 8,192 | **2.00** | 768 KiB | 1.7% |
| `(*Replay).AddFrame` | 4,106 | **1.00** | 21,509 KiB | 47.6% |
| `internal/reflectlite.Swapper` (the `sort.SliceStable`) | 4,096 | 1.00 | 128 KiB | 0.3% |

**6 objects per tick and per SESSION, none per peer**, the same as the 38.3. The three of
`PopTick` are the ack body, the `sort` closure and the `sort`
swapper. And the bytes: **95.4% are the ack body and the copy the store makes**,
which is exactly what the 38.3 proved that **cannot** be pooled by property (it leaves
the file, N peers receive it, and nobody outside says "I do not need it any more").

Long (`inuse_space` after two collections, the store full and the fixture alive):

| | 8 peers | 32 peers |
|---|---|---|
| `(*Replay).AddFrame` | **480 KiB (99.1%)** | **2.12 MB (99.7%)** |
| everything else | 4 KiB | 2 KiB |
| live objects, in total | 391 | **421** |

That is: **everything that lives is the store, and it is 421 objects in the whole session.** There is
no leak, no accumulation per tick, and the only long-lived thing is the structure that is
already capped in bytes and has already been measured (37.1). **The conclusion about memory is that there is
nothing left to attack**, and now it is MEASURED and not assumed: the three sites that remain
are the ack body (not poolable, proven), its copy in the store (it is the store) and
two objects of the `sort` and two of `queueEntries` that the 38.3 wrote down with their
reason. The 45 MiB of a 32-peer window are CHURN, not heap.

### 39.3 WHAT CHANGED: THE THREE `AckState` MAPS

`queued map[byte]*slotQueue`, `perPeer map[byte]*checksumQueue` and
`lastSync map[byte]time.Time` are now `[256]...`. The key is a byte and the domain is
256: a map is a hash of a value that is already a perfect index. It is the same
argument as `rank [256]int16`, which the 38.3 already applied for the same reason.

Measured on this machine, the same test run back to back, with and without the change:

| peers | ns/tick before | ns/tick after | delta | allocs/tick | B/tick | live heap |
|---|---|---|---|---|---|---|
| 1 | 1,699 | 1,238 | **-27%** | 4.00 | 422 | +15 KiB |
| 2 | 2,702 | 2,244 | -17% | 6.00 | 777 | +16 KiB |
| 4 | 4,164 | 3,225 | **-23%** | 6.00 | 1,497 | +16 KiB |
| 8 | 8,926 | 5,927 | **-34%** | 6.01 | 2,667 | +15 KiB |
| 16 | 19,884 | 11,266 | -43% | 6.02 | 5,535 | +14 KiB |
| 32 | 36,070 | 20,932 | **-42%** | 6.06 | 11,017 | +12 KiB |

**slope per peer: 1,114 ns -> 622 ns (-44%)**. Allocations and bytes do NOT move,
which is what fits. The two-hour match: 0.240 s -> **0.186 s** of CPU at four
players, and 0.514 s -> **0.341 s** at eight.

What is paid is **+15 KiB of live heap per session, and it is FLAT in the peers**
(the column above), because an `AckState` is one per match and not one per client: the
per-peer slope that the 37.7 and the 38.1 bound stays at 0.00 objects per peer and tick.

**The falsification**, on both sides that exist:

- With the maps back in (the defect reintroduced by hand, same test and same
  machine): 36,070 ns/tick at 32 peers and a slope of 1,114 ns/peer, and the profile brings
  `mapaccess2` back to 36.4% and `memHashAES` to 10.2%.
- With the change in: **the five committed captures of `internal/sim`, which is the
  byte-for-byte comparison against the original.** An array where there was a map changes what is
  read if the map did not return the same thing, and that comes out as a must-match. It did not come out.

### 39.4 GOROUTINES AND CONTENTION: YES, THERE IS, AND IT DOES NOT MATTER

Real server, 4 real TCP clients, 3 s, `-simulationPeriod 125`, with
`SetMutexProfileFraction(1)` and `SetBlockProfileRate(1)` (both records).

**Goroutines**: 5 with no client at all, **17 with four**, 17 after the window. The
17, from the goroutine profile:

| n | where |
|---|---|
| **4** | `(*Client).readLoop` (client.go:106) reading: one per connection |
| **4** | `(*Client).writeLoop` (client.go:129): one per connection |
| 4 | THIS file's own client readers (serverprofile_test.go:331) |
| 1 | the pprof profile's own writer |
| the rest | main loop, runtime, GC workers |

That is **TWO server goroutines per connection**, which is what the 35 already
said, now with the number in front of it.

**Contention, which is the question**: out of **95 contended mutex acquisitions** in
3 s, **8 are ours** and the other 87 are runtime locks (GC and scheduler):

| mutex | events | total |
|---|---|---|
| `runtime.gcMarkTermination -> forEachP -> unlock` | 4 | 0.777 ms |
| `runtime.findRunnable -> unlock` | 17 | 0.584 ms |
| `runtime.stopm -> unlock` | 3 | 0.468 ms |
| **`(*AckState).QueueDataSub` (ack.go, `a.mu.Lock()`)** | **8** | **0.306 ms** |
| `(*Client).writeLoop` | 1 | 0.004 ms |

And the block profile, with its 153 events, says the same from the other side: the only
block of ours is the **same** `(*AckState).QueueDataSub`, 8 events, 0.234 ms.
The two big ones are not contention but **idle waiting**: `(*Client).writeLoop`
waiting for work (`selectgo`), 12,472 ms spread over 104 events, and
`(*Server).mainLoop` waiting for the next tick, 3,122 ms over 25 events.

**The answer, with numbers**: 8 contended acquisitions of `AckState.mu` over 24
ticks (3 s at 125 ms), 0.31 ms in total, **0.010% of the window**. And
`Session.Tick`, `updateObserversLocked` and the peers map **do not appear even once**
in the mutex profile. At this scale contention is negligible, and now that is a
figure and not an opinion. What does appear, and with more weight, is 87 events on the runtime's
own locks, which is the collector doing its job.

### 39.5 READING A pprof PROFILE IN GO 1.27, AND ITS FOUR WAYS OF FAILING

A whole package (`internal/profiletest`) was needed to be able to ASSERT that a profile
has samples, which is the only way a profile of an empty window does not produce a
spotlessly false top-10. The four things that go wrong, and all four were
measured:

1. **Go 1.27 compresses with GZIP** what `StartCPUProfile` writes. The file starts
   with `1f 8b`; a reader that takes it for protobuf counts **0 samples** in a file
   that `go tool pprof` reports as full. Measured: **0 against 958**.
2. **`Sample` is field 2 of `Profile`**, not 1 (1 is `sample_type`). Counting field
   1 counts the two value types.
3. **The number of Sample RECORDS is not the number of samples.** Go merges identical stacks
   inside one profiling cycle: measured **678 records with 958
   samples**. The count comes out of `Sample.value[0]`, and it has to be the **first** of
   the two (Go writes `[samples, nanoseconds]`, and the second here is 958 million).
4. A field skipped with `continue` instead of with **its payload** desynchronises the
   walk and counts TOO MANY, instead of failing loudly: **1192 against the 1076 that
   `pprof -sample_index=samples` says**. Without that comparison, the counter would have passed
   for good.

`profileSampleCount` is checked against `go tool pprof -sample_index=samples` on
the six files of the 39.1: **both give the same number**.

And there is a fourth assertion that a sample count cannot give, and it is the one that
really matters: **Profiles have to MENTION the measured code.** The fixtures
run in the same process as the code they drive, so a profile of a window in which the code
never ran is a well-formed profile of the harness. Both
drivers check that the profile names `(*AckState).PopTick` and
`internal/server.` respectively.

### 39.6 THE TWO ASSERTIONS THAT HAD TO BE FIXED, AND THAT THIS SECTION WAS NOT CREATED FOR

Both were uncovered by this work and both are its own, and both have the same
shape: **cheaper work made more sensitive an assertion that had been written
in terms of the more expensive one.** That is the price of optimising and that is why it is written down.

**39.6.1 A profile's sample count cannot depend on the machine.**
`internal/profiletest` required 100 samples from a 250 ms window. It failed **1 out of 1 in
a COMPLETE suite run with `-race`**, and it did not reproduce either 8 times in a row
alone nor in 5 runs with other packages alongside (0.57 s against 0.28 s: the
thread had been descheduled). **The cause is not established and is not guessed.** What
can be said is what it was: how many samples the profiler hands to a thread whose CPU is
taken away, and the full suite is the only place in this repo where four
packages run flat out for thirteen minutes. The count's magnitude **is not asserted any more**;
the decoder is proven with a hand-built profile (exactly 300), with `Contains`
finding the test's own function inside a runtime profile, with
stability between two reads, and with the ceiling of 5.000, which is
independent of load.

**39.6.2 The per-stage split's tolerance cannot be a fraction of the
reference.** `TestTickCostSplitsIntoStages` used `ref.nsPerTick * 0.15`. The tick got
42% cheaper, the tolerance went down with it (from ~2,800 ns to ~1,700) and **the noise it
has to absorb did not go down**, because it is the spread between two windows of the same
mode and that is the clock and the scheduler. Measured on a machine with an unrelated game and IDE:
**with maps 0 failures in 6 runs, with arrays 2 failures in 6**, with this
failure:

    8 peers: 1 upload + 1 checksum record a peer REMOVES work and still came out
    3111 ns a tick dearer than the reference (14433 against 11322, tolerance 1698)

The tolerance is now the **maximum of the fraction and the spread THAT THIS RUN
MEASURED** (`costStageTolerance`): each mode already takes its three windows in round-robin, so
the distance between its fastest and its slowest window is the machine's drift over
the length of the test. Measured: **7,182 ns** of tolerance with the fraction at 1,541, and **0
failures in 8 runs**.

A tolerance that grows cannot swallow the check, so the sign rule
goes to `costStageSignVerdict` and the tolerance to `costStageTolerance`, and
`stagetolerance_test.go` handles them with numbers and without a clock. **And what makes it not
decoration: the tests call the two functions and not a copy of the formula.** The
first version wrote the `max()` in the test too, and with the defect in place (the
fraction alone) **the whole file came out green**, because it was checking its own
arithmetic. With the defect in place now:

    costStageTolerance returned 1541 ns, which is under the noise it has to absorb

1541 is exactly the failing run's tolerance. And there is one more line: if the measured
noise is the size of the smallest stage, the test prints **NOT RESOLVED** with the
two figures and says that the checks pass because the tolerance covers the noise and not
because the stages were measured. **The test is NOT skipped**, because a rule that decides to
NOT run depending on the machine is exactly the class of rule this repo rejects (§28,
`TestFastPathIsNotTheFullSuite`). It is measured, said, and passed.

### 39.7 WHAT WAS LEFT UNTOUCHED, AND WHY

**`Session.Peers` is still `map[byte]*Peer`, with 17% of the window inside**
(13.4% of lookups and 4.2% of traversal). It is not a representation change but a change
of **order**: `Peers` is exported and walked with `range` in five places
(`GetPeers`, `ForEachPeer`, `LeaveSimulation`, `PeerByKey`, `RemovePeerGracefully`,
`Close`), and the order those five see is today a map's order, which in Go is random.
Converting it makes it ascending by slot, which is an observable fact, and
`ForEachPeer` is the one that sends the ack to each peer. The justification that would be needed
is "the send order does not reach the wire", and that **is not proven by the profile nor by
any capture**. With rule 1 and with the byte-for-byte comparison, it is left measured and
written down.

What is left proven is that the path that DOES reach the wire does not depend on the
map's order: **`queueEntries` (session.go:464) walks `s.peerOrder`**, an explicit list in
join order, and not the map. The `0x0A` pending list has never been exposed to this.

The ack body and its copy in the store (95.4% of the bytes) were not touched either: the
38.3 already proved that pooling them is the defect this project cannot afford.

### 39.8 WHAT WAS MEASURED (suite), AND WITH WHAT LOAD

**Warning that goes before the figures**: this section's latest measurements were
taken with the machine at **87-100% CPU from an unrelated process** (a game, 16,700 s of
accumulated CPU, and an IDE). That stretch's clock figures **are not comparable**
with the 38.5's, which were taken on a quiet machine. Both are given and it is said which
is which.

| | 38.5 (quiet machine) | this section, with external load | note |
|---|---|---|---|
| `go test -short ./... -count=1` | 43.6 s / 110.8 MB | **97.3 s / 117.4 MB** | under load; **46.3 and 47.9 s / 117.4 and 128.3 MB** earlier, with less load |
| `go test ./... -count=1 -timeout 25m` | 770.0 s / 115.6 MB | **791.5 s / 116.7 MB** | `internal/sim` 772.5 s |
| `go test -race ./internal/server/ ./internal/session/` | (not measured like this) | **157 s / ok, 0 races** | the 38.5 measures server+session+sim in 790 s |
| `go run ./cmd/gensim -check internal/sim/testdata` | ok | **ok (exit 0)** | |

**What is comparable, and why**: `internal/sim` is 768-772 s and **does not move
with the load nor with this change**, because they replay captures in real time and their clock
is the recording's duration, not the server's CPU (the 38 says it: the replay goes in
real time on purpose). The +2.5 s over the 38.5 are machine drift. And the
`+1 s` the new packages bring (`internal/profiletest`, plus the two profile
drivers that are skipped without `BS_PROFILE_DIR`) is 0.13% of the suite.

The sampler's filter is the 38.5's, and the sampler repeats as a lesson: **the sampler itself perturbs
what it measures**, because `Get-CimInstance Win32_Process` every 400 ms is a heavy WMI
call. The RAM figures are trustworthy (sum of working sets, attributed by the parent's
command, 0 unattributed samples in most runs); the clock figures of the
runs with the sampler on top are better without it. In runs with two or more
other projects' `*.test` on the machine, the sampler leaves them OUT and says so
(`samples_with_unattributed_test_binaries` from 3 to 78 of 774), which is correct: it does
not sum them.

### 39.9 THE PROFILES ARE NOT COMMITTED

They are regenerated, and they do not take up space in the repo:

```
BS_PROFILE_DIR=<a directory> go test ./internal/session/ -run TestTickProfile       -count=1 -v
BS_PROFILE_DIR=<a directory> go test ./internal/server/  -run TestServerRunningProfile -count=1 -v
```

Each one writes to `<directory>/session/` or `<directory>/server/`. Neither runs in
the suite without the variable (they are skipped), both have their budget written down (1.2 s and
3 s of window, with tick and frame caps as named constants), and both
**settle** that the real window did its work: more than enough samples, the fixture
in its stable state, one tick per tick with `PopTick` in the profile, and the goroutines
matching the count. A profile of an empty window has to be a failure, not a top-10.
---

## 40. The stash, the peers table, the last port, and the observer step (2026-10-04)

Four things, in the order they were asked for, and all four with their why.

### 40.1 THE STASH THAT WAS LEFT: IT WAS A REVERT, AND ALL IT BROUGED DOES NOT SURVIVE

`stash@{0` said: *"REVERT-IN-PROGRESS of HEAD~2..HEAD: unsigned, unverified,
NOCommitted. I leave it intact here: it is a portability distinction that the
Stop() panic fix and unfinishedReplay() carry with them"*.

Measured before deciding anything: the stash's tree is `9c8144b` (the Stop()
panic's commit) with `393d528` reverted on top, and `git diff HEAD stash@{0}` does not contain
one line that HEAD does not have. That is, **there is no new code to rescue**. The only
thing the stash offered over the current state was demoting three things:

| | what is demoted | why not |
|--|--|--|
| 1 | `Stop()` goes back to nilling `clientSet` instead of emptying it | it is the measured panic of 1 in 40, and the fix is inside |
| 2 | `unfinishedReplay()` disappears | without it, a connection the harness could not drive goes back to being reported as a server must-match divergence (the 38, with 24 burners: 3 of the 5 `TestCroppedCaptureStillAnswersTheSame` failures) |
| 3 | `EphemeralPorts` (a set) goes back to a `bool` that turns on **both** transports | it is the label's "portability distinction", and it is the only thing that had to be decided |

Point 3 does not survive, for two reasons that do not depend on anyone saying so:

- **It is not a portability distinction.** Asking the system for the port at bind
  time (`:0`) is the same mechanism on every platform. The only platform fact in the whole of
  `Config.EphemeralListenPorts`'s comment is that a listener bind on Windows is exclusive, and that
  is what makes a port picked by binding and releasing a problem HERE; it says nothing about a set against
  a bool.
- **The bool is worse everywhere**, because it changes what "this transport is
  off" means: a caller that asks for an unknown `webSocketPort` finds
  a `bs` listener it did not ask for, on a port it is not going to read, and a test that measures
  one transport ends up measuring two. The assertion that holds it up already exists
  (`TestBoundPortsIsZeroForATransportThatIsOff` in
  `internal/server/ephemeralports_test.go`): it asks for the ephemeral `bsPort` and checks that
  `BoundPorts` names no `webSocketPort` **and that the WebSocket listener has not
  been built**. With a bool that assertion cannot even be written.

The why is left in `Config.EphemeralListenPorts`'s comment, which is
where someone would look again, and the stash is thrown away. **No stash is left in the
tree.**

### 40.2 `Session.Peers`: THE ORDER DOES NOT REACH THE WIRE, AND THE PENDING LIST DID NOT READ THE MAP

The 39.7 left `Session.Peers` as `map[byte]*Peer` with 17% of the tick's window
(13.4% of lookups, 4.2% of traversal) untouched, and wrote it down with a reason that
was not proven: *"the send order does not reach the wire"*. That was proven neither by
the profile nor by any capture. Here it is proven, in both directions.

**First, the `0x0A` pending list did not read the map, and it is checked.** The
list comes out of `QueueEntry`, and `queueEntries` (session.go:510) walks `s.peerOrder`.
That was already there, and `TestPendingListFollowsPeerAddOrderNotSlotOrder` has carried it since
`6410e06`: with the peers registered in the order 2, 1 the list comes out 2, 1. If it read the
map, that test would pass half the time.

**Then, the five places that walked the map, one by one:**

| site | what it does | is the order visible? |
|--|--|--|
| `GetPeers`, `ForEachPeer` | the ack and the desync notice (`handlers.go:360`, `:365`), the relay's `MsgPeer` (`:728`), `Broadcast`, the liveness check (`server.go:531`), `RemoveSession`, the expired sweep | **no**: they all pass **the same body** to every peer, and the ack and the notice are two **sequential** passes, so the order between peers cannot change either the order inside one connection. And the harness's comparison is **per connection** (`compare.go:527`, `compareConn` with a `want` and a `got`) |
| `LeaveSimulation` | existential ("is any still participating") | no |
| `PeerBySecret` | "the first whose `peer+80` equals X" | **yes**, when two peers share the secret. And there is the binary: `sub_1400E5CB0:33-39` (`src/sub_1400E5CB0_1400e5cb0.cpp`) walks a vector eight bytes at a time and returns the **first** peer, that is, the one that joined earlier. With the map it was a coin toss on every run |
| `PeerByKey` | the same by `key` | yes, and in the original it **cannot** happen: its table is indexed by identity (`sub_1400E5C40_1400e5c40.cpp:19-30`) |
| `Close` | closes everything | no |

**What changes, and why it is the original's shape and not a speed
choice**: `Peers` becomes `[256]*Peer`. `sub_1400E5670:100` is the only writer of
a peer in the binary and it writes `session + 24 + 8*peerID`: an 8-byte slot per id,
with id 0 unused. There is no hash table on that path. 255 slots are 2 KiB per session,
which is what the original reserves too.

What that forces, and which is half the work: **a slot table has no
order**, so every iteration has to say which one it wants, and there is exactly one
answer: `peerOrder`, which is the `2192` vector. The five places from before
walk `peerOrder`; `PeerCount()` is `len(peerOrder)`, and the two "the session
has no peers" comparisons too (they are the `v5` ones in `updateObserversLocked`, and
there it matters: see 40.4).

Measured on this machine, the same test run back to back and with the same load (see 40.5):

| peers | ns/tick before | ns/tick after | delta |
|--|--|--|--|
| 1 | 1205 | 1000 | -17% |
| 2 | 2534 | 1550 | -39% |
| 4 | 4383 | 2385 | -46% |
| 8 | 7302 | 4043 | -45% |
| 16 | 12871 | 7996 | -38% |
| 32 | 25271 | 14636 | **-42%** |

Slope per peer: 748 ns -> 441 ns. The allocations and bytes per tick **do not
move** (4.00 / 6.05 objects and 422 / 11,016 B), which is what fits: this does not touch the
hot allocation path, only the table.

**Falsification, on both sides.** With the defect in place (the iteration walks the 256
slots ascending) `peerorder_test.go` fails on its first assertion: *"GetPeers
walked [1 2 3], want the join order [2 1 3]"*. With the change in: **the five
committed captures of `internal/sim`**, which is the byte-for-byte comparison against the
original, 770.4 s and MUST-MATCH 0.

### 40.3 PORTS: THE INVENTORY, AND THE TWO PLACES THAT CANNOT USE `:0`

The brief for this work said that `internal/server` had a `freePort` in the
old shape. **It does not**: `393d528` (the same commit that fixed the harness) removed
it from the whole package. The complete inventory is now written in the header of
`internal/server/ephemeralports_test.go`, and it is four places that bind and **hold**
(`internal/net/tcp.go:63`, the `bs` one, the `ws` one at `ws_listener.go:176`, and the proxy's
at `cmd/proxy/tcp_proxy.go:389`, and each one reports what it took), a test that binds a
literal (port 1, and says why: nothing that asks the system for a port can
receive it) and **two that cannot use `:0`**:

| site | why it cannot |
|--|--|
| `cmd/proxy` `TestWaitForPortGivesUp` | it needs a port where **nobody** is listening; asking for it and holding it would make `waitForPort` come back immediately |
| `cmd/proxy` `TestClientPortIsTheConfiguredOneWhenThereIsOne` | its assertion IS that a **configured** port is the one left listening; asking the system would compare the system's answer with itself |

Both pick a number with a `:0` bind and release it, which is the shape that already cost a
whole capture. Neither can eliminate the window, so what is left for them is to
close it the only other way that exists: **check the port and retry with a
written cap**. That is `takeUnusedPort` (`cmd/proxy/process_manager_test.go`): it asks the
candidate for a number and only returns it when a connect to it is **REFUSED**. A
port whose "nobody is listening" cannot be established is not returned: it is another attempt,
and eight exhausted candidates is a failure that says so.

The second site also changes one more thing: it used to do `t.Skipf` when `Bind` did not
get the port. A `Skip` is the same fact reported as "this test did not
run", and in a bug that shows up as a timing flake a skip is the
worst of the outcomes because it is not visible. Now the miss is counted and, if it happens eight
times, it fails.

**Falsification, deterministic and without waiting for the machine to have a bad streak**:
`TestTakeUnusedPortRefusesAPortInUse` gives the helper a port that is **really in use**
(a listener the test itself has open) as the first candidate, and a
free one as the second. With the defect in place (returning the candidate without checking) the
file fails with all three assertions:

```
takeUnusedPort refused 0 candidate(s), want 1: the port 13018 was being listened on
takeUnusedPort returned 13018, want the candidate that was free (13019)
takeUnusedPort returned a port after refusing every candidate
```

### 40.4 `TestCroppedCaptureStillAnswersTheSame`, AND WITH IT THE `TestSim_synthetic` FLAKE

**About the trim**, which is what was asked, and here is the short answer: the
cause was never established and **still is**. What has been done is read
the path, and reading it produced one concrete thing and one non-concrete one.

**The one that is not touched (the close).** `replayConn` closes the connection at `start+end` and
**waits for the reader afterwards** (`replay.go:604-609`). A close discards what is in the
receive buffer unread, and the wait that comes after cannot recover anything:
it starts when the socket is already closed. The real margin is one simulation
period minus how far behind the reader is running, so the reader has
to be unable to run for more than 125 ms. With 24 burners that order of magnitude is
possible; **it has not been measured and is not going to be measured on purpose** (this machine has been at
100% CPU with an IDE alongside, and reproducing a flake with the scheduler already
busy is not a measurement).

**The one that is fixed (the classification).** The three signals of "the harness did not collect
everything" were three **notes** on a connection that was then compared as if it were
whole:

```
server->client: the reader did not finish within 2s of the connection closing
server->client: N trailing bytes are an incomplete frame at the end of the replay
server->client: <what scan could not frame>
```

It is the same fact, and it is the fact `unfinishedReplay` (393d528) was written to
refuse: "the server answered nothing" and "the harness stopped collecting" look the same in the
connections table. That commit closed it on the side of **driving** the connection and left
it open on the side of **collecting**. And there is an assertion that says so without having to
look at the code: `startObserveStatus` is must-match also when Ours is shorter
than Their (*"the client is blocked on this byte"*, `compare.go:609-624`), that is, a
truncated end was being reported as a server divergence that does not exist.

Now all three are a replay error, through the same `unfinishedReplay`. No
assertion is touched: nothing is relaxed, something that is not known to be whole stops being
compared. **Measured before touching anything**: the seven capture replays of the full suite do not
produce any of the three (zero occurrences in the 770.9 s run with `-v`), so this does not move any
verdict that exists today.

**Falsification**: `truncatedstream_test.go` stands up a real socket (the
server is not needed: what is tested is the harness's *collection*) that answers with two
whole RLink frames and then four bytes of a third that is cut in half through the
body. With the defect in place the test fails with *"replayConn accepted a stream whose
last frame stopped mid-body and returned 1 message(s) and no error"*, and with the defect
out both sides pass: the truncated stream is rejected with the reason, and the stream of
whole frames (the case the seven captures are in) is accepted.

**And the big finding, which was not in the brief: the observer step.**

`TestSim_synthetic` failed in one full-suite run with a must-match, and the
cause is neither clock nor ports. It has been reproduced (1 in 60 replays of
`testdata/synthetic`, with instrumentation in `tickLocked` and in `updateObserversLocked`) and
**it is not new**: it reproduces the same in `ed68bca`, the commit before the four of this
session, with the same must-match (1 in 60 on the parent, 2 in 120 at the head).

The mechanism, which is written in `updateObserversLocked`'s comment:

- The default interval is **8 TICKS**, not 8 ms (`internal/server/ticks.go`, with
  `main_1400d1310:299` and `:737`, and the 200 frames per second that the capture
  `2026-09-29_012528` measured on the original).
- The turn starts at the counter and advances interval by interval, and the counter
  only wraps when it is **strictly** greater than the interval. With a single
  observer (index 0) and interval 8, index 0 is visited on the ticks whose
  counter is 0, 9, 18...
- The counter advances **also on a tick in which nobody was visited**, and the tick
  in which the observer is registered but not yet active is one of those.

Instrumented, the failing run is exactly this:

```
observers 1 pageCount  8  0:active=false   <- registered, the Activate has not entered
observers 1 pageCount  9  0:active=true    <- and from here the counter goes 1, 2, 3
observers 1 pageCount 16  0:active=true    <- so index 0 is never reached
```

and the harness is left with the `0x19` and the `0x1E` and no frame, which is what
`replay-page-order` says as must-match.

That the peers-table change has nothing to do with it is checked with the same
instrumentation: `peers 1` in that fixture, and then `v5` (which with the change reads as
`len(peerOrder) == 0` instead of `len(Peers) == 0`) is **false before and after**.

**What is not touched, and it is the decision**: the step. Whether the original serves the
freshly activated observer in the same tick or whether it waits for its turn **is not
established here**, and changing it so that a fixture is deterministic would be guessing the
original, which is what rule 1 does not allow. And the synthetic fixture, which is what
uncovered this, is left as it is: enlarging it (more seconds of recording after the
activation, or a one-tick interval) removes the flake without removing the coverage, but it is
a change in `testdata` that needs its own measurement and its own chapter.

What can be said, and it is the only thing that is asserted: `TestCroppedCaptureStillAnswersTheSame`
and `TestSim_synthetic` **share the shape** (one observer, the 8-tick interval, and a
one-second replay window after the activation), so the first one's flake has a
candidate mechanism here that can be checked at any time.
**It is not claimed to be its mechanism**: the failure described here is from another test and with
another signature.

### 40.5 WHAT WAS MEASURED, AND WITH WHAT LOAD

**Warning that goes first**: this session found the machine at **100% CPU** with
an unrelated process (an IDE with a 3.5 GB working set) during the first half, and at
**15-38%** at the end. The clock figures of the two stretches **are not comparable with each
other**; the RAM ones are, because they are sums of working sets.

| | 39.8 (loaded machine) | this section | note |
|--|--|--|--|
| `go test -short ./... -count=1` | 97.3 s / 117.4 MB | **43.3 / 44.1 / 46.7 / 58.6 s** over four runs / **113.2-117.4 MB** | the spread IS the load: 43 s with the machine quiet, 58.6 s in the last one |
| `go test ./... -count=1 -timeout 25m` | 791.5 s / 116.7 MB | **770.4 s / 115.7 MB** | `internal/sim` 765.1 s |
| `go test -race ./internal/server/ ./internal/session/ -count=1` | 157 s | **38.7 s + 52.5 s, 0 races** | the 39.8 measures server+session+sim in 790 s |
| `go run ./cmd/gensim -check internal/sim/testdata` | ok | **ok (exit 0)** | |

**The sampler, and why this time it did measure**: the filter is the 38.5's (each binary by
**its own executable path**, a temporary `go-build` one, read with
`Process.MainModule.FileName`, which is `OpenProcess` + `QueryFullImageName` and **not**
WMI), and the sampling runs **in this shell** while `go test` goes in a
`Start-Job`. The other two forms were measured and failed, and all three fail **without
warning**, which is the worst thing a sampler can do:

| form | result | why |
|--|--|--|
| `Start-Job` with the filter inside | **0 MB**, 103 samples, 0 name matches | a job sees the processes, but the array of names does not cross the job's serialisation: it arrives as `System.Collections.ArrayList`, and the comparison is against that string |
| `Start-Process pwsh` with the sampler inside | **0 MB** | the child sees **one** process, its own |
| the package names as they are (`sim`) | **0 MB**, 143 matches, 0 summed | Windows strips `.exe` but **keeps** `.test` (`36.6` partly): the name is `sim.test` |

All three are 36.6 and 39.8 repeated: a filter that matches nothing reports a
zero as confidently as the thing it measures.
---

## 41. Splitting the journal into files by subject (2026-10-04)

This section is the **record of the moves**. Its only purpose is that nobody has to
guess what moved where, or believe something was lost.

### 41.1 WHY IT WAS SPLIT, AND WHY `SESSION-JOURNAL.md` IS STILL ALIVE

The journal had reached **43 sections, 428 KB and 6,686 lines**, and it was doing two
opposite things at once: it was a **chronological** record that was also the **current design** of every
part of the system. That made a new session find nothing without walking 6,686
lines, and it made the current design duplicated inside a work round.

The split separates the two by their nature:

- **The current design** is **living files**, one per subject, updated when
  the design changes and read whole.
- **The evidence** —every measurement, every falsification, every binary address— stays
  here, in the place and on the date it was produced. A measurement that lives only in a
  design file is lost the moment the design changes, and with it the reason for a
  decision.

**The deletion criterion has been: never delete a measurement.** When a section became
superseded, what was done was **replace it with a pointer** that says where it is now and why
it moved, and the original text stayed in the git history. In a single exception
bytes were deleted without a replacement: the four superseded markdowns of the root, which were not deleted,
and the five scripts of the dead capture format, which were.

### 41.2 WHAT WAS CREATED

| file | what it carries | lines |
|---|---|---|
| [`INTENTIONAL-DIVERGENCES.md`](INTENTIONAL-DIVERGENCES.md) | everything we depart from the original **on purpose**, with the reason and the evidence | 370 |
| [`UNINTENTIONAL-DIVERGENCES.md`](UNINTENTIONAL-DIVERGENCES.md) | open bugs, unexplained gaps and the `UNKNOWN`s that remain | 321 |
| [`PENDING.md`](PENDING.md) | what is missing, split by who can resolve it | 184 |
| [`CLOSED-QUESTIONS.md`](CLOSED-QUESTIONS.md) | the two questions the user closed, with their evidence | 67 |
| [`PERFORMANCE.md`](PERFORMANCE.md) | CPU, RAM, allocations, goroutines, methodology and numbers | 523 |

The two divergence files are **mutually exclusive** and the separation criterion is only
one, written in the header of both: **if there is a reason to accept the difference, it is
intentional; if not, it is a bug.** A "do not touch" with no written reason does not count as a
reason, so the `LobbyConnect` with the same bytes and the ping where the other answers with a
pong — which §25 listed as "the only thing left different, and not touched" — are in
the **un**intentional file, because they have no written reason.

### 41.3 WHAT MOVED FROM HERE, AND WHERE IT IS

| section from here | what it was | where it is now |
|---|---|---|
| Index, "WHERE EVERYTHING IS" | 73 topic -> section rows | the index's "By subject" and "By section", plus the rows that point at the living file |
| Index, the two closed questions | the summary | [`CLOSED-QUESTIONS.md`](CLOSED-QUESTIONS.md) |
| §9 Pending | the August list | [`PENDING.md`](PENDING.md) |
| §11 Analysis scripts | an inventory of `Temp` files | a "what replaced it" table |
| §12 Existing documentation | the inventory of 15 markdowns | `AGENTS.md`'s table, corrected |
| §16.1 Decisions taken | the register with reasons | [`INTENTIONAL-DIVERGENCES.md`](INTENTIONAL-DIVERGENCES.md) |
| §16.2 and §16.2b, the replay upload | the doubt and the "we do not change it" | the doubt to the **un**intentional (1); the "we do not change it" to the **intentional** (4.2) |
| §16.4 The divergence table | ~40 rows with their state | the two divergence files |
| §16.5 What has not been touched | the register of exclusions | [`INTENTIONAL-DIVERGENCES.md`](INTENTIONAL-DIVERGENCES.md) |
| §16.6 The next capture | a runbook | `CAPTURES.md` (done) |
| §25 "The only thing left different" | three transport divergences | [`UNINTENTIONAL-DIVERGENCES.md`](UNINTENTIONAL-DIVERGENCES.md) 4, and **§44.3** for the "ping", which was never one |
| §27 "What is left from this session" | the extra LobbyRelays | [`UNINTENTIONAL-DIVERGENCES.md`](UNINTENTIONAL-DIVERGENCES.md) 3 |
| §29 "What is NOT known" | the second peer's question | [`UNINTENTIONAL-DIVERGENCES.md`](UNINTENTIONAL-DIVERGENCES.md) 2 |
| §30 "What cannot be established" | the reconnection and the `peerIdentifier` | [`UNINTENTIONAL-DIVERGENCES.md`](UNINTENTIONAL-DIVERGENCES.md) 7 and 8 |
| §32.4 "Verified-different" | four unimplemented differences | [`INTENTIONAL-DIVERGENCES.md`](INTENTIONAL-DIVERGENCES.md) 2.1, 2.2, 3.1, 3.2 |
| §32.5 and §32.8, the `UNKNOWN`s | eleven fields, nine with a verdict | **§44.1**: nine left the register, only `conn+88` stayed ([`UNINTENTIONAL-DIVERGENCES.md`](UNINTENTIONAL-DIVERGENCES.md) 10.2) |
| §33 "What is left undone" | the cap of 17 chunks and the three differences | [`INTENTIONAL-DIVERGENCES.md`](INTENTIONAL-DIVERGENCES.md) 2.2 and 4.1 |
| §34 to §40, the cost measurements | all of the performance | [`PERFORMANCE.md`](PERFORMANCE.md), **without losing a single figure**: the original measurements stay here with their context |
| §36.4 The documentation audit | the verdict on 15 markdowns | `AGENTS.md`'s table, and §36.4 intact here |
| §37 "A flake that is not from this round" | the 1-in-8-to-10 flake | **§44.4**: closed, and it was an assertion that needed a count of ticks, not a tolerance |
| §40.4 "The observer step" | the flake's mechanism | **§44.4**: the settle was shorter than the step's period |
| The file's header | "a document that persists EVERYTHING" | "this is history; the design lives in living files" |

### 41.4 WHAT WAS DELETED, AND WHAT REPLACED IT

No measurement. Everything deleted was **dead format** or **dead procedure**.

| what was deleted | what replaced it |
|---|---|
| `scripts/csvtool.py`, `scripts/dump.py`, `scripts/parse_logs.py`, `decode_proxy_log.py` | the `internal/sim` harness. All four read `main_log_*.csv`, the layout the proxy stopped writing |
| `scripts/OBSERVER_CAPTURE.md` | `CAPTURES.md` + the harness. It was the runbook of a capture **that was already made**, and it described a format that is not written and environment variables that the proxy **deliberately ignores** |
| `scripts/bctest/` | nothing. A 25-line UDP probe with its own `go.mod` |
| `bsmunin_00001` | nothing. 0 bytes; the `.gitignore`'s `bsmunin_*` rule stays |
| the root `metadata.json`, **stops being tracked** | `src/`. It is 67 MB of decompiler output (`total_functions: 7389`), and the 7,389 files it wrote are in `src/`, import thunks included. It stays on disk |

### 41.5 WHAT IS KEPT ON PURPOSE, EVEN IF IT LOOKS DEAD

| what | why |
|---|---|
| `battle-server.txt` (47 KB) | the **original's** console output from 2026-08-11. It is the only runtime measurement of its animation loop (`Socks::AnimateAll: timeoutMS=1300`, 78 times), which is the accepted cause of the ack-cadence residual. The literal is in `src/`; that the server **runs** at 1300 ms is not |
| `headers.hpp` at the root | it is included by **7,450 lines of 7,389 files of `src/`** with `#include "../headers.hpp"`. It is a live include of the decompilation. Moving it into `src/` would mean touching the 7,389, and `src/` is not touched |
| `scripts/compare_observer.py` | its CSV CLI is dead, but the **module** is imported by `scripts/export_vectors.py`, which is the step documented in `CAPTURES.md`. And it already reads the new layout (`load_main`, `load_secondary`) |
| `scripts/csvlib.py` | it is imported by `compare_observer.py` **and** by `scripts/fidelity.py`, which is a parent's change that rule 5 forbids touching. Deleting it would break the import |
| `uploadVia-mysql.md`, `uploadVia-odbc.md`, `PROXY-CAPTURE-VALIDATION.md`, `WEBSOCKET-BSPORT-COMPARISON.md` | all four are linked from `README.md`, which is the user's work and is not touched, and from `webSocketPort.md` and the three lobby files. Deleting them would leave broken links in other people's files. They have been given **a line at the top** saying they are superseded, why and what replaces them: the second of the two options §36.4 recommends |
| `go1.27.md` | a copy of the Go 1.27 release notes. It is not a project document, but §35 and §39.5 cite it for the leak profile |

### 41.6 WHAT WAS NOT DONE, AND WHY

- **No section of the journal was deleted.** They were replaced by pointers, which is
  what leaves the trace. The only content change is the header and the index.
- **The body was not reordered.** Chronological order is what makes "the 34 before
  the 38" mean something.
- **`src/` and `src_go` were not touched.** `src_go` is a submodule with its own commits and
  is at `200dfc8`, clean.
- **`README.md`, `relay.md`, `cloc-2.10.exe`, `gsa.exe`, `scripts/fidelity.py`
  and the deletion of `build/BattleServerReimpl.exe` and `docs/architecture_refactor.md` were not touched.** They are
  the parent's changes (rule 5).
- **`captures/` was not touched.** It is 92.6 MB in one folder, already ignored by `.gitignore`, and
  with **zero** tracked files. What matters from there are the trimmed copies in
  `src_go/internal/sim/testdata/`, which are committed.

---
---

## 42. The divergence criterion, tightened to the black box (2026-10-04)

The 41 split the two divergence files with only **one** criterion: *"if there is a reason
to accept the difference, it is intentional; if not, it is a bug"*. That criterion falls
short in one case, and the case is big: **it treats any implementation
difference as a divergence**, and an implementation difference is invisible from outside. The user
put it like this:

> **Understand ONLY as a divergence when the functionality changes from the users' point of
> view** (not that it uses a pool or a data structure), that is, **as a black box**.

That is, two halves and both have to be applied: **(1)** a difference is a divergence if a
client, a peer or the replay host can tell it apart **on the wire**; **(2)** if it is visible from
outside, then you ask whether there is a written reason. What only changes inside is a
design decision, and it lives in [`PERFORMANCE.md`](PERFORMANCE.md) if it is about cost and here if it is
about why it was decided. Section 7 of
[`INTENTIONAL-DIVERGENCES.md`](INTENTIONAL-DIVERGENCES.md) lists what left by that.

### 42.1 WHAT LEFT THE DIVERGENCE REGISTER

Fourteen points. Three of them had a **false claim**, and the rest are internal
decisions or paths with no producer. Section 7 says which is which; here is how they
were checked:

| what was retired | how it was checked that it is not visible |
|---|---|
| The `0x08` to the observer | `sub_1400E2590:360` builds with `MultiGameDataMessage__vmethod_18`, which serialises with **type 10**, and `:362-371` walks the **peer slot vector**, not the observer list. The four class vmethod_18s in `src/` give 9, 10, 23 and 34: **there is no type-8 class at all** |
| The `DropoutPeer` gate (`*(u16*)(peer+68) < 3`) | `sub_1400DAAF0:141` requires `v63 == 8` and the switch at `:133-140` only has `2, 3, 5, 6, 7, 9`: dead code. And `peer+68` is **not** a transition counter: it is the `LobbyConnect`'s `protocolVersion` (`EB5F0:470` and `:491`, with the field read in `LobbyConnectMessage__vmethod_20:64-70`) |
| On a connection dying its observer entry is not removed | the original **does** remove it: `sub_1400E52D0:227-242`. The intentional file's 5.4 denied it and its 5.5 said it right |
| `Session.Peers` is still a map | it has been `peerSlots [256]*Peer` since the 40.2's commit, and every iteration goes through `peerOrder` |
| The `titleID` is accepted even when it does not match | the lookup **does** use the pair: `GetSession(SessionKey{TitleID, SessionID})`, which is what `sub_1400EC060:58-77` compares |
| The observer step | here it is done **the same as the original**, with the same arithmetic (`E4EC0:199`, `:206-209`). What is not established is whether the original serves the freshly activated one in the same tick, and that is a question, not a divergence: to the unintentional file, as an `UNKNOWN` |
| The rest (peer header, endianness, `maxreplay`, relay host-only, the poolings, the observer list emptied, the `E2590:483` byte) | either they are memory, or they are paths with no producer, or they are not sent, or only the munin reads them |

### 42.2 THE CITATIONS THAT DID NOT SAY WHAT THEY SAID

| where | what it said | what it says |
|---|---|---|
| intentional 4.1 | the cap of 17 chunks is `sub_1400E34F0:124-231` | it is `sub_1400E2590:314-320` (the call with the flag and the sixteen without it). `E34F0:124-231` is the copy of one chunk, and what it does bound is the body to `< 0xF7FF` and the sub list to `< 0xFF` (`:152` and `:155`) |
| intentional 5.1 | the peers map is still untouched | changed in the 40.2, with its before/after table |
| intentional 5.4 | the original does not remove the entry when the connection dies | it does, `E52D0:227-242` |
| intentional 4.4 | `maxClientDosAttackCount` is consumed in `sub_1400D8E10:1146` | that line decrements a counter in the TLS's **filename table** (`strrchr("D:\\a\\1\\s\\Server.cpp", 92)` at `:1112`) |
| unintentional 1 | the sessionID lives in `session+2096` | it lives in `session+2104`; `session+2096` is the titleID (`E1DE0:33` copies both from the key, `E9E90:289-290` prints them in that order) |
| unintentional 2, and `AGENTS.md` | `sub_1400EA680:193` | `:259`; and `sub_1400E5670`'s only caller is `sub_1400EAE60:29`, and `EAE60` is only called from `DAAF0:298` |
| `AGENTS.md` | the observer stream "7204 messages against 7208" | it is from a **WebSocket** capture (§27), not from the bsPort one: **the bsPort one has no observer**, and its two connections are two players' |
| `AGENTS.md` | the four bytes of the `0x19` are four rejections | the `0` is the **acceptance**: it is the only arm that registers the observer (`E3B50:237-254`) |
| `AGENTS.md` | `2026-09-28_205830` has ours at the front | it has the **original** at the front: `metadata.json`, `primary.writer = original` |
| `AGENTS.md` | the harness has been seen against four captures | against **five** |
| `PERFORMANCE.md` 3 and 4 | the figures of the per-tick cost and the profile | those of 3 were from the 38.2 and those of 4 are the **before** of the two representation changes (the 39.3 and the 40.2). Both have been measured again: `TestTickCostAtEveryPeerCount` gives 1,066 ns at one peer and 15,931 at 32, with a slope of **492 ns/peer**, and the extrapolation to 2 h is 0.142 s at four players and 0.232 s at eight |
| `PERFORMANCE.md` 5 | "the four caps" of the long match | there are **seven**, plus an eighth in the residue test. The first four do match the code (4 MiB, 8 B/tick, 32 MiB, 24 per tick) |
| `PERFORMANCE.md` | the probe comes out 78 times | `Socks::AnimateAll` comes out **78** times, of which **70** carry `timeoutMS=1300`; the other 8 are `sendwait` and `shutdown request` |
| `CAPTURES.md` | 012528 with 22 informative, 010531 with 15, 183120 with 15, 205830 with no number | the 2026-10-04 run gives **25**, **16**, **19** and **12**. The informative number is not a constant: it grows with the number of informative rules the harness has |

### 42.3 WHAT WAS ADDED

One new point in the unintentional file: **a top-level `0x01` is handled here and the
original does not have it**. The original's dispatcher has no case for `0x01` (its labels
are `0x09, 0x0F, 0x10, 0x13, 0x14, 0x17, 0x18/0x1C, 0x1A, 0x1B, 0x1D, 0x1F`, plus the `0x15`
that is handled before the switch, `DE050:136-154`), and a type outside the list is only
logged (`DE050:948`, `"!!!Unknown message type %d received!!!"`). We create the session,
register a peer and emit an `AddPeer`. No capture sends it (the game sends it
inside a `0x13`, compressed), so **there is no written reason** for handling it: by
the criterion it goes in the unintentional file, and the decision to remove it or to write the reason is
in section 3 of [`PENDING.md`](PENDING.md).

### 42.4 WHAT WAS VERIFIED, AND HOW

- **The four bytes of the `0x19`**, with their arm: `3` at `:93`/`:113` (store already cycled, **does not**
  register), `0` at `:170`/`:219` (registers, `:237-254`), `2` at `:295`/`:344` (list full),
  `4` at `:403`/`:452` (observers deactivated for the session).
- **The `0x1E`**: `sub_1400E43E0:54` sets type 30, `:55` the status byte (0), `:118` writes it,
  `:142` the 8 bytes of the time, and `:56-65` are the two time branches (lagged and
  not lagged).
- **The store's formula**: `E2590:221` (vector length), `:235` (floor of 8), `:241`
  (`ceilf(peers * f)`, with `f` in `session+2168`), `D4E10:18-19` (blockSize `session+2152`,
  maxBlocks), `main_1400d1310:346` (f default 1.0), `:347` (`-pageSize` 0x10000), `:617`
  (`-maxPagesToStorePerSession / 8`), `:818` (the bits go down to the store).
- **The two per-peer caps of the upload queue**: `sub_1400E2E00:53-54`, `>= 0x12C` entries
  and `>= 0x80000` bytes. The codes have the same figures.
- **The observer step**: `E4EC0:54` (the cursor), `:57` and `:203` (the turn), `:199`
  (the advance, which is outside the `if` of the active entry at `:62`), `:206-209` (the
  wrap, only if `v2 + 1 > interval`), and the default of 8 at `main_1400d1310:298` with the
  two paths of `:737` (the flag conversion) and `:749` (without the flag, as it is).
- **The pending list**: `E34F0:86-119`, and the cut at `:96-98` is the bottom of the loop, not
  a `break` (the loop ends at `:325`).
- **The version that decides the tail block**: `E2590:233-234` caches it, `:301` copies it to
  the message's version field, `vmethod_18:146` emits the block if it is `> 1`. And **that real
  clients announce 5** is measured in the committed captures: the four
  `LobbyConnect`s of `2026-09-30_201122` and the echoes of `2026-09-30_010531` carry `05 00` at
  offset 17 of the body.
- **The original's dispatcher** has no case for `0x01`, and `sub_1400DF920:136-153` maps the
  relay's verdict (`1` to everyone, `2` host only, and nobody produces the `2`).
- **The suite**: `go build`, `go vet`, `gensim -check` and all of `internal/sim` (764.7 s, the
  seven captures green) on 2026-10-04, plus `TestTickCost*` for the cost figures.

### 42.5 WHAT WAS NOT TOUCHED

Neither `src/` nor `src_go`, nor a line of code. The `src_go` submodule is still at `200dfc8` and
clean, and this session's changes are **documentation only**: the two divergence files,
`PERFORMANCE.md`, `CAPTURES.md`, `PENDING.md`, `AGENTS.md` and the rows of
this index that pointed at section numbers that have changed.

## 43. The documentation files get their names in English

This is a **rename and nothing else**. The documentation was translated into English
(the 41 and the 42 are the two halves of that work) and the file names were left in
Spanish, so the tree had two languages in it: English inside every sentence and
Spanish outside every file. A reader who arrives at the index by name and finds
`DIVERGENCIAS-NO-INTENCIONALES.md` cannot tell from the name what the file is, and a
name has nothing else to do.

**The ten renames.** `git mv` for all of them, so the history of each file follows it:

| was | is |
|---|---|
| `DIVERGENCIAS-INTENCIONALES.md` | `INTENTIONAL-DIVERGENCES.md` |
| `DIVERGENCIAS-NO-INTENCIONALES.md` | `UNINTENTIONAL-DIVERGENCES.md` |
| `PENDIENTES.md` | `PENDING.md` |
| `PREGUNTAS-CERRADAS.md` | `CLOSED-QUESTIONS.md` |
| `RENDIMIENTO.md` | `PERFORMANCE.md` |
| `SESSION-JOURNAL.md-SESION.md` | `SESSION-JOURNAL.md` (this file) |
| `mensajes.md` | `MESSAGES.md` |
| `protocolo-lobby.md` | `LOBBY-PROTOCOL.md` |
| `validacion-capturas-proxy.md` | `PROXY-CAPTURE-VALIDATION.md` |
| `webSocketBsPortsComparison.md` | `WEBSOCKET-BSPORT-COMPARISON.md` |
| `uploadVia-lobby-conexion.md` | `uploadVia-lobby-connection.md` (**§46**) |
| `uploadVia-lobby-estado.md` | `uploadVia-lobby-state.md` (**§46**) |

The last one of the ten is the only name that is **not** the literal translation, and the
reason is that the original name was not a word: `webSocketBsPorts` is neither
`webSocketPort` nor `bsPorts`, it is the two port names run together.
`WEBSOCKET-BSPORT-COMPARISON.md` names the two documents it compares, which is what a
reader looks for.

> **Corrected on 2026-10-05, by §46. The table above was ten rows and it should have been
> twelve.** The criterion the 43 set for itself — a file name that is a Spanish word while
> its prose is English — was applied to the *documents* and not to the two `uploadVia-*`
> files, on the grounds recorded below ("the five `uploadVia-*` files carry the vocabulary
> of the protocol"). That reasoning is right about the stem and wrong about the suffix:
> `uploadVia-lobby` is the protocol vocabulary and stays, and `conexion` / `estado` are
> Spanish words describing what the file is about, exactly like `protocolo-lobby.md` and
> `mensajes.md` in the same table. The two names are now English, and **the check that
> missed them is committed** (`scripts/check_language.py`), because "we translated
> everything" is not verifiable from the tree and a name is the first thing a reader sees.

**What was NOT renamed, and why it must not be.** `bsPort.md`, `webSocketPort.md`,
`oob.md`, `replays.md` and the five `uploadVia-*` files carry **the vocabulary of the
protocol**: the flag is `-bsPort`, the class is `RLink::SessionMessage`, the method is
`uploadVia=lobby`. A name that is the thing itself can be searched against the binary,
against a flag line and against a capture, and a name that is a translation of it cannot.
`go1.27.md` is a copy of a third-party document and keeps its upstream name.

**Every reference moved with the file, in the same commit.** Markdown links, bare names
in tables, the index of `AGENTS.md` and this index, the `§N` citations into this journal
and the two `uploadVia-*` files that cite it. A commit that renames a file and leaves a
link pointing at the old name is a commit nobody can bisect to, and the whole point of
keeping the original text in the history is to be able to go back to a working tree.

**What the rename does not touch.** The `§N` cross-references, because they are by number
and not by name: there are 88 distinct ones across the root markdown and 87 of them
resolve to a `## N.` heading here. The one that does not is **§5**, and it is a hole that predates
this rename: the index lists "5. Root cause of the crash, the fix and its regression
test" and the text is there, as `### Cause`, `### Fix` and `### Regression test` **under
§4**. There is no `## 5.` heading. It is left as it is because fixing it is an edit to
the journal's own structure and not a reference, and because a `§5` that points at
content that exists is not a broken link.

> **Corrected on 2026-10-05, by §45.1.** The 88/87 count above was measured wrong, and the
> decision to leave the hole was reversed. Measured again, document by document: **95**
> distinct `§N` numbers across the root markdown, **91** of them resolving. The **39
> distinct bare `§N`** references resolve **38**, and the one that does not is `§5`, so the
> conclusion above was right about `§5` and right that it was the only bare one. The count
> was not: a reference is not necessarily to *this* file, and four of the ones that look
> dangling resolve into another document or are not journal references at all. The index
> row is now corrected, and **§45.1** carries why the index was corrected instead of a `##
> 5.` being created.

**No content was rewritten.** Not one sentence of prose changed: a file name inside prose
is a reference and moved, and everything else stayed.

## 44. The criterion applied to every entry, and what it moved (2026-10-05)

The 42 tightened the divergence criterion to the black box: **a divergence is a difference a
client, a peer or the replay host can tell apart on the wire**. It tightened it and then did
not apply it to the entries that were already in the register. This round is that
application, entry by entry, and the entries that fail it leave.

### 44.1 10.1 and 10.3 were never divergences

The user named them: `session+2140` and `session+2176`/`+2224`/`+2232`/`+2240` are listed in
[`UNINTENTIONAL-DIVERGENCES.md`](UNINTENTIONAL-DIVERGENCES.md) as open divergences, and the
same file says they never reach the wire. **A finding that says of itself that no consumer can
tell it apart is not a finding of this register**: it is a fact about the inside of the
process. They were filed under the pre-42 criterion, which only asked "is there a written
reason to accept it", and a field with no reader trivially has no reason to accept it.

The check, done against the binary rather than against the claim:

| field | readers | on the wire? |
|---|---|---|
| `session+2140` | `sub_1400E7600`/`sub_1400E7610`, called **only** from `sub_14010F4F0:99` and `:273`, which is the replay uploader's "already published" latch | no |
| `session+2176` | the munin itself (`E9E90:153` writes it, `:66-71` walks the sessions) | no |
| `session+2224` | the munin; written by the tick when it serialises the `0x0A` (`E2590:361`) | no |
| `session+2232` | the munin; written by the `LobbyConnect` handler (`EB5F0:136`) | no |
| `session+2240` | the munin; written by the observer frame-resend path, keeping the **widest span anyone asked for** (`E3260:32-36`) | no |

The two readers of all five are `sub_1400E9E90` (the munin) and `sub_1400EBE60` (the state
dump), and **neither of the two produces a protocol message**. That is the whole argument,
and it is the same argument as for the fourteen points the 42 retired.

**`conn+88`, the one that stays, is the control.** It is in the same audit and in the same
section, and it survives because the **dispatcher's own gate reads it**
(`sub_1400DE050:136`: `if (*(_BYTE *)(a3 + 88) == 0) { if (v128 != 21) return 0; ... }`). Its
value decides **which message types that connection is answered to at all**, so a client can
tell the two implementations apart: the original says nothing, we ack. Two fields in one
audit section, one divergence and one not, and the difference between them is precisely
whether a byte the client receives depends on it.

**Where the knowledge went, so that it is not lost with the classification.** Each field,
what writes it and what it costs is now in [`PERFORMANCE.md`](PERFORMANCE.md) 5, "The
counters that reach nothing". The audit record stays where it was made, §32.5 and §32.8. And
the two questions that are still open — the **name** of the metric behind `+2140`, and the
name of `+2240` — stay on the work list in [`PENDING.md`](PENDING.md) 2.1 and 2.3, retitled
there so that nobody reads them as wire differences.
### 44.2 A top-level 0x01: the handler is gone and 0x01 is an unknown message type

The 42.3 added the finding; this is the fix, and it is the first change of protocol
behaviour in this round.

**The binary, read case by case.** `sub_1400DE050` reads the type with
`sub_1400EF1F0(a2, &v128)` at `:135` and dispatches on it at `:266`. The labels, in order:
`:268` `9u`, `:366` `0xFu`, `:442` `0x10u`, `:544` `0x13u`, `:695` `0x14u`, `:727` `0x17u`,
`:771` `0x18u` with `:772` `0x1Cu`, `:810` `0x1Au`, `:838` `0x1Bu`, `:866` `0x1Du`, `:920`
`0x1Fu`, and `:948` `default`. **There is no `0x01`**, and the `default` arm between `:948`
and `:1026` is the TLS statics dance and one call:

```
sub_140099DE0(&qword_1402F35C8, "!!!Unknown message type %d received!!!", v128);
```

then `return 0`. It logs and closes nothing.

**And the one function that would have done it is unreachable from there.**
`sub_1400EA680` — the REQUESTJOIN case that looks the session up or creates it and calls
`sub_1400E5B00` for the `session+2080` tree — has exactly **one** caller in the decompiled
set: `sub_1400DAAF0:193`, the type-2 case of the internal peer switch (labels `2, 3, 5, 6,
7, 9`, `:133-140`). `sub_1400DAAF0` is reached from `sub_1400DE050:399` (the `0x0F` case) and
from `sub_1400DF920:136` (the relay). Both carry a **peer** message. So a top-level `0x01`
has no path to a session, a peer or an `AddPeer`.

**The negative measurement is not the argument.** Finding `_7` recorded that no `01...`
frame was ever seen on the wire. That says the path is not exercised; it does not say that
answering it is right, and under rule 1 it cannot be the reason to keep a handler whose only
producer does not exist.

**The falsification**, and it is the interesting part of the commit:
`TestTopLevelRequestJoinIsAnUnknownMessageType`. Written before the change, it fails with

```
a top-level 0x01 created 1 session(s): the original's dispatcher has no case for it and reaches nothing
```

and it also asserts the two things a lazy fix would get wrong: the connection stays **open**
(the original's `default` returns 0 and closes nothing) and the client is not marked an
observer. `TestTopLevelRequestJoinIsNotOneOfTheGateTypes` closes the other way in: `0x01` is
not one of the two types that pass the guards before the switch (`conn+88`, which admits
only `0x15`, and the observer bitmask at `:250`), so removing the handler cannot be worked
around by another gate.

**What it cost.** Twenty tests used `handleRequestJoin` as a scenario fixture, which is
the only reason it existed. They now call `testSessionWithPeer`, a **test fixture** that
creates the session and the peer slot and deliberately does **not** broadcast the
`AddPeer`: the original's internal `REQUESTJOIN` emits none either (`sub_1400EA680` returns
after `sub_1400E5B00`, and there is no `AddPeerMessage` construction on that path).
`wsJoinHost`, which created a session over a real WebSocket connection by sending a `0x01`,
now sends a `LobbyConnect` — the path the game actually takes (`sub_1400EB5F0`, into
`GetOrCreateSession`), which makes those two integration tests closer to a real client than
what they were.

**Not one byte of the five committed captures changes**: no capture contains a top-level
`0x01`, which is the same fact as before and is why the replays pass unchanged. That is
recorded rather than assumed — the full suite was run.
### 44.3 The "WebSocket ping 0x65 against the pong 0x6B" was never a ping

`SESSION-JOURNAL.md §25` listed three things left different and not touched: the ack grouping, the
`LobbyConnect` carrying the same bytes, and "a ping where the other answers with a pong".
The 42 retired nothing of those three, because each one is wire-visible. This is the
resolution of the third, and it is a **false claim**, not a fix.

**Where the two bytes come from.** The original's `101 Switching Protocols` answer arrives
**split in two TCP writes**, 86 bytes and then 70 — measured on
`2026-09-30_010531_a553ef8fd63c`. The harness used to throw the upgrade response away
entirely and read the stream from the first byte, so the second record was handed to the
framer as if it were a message and **its first byte was reported as a message type**. That
byte is `0x65` in our record and `0x6B` in the original's:

| record | our side | the original |
|---|---|---|
| 1st | `...Connection: Upgrade\r\nSec-WebSock` | `...Sec-WebSocket-Accept: ICX+Yqv66` |
| 2nd | **`65`**`et-Accept: ICX+Yqv66kxgM0FcWaLWlFLwTAI=\r\n\r\n` | **`6B`**`xgM0FcWaLWlFLwTAI=\r\nServer: WebSocket++/0.8.2\r\nUpgrade: websocket\r\n\r\n` |

`0x65` is the **`e`** of `Sec-WebSocket-Accept` and `0x6B` is the **`k`** of the accept key.
They are HTTP header continuation bytes of the same `101`. As a WebSocket frame `0x65` would
be FIN + opcode 0x5 (ping) and `0x6B` would be FIN + opcode 0xB (pong), which is why the
claim looked like a ping; neither frame was ever sent.

**The control measurement**, over the five committed captures, both sides, every logged row,
first byte in `0x80..0x8A` and not a binary frame:

| capture | main.log (truth) | secondary.log |
|---|---|---|
| `2026-09-28_205830` (bsPort) | `0x88` x2 | `0x88` x2 |
| `2026-09-29_012528` (bsPort) | `0x88` x4 | `0x88` x2 |
| `2026-09-30_010531` (ws) | none | none |
| `2026-09-30_183120` (ws) | none | none |
| `2026-09-30_201122` (bsPort) | none | none |

**No `0x89` and no `0x8A` anywhere, on either side of any capture.** The only close frames
are on the two bsPort captures.

**And the binary agrees.** `WebSocketManager::SockSendPing` exists
(`sub_1400A5D80`, with its own log lines at `:209`, `:266`, `:292`, `:338`, `:519`), but
**there is no call to it anywhere in the decompiled set**: its only caller is `unk`, which is
not among the decompiled sources. The path exists in the binary and nothing in it is
exercised by a session this server has ever seen.

**What the misreading was hiding, and it was real.** Our `101` carried **no `Server`
header**; the original's carries `Server: WebSocket++/0.8.2`. That is now sent
(`internal/server/ws_listener.go`, `ResponseHeader`) and the **whole upgrade response is
compared** by `ruleUpgradeResponse`: status line, header set, and each side's
`Sec-WebSocket-Accept` checked against the derivation of the key **that side's** request
carried. The order still differs and is reported as informative, because RFC 6455 gives the
response no ordering requirement and the client is websocketpp, which collects headers into
a map.

**The falsification** is `TestUpgradeResponseCatchesTheMissingServerHeader`: it feeds the
reader the two real responses and fails if the rule says nothing.
`TestUpgradeResponsePassesOnceTheHeaderIsThere` is the other direction, and
`TestUpgradeResponseDoesNotCompareTheTwoAcceptStrings` plus
`TestUpgradeResponseCatchesAWrongDerivation` keep the rule from failing on correctness
(the first version of that file compared the two accept strings, which are the correct
derivations of two different keys, and produced two must-match divergences out of it).
`TestUpgradeHeadersReadsSplitRecords` pins the two-record shape that caused the whole thing.

### 44.4 The observer step: answered from the binary, and the two flakes with it

#### The question, and the answer

`UNINTENTIONAL-DIVERGENCES.md` 6 and `PENDING.md` 1.7 asked: **does the original serve a
freshly activated observer in the same tick it activates, or does it wait for its turn at the
step?** It has been open since the 40.4 and it blocked `SESSION-JOURNAL.md §11.1`. It is answered, and
the answer is that **it waits, and no arm in the binary could do otherwise**:

- `sub_1400EC580`, the `0x1A` case, is nine lines and does one thing: `v2 = a2[5]; if (v2)
  sub_1400E43E0(v2, a2)`. It passes the **connection** straight through.
- `sub_1400E43E0:39-48` walks the observer list at `session+2272/2280` comparing
  `**(_QWORD **)v4 != *a2`; `:49` writes `entry+12 = 1`; `:50-69` computes the time offset;
  `:54` and `:116-142` build and send **one** message, the `ActivateObserverResponse` whose
  wire type is 30 (`0x1E`). That is the whole function.
- The only two functions in the binary that fetch frames out of the store and write them to
  an observer are `sub_1400D5560` (fetch) and `sub_1400D4A70` (write). **Neither is called
  from `sub_1400E43E0`.** Their only callers are `sub_1400E4EC0:79` and `:83` — the tick's
  observer step, which `sub_1400E2590:97` calls — and `sub_1400E3260:43`, the reconnect
  resend.
- Inside the step nothing distinguishes a new arrival: the body opens at `:62` with
  `if (*(BYTE *)(v4 + 12) != 0)`, the cursor picks the entries (`:57`, `:199`, `:203`), and
  the cursor advances on every tick whether it visited anybody or not.

#### What the step costs, measured

The cursor at `session+2296` takes `interval+1` values — `:206-209` wraps only when it is
**strictly** greater than the interval — and index 0 is served on the ticks where it is 0.
So a freshly activated observer waits **up to `interval+1` ticks: 9 ticks, 1,125 ms** at the
defaults. The **worst phase is 1 and not `interval`**, which nobody had worked out: from
`interval` the cursor wraps to 0 on the very next tick, and from 1 it has to climb all the
way round. Two new tests walk every phase (`interval+1` of them) and measure the bound:
`TestAFreshlyActivatedObserverIsServedWithinOneStepPeriod` and
`TestTheWorstPhaseCostsTheWholePeriod`, in `internal/session`.

#### What that cost, and it was not the fixture's fault

The replay's settle was **three simulation periods, floored at 400 ms** — less than a third
of the step's own period. The settle's job is to keep reading after the last captured input
"so the session ticks have delivered everything they owe an observer", and what it was not
covering is exactly that. On the phases where the first visit came after 400 ms the replay
**closed the connection before the frame was written**, and `replay-page-order` came out
**must-match with no frame in it**: the signature of 11.1, and, with a different signature,
of 11.2.

**Enlarging the fixture would not have fixed it, and that is worth saying out loud** because
it is the remedy 11.1 proposed. The settle is always derived from the last **whole second**
of the recording (`settleForTruth`, `cmd/proxy` stamps whole seconds), so a fixture of any
length keeps landing on the same 1-second tail, and the phase coin-flip stays. The defect is
in the settle, and the fix is in the settle.

The floor is now `interval+2` ticks — the step's period plus the tick the write happens on —
read from the server's own `ObserverUpdateIntervalTicks`, which is **exported for exactly
this** so the number the harness waits for and the number the step walks with cannot drift
apart. The report carries the arithmetic as a `settle window` deviation, and
`TestTheSettleIsExplainedInTheReport` reads it back off a real 48-second replay. **A settle
is not a tolerance on an assertion**: it is how much of the run gets recorded, and widening
one would have hidden the defect instead of closing it.

#### 11.2, and why it was not reproduced

`SESSION-JOURNAL.md §37` wrote down "once every 8 to 10 runs of `internal/sim`, not investigated, looks
like clock under machine load". The hypothesis was **right**, and the reason is more precise
than "load". The assertion that failed was *"the player connection's extras are interleaved"*,
and that needs **more of our messages before the window than the original had**. Our server
emits one ack per simulation tick; the fixture's original recorded **three** for that second;
so the assertion needs at least **five of our eight 125 ms ticks to land inside the first
wall-clock second**. That is how many ticks the host's scheduler managed. It is not a
property of the protocol, and it is not a boundary that a tolerance could absorb — it is a
count of ticks.

It was **not** reproduced on purpose. Loading the machine to make a 1-in-9 flake appear is
what a previous attempt paid for, and a flake reproduced that way teaches nothing new. It was
settled by reading the assertion and the arithmetic behind it, and it is **proven** by the
two falsifications, which fail on the defect and not on the host:

| defect reintroduced | what the test says |
|---|---|
| the cursor wraps at `>=` instead of `>` | both new session tests fail, on the bound |
| `ActivateObserver` sends one frame | `phase 0: ActivateObserver served 1 frame(s); the original's handler builds one message, the 0x1E` |
| the settle floor removed | `settle = 400ms, but an observer activated at the end of the recording waits up to 9 ticks (1.125s) for the step's first visit and the write happens on that tick` |
| `SettleOnly` inverted | `the observer connection (port 51002) reached "interleaved", want "settle"` |
| the totals line prints one count too few | `the totals line says 4 of ours before the window against the original's 4, which is 0 extras before it, and the split line says 4` |

The tree was restored after each one.

**What the assertion is now**, which is stronger and deterministic: the **observer**
connection is settle-only (structural: its whole handshake is stamped at the instant the
window opens, and the longer settle now guarantees the frames are collected), and for
**both** connections the verdict word, the two counts in the verdict sentence, the counts in
the totals line and the "the split is exact" line must be **four views of the same
numbers**. That is what "a report whose own lines disagree is worse than no report" means
as a test. Both conclusions stay covered end to end **without a replay** by
`TestCompareConnReportsTheTimes` (interleaved) and `TestMessageTailIsNotAMustMatch`
(settle), and at the unit level by `TestJudgeTailSettlesOnly` and
`TestJudgeTailInterleaved`.

#### 8. The observer pairing, settled from the binary

The other stale entry. `sub_1400E43E0:43` and `sub_1400E46C0:25` are **the same one line**,
`while (**(_QWORD **)v4 != *a2)`, and what it compares is the connection's own u64 at offset
0 — which `sub_1400E3B50:240` writes from the request's third u64
(`*(_QWORD *)a2 = *(_QWORD *)(a3 + 64)`). So **the first entry whose connection carries the
caller's identifier wins**, activation and removal alike, and with two observers sharing an
identifier the second one arms the first entry and a `StopObserve` from it removes the first
one. The entry's `+0` is the connection pointer, so the entry that gets armed and the
connection that gets the answer are two different objects in that case. That is already what
`internal/session/observeridentity_test.go` does
(`TestSameIdentifierArmsTheFirstEntryAndAnswersTheCaller`,
`TestStopObserveRemovesTheFirstEntryOfASharedIdentifier`) and it is what the comment on
`Session.ActivateObserver` says. **The entry was open in the register because nobody wrote
the binary reading down; the behaviour was settled in the 31.**

#### The full criterion audit, and where the twelve entries went

Section 0 of `UNINTENTIONAL-DIVERGENCES.md` is the table. In short: **four stay blocked on
gameplay** (1, 2, 3, 4), **four are gaps with no capture** (7, 10.2, and the two "the
original's own question is unanswered" cases folded into them), **five were closed this
round** (5, 6, 8, 9, and 10.1/10.3 as non-divergences), **two harness entries moved to
where harness limits belong** (11 to the register's own record and `PENDING.md` 5, 12 to
`CAPTURES.md`). The other divergence file needed no move: every entry in it is wire-visible
or is a "not a difference" note, and each carries a written reason.

#### What the five committed replays did

`MUST-MATCH 0` on all five, `internal/sim` 772,2 s, whole suite green. The settle grew by
850 ms on a capture whose own span is seconds, which is inside the suite's budget and is
the price of not reporting the harness's own truncation as a divergence of the server.
## 45. The names, the dangling `§5`, and the `conn+88` premise (2026-10-05)

Four things that had been left, and one that was measured and **not** changed because
changing it is not this round's to change.

### 45.1 `§5` does not exist, and the index is corrected instead of a heading being created

The 43 recorded that the index lists a section 5 and that no `## 5.` heading exists, and
it left it. **Reversed here, and the index row is what changed.**

**Measured first, because the 43's count was wrong.** Resolved document by document — a
`§N` in a living file is not necessarily a reference to *this* file. **Before** this change,
of the root markdown:

| | distinct | resolve | does not |
|---|---|---|---|
| bare `§N` | **39** | **38** | **1**, and it is `§5` |
| `§N.M` | **56** | 52 | 4 |
| all | **95** | **91** | 4 |

The 43 said 88 distinct and 87 resolving. It was right that `§5` is the only unresolvable
**bare** reference, and wrong about the totals: the counting treated every reference as a
reference to the journal. Of the four `§N.M` that look dangling, **one resolves into
another document** (`UNINTENTIONAL-DIVERGENCES.md` 11.1, which the journal's `:6830` and
`:7087` both mean), **one is not a journal reference at all** (the journal's own `§3.10` at
`:1329` cites **plan v3's** `§3.10`, a document not in this repository), and **two are real
holes** this round fixed: `§3.9b` in INTENTIONAL-DIVERGENCES.md (a **plan's** row, so it is
now named and not linked) and `§10.1` in PENDING.md, which pointed at a subsection of the
divergence register that **was deleted** the day before.

**The tool itself got it wrong first, and that is worth recording.** The first version of
the checker kept one variable across the loop instead of resetting it per reference, so a
reference that resolved left its verdict behind for the next one: it reported "39 of 39
resolve" on the very run that listed `§5` as unresolvable. **Both cannot be true and the
summary line was the wrong one.** Re-measured with the variable reset per reference and
with every living file's own numbering in scope.

**Why the index and not a `## 5.`.** Promoting `### Cause` / `### Fix` /
`### Regression test` to a real `§5` **would not have renumbered anything**: `6` to
`44` are already correct and nothing moves, so the danger that made renumbering
forbidden does not arise for the promoting option. It was still not done, for a reason that
only showed up on reading the region: `### Evidence` at `:461` is part of the same
argument as `### Cause` at `:467`, and the capture's own findings (`### Finding
_10`, `:448`) sit above them. A clean cut has to go **above** `### Evidence`, which
drags the `_10` capture out of `## 4.` — and the index cites `§4 to §15` as the
captures. So promoting means either splitting `Evidence` from `Cause` or moving `_10`,
and **both change what the section means**. Correcting one index row changes **no prose and
no structure**, and a reader who follows it lands on the right text by name:

> `| The _5 crash and its cause | §4, and inside it ### Cause / ### Fix / ### Regression test |`

**After the change, measured with the corrected checker**: **39 of 39 bare `§N` resolve, 0
missing** — `§5` included, because it now resolves to a section that says out loud that it
does not exist. `§N.M` is **60** distinct (this round added `45.1` to `45.4`) and **57**
resolve. The **three numbers** that do not resolve are each **named as what they are**, in
this paragraph and in the two living files, and are not broken links: `§3.10` is **plan
v3's** own subsection — the pre-existing site is the journal's `:1329`, which is where the
whole confusion started — `§3.9` is that plan's row, written as a name and not as a link,
and `§10.1` is the deleted subsection of the divergence register. The count of sites is not
quoted because this paragraph is one of them: **every `§N` that means "go and read that"
now resolves, and the ones that do not are the ones that mean "this does not exist".**

### 45.2 `conn+88`: the premise was a misreading of the nesting, and it is settled

UNINTENTIONAL-DIVERGENCES.md 10.2 asked **which connections set `conn+88` to 1**, and
said it would close with *"a real match that puts the case in the log, or the binary in the
nested region"*. The second half was available. **It is closed.**

The entry's premise: *"the only two writers are in the branches nested under 'the secret
already found a peer' (`if (v15 != 0)`, `EB5F0:131`), and that cannot be true for a first
connect"*. **The premise is false, and the decompiler is right.** Brace depth, which is the
only thing that settles it, from `sub_1400EB5F0_1400eb5f0.cpp`:

| line | depth | what it closes |
|---|---|---|
| `:124` `__wind` | 1 | opens the EH scope |
| `:127` `{` | 2→3 | **`if (v13 != 0)`** — opened at `:126` |
| `:132` `{` | 3→4 | `if (v15 != 0)` — opened at `:131` |
| `:134` `{` | 4→5 | `if (v16 != 0)` — opened at `:133` |
| `:462` `}` | 5→4 | closes `if (v16 != 0)` |
| `:464` `}` | 4→3 | **closes `if (v15 != 0)`** |
| `:465` `if (v16 == 0)` | 3 | a **sibling** of `if (v15 != 0)`, still inside `if (v13 != 0)` |
| `:485` `}` | 3→2 | **closes `if (v13 != 0)`** |
| `:486` `else if` | 2 | **the `else` of `if (v13 != 0)`** |

So the two writers are at **depth 4** (`:469`) and **depth 3** (`:490`), and **neither
is inside `if (v15 != 0)`**:

- **`:490` is under `v13 == 0`** — *no session in the tree at all*. That is the first
  connect, and it is the branch the entry said could not be true.
- **`:469` is under `v13 != 0 && v15 == 0`** — the session is known but **the secret has
  no peer yet**, which is also a first connect.
- Both require the `LobbyConnect`'s dword at `msg+80` to be **0**. `v16` is
  `*(_DWORD *)(a3 + 80)` (`:130`), so `:465`'s `if (v16 == 0)` and `:486`'s
  `*(_DWORD *)(a3 + 80) == 0` are **the same test**; a non-zero one is a reconnect, and
  `:506` names what is left over: `"Attempt to reconnect to a non-existent sessionID"`,
  reached when `v13 == 0 && msg+80 != 0`.
- **`v13` is the session, and that is positive, not an assumption** — the one thing this
  reading needed and did not have at first. The lookup at `:87-111` searches a tree on the
  composite key **(`*(_WORD *)(a3 + 48)`, `*(_QWORD *)(a3 + 56)`)**, which are the
  `LobbyConnectMessage`'s **`titleID` and `sessionID`** (`MESSAGES.md:149-150`), and
  yields `v11[6]` — or **0** when the key is not there (`:110`). And `:506`'s string,
  reached exactly when it is 0, names the same thing.

**There is no third writer, and zero-initialisation is not the answer.** Every byte-sized
store to `+88` in the **7,450** files of `src/` is **three lines**: `EB5F0:469` and
`:490` (both `= 1`) and `sub_140233400:61` (`= 0`). That third one is **not a connection**:
`sub_1400CD030(v3, 8, (const char *)(a2 + 88))` at its `:46` passes `a2 + 88` as the
**`const char *`** of a `std::string`, whose capacity sits at `+112` (`:47-48`, the
`>= 0x10` heap test) and is set to **15**, the small-string marker, at `:60` — it is the
`std::terminate` helper that carries `"termination_handler call failed"` at `:31`. A
zero-initialised object is **0**, not 1, and the only two writes of 1 in the binary are the
two above, so the field cannot be 1 without one of them running.

**The gate, read whole** (`sub_1400DE050:135-247`, brace depth `d1` for the `if` at
`:136`): read the message type; **if `conn+88 == 0` and the type is not 21, return 0**;
otherwise build the `LobbyConnect` and call `sub_1400EB5F0` (`:185`), which is what
sets the byte; then `:247` closes the `if` and `:248` is **the next statement, not an
`else`**, so the `LobbyConnect` falls through to the normal dispatch as well. `21` is
`LobbyConnect` — `0x15` = 21 decimal, `MESSAGES.md:60`. **The field means "this
connection has completed its `LobbyConnect`"**, and the dispatcher drops everything else
until it has.

**One more thing the entry had wrong.** It cited `sub_1400EB1E0:33-34` as a second reader
using it as a slot index. There `a3` is the **message**, not the connection: `a2` is the
connection, because its `+8` is what `:44` prints as `peer=`, and `a3 + 88` is the
**removing peer's id**, used as `session[24 + 8*id]` — the peer slot table, the same shape
as `sub_1400E5670:100`. Same offset, different object. It read as a second confirmation of
a layout and it is a coincidence.

**Why it left the register.** Reaching a difference needs a connection whose **first**
message is not a `LobbyConnect`. Every connection of every committed capture sends its
`0x15` first — **the two observers of `012528` and the observer of `010531` and
`183120` included**, each sending `0x15`, then `0x1C`, then `0x1A` (`trimsim
-inspect`, which prints the first message per connection). So the drop path is never taken,
no byte of any capture is compared differently, and **nothing is implemented and nothing
needs to be**. Implementing the gate *wrongly* is what would have rejected those observers,
which is what the entry warned about; the warning was right and its reason was not.

**What the binary could not close, and it is not this question**: the **reconnect** branch
(`EB5F0:133-463`) **never writes `conn+88`**, and that branch is the one that swaps the
peer onto the new connection (`:137-139`: `v18 = *(v15 + 24); *(v15 + 24) = *(a2 + 24);
*(a2 + 24) = v18;`). So a reconnect can only work on a connection that already carries the
byte. That is PENDING.md 2.2 and it needs a real reconnection, like 1.5.

### 45.3 The six literals, and the seventh

Six test failure messages in `src_go` still said `SESSION-JOURNAL.md-SESION.md`, the file's name
before the 43, in a string that names the evidence for an assertion:
`internal/server/goroutineleak_test.go`, `internal/server/serverprofile_test.go`,
`internal/session/peersweep_test.go`, `tickcost_test.go` and `tickprofile_test.go`
(twice). A failure message is read by whoever the assertion fails, and it was pointing them
at a file that has not existed since the 43.

Verified afterwards that **nothing in `src_go` names the old file any more**, and that the
one surviving mention in the root markdown is the 43's own rename record at `:6600`, which
is **supposed** to name it: it is the row that says what it was renamed from. The seventh is
the one the list did not have: `internal/sim/replay.go:416`, a **comment** this time, not a
string, reading `SESSION-JOURNAL.md §11.1` — and it was pointing at
`UNINTENTIONAL-DIVERGENCES.md` 11.1, so the citation was right and the file name in front
of it was not.

### 45.4 The suite is over budget, and it is measured and not changed

The full suite is **over the 10 minutes** of AGENTS.md's rule 7 and the options are in
[PERFORMANCE.md](PERFORMANCE.md) 9, with the figures. **Nothing about what the suite runs
was changed**: which tests execute is the user's decision and this round did not take it.
What the measurement settled is that **the saving is not where it looks**: the five
committed captures have **no trim slack** (`trimsim -inspect`: the observer of `010531`
connects at `t=4m2s` of a 4m9s capture, the second attempt of `012528` at `t=3m13s` of
3m19s, the observer of `183120` at `t=1m16s` of 1m29s, and `201122` is pinned by its
first `0x0A` at 47 s; `205830` has traffic in every 10 s bucket to the end), so the
678 s of recorded clock cannot be shortened without dropping compared messages. **The
capture on disk that is not imported would add, not save**: trimmed to 1 m it is about
61 s, and it is 24m44s of one connection doing the same thing throughout.

## 46. The English sweep finished (2026-10-05)

The standing instruction is **file names and content, all in English, everywhere**. The 43
renamed the documents and left **two behind**, the `src_go` tree was never swept at all, and
the numeric-integrity checker the translation was verified with (`numcheck.py`) was written
**outside the tree**, so nothing about it survives for the next person to re-run. This section
is the three of those closed, and it is closed with a **committed tool** rather than with a
claim.

### 46.1 The two names the 43 missed

| was | is |
|---|---|
| `uploadVia-lobby-conexion.md` | `uploadVia-lobby-connection.md` |
| `uploadVia-lobby-estado.md` | `uploadVia-lobby-state.md` |

`git mv`, so the history of each file follows it, and **every reference in the same commit**:
`AGENTS.md` (the inventory), `README.md` (the index, two rows), `replays.md` (five mentions,
two of them inside a markdown link) and `uploadVia-lobby-state.md`'s own cross-link to its
sibling. The journal's own index row moved with it. **There are no references in `scripts/` or
in `src_go/`: both were grepped and both are empty**, which is the answer and not an assumption.

**Zero-hit proof, and what the surviving mentions are.** Grepping both old names over the whole
tree gives **4 hits, and all four are rename records**: the `was` column of the two rows of §43's
table above and the two rows of the table in 46.1. That is the same shape as the
`SESSION-JOURNAL.md-SESION.md` case in §45.3 — a rename record is **supposed** to name what it renamed
from, and a reader who greps for it should land on the row that explains it. Outside those four
rows the count is **zero**, and `src_go/` and `scripts/` are two of the places that count as zero
because they were checked, not because they were forgotten.

**The rule the 43 wrote and then broke.** Its own table says a name is not renamed when it
carries *the vocabulary of the protocol* — `-bsPort`, `RLink::SessionMessage`, `uploadVia=lobby`.
That is the right rule and it does not cover these two: the stem `uploadVia-lobby` **is** the
protocol vocabulary and stays, and `conexion` / `estado` are **Spanish words describing the
contents**, in the same position as `protocolo-lobby.md` and `mensajes.md`, which were renamed.
A suffix that names what is inside the file is not a protocol token, and the fact that it sat
next to one that is did not make it one.

### 46.2 The Spanish in `src_go`, and why 246 lines and not 177

**246 flagged lines in 30 files, measured by the tool of 46.3 over `src_go` at `c51bd29`:**
202 `SPANISH` and 44 `AMBIGUOUS`. The brief's **177 lines in 8 files** is a count of
Spanish TOKENS, and two things follow from counting tokens instead of prose.

**It is smaller than the truth.** `internal/loadgen/report.go` was on the list with 7 lines,
and the 7 are the Go **identifier** `el`, short for `elapsed`, which appears 10 times in
that file and is not Spanish. A word-level tool cannot tell an identifier from prose.

**And it is in the wrong place.** 22 of the files it missed are the ones whose Spanish
is a single content word on a line with no function word: the two `ORIGEN.txt` fixtures,
the four `metadata.json` trim records, and six comments that cite **this journal by a
section title that was still written in Spanish** -- section 22's own heading, and the
one on capture identity. That last group is the `SESSION-JOURNAL.md-SESION.md` defect of §45.3
again, on the same kind of literal: a comment pointing at a file or a section by a name
it no longer has. §45.3 fixed seven of them and the citation titles were the other half
of the same hole.

**Translated, because it is the same prose in the same tree:** the comments of `.gitignore`,
`scripts/compare_observer.py` (its `--selftest` still passes) and `scripts/deploy.ps1`
(which still parses with 0 errors). Neither script is on rule 5's list. A checker that
reports a file we were told not to touch is a checker whose output gets read past, so
the two lines in `fidelity.py` are exempted with rule 5 as the written reason instead.

**That no code moved is measured, not asserted.** Both revisions of every changed `.go`
file are reduced to their CODE (comment text deleted, string-literal contents blanked,
everything else verbatim), tokenised whitespace-insensitively and compared: **25 files
changed, code identical in 25 of 25.** No expression, no assertion threshold, no number.
The number of string-concatenation operands in each literal is also unchanged, which is
not automatic when prose is re-wrapped: two `t.Logf`/`t.Fatal` literals were first
rewrapped into four operands instead of three and the check caught it.

**Two mistakes, both caught by measuring instead of reading.** `primera` and `segunda`, the
two fixture capture names in `cmd/gensim`'s `TestRunAgainstOneRootKeepsTheOtherTests`, are
Spanish and were renamed, and **the test then failed on its own hard-coded expectation**:
`sim_primera_test.go`. The assertion moved with the fixture and still asserts the same
property. That is what a rename inside a test costs and the only way to see it is to run
the test.

#### WHAT WAS MEASURED

With the sampler filtering each test binary by its **parent command line**, which it prints
because a previous sampler reported a confident 0 MB three different ways:

| | time | peak working set |
|--|--|--|
| `go build`, `go vet`, `gensim -check` | clean | - |
| `go test -short ./... -count=1` | **51.3 s** | **112.6 MB** |
| `go test ./... -count=1 -timeout 25m -v` | **777.9 s** | **116.0 MB** |
| `go test -race server+session+sim+proxy+gensim` | **801.4 s** | **482.4 MB** |

The machine was at **42%** before the short loop and had **7.7 GB of 32.7 GB** free. The
short loop is within noise of the 45's 49 s; the full suite is against its 775.6 s and the
45's own note that a machine at 54-71% makes the full suite's clock incomparable applies
to nothing here: this one ran on a quieter machine and came out **2.3 s slower**, which is
inside the run-to-run spread and is stated as such rather than as a movement.

**The five committed capture replays are MUST-MATCH 0, byte for byte.** Measured on this
revision, from each report in the full run:

| capture | MUST-MATCH |
|--|--|
| `2026-09-28_205830_314309565d6e` | **0 divergence(s)** |
| `2026-09-29_012528_03cb67c7cf19` | **0 divergence(s)** |
| `2026-09-30_010531_a553ef8fd63c` | **0 divergence(s)** |
| `2026-09-30_183120_c4e7c284ff1c` | **0 divergence(s)** |
| `2026-09-30_201122_c4e7c284ff1c` | **0 divergence(s)** |

**WHAT WAS DELIBERATELY LEFT, WITH THE REASON**

| left | why |
|--|--|
| `cmd/gensim/main.go`'s `return "captura"` | the fallback of `ident()`. It lands in the **generated** test's function name, so renaming it changes generated output. That is a behaviour change and this is a language sweep. Dead in practice: the proxy always names a capture `<date>_<time>_<commit>`. |
| `scripts/fidelity.py` lines 94-95 | **rule 5**: the parent's file, named by the rule that protects it. Two Spanish strings. |
| `SESSION-JOURNAL.md`'s `"No se ha dado ninguna frase contrasena"` | the original's own error string, quoted verbatim. Translating it would make the document stop describing the binary. |
| `SESSION-JOURNAL.md`'s `Solo se permite un uso de cada direccion de socket` | Go's own runtime error, inside a fenced block: a test's real output. |
| `SESSION-JOURNAL.md:1876`'s `` an exporter `captura -> testdata` `` | a name quoted from the round that proposed the exporter. The exporter that exists is `cmd/gensim`; rewriting the history to say otherwise would be a worse lie. |
| `internal/loadgen/report.go`'s `el` | a Go identifier, and the brief says identifiers are not touched. |
| the `0x19`, `0x1E`, `0x0A`, `sub_1400E2590:233` forms | the same in both languages. A checker that flags them is a checker that gets switched off. |

### 46.3 `scripts/check_language.py`, and why a claim needs a tool

The previous round verified its own translation with `numcheck.py` and **left the file
outside the tree**, in a temp directory. Nothing about it survives, so the next person
cannot re-run it, cannot extend it, and cannot tell whether the numbers in its report were
reproducible. A check that cannot be repeated is not a check; it is a story.

**What it checks.** Every remaining Spanish occurrence in the documentation and in the
reimplementation, **with file and line**: the root markdown, `scripts/`, `.gitignore`, and
`src_go/` (Go comments and string literals, JSON and text test data). **The file NAMES too**,
because that is the check that would have caught the two of 46.1 and it was not there when
they were missed. Not checked, on purpose: `src/` (7,450 decompilation files, not our
content), `captures/` and `build/` (gitignored artefacts), and `go1.27.md` (a copy of an
upstream release note).

```
python scripts/check_language.py            # the whole tree
python scripts/check_language.py --strict    # AMBIGUOUS also fails
python scripts/check_language.py --quiet    # counts only
python scripts/check_language.py --selftest # detection + false positives
python scripts/check_language.py --list-allow
```

Exit 0 clean, 1 when `SPANISH` was found, 2 when only `AMBIGUOUS` was. It enumerates
**git-tracked files per repository**, and descends into `src_go` as its own repository:
the alternative, walking the filesystem, sweeps in the gitignored 67 MB `metadata.json`
and buries the 20 real hits under 7,000.

#### WHY IT WORKS ON LINES AND NOT ON WORDS

Spanish and English share a lot: `no`, `a`, `si`, `so`, `an`, `one`, `final`, `normal`,
`solo`, `nuevo`, `dato`, `byte`. A checker keyed on isolated words is either silent or
noisy on English, and a noisy checker is ignored, and an ignored checker is worse than
none. So it scores a **whole line** by how many Spanish **function words** it carries, and
the words it scores are the ones with **no English homograph at all**: `de`, `la`, `el`,
`que`, `con`, `por`, `para`, `del`, `un`, `una`, `se`, `es`, `son`. A Spanish sentence has
several in a row; an English sentence has none. The words that ARE shared carry **zero
weight** and are listed in the source, so a reader can see they were excluded on purpose
and not forgotten.

It flags **prose**, not code: not fenced blocks, not inline code spans, not link targets,
not URLs, not hex literals, not file paths, not identifiers. In a `.go` file the prose is
the comments and the string-literal contents, and getting that right needed a real scanner
(rather than a regex per line) because two things went wrong first and both made the tool
lie: a `//` comment leaked its line-comment state into **every line after it**, so a
variable named `el` was reported as the Spanish article, and the string scanner read the
**opening** quote as the closing one, so the test failure messages it exists to find were
invisible. Both were found by running it against the tree, not by reading it.

#### THE FALSE-POSITIVE RATE, MEASURED ON 3,322 LINES OF ENGLISH

| corpus | prose lines | SPANISH | AMBIGUOUS | flagged |
|--|--|--|--|--|
| the 20 documents translated in the 41/42 | **3,322** | **0** | **0** | **0 = 0.000%** |
| `src_go` before this sweep, byte-exact `c51bd29` | 8,383 | 201 | 44 | 2.92% |
| `src_go` after it | 8,383 | **0** | **0** | **0** |
| the whole tree as it stands | - | **0** | **0** | **0** |

The English corpus is the 20 files the previous round left in English, **minus**
`SESSION-JOURNAL.md-SESION.md`: that one is a **stale Spanish copy** in the temp directory (the journal
was translated later and the working file was never replaced), and including it would make
"the English corpus" a measurement of nothing. Every one of the false positives that
existed during development is now a rule in the source with its reason:

| it said | why it was wrong |
|--|--|
| `Suspicion of f2e956e` | correct English, 9 letters, caught by the `-cion` pattern. **`-cion` is now supporting evidence and never a verdict on its own**, and the English `-cion` words are listed beside the pattern. |
| `the **un**intentional file` | English derivational use of `un`. `un` now carries 1.0 instead of 2.0 and does not count as a function word. |
| `AoE2DE` -> `DE` | an all-caps token is an acronym, not an article. Capitalised prose (`El`, `Las`) still counts. |
| `Option C`, `C++ exception` | a one-letter token is never a Spanish content word. |
| `un-inflated` | an English compound: a function word followed by `-` is skipped. |
| `/ Component / C++ Class /` | a table header; fixed by the two rules above together. |

`--selftest` is the false-positive and detection test, **committed and runnable**: 7 known
Spanish lines that must be caught, 9 known English lines that must not be, and a
consistency check over the word lists themselves, which is how `solo`, `ha` and `son` were
found sitting in two lists at once.

#### IT IS FALSIFIABLE, AND HERE IS THE FALSIFICATION

A checker that reports zero on known-Spanish input is worse than none. Against a **byte-exact
snapshot of `src_go` at `c51bd29`**, written outside the tree with `git show`, it reports:

```
files scanned: 175SPANISH: 201   AMBIGUOUS: 44     (exit 1)
crosscheck_test.go:221: SPANISH [13.0 f=5 conexion,de,hay,mas,ninguna,una]
    t.Fatalf("una conexion de cliente y ninguna mas, hay %d", len(cap.Conns))
observeridentity_test.go:5: SPANISH [11.0 f=4 el,la,lista,por,se]
    // EL IDENTIFICADOR DEL OBSERVADOR: la lista se ordena por el peerIdentifier, no
```

and against the same tree as it now stands, **0 and 0** with `--strict`. The difference is
the whole claim, and it is one command either way.

**Detection is 66% of Spanish prose lines and 75% flagged at all** (6,035 `SPANISH` of 9,096
lines, 6,835 of 9,096 with `AMBIGUOUS` included). The 25% it does not flag are lines with
no Spanish function word in them at all: table rules, `---`, and short fragments. That is a
deliberate trade: a line with one Spanish word is reported as `AMBIGUOUS`, which is a
**report and not a verdict**, because guessing on one word is what produces the false
positives that get a checker switched off.

#### THE SIX EXEMPTIONS, AND WHY THERE IS NO SILENT SUPPRESSION PATH

`ALLOWED` in the source is a table of **(pattern, reason)** and nothing else: a line is
exempt only if it matches a pattern that has a written reason next to it, and
`--list-allow` prints them. There is no config file, no per-line number and no way to
quiet the tool without editing code a reviewer can see. Four of the six are the
original's own quoted log strings, `fidelity.py`'s two lines are one grouped pattern under
rule 5, and the sixth is gensim's generated identifier.

**And one file is skipped whole, which was found the hard way.** The commit that added the
checker made it report **itself**: 30 `SPANISH` in its own source, because its word lists
ARE Spanish and its `--selftest` fixtures are Spanish sentences it has to catch. That is
the worst failure a checker can have, because its verdict now depends on its vocabulary
and the obvious "fix" is to delete the detector. So `SKIP_WITH_REASON` is a second table,
right beside the first, and `--list-allow` prints both. One entry, `check_language.py`,

> this file: its word lists and self-test fixtures are the detector, and a checker must not
> report its own vocabulary

and the reason is the whole argument. **A checker that flags itself is not wrong about
Spanish, it is wrong about what it is for**, and writing the exemption down is what keeps
the next reader from "fixing" it the only way it looks fixable.

---

## 47. The Go source stops citing the decompiler (2026-10-06)

`DERIVATIVE-WORK-RISK.md` arrived as a review of `src_go`: a source-level assessment
concluding that the Go comments carried decompiler symbols, memory-offset narration,
embedded pseudocode and six log strings copied verbatim out of the original, and that
this makes a copyright argument easy to raise. **It does not conclude that the project
infringes**, and neither does anything written since. What it asks for is a decision
about what this repository says about itself and what may leave it.

The review understated the scale, and the first thing measured was the scale.

### 47.1 The number was 689, not "many"

`sub_1400` in `src_go`:

| | files | references |
|--|--|--|
| review's phrasing | "many comments in `internal/`" | — |
| **measured** | **48 of 144 `.go` files** | **689** |

And it was not only comments in `internal/server` and `internal/session` as the review
listed. `internal/session/session.go` alone carried **161**, `internal/server/handlers.go`
**69**, and the `_test.go` files — which the review did not list at all — carried **343
of the 689**, more than half. Six test files carried 20+ each.

**Why the number matters more than the review's estimate.** A reviewer who reads "many
comments cite the decompiler" and finds three of them forms an opinion. A reviewer who
counts 689 across 48 files, with more than half in tests, concludes that citing the
binary was this project's normal way of writing, and that is a different and much less
defensible position. The measurement is what turned this from a cleanup into a decision.

### 47.2 What was removed, and what was not

Everything named above, measured after the change:

| artifact | before | after |
|--|--|--|
| `sub_1400XXXXXXXX` / `main_1400XXXXXXXX` in `src_go/**/*.go` | 689 | **0** |
| recovered vtable slots (`PeerMessage__vmethod_18`) | 11 | **0** |
| embedded pseudocode (`*(_QWORD *)`, `goto LABEL_`, `vN =`, `__int64`) | 25 | **0** |
| offset narration (`session+2192`, `peer+0x5A`, `client+156`, `store+80`) | 191 | **0** |
| log strings verbatim from the binary | 6 | **0** |
| logger module/method labels copied from the original's class names | 64 | **0** |

**What was NOT removed**, and this is the half that matters:

- **The knowledge.** Every one of those 689 citations was attached to a real finding: a
  branch ordering that inverted what the code used to do, a status byte that is not what
  its number suggests, a stride that is a phase and not a rate. Deleting the comments
  would have deleted the project's own memory of why its code looks the way it does, and
  this project exists because of those findings.
- **The evidence trail elsewhere.** `SESSION-JOURNAL.md` keeps every `sub_...`
  citation, because it is the research record and its citations are what make its
  measurements checkable. `src/`, `battle-server.txt` and the captures are untouched.
- **One byte of behaviour.** Every edit was a comment, a log string, a test failure
  message or a logger label. No executable statement was touched anywhere, and the
  evidence is the suite: five capture replays, 770.6 s, green before and after.

The 689 became **0** and the behaviour did not move. That is the whole claim, and it is
falsifiable by diffing a commit: there is no hunk in this round that changes a statement.

### 47.3 The transformation is a rewrite, and the proof is in the cases where deleting would have been wrong

A citation is usually a parenthetical aside, and deleting it leaves the sentence whole.
**That is the easy case and it is the minority.** These were not:

```
// sub_1400E3B50:237-254 marks the connection observer (client+156), appends a
// 16-byte entry with progress 0, active 0 and the lagged flag
```
becomes
```
// The accepting arm marks the connection observer, appends an entry with
// progress 0, active 0 and the lagged flag
```

```
// v27 = session+2336;                        // the u64 the players published with 0x1D
// if (v27 == 0 || (v28 = msg+72 == v27, entry.lagged = 0, !v28))
//     entry.lagged = 1;
```
becomes
```
// lagged unless the request presents EXACTLY the u64 the session published with
// 0x1D: no published hash means lagged, a wrong hash means lagged, and only the
// matching one clears it
```

In the second case **deleting the pseudocode and nothing else would have left a sentence
with no predicate**. The rule the whole round was run under is therefore not "remove the
citation" but "**state the behaviour the citation was citing**", and the second case is
the one that shows why. 31 files were edited in `internal/session` and `internal/server`
alone; the three cases where a shortcut would have produced a wrong comment are recorded
here because they are what a reviewer should check first.

**And four comment blocks were deleted whole rather than rewritten**, because their
content *was* a table of recovered statements (`// sub_1400D4770:20 *(_QWORD *)a1 = 0;
// constructor`) and keeping the statement with the symbol gone would leave a line nobody
can act on.

### 47.4 Six log strings, and how they were found

The review said to "search for exact strings copied from the original console output and
review each one individually". That is done, and it is a two-command check rather than an
opinion:

```
# every string literal in src_go of 12+ chars, searched against the decompilation
$all = Get-ChildItem -Path src -Recurse -File -Include *.cpp,*.hpp | ForEach-Object { Get-Content -Raw $_ }
$j = $all -join "`n"
# ... then filter out %-verbs and hex
```

**Six matched**, and all six were the original's diagnostic wording reproduced as our own
log output: the unknown-message warning, the observer-kick notice, the store-refusal
notice, the sync-timeout reason, and two ordering-case messages. All six are now in this
project's own wording. What remains matching — `-maxPagesToStorePerSession`,
`Sec-WebSocket-Accept`, `WebSocket++/0.8.2`, `M#%08u Sack%08u` — is enumerated in
`INTENTIONAL-DIVERGENCES.md` §8 with the reason each one has to stay, because each is a
protocol or library fact rather than a developer diagnostic.

**The distinction that decided each one**, and it is the review's own: *a log string a
client reads is a compatibility fact; a distinctive developer diagnostic is not*. None of
these six travel on the wire. `syncDataTimeoutReason` is now a named constant in
`internal/session` **because** it is the only one of the six that a test asserts on, and
giving it a name made the tests read as what they are.

### 47.5 The decision, in a file: `PROVENANCE.md`

The review's Phase 1 is "identify who owns each original binary, decompilation, capture,
log and golden file, and decide what may be distributed". That is a decision, not a code
change, and it is now `PROVENANCE.md`: a **per-path** list, because a list by category is
a list somebody will argue with.

**And its first paragraph is a self-classification, which is the part with the most
weight and the part most easily got wrong**:

> This is an independent reimplementation built by observing a proprietary server's
> network traffic and its behaviour on the wire. It is not a clean-room implementation.

The review's "What should not be done" includes *"do not remove evidence of provenance and
then claim the implementation was clean-room"*, and that warning is correct and it
constrains this round as much as it guided it. Removing 689 citations from the source
while the repository said nothing about how it was built would have left exactly the
position the review warns about. So the repository now **says what it is**.

**And `PROVENANCE.md` is itself on the do-not-distribute list, which is the correction
this section records.** The first version of this round wrote the opposite: it said the
one thing that must survive any publication split is `PROVENANCE.md` itself, *because*
the file declares the derivation. **Those two positions cannot both hold.** A file whose
first paragraph is "this is not a clean-room implementation" is not a document to ship;
publishing it converts a defensible position — *an independent implementation, verified
against observed traffic* — into an acknowledged derivative-work claim, in the project's
own words.

So the distinction this round actually turns on is one the review never drew:

| | |
|---|---|
| **an honest INTERNAL record** of how the code was built | unambiguously right, and it is why the file exists: it is what stops the research from lying to itself |
| **a voluntary PUBLIC admission** of the derivation | a different act, it is not required by anything, and it is the one that creates the exposure |

**Only the first one is free.** The 689 citations still came out of `src_go` — that part
stands, and it stands on its own, because a source file full of decompiler addresses is a
worse thing to hand to anyone than a source file that states its behaviour. What changed
is where the admission lives: **internal, and it stays internal.**

Nothing is staged for distribution and nothing has been distributed. What a public tree
would need is a *different* statement — what the code implements, against which protocol,
and that it contains no third-party or original material — and that is a legal decision
this repository has no authority to make and no reviewer has been asked to make.

`DERIVATIVE-WORK-RISK.md` is kept as history and marked as such: its recommendations were
carried out, so reading it as a description of the tree's current state would be wrong.

### 47.6 `BEHAVIOUR-NOTES.md`: what a maintainer reads instead

The 689 citations were attached to findings. Removing them removed the project's memory of
why its code is shaped the way it is, which is not acceptable for a codebase whose whole
method is *"a measurement that lives only in a design file is lost the moment the design
changes"*.

So the knowledge moved, not into a git archaeology exercise but into a file written in
behaviour: 22 sections, no recovered symbol in any of them, each stating what the protocol
requires and why the code follows it. It is the distributable-facing summary;
`SESSION-JOURNAL.md` is still the record with the evidence.

**What did not go into it**, and that boundary is the point: which recovered function
wrote which offset. That is the material `PROVENANCE.md` says must not be redistributed.
What is in `BEHAVIOUR-NOTES.md` is the behaviour that structure produced, which is both
what a maintainer needs and what a client can observe.

### 47.7 What was deliberately not done, and the one cost worth naming

- **No protocol constant was changed.** `0x1E` came from the binary and it stays. Changing
  it would break interoperability and answer nothing about provenance.
- **No identifier was renamed.** `NewReplay`, `updateObserversLocked` and every other name
  are the project's own. The risk the review warns about — renaming while preserving the
  recovered implementation — does not apply to names the project chose for itself.
- **The research record was not thinned.** §47.2 says this and the reason is that
  `CLOSED-QUESTIONS.md` exists: two conclusions are closed on the strength of evidence, and
  evidence that is deleted cannot hold anything closed.

**The cost, named rather than discovered later.** 55 files changed in `src_go` and the diff
is **1,774 insertions against 1,997 deletions**, and it is almost all comments. `git blame`
on a behaviour line now points at a round that only reworded the comment above it. That is
real: the next reader of the history has to read one commit to learn that 1,000 deleted
lines were citations and not findings. The journal section and `BEHAVIOUR-NOTES.md` are
what pay that cost; they are why this round is a section and not a commit message.

**And the honest limit of the claim.** This changed what the distributed artifact says
about itself. It did not change how the implementation was built, and nothing here should
be read as saying it did. Whether that is sufficient is a legal question, about jurisdiction
and about rights this repository does not hold, and `PROVENANCE.md` says in its own text
that **a qualified lawyer must review this before publication**.

### 47.8 The verification, all of it after the change

```
$env:CGO_ENABLED=0; go build ./...                    exit 0
$env:CGO_ENABLED=0; go vet ./...                      exit 0
go test ./internal/sim/ -count=1 -timeout 25m         ok 770.589s
go test -race ./internal/server/ ./internal/session/  ok 39.580s / ok 55.658s
go test -short ./... -count=1                         16 packages ok
python scripts/check_language.py --strict             files scanned: 195  SPANISH: 0  AMBIGUOUS: 0
```

The four artifact counts in §47.2 were measured by regex over all 144 `.go` files, case
sensitive, and each is **0**. Four hits survived the first pass and all four were
`kilobytes` matching `LOBYTE`: the same class of false positive as §45.3, found here by
running the check instead of trusting the edit.

`gofmt -l` reports 34 of the 55 changed files, and **all 55 were already unclean at HEAD**
(CRLF in the working copy), so nothing here introduced a formatting change. That was
measured by checking the HEAD version of each file, not inferred.

---

## 48. Performance, measured against the original instead of only against us (2026-10-06)

`PERFORMANCE.md` had eight sections and every measurement in them was of **our** code: our
tick, our heap, our profile, our goroutines. That is the right way to work and it has a blind
spot that is worth writing down rather than fixing quietly: **nothing in that file says
whether the original spent more or less.** A Go tick made 42 % cheaper and a reader learns
nothing about the binary it is reimplementing.

The original binary is in the tree (`BattleServerOrig.exe.pristine`, untracked), and the first
thing to establish was whether it can be **launched and observed** at all, because a
comparison that cannot run is not a plan. It runs: with `-bsPort` and `-webSocketPort` on
free ports it binds both and prints its version line. That is all this section needed.

### 48.1 `benchpeer`, and why the tool is deliberately ignorant

`src_go/cmd/benchpeer`. It launches two binaries, applies **identical measurement rules** to
both, and reports what each costs **as a running process**: time from launch to the port
accepting a connection, idle working set / commit / handles, CPU burned while holding a port
with nobody on it, and the cost of one held connection.

**The design constraint that makes its output comparable is that it knows nothing.** It takes
a path, an argument list and a port. It does not know which of the two binaries is the
reimplementation, it does not read either one's internals, and the same code path produced
both columns. That is what lets it run against a binary this project has no source for, and
it is the reason its numbers can be put in a table next to each other.

### 48.2 The result, and the one figure that is ours to fix

N=5 per binary, medians with ranges. The full table and the reading of each row are in
`PERFORMANCE.md` §10. What belongs here is what changes how the other sections are read:

| | ours | original | ratio |
|---|---|---|---|
| **idle commit** | **46,8 MB** | **8,9 MB** | **5,3×** |
| idle working set | 11,9 MB | 10,1 MB | 1,18× |
| commit per held connection | **0,03 MB** | **0,11 MB** | **0,27×** |
| idle CPU (% of machine) | 0,04 % | 0,02 % | both ≈ 0 |
| startup to listening | 140 ms | 81 ms | 1,7× |

**The idle commit charge is the finding.** 5,3× and the most stable figure in the table
(range 46,7-47,1, ±0,4 %), so it is a property of the binary and not of the run. It is also
**the Go runtime's arena reservation and GC metadata, not our data structures**: §5 measures
our own heap as an order of magnitude smaller than this figure. **An operator sees it**, it
is the one place the reimplementation is unambiguously more expensive than the thing it
replaces, and §48.3 says what would and would not move it.

**And the per-connection row is the one that answers a question §5 could not.** §5 measures
what a connection costs *us* with nothing to compare against. Against the original it is a
tie on working set and handles and **3,7× cheaper on commit**. So the connection path is not
where a reimplementation costs more, and §5's per-connection figures can be read as ordinary
engineering rather than as a suspicion.

**Idle CPU is negligible for both**, 0,04 % and 0,02 %. That settles a question §3 could
not: the tick's cost is entirely tick-bound work, neither implementation pays for idle
polling, and the 1,3 s animation loop does not appear as CPU on this machine.

### 48.3 Two rows that are NOT numbers, and why they are printed anyway

**The handshake row is bimodal on the original and is marked not comparable.** 0,55 ms on
one repetition, 134,7 ms on another, median landing on the slow one. A first-connection cold
path on our side of the measurement is the obvious candidate and **was not isolated**, so
`benchpeer` prints the row with its range and the section says what can be said: neither side
has a slow path that is slow *consistently*.

**Thread count is `n/a`, not zero.** `Thread32First` returns `ERROR_BAD_LENGTH` on this
machine for a `THREADENTRY32` whose `dwSize` is the correct 48 bytes, for every
`TH32CS_SNAPTHREAD` flag combination tried — and the struct size was ruled out as the
variable by a standalone probe before this was called a limitation. **A zero thread count is
not a reading either binary can produce**, and printing one would be exactly the §2 failure
in a different layer: an API that matches nothing, reporting a zero as confidently as the
thing it measures.

### 48.4 It is a comparison of PROCESSES, and the sentence that has to travel with it

**This is not a comparison of algorithmic efficiency.** It includes each language runtime's
idle footprint and GC headroom. The 5,3× on commit is Go reserving arenas and the original's
C++ not doing so; no work in `internal/session` would change it, and a future reader who
"fixes" it by tuning our heap would be chasing a runtime property.

**And it does not measure a match.** Both columns are servers holding an open port with no
clients. The figures that would need a real game — §3's tick and §5's heap curve — **still
have no original counterpart**, and this section closes the gap halfway and names it. A match
needs the game, and a capture replays the game against us only.

### 48.5 Three API facts, measured because they are not knowable by reading

The memory counters were wrong on the first attempt and **wrong invisibly**: working set and
commit both came out as **0,00**, and a zero working set is a legitimate reading. Two facts
had to be established:

1. The call must be `kernel32.dll!K32GetProcessMemoryInfo`. The same-named **`psapi.dll`
   export returns `ERROR_BAD_LENGTH` for every `cb` value tried, including the correct one.**
2. `cb` must be the size of the **plain** `PROCESS_MEMORY_COUNTERS`, **72 bytes**, not the
   `_Ex` structure's 80 — which the plain call also refuses with the same error.

Both failures return a **zeroed structure**, which is why `benchpeer` carries an error out of
every counter read and prints it as `UNAVAILABLE` rather than as 0. §2 exists because a
sampler reported 0 MB three different ways; this is the same class of failure twice more, at
the Win32 layer, and the fix was the same: **make the figure and the reason travel together.**
The reasoning is in `cmd/benchpeer/procinfo_windows.go`, where the next reader will find it.

### 48.6 Why the tool repeats, which is a finding and not a refinement

An early single run of this comparison reported **51,7 MB** idle working set for ours. The
next reported **11,9**. A single-run tool prints whichever it saw and calls it the figure.
That is why `benchpeer` has `-repeat` and reports the median with the range, and why two rows
in §10 are only readable as a range.

**This is the §2 lesson at a third level**, and it is the strongest argument in this section
for the tool existing rather than a shell one-liner: a shell loop would have produced the
same numbers and a single-run tool would have produced a wrong headline.

### 48.7 Verification, and what was not committed

`go build ./...` and `go vet ./...` exit 0; `go test -short ./... -count=1` is green (no `go
test` runs `benchpeer`: it launches processes and would not fit the budget of §0).
`python scripts/check_language.py --strict` reports **195 files, SPANISH: 0, AMBIGUOUS: 0**.
The measurement is `PERFORMANCE.md` §10 and the invocation is in `AGENTS.md`; the figures
were taken with `-repeat 5 -window 15s` on this machine at **20,8-39,8 %** system CPU, which
is printed beside every row for the same reason §1 prints it.

Nothing is committed: rule 4 needs the user's GPG authorisation, and the tree still carries
their uncommitted work.

---

## 50. Two decisions, and one divergence closed without a capture (2026-10-06)

### 50.1 The timing criterion, decided by the user, and what it moved

**The gap, which was real.** `INTENTIONAL-DIVERGENCES.md`'s criterion says a
divergence is a difference visible "in the bytes or in the message sequence". It did not
say what happens when **no byte differs** and only the clock does, and one entry was sitting
on exactly that gap: the WebSocket upgrade latency, where the original defers its `101` to
the next main-loop tick and we answer on the connection's read path.

**The decision, taken by the user**: *"the clients work the same, it is only the connection
and it affects nothing"* **and put it in the intentional file.** That is a criterion answer
and not only a filing decision, so **it is written into the criterion itself** rather than
left implied by where one entry happens to sit:

> timing is NOT part of this criterion. A difference in *when* a client receives something,
> with the bytes and the sequence identical, is not a divergence.

**The reason, and it is the reason the user gave**: the consumers are the game and an
observer and neither reads the wire with a stopwatch. **The boundary clause matters more
than the ruling**, and it is in the same paragraph: *observable timing that changes the
content of what is sent* is still an ordinary divergence. A delay that moves a message into
a different second, changes which peer a message names or changes a count differs in the
bytes or the sequence, and this clause does not cover it. **The observer step's period is
the case that keeps its place** `` it decides **which frame** an observer is served first,
which is content and not timing.

**Nothing was lost in the move.** Every measurement went with the entry: the three-point
`-simulationPeriod` sweep that shows ours does not move with the period and the original's
does, the two-cluster sample distribution, the 509 µs clock floor, and the reason
the row read as "not comparable" for a week (a tool defect, not either server). `UNINTENTIONAL-DIVERGENCES.md`
12 is now a pointer, which is this repository's convention for a move.

### 50.2 A pending item that was refuted by its own evidence

`PENDING.md` 3 listed "enlarging the synthetic fixture" as a way to remove the
`TestSim_synthetic` flake, addressed to **next session**. `UNINTENTIONAL-DIVERGENCES.md`
§11.1 says in its own words that enlarging the fixture **would not** fix it: the settle is
always derived from the last whole second of the recording, so a fixture of any length
keeps landing on the same 1-second tail and the same phase coin-flip. The flake was the
harness's own settle and the settle's floor was already fixed.

**It survived that long because its "who" column said next session**, which is the failure
mode this file keeps warning about: a work item nobody owns is a work item nobody checks.
Now marked refuted, with the reason next to it so it cannot be re-filed.

### 50.3 The upload gate: a wire-visible divergence closed with no capture

**What was missing.** The reconnect section said the *other half* of the sync-message count
guard was not modelled. Reading the guard settled what it is in one line: it wraps four
effects, not one. A `0xFE` upload is taken only when **the peer exists** AND, once the
session is recording, **that peer is participating**. When it is not taken, four things do
not happen: the sub-message is not queued, the accepted-sync counter does not move, the peer
is not marked as having sent data, and its last-sync time is not stamped.

**Why it was a divergence and not a nit.** The queueing half was missing, so a peer that
joined **after** the frame store was created — or one that had sent
LeaveSimulation — had its data put into the next `0x0A`. **That is a sub-message on the
wire that the server never sends**, and it is the kind of difference this register exists
for.

**The change, and its shape.** The counter moved into `Session.QueueGameData`,
because a caller that increments before asking is what the code used to do, and one rule in
one place is the fix rather than the tidying. The handler no longer counts.

**How it was closed without a capture, which is the part worth keeping.** All five committed
captures replay with **no must-match divergence before and after** (770.3 s), because none
of them has a late-joining peer uploading into a recording session. So the change is a no-op
on the recorded evidence and the evidence does not contradict it. **It is not verified
against a recording of the case**, and saying so is part of the entry rather than a footnote
to it.

**And the test that had been asserting the impossible.** The guard's `peer exists` half
failed `TestTickIsOneAckOnceAndExactlyOneUntilItStops`, which turned out to be the
useful failure: that fixture was `NewSession(1)` with **no peer**, and it queued from
slot 1 anyway. **It was asserting an ack cadence the protocol cannot produce** — a
session's tick cannot start before a real peer has sent something. The fixture now has the
peer, and a new test pins all four effects. **It is falsifiable and was checked so**: with
the participating half of the guard disabled, the new test fails on the refusal, on the
counter and on the stamp.

### 50.4 What is left, and who owns it

Of the open entries in the register, **none is closable without a run**: the replay byte
range needs a replay host, the bsPort second peer and the extra relays need a three-client
match, and the two reconnect ones need a capture with a reconnection. That is the honest
answer to "what can I do alone", and it is why this section is mostly one closed item and
two decisions.

**Verification**: `go build ./...` and `go vet ./...` exit 0; the short loop is
green; `internal/sim` replays all five captures at **no must-match divergence**;
`check_language.py --strict` reports 196 files, SPANISH: 0, AMBIGUOUS: 0. Nothing is
committed: rule 4 needs the user's GPG authorisation.
their uncommitted work.

---

## 49. The 5,3x is not the arena, and the handshake row is a tick (2026-10-06)

This section corrects two claims that were in `PERFORMANCE.md` §10 and in `AGENTS.md`, both
of which had been carried since the 48. It adds a tool, `src_go/cmd/vadbreak`, and it adds
one entry to `UNINTENTIONAL-DIVERGENCES.md`. Nothing in `internal/` was changed.

### 49.1 The arena reservation is not a cost, and never was

`PERFORMANCE.md` §10 point 1 said the 5,3x on commit is "the Go runtime's arena reservation
and its GC metadata". A reservation is not a cost, and the tool written to attribute the
figure says so on its first run: the runtime reserves **5,25 GB of address space** for the
heap and commits **0 bytes** of it, and the original reserves 4.136 MB and commits 0 as
well. Both are free on x64.

The reason there was never a knob to turn is that on Windows 64-bit the arena is a
**compile-time constant**: `runtime/malloc.go:252`, `heapArenaBytes = 1 << logHeapArenaBytes`,
where `logHeapArenaBytes = (2+20)*(_64bit*goos.IsWindows)`, i.e. **4 MB**, and `arenaL1Bits = 6`,
`arenaL2Bits = 20`. Scanning `runtime/*.go` for what it reads from the environment, the only
two memory variables that exist are **`GOGC` and `GOMEMLIMIT`**.

### 49.2 Every knob, measured, and none of them moves it

Idle commit charge, `benchpeer`, this machine:

| setting | commit charge |
|---|--:|
| baseline | 46,99 MB |
| `GOGC=50` | 48,24 MB |
| `GOGC=20` | 48,53 MB |
| `GOMEMLIMIT=32MiB` | 46,96 MB |
| `GOMEMLIMIT=16MiB` | 47,36 MB |
| `GOMEMLIMIT=8MiB` | 47,79 MB |
| `-ldflags="-s -w"` (10,67 -> 7,39 MB on disk) | 46,31 MB |
| `GOMAXPROCS=1` / `2` / `12` | 45,22 / 45,79 / 46,80 MB |

**The collector settings make it slightly worse, not better.**

### 49.3 The manual GC, which was the obvious idea, is falsified

The proposal was: force a collection after start-up. Tested with the most aggressive settings
physically possible, `GOGC=1 GOMEMLIMIT=8MiB`, which makes the collector run continuously and
hold the heap under 8 MiB. `GODEBUG=gctrace=1` for 15 s at idle:

```
gc 1 @0.004s 3%: ... 0->0->0 MB, 0 MB goal, ... 12 P
```

**Two GC cycles, a managed heap of 0 MB, and a commit charge of 47,79 MB.** There is nothing
to collect, so `runtime.GC()` after start-up would free exactly nothing. `GODEBUG=inittrace=1`
attributes **99 KB across 97 packages** to our `init()` functions, so it is not our start-up
code either — it is the runtime's and the loader's.

### 49.4 The decomposition, and the two controls that earn it

`vadbreak` walks the address space with `VirtualQueryEx`. The control that makes the row
mean something is **an empty Go program**, because "our data" is only knowable against a floor:

| committed, by region | ours | empty Go | original |
|---|--:|--:|--:|
| `MEM_IMAGE` write-copy | **32,55 MB** | 0,03 MB | **0,07 MB** |
| `MEM_PRIVATE` read-write | 13,21 MB | 10,86 MB | 7,44 MB |
| `MEM_IMAGE` read-only + execute-read | 25,80 MB | 12,50 MB | 27,10 MB |
| `MEM_MAPPED` (DLLs) | 4,83 MB | 1,57 MB | 6,57 MB |
| **whole address space, committed** | **76,74 MB** | 25,39 MB | **41,52 MB** |
| **commit charge** (benchpeer's figure) | **46,71 MB** | 11,70 MB | **8,42 MB** |
| `MEM_PRIVATE` **reserved** | **5,25 GB -> 0 committed** | 5,24 GB -> 0 | 4,14 GB -> 0 |

Reconciliation: 13,21 private + 32,55 dirty image = 45,76 against the 46,71 the first
instrument reports, and the remainder is the writable image sections and the private share of
the mapped DLLs. **Two instruments, same figure, and that is what earns the table.**

So the 5,3x is **32,55 MB of dirty executable pages plus an 11,7 MB Go runtime floor**, and
**our own data is 2,35 MB** (13,21 less the 10,86 an empty Go program already commits) against
the original's 7,44 MB. **On the part that is data we are cheaper than the thing we replace.**
The 32,55 MB is 3x the size of the binary carrying it, so it is not file-backed content: it is
the writable image sections, dirtied by the loader and by start-up.

### 49.5 `vadbreak` found two bugs in itself, which is the argument for it

- **It read the wrong counter.** The first version summed
  `QuotaPagedPoolUsage + QuotaNonPagedPoolUsage` and reported **0,16 MB** for a process whose
  commit charge is 46,71 MB. The column `PERFORMANCE.md` had been calling "commit / private" is
  `PagefileUsage`: `PrivateUsage` is in the `_Ex` structure and the plain 72-byte call that
  `benchpeer` uses refuses the 80-byte one. **The column has been renamed.**
- **It read `SYSTEM_INFO` at the wrong offset.** `dwPageSize` is at **offset 4**, after two
  `WORD`s. Read through a pointer-sized field it lands on `lpMinimumApplicationAddress`, which
  on 64-bit Windows is `0x10000`, so the walk stepped over **64 KB** wherever
  `VirtualQueryEx` declined to describe a region. The symptom was this repository
  contradicting itself: the same binary measured at **32,55 MB** and at **0,29 MB** on the same
  machine. **The totals reconciling with `benchpeer` is what made the difference visible
  instead of a matter of opinion.**

Four tests exist because of those two: `TestCommitChargeIsTheFigureBenchpeerReports` rejects a
value under 8 MB, `TestPageSizeIsReadFromTheRightOffset` pins the offset,
`TestWalkOnAbsentPIDIsAnError` refuses to report nothing for a process that does not exist, and
`TestRenderSaysUnavailableRatherThanZero` keeps "absent" distinct from 0.

### 49.6 The handshake row: "not comparable" was wrong, and it is a divergence

`PERFORMANCE.md` §10.3 said the original was bimodal, that a first-connection cold path **on
our side** was the obvious candidate, and that it was **not isolated**. Both halves were wrong.
`-repeat 15 -json`, one upgrade sample per repetition:

```
ours     (n=15)  0.55 0.56 0.57 0.58 0.61 1.08 1.12 1.12 1.14 1.14 1.77 1.98 2.63 5.38 15.07
original (n=15)  0.54 0.95 1.14 1.97 3.14 | 124.90 127.60 128.16 128.19 131.34 132.75 133.88 135.50 135.52 135.84
```

**Two clusters and an empty gap.** Not a slow path: a whole number of periods. The prediction
was that the deferral follows `-simulationPeriod`, and moving it settles it:

| `-simulationPeriod` | ours, median | original, median |
|---|--:|--:|
| 30 ms | 1,12 ms | 36,9 ms |
| 125 ms (default) | 1,12 ms | 128,2 ms |
| 500 ms | 1,09 ms | 513,3 ms |

**Ours does not move. The original's does.** The original serves the upgrade from its main-loop
tick; ours is served on the connection's read path (`internal/server/ws_listener.go`, against
the 125 ms tick in `internal/server/server.go`).

**What actually made the row unreadable is the tool**: `benchpeer` takes **one** sample per
repetition and reports the **median of a bimodal distribution**, so the number in the table
depends on the tick phase and on N, and **at N=5 the row can invert**. That is a defect to fix
in the tool, and it is not fixed.

**No byte differs** — a client can only time it — and whether latency is "on the wire" is not
decided anywhere in this repository. So it is filed in `UNINTENTIONAL-DIVERGENCES.md` 12
**because no reason is on record**, not because the criterion has been read to include timing.

### 49.7 Verification, and what was not committed

`go build ./...` and `go vet ./...` exit 0. `go test ./cmd/vadbreak/ -count=1` is green (11
tests, 1 skipped: the `!windows` branch cannot run on this machine, which is what it is there
for). `go test -short ./... -count=1` was green before the tool was added and its own package
is the only one that changed since; the full suite is not re-run here because it is 775 s and
nothing it covers was touched. `python scripts/check_language.py --strict` reports
**196 files, SPANISH: 0, AMBIGUOUS: 0**.

Not committed: rule 4 needs the user's GPG authorisation for the parent, and `src_go` is a
submodule whose own history is a separate decision. The tree still carries their uncommitted
work.

### 49.8 What the samples row found, twice, and the clock on this machine

Printing the samples instead of one number found two more things, and neither of them is
about either server.

**The clock is coarser than the figure it is used to measure.** A standalone probe of 200
loopback TCP round trips:

```
n=200  zero=94  min=0 p25=0 p50=509 p75=556 max=11333  (microseconds)
```

**47% of the readings are zero.** `time.Since` on this machine cannot resolve a sub-millisecond
interval reliably. So the six `0.00` samples the new samples row printed on its first run were
**a clock, not fast handshakes**, and they were being stored as `0.00` and then averaged. The
tool now:

- returns the handshake timing in **microseconds**, as an integer;
- **counts** a sample it could not resolve in `UpgradeUnresolved` and **excludes** it, because
  an untimed reading is not a fast handshake — that is the tool's own second rule applied to
  its own new figure;
- prints `(+N the clock could not resolve)` beside the samples.

The unit change is not cosmetic. Stored in milliseconds and printed to two decimals, **every
reading under 10 µs collapses into the same `0.00`** and cannot be told apart from a reading
that never happened: a tenfold loss of exactly the range the samples exist to show.

**What this does and does not invalidate.** `startup` at 140 ms is three orders of magnitude
above the floor and stands. **Our own upgrade figure does not**: the samples are 441-1210 µs
with a median of ~550 µs, so **0,55 ms sits right on the clock's noise floor** and must not be
quoted as a precise number. The comparison survives because the original's 125.000 µs is two
and a half orders above the floor and ours is not close to it, and because the claim rests on
the `-simulationPeriod` sweep. **The honest sentence is "ours is around a millisecond and does
not move with the period", not "0,55 ms".**

**How often each cluster occurs is not reproducible.** With `-upgrades 5 -repeat 3`:

```
upgrade samples ours (us):      441 469 535 544 548 549 553 569 804 1087 1210
upgrade samples original (us):  115079 118454 119954 120420 122125 123899 124730 126423 127135 128140 129753 132942 133106 133112 133741
```

**15 of 15 in the slow cluster for the original**, where the run in 49.6 had 5 of 15 in the
fast one. The medians are stable and the frequencies are not, which is the argument for
printing the samples instead of one number: with a single sample per repetition the reported
figure is decided by which cluster got the majority, and the majority is a draw. **The
published table's `0,55` and `129,4` are two draws from that distribution, and the row should
never have been a single number.**

### 49.9 Verification

`go build ./...` and `go vet ./...` exit 0. `go test ./cmd/vadbreak/ -count=1` green, and
`go test -short ./... -count=1` green at **50,2 s** with the new package in it. `benchpeer` has
**no test file**, so its two changes are verified by what they print: `-upgrades` shows the
samples, and the unresolved counter shows `(+N ...)` when the clock could not time one.
`python scripts/check_language.py --strict`: **196 files, SPANISH: 0, AMBIGUOUS: 0**.

`gofmt -l` reports **65 files** in this tree and it is not a finding: the working copy is CRLF
throughout and `gofmt` normalises to LF, so it flags files nobody has touched
(`internal/session/session.go`, `cmd/proxy/main.go`, `battleserver.go`). The files written this
round were checked to match that convention and to carry no delta of their own beyond it.

---

## 51. `-maxPagesToStorePerSession` keeps the original's lack of a cap (2026-10-07)

### 51.1 The question, and what it was not

`PENDING.md` 3 listed it as a decision for the user: *"`f = flag / 8` with no upper bound: with
a very large flag the ceiling is the operator's. Is that what the operator wants?"*

The first thing to settle was whether the flag is ours at all, because the phrasing invited
the wrong reading. **It is the original's, under the same name.** `src/main_1400d1310.cpp:615`
compares the argument against `"maxPagesToStorePerSession"` and `:617` stores
`(float)atoi(v49) * 0.125`. Ours registers it as an **alias** of `-maxPagesPerMatch`
(`internal/config/config.go:63`, with the reason in the comment at `:59`), because a launcher
that passes the upstream spelling must not be rejected by the parser. That aliasing is already
filed as intentional (`INTENTIONAL-DIVERGENCES.md` 7, the table row for the flag).

So the question was never "does the flag exist". It was only **whether a ceiling goes on top of
it**.

### 51.2 The decision, and why there was nothing to implement

**Keep the original's behaviour: no cap.** With a large flag the ceiling is the operator's, and
it is the operator's on the original too.

That is worth one line of arithmetic, because the scaling is the same on both sides and no
change is a change of behaviour by definition: the original multiplies by `0.125`
(`main_1400d1310.cpp:617`) and `internal/session/session.go:1045` multiplies the same value by
the same `0.125`; the default is `0`, which is `f = 1.0` on both sides
(`SESSION-JOURNAL.md` 28, the flag table). **Adding a cap would therefore have been a
divergence, and not a safety feature**: it would have made a late observer get fewer pages than
the original serves it, which is the one thing this repository is not allowed to do quietly.

**No line of production changed, and none will.** The row is closed in `PENDING.md` and
mirrored in its section 5, with the reason next to it, so that the next reader does not
re-open it as a hardening opportunity.

---

## 52. The byte range of the replay upload, closed by a negative measurement (2026-10-07)

### 52.1 What could not be measured

PENDING.md 3 listed the byte range of the replay upload as the one open item that only closes by playing, with a replay host it called a real POST. The test was run:

- a listener on the replay host, the original started with -uploadVia=lobby (which is the only spelling that constructs the upload object, per sub_1400D6790 2: v32 == 2), the default putting it at 127.0.0.1:8443;
- the match was played to the end, so both the per-message latch (sub_14010F4F0:99, gated on session+2140) and the end-of-session one (sub_1400D8E10:1209 -> sub_1400EA460:51 -> OnAddReplayData) had the chance to fire.

What the listener saw: the TCP connection arrived, and no request ever followed. Not in a short warmup, and not with a full match done. battle-server.txt, the only recorded run of the original, agrees: the uploader is built (:18) and its connect is deferred and never completed (:23), with no request in the remaining 423 lines.

### 52.2 What the negative settles, and what it does not

This moves the entry from UNINTENTIONAL-DIVERGENCES.md 1 to INTENTIONAL-DIVERGENCES.md 6, and it changes the status of the question: it is no longer blocked on a measurement, it is unmeasurable by this method, because the original does not upload in any observable run. With nothing to be byte-identical to, the byte range stops being a fidelity question and becomes a design choice, and the choice that does not throw the replay away is taken.

What stays a reading, and is said so: the slice of bytes that goes into the compressed blob. The original builds the request by hand with nothing but POST /game/replay/setReplayData?matchid=<u64>&sequenceid=<u32> and a Content-Length (sub_14010F480, the request builder), reuses the same path and parameters, keys every page by matchid+sequenceid with the counter starting at 1, and base64-encodes the zlib stream. All of that is pinned by TestReplayUploadRequestShapeAndPayload. Only from-the-type-byte-onwards versus framing-included is ours and unverified.

### 52.3 The two things the test added while it tried to decide the range

TestReplayUploadRequestShapeAndPayload was meant to compare the body, and instead found two things ours did wrong against the contract read from the binary:

1. We sent a Content-Type: application/octet-stream header. The original request builder writes only the request line, Host: and Content-Length:; an extra Content-Type is a header the replay host does not expect. Fixed in internal/replay/uploader.go pushReplayData with SetNoDefaultContentType(true) the serializer default-substitution is what is disabled, because deleting an absent header leaves the substitution on.
2. The assertion that the compressed-and-base64 payload inflates back to the message body is falsifiable and was checked by breaking it both ways: prefixing the payload with the RLink framing fails the base64 step, and writing a Content-Type: multipart/form-data fails the header step. The test fails on both corruptions.