# PROVENANCE: what this repository is, and what may not leave it

> **THIS FILE DOES NOT LEAVE THIS REPOSITORY.** Not in a published tree, not in a
> release, not in a tarball handed to anyone. It says in its own words that the
> implementation was informed by disassembling the original executable, and a
> document that says that is a record, not a publication artifact. The reasoning
> is in **"THIS FILE IS NOT FOR DISTRIBUTION"** below, and the short version is:
> **an honest internal record of how the code was built is free; a voluntary
> public admission of the derivation is not, and nothing requires the second.**
> Its own table lists it as not distributable, next to `src/` and the captures.

**Date:** 2026-10-06
**Written because:** `DERIVATIVE-WORK-RISK.md` reviewed the Go source and
concluded that the tree carried material making a copyright argument easy to
raise. This file is the answer to the part of that review that is a decision
rather than a code change: **which parts of this repository are ours to
distribute, and which are not.**

## The short version

This is an **independent reimplementation built by observing a proprietary
server's network traffic and its behaviour on the wire.** It is not a clean-room
implementation, because the behaviour it reproduces was learned partly by
disassembling the original executable. **It says so here, and here stays.**

Two things follow, and they are easy to run together:

- **The repository must not lie to itself.** Keeping the derivation written down,
  in the place where the work happened, is what stops a later reader — or a later
  session — from assuming a clean-room provenance that never existed.
- **The repository must not volunteer the derivation outward.** Accuracy about
  one's own history is not the same act as publishing an admission, and only the
  first is free. A tree full of decompiler citations is also a worse thing to hand
  anyone than a tree that states its behaviour, which is the other half of what
  §47 changed and is independent of this file entirely.

The review's own "What should not be done" lists *"do not remove evidence of
provenance and then claim the implementation was clean-room"*. That warning is
correct, and this file satisfies it — **by being accurate, not by being published.**

## What was done, measured

Before this file, `src_go/` carried **689** references to decompiler symbols
across 48 of its 144 Go files, plus memory-offset narration of the original's
object layout, plus decompiled C++ quoted as pseudocode, plus six log strings
copied verbatim out of the original binary. Measured after:

| artifact | before | after | how it was verified |
|--|--|--|--|
| `sub_1400XXXXXXXX` / `main_1400XXXXXXXX` references in `src_go/**/*.go` | 689 | **0** | regex over every `.go`, case-sensitive |
| recovered vtable slots (`PeerMessage__vmethod_18`) | 11 | **0** | same |
| embedded decompiler pseudocode (`*(_QWORD *)`, `goto LABEL_`, `vN =`, `__int64`) | 25 | **0** | same; 4 hits remained and were `kilobytes` matching `LOBYTE` |
| memory-offset narration of the original's layout (`session+2192`, `peer+0x5A`) | 191 | **0** | same regex; the 4 remaining hits were `kilobyte` matching `LOBYTE`, case-insensitively |
| log strings verbatim from the original binary | 6 | **0** | each string searched against `src/**/*.cpp` |

The last row is the only one that is a claim about *expression* rather than
about *provenance evidence*, and it was checked the only way it can be: every
string literal in `src_go` of 12 characters or more was searched against the
concatenated text of the decompilation. Six matched and were rewritten in this
project's own wording. What remains matching is enumerated in
`INTENTIONAL-DIVERGENCES.md` §8 with the reason each one has to stay.

**What did not change: one byte of behaviour.** Every edit was a comment, a log
string or a test failure message. No executable statement was touched, and the
test suite is the evidence: the same tests, the same capture replays, the same
comparisons.

## The rule the Go source now follows

Every comment in `src_go/` answers one of three questions:

1. **What does the wire require?** The message, the field, the byte order, the
   status value, the timing that a client can observe.
2. **Why is this code shaped like this?** The performance measurement, the bug
   this ordering prevented, the constraint that forced it.
3. **Where is the evidence?** `SESSION-JOURNAL.md §N`, a capture directory name,
   `src/sub_*.cpp`, `battle-server.txt`.

What no comment in `src_go/` does any more is explain the original's *internal
structure*: which function wrote which offset, which branch of which recovered
switch produced which constant. That knowledge was real, and it was removed from
the distributed tree, but **it was not deleted** — see
[`BEHAVIOUR-NOTES.md`](BEHAVIOUR-NOTES.md), which holds the recovered structure
in behavioural form, and `SESSION-JOURNAL.md`, which holds it with its evidence.

## THE LIST: what must not be redistributed

Everything in this section is either the original owner's expression, or a
recording of it. None of it is ours to publish. This is a per-path list because
a list by category is a list someone will argue with.

