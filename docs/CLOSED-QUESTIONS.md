# The two questions closed by the user

> **What this file covers.** The two questions the user **declared closed** and that
> are not reopened, with the evidence that closes them.
>
> **Status.** **Closed by authority.** They are not "closed by measurement": they are
> decisions by the user about what gets investigated, and that is why they are here, in a
> **living** file, and not in the journal. `AGENTS.md` repeats them in its first section
> because that is the front door of the repo.
>
> **Evidence.** Each one says what was closed **and** what stays open inside
> it, so that "closed" is not read as "nothing is left".

Both are in the index of `SESSION-JOURNAL.md` from the section that closed them, and the
text of that section is still there: `SESSION-JOURNAL.md §25` (the zoom) and the index's own
closure by authority. This file takes them out of there so that they **do not become
history**.

---

## 1. The BattleServer is DETERMINISTIC

- **Closed by authority**, not by measurement: **it is not measured, no run that would
  contradict it is looked for**. It is the BattleServer that Age of Empires ships, so the
  original's output is reproducible.
- **What does keep being done** is comparing against the **recording**, which is the available
  copy of that output. That distinction is what avoids the misunderstanding: we are not
  looking for a better run than the recording, we compare with the one there is.
- **What stays open inside this closure**: the measurement that §25 asked for —the zoom with
  the **original at the front**— is **voided** by this closure. It is not that it came out
  badly: it is that it is not going to be done.

## 2. The observer's excessive zoom is NOT a BattleServer bug

- **Closed by the user, confirmed from outside.**
- **What the measurement closed, and what holds it up**: the **observer's stream is
  indistinguishable from the original's, byte for byte**. The harness **did not have the
  question**: it did not compare the observer's stream, so it could not answer.
- **What stays open inside this closure, and is informative, and is not touched**:
  - the **grouping** of the acks (how much fits in a tick);
  - the `LobbyConnect` with the same bytes on both connections.
- **The third one was never a ping, and it is closed as a false claim** (2026-10-05,
  `SESSION-JOURNAL.md` §44.3). §25 read two bytes of the original's `101` response, taken
  from the second of the two TCP records it arrives in, as if they were the opcode of a
  WebSocket control frame: `0x65` is the `e` of `Sec-WebSocket-Accept` in **our** response
  and `0x6B` is the `k` of the accept key in **the original's**. Neither side ever sent a
  ping or a pong. The difference the misreading exposed was real and is fixed: we did not
  send a `Server` header and the original does.
- **A warning that has to be written down**: of the two that remain, the `LobbyConnect` one
  **is not** the zoom and is **not** covered by this closure. It is a transport divergence
  and is still alive, and that is why it is in
  [`UNINTENTIONAL-DIVERGENCES.md`](UNINTENTIONAL-DIVERGENCES.md) (point 4).
  "The zoom is closed" **does not** mean "everything about the observer is closed".

## What was stopped being done, and why

- The control match with the original at the front **for the zoom** is **not** done.
- The lobby is **not** chased for the zoom.
- The observer's stream is **not** investigated any further **for the zoom**. The observer's
  stream is still compared, but under the harness's normal rules
  (`replay-page-order` is must-match), not for this question.

## Where the detail is

- `SESSION-JOURNAL.md §25` — "The observer's zoom: measured, and not in the stream", and its
  subsection "The only thing left different, and not touched", which is where the three
  points above come from.
- `SESSION-JOURNAL.md §23` — why the harness did not have the question, and the three must-matches that
  uncovered it.
- `AGENTS.md` — the section "TWO QUESTIONS THAT ARE CLOSED. THEY DO NOT REOPEN.", which is the
  short and authoritative version.
- `SESSION-JOURNAL.md §26`, `§27` — the zoom is not the reason why row 1 of the `CAPTURES.md` matrix is
  still pending; that row is pending for the full comparison with the original at the front.