| path | what it is | why it cannot be redistributed |
|--|--|--|
| `BattleServerOrig.exe.pristine` | the original executable | the copyright owner's binary |
| `src/**/*.cpp`, `src/**/*.hpp` | decompilation of it, 7,450 files | the owner's expression, recovered |
| `headers.hpp` | the decompilation's own type header, in compilable form | same |
| `battle-server.txt` | 47 KB of the original's console output | a verbatim recording of the owner's expressive output |
| `captures/**` | recordings of the original's traffic | exact captured output of the owner's program |
| `src_go/internal/sim/testdata/**` | five committed captures + two synthetic fixtures | same; the synthetic ones are the only ones that are ours |
| `src_go/internal/server/testdata/observer_vectors.json` | `0x19`/`0x1E` bytes taken from the original | exact captured output |
| `SESSION-JOURNAL.md` | the research record; 7,118 lines quoting the decompilation | quotes the owner's expression throughout |
| `BEHAVIOUR-NOTES.md` | the recovered structure written as behaviour, 22 sections | **no** — it is a summary of the same recovery and exists because of it |
| **`PROVENANCE.md` (this file)** | **this file** | **no** — see the banner at the top and the section below |

`src_go/internal/sim/testdata/synthetic/` and `.../synthetic-rollover/` are the
exception and the reason to read the rest of this table carefully: they were
**written by this project**, they encode the behaviour as an input and let the
harness produce the output. They are the only fixtures here with no recording in
them.

## What the review did not claim, and this file does not claim either

- That this project infringes. That is a legal question, it depends on
  jurisdiction and on rights this repository does not hold, and it is not
  decided here. **A qualified lawyer must review this before publication.**
- That removing the citations changed the provenance. It did not. It changed what
  the Go source itself contains.
- That the wire protocol is the risky part. It is not. Message numbers, field
  widths, byte order and port numbers are interoperability facts; a client
  cannot be written against them otherwise. They stay as they are. What was
  removed is the account of how each one was *found*.

## THIS FILE IS NOT FOR DISTRIBUTION

**It is a research record, and it stays here.** That is not a compromise and not a
pending decision: it is the answer, and it is stated at the top of the file because
the whole point of the record is defeated if it is read as a publication artifact.

The reason is the first paragraph. **This file says in its own words that the
implementation was informed by disassembling the original executable.** Publishing
it does not make the other files honest; it converts a defensible position — *an
independent implementation, verified against observed traffic* — into an
acknowledged derivative-work claim, in the project's own words, on its own letterhead.
Honesty about **how a project was built internally** and **what a project volunteers
about its own derivation** are different acts, and only the first one is free.

So, corrected and final:

| file | distributable? |
|--|--|
| `src_go/` Go source | possibly, subject to the legal review and to the fixtures being excluded |
| `src/`, `headers.hpp`, `BattleServerOrig.exe.pristine`, `battle-server.txt`, the captures | **no** |
| `SESSION-JOURNAL.md`, `BEHAVIOUR-NOTES.md` | **no** — both carry the decompilation's citations by design |
| **`PROVENANCE.md` (this file)** | **no** |

**What the public tree needs instead is a short statement of what the code is**, written
without the derivation: what it implements, against which protocol, and that it contains
no third-party or original material. That is a different document, it is a legal
decision, and it is not written here — because this repository has no authority to decide
it, and no lawyer has reviewed it.

**Nothing has been distributed, and nothing is staged for distribution.** The separation
is recorded and nothing else is pending.

## If publication is decided on later, the order that matters

1. Get the legal review. Everything below is worthless without it, and the review is
   also what says what the public statement may and may not claim.
2. Decide, with the review's answer, what the public tree is: at minimum the `src_go`
   Go source and `go.mod`, with the captures and the fixtures excluded.
3. Write the public statement **for** what the tree contains, not by reusing this file.
4. Keep this file, the journal and the notes private, and say in the public tree
   nothing about how the implementation was derived.

## What was deliberately NOT done

- **No renaming of protocol constants.** Changing `0x1E` because it came from the
  binary would break interoperability and answer nothing.
- **No deletion of the research record.** The evidence is what makes the
  behaviour claims in `INTENTIONAL-DIVERGENCES.md` and `PENDING.md` auditable.
  Removing it would make this repository's own conclusions unfalsifiable, and
  `CLOSED-QUESTIONS.md` exists precisely because they are held to that standard.
- **No claim of clean-room.** See "The short version".
- **No distribution, and none staged.** That is a completed decision, not a
  pending one, and no artefact in this repository is arranged for publication.

## The files that carry the boundary

All three are internal. There is no fourth.

| file | what it is for | who it is for |
|--|--|--|
| this one | the decision: what may be published, and the record of how the code was built | whoever picks this repository up next, and whoever later decides on publication |
| [`BEHAVIOUR-NOTES.md`](BEHAVIOUR-NOTES.md) | the recovered structure, written as behaviour | a maintainer who needs the "why" the Go comments no longer carry |
| `SESSION-JOURNAL.md` | the same knowledge with its evidence, plus everything else | a maintainer auditing a specific decision |

`DERIVATIVE-WORK-RISK.md` is the review that asked for this. It is kept as history:
it is a point-in-time assessment and its recommendations have been carried out, so
reading it as a statement of the tree's current state would be wrong.