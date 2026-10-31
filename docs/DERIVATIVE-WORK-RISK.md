# Derivative-work risk review for `src_go`

**Date:** 2026-10-06  
**Scope:** the Go source under `src_go/`, with the repository documentation and
`battle-server.txt` used only as context.

## Important limitation

> **STATUS: REVIEWED, AND CARRIED OUT (2026-10-06). Read this as history.**
>
> This file is a point-in-time assessment of `src_go`. Its recommendations were
> acted on in `SESSION-JOURNAL.md` §47, and reading it as a description of the tree
> as it now stands would be wrong. **What changed**, all of it measured:
>
> | | before | after |
> |---|---|---|
> | `sub_1400` / `main_1400` references in `src_go/**/*.go` | 689 (48 of 144 files; **343 in `_test.go`**, which this review did not list) | **0** |
> | recovered vtable slots | 11 | **0** |
> | embedded decompiler pseudocode | 25 | **0** |
> | memory-offset narration | 191 | **0** |
> | log strings verbatim from the binary | 6 | **0** |
> | logger labels copied from the original's class names | 64 | **0** |
>
> The two files this review asked for now exist:
> [`PROVENANCE.md`](PROVENANCE.md) (Phase 1: the per-path list of what may not be
> redistributed, and the honest statement that this is **not** a clean-room
> implementation) and [`BEHAVIOUR-NOTES.md`](BEHAVIOUR-NOTES.md) (Phase 3's comments,
> rewritten as behaviour so the knowledge survives the removal).
>
> **One finding here is still open and is not the author's to close**: whether the
> result is sufficient is a legal question about jurisdiction and about rights this
> repository does not hold. `PROVENANCE.md` says so in its own text.

This is a technical source-review, not a legal opinion. Whether a reimplementation
is a derivative work depends on the applicable jurisdiction, the rights held by the
copyright owner, the project's licence and distribution model, the provenance of the
implementation, and the purpose for which it is used. A qualified lawyer should review
the project before publication or redistribution.

The review does not conclude that the project infringes copyright. It identifies source
material and implementation practices that make such an argument easier to raise, and
technical changes that would reduce that risk.

## Executive summary

The strongest risk factors are not the fact that the project speaks the same network
protocol or has compatible message numbers. They are:

1. production comments in `src_go` that reproduce decompiler addresses, function names,
   field offsets, pseudocode fragments and implementation decisions from `src/`;
2. a stated goal of a **byte-for-byte reimplementation**, combined with a one-to-one
   translation of many recovered internal behaviours;
3. source-level names and explanations that explicitly map Go fields and branches to
   particular functions and offsets in the original executable;
4. exact original log strings and other expressive text reproduced in the Go server;
5. generated tests and fixtures that preserve exact captured output, if those captures
   are distributed without permission.

The lowest-risk parts are ordinary Go infrastructure, independently written socket
management, generic concurrency, standard WebSocket/TLS/zlib/base64 use, and
compatibility code that implements a documented or observable wire contract without
copying original source expression.

Changing names alone would not solve the central issue. The meaningful mitigation is
to create an independent implementation boundary: keep a separately reviewed,
minimal protocol specification; remove decompiler-derived material from production
source; reimplement from that specification and black-box observations; and retain
the original binary/decompilation only in a non-distributed research or evidence
repository where legally permitted.

## What “derivative” can mean in this review

The word is used here in a practical risk sense, not as a legal finding. There are
three different questions that should not be conflated:

- **Ideas, facts and interoperability requirements:** message type numbers, field
  widths, byte order, port numbers and externally observable state transitions may be
  necessary for compatibility. Their necessity does not automatically make the Go
  source a copy of the original expression.
- **Expression:** source text, distinctive comments, log messages, names, algorithms
  and detailed implementation structure can be protected independently of whether
  the resulting packets are compatible.
- **Provenance and intent:** a project that openly translates a decompilation and
  aims to reproduce every observable detail presents a different risk profile from a
  clean-room implementation based only on an independently prepared protocol
  specification.

The same code can therefore be technically necessary for interoperability while still
having a provenance or expressive-copying problem that needs legal review.

## Findings by risk level

### High priority: decompiler-derived material embedded in production Go

Many comments in `src_go/internal/` cite identifiers such as
`sub_1400EB5F0`, source line ranges, object offsets and recovered branches. Examples
are in:

- `src_go/internal/protocol/messages.go`;
- `src_go/internal/server/handlers.go`;
- `src_go/internal/session/session.go`;
- `src_go/internal/config/config.go`;
- the observer, replay and peer-management implementations.

The comments explain not merely the public wire contract, but how particular fields
correspond to offsets in the original object layout and how individual decompiled
branches were translated. Phrases such as “the original writes it exactly once”,
references to `peer+68`, and embedded decompiler snippets are especially useful
evidence that the Go implementation was derived directly from the binary's recovered
expression.

**Recommended change:** remove these citations and decompiler-specific explanations
from distributed production files. Replace them with short, independently written
statements of the externally required behaviour, for example:

> “The client message version is retained for the lifetime of the connection and is
> used when sending the observer completion marker.”

Keep detailed reverse-engineering notes in a separate, access-controlled research
record only if the project has a lawful basis to keep and use them. Do not replace
one `sub_...` reference with a renamed reference that preserves the same recovered
algorithm.

### High priority: one-to-one recovered internal structure

The Go code models concepts that correspond closely to recovered original objects:
peer fields, session tables, observer registration, reconnect counters, timeout
state, replay windows and message dispatch branches. Some of those concepts are
unavoidable for a compatible server, but the current comments frequently state the
correspondence explicitly and preserve original field-offset semantics.

A direct translation of an internal structure is more exposed than an independent
state machine that merely satisfies the same public contract. The risk is highest
where a Go field exists solely because a particular original offset or branch was
recovered, rather than because the public protocol requires it.

**Recommended change:** define the Go state model from externally testable invariants:

- what a client may send;
- what response it receives;
- when a connection is closed;
- what replay data is retained and delivered;
- what timing and ordering guarantees are observable.

Then simplify or reorganise private fields where possible. Preserve wire-compatible
behaviour, but avoid mirroring private object layout, recovered names or branch order
unless an independent protocol requirement needs it. Add a design note describing the
public invariant, not the original memory layout.

### High priority: exact original strings and expressive diagnostics

The server and tests contain original-looking diagnostic text such as unknown-message
warnings and protocol-specific log phrases. A log string that is required by a client
is a compatibility fact; a distinctive developer diagnostic usually is not.

**Recommended change:** retain only strings that are demonstrably part of the public
interface or required for existing tooling. Rewrite private diagnostics in original
project terminology and style. Search for exact strings copied from the original
console output or decompilation and review each one individually.

### Medium priority: wire formats and protocol constants

The definitions in `src_go/internal/protocol/messages.go` include message identifiers,
field layouts, byte order, relay formats, lobby-connect parsing and observer response
formats. These are primarily compatibility facts and are generally lower risk than
copied source expression. The risk rises where a comment includes the original
function name, exact offset, or a long recovered rationale.

**Recommended change:** keep the minimum information needed to implement the
protocol, but source it from a project-owned protocol specification written in terms
of observed inputs and outputs. Remove original function names and decompiler
addresses from the Go package. Use neutral names for private helpers and document
ambiguous behaviour as a tested compatibility requirement.

The numeric constants themselves should not be changed merely to look different:
changing a wire value would break interoperability and would not meaningfully address
the underlying provenance question.

### Medium priority: byte-for-byte goals, captures and golden tests

The repository explicitly describes `src_go` as a “byte-for-byte reimplementation” and
compares it against the original. The comparison harness, captured streams and golden
observer vectors preserve exact original output. This is strong evidence of the
project's objective and provenance, even though a packet capture is not source code.

**Recommended change:** change the public project description to state the narrower
technical goal: “protocol-compatible implementation validated against black-box
observations.” Keep exact comparisons as internal conformance tests where permitted,
but review whether original captures, logs, binary artifacts and golden data may be
redistributed. Prefer reduced, synthetic, non-substantial fixtures for a public
repository. Do not claim that a test fixture is safe to distribute without checking
its provenance and applicable law.

### Medium priority: tests that assert recovered internals

Some tests preserve details that are not needed to establish a public contract, such
as exact private sequencing, internal counters or behaviour named after decompiled
functions. Tests that assert client-visible bytes, message order, connection state and
bounded timing are much easier to justify as interoperability tests.

**Recommended change:** classify each test as one of:

1. public protocol conformance;
2. security or robustness;
3. implementation invariant;
4. historical reverse-engineering evidence.

Keep categories 1–3 in the distributable project after rewriting their descriptions
independently. Move category 4 to the research record or remove it from the public
tree. A test should not cite a decompiled function merely to explain why an externally
observable assertion exists.

### Low priority: ordinary implementation mechanisms

The following are not, by themselves, meaningful evidence of copying:

- Go package structure and idiomatic error handling;
- TCP, UDP, TLS, WebSocket, zlib and base64 APIs;
- mutexes, atomics, channels, pools and cancellation;
- generic replay queues and bounded buffers;
- ordinary serialization helpers needed for a public wire format;
- independently written test utilities and load-generation infrastructure.

They should still be reviewed for third-party licence compliance, but changing them
would not be a sensible response to the derivative-work concern.

## Files and areas to review first

This is a priority order, not a claim that every line in these files is copied:

1. `src_go/internal/server/handlers.go` — dispatch and comments that map branches to
   recovered original functions.
2. `src_go/internal/session/session.go` and the other session files — private fields,
   offsets, observer state and peer lifecycle explanations.
3. `src_go/internal/protocol/messages.go` — protocol definitions mixed with original
   function references and recovered rationale.
4. `src_go/internal/config/config.go` — original flag names and recovered default
   calculations.
5. `src_go/internal/replay/` — upload behaviour and comments describing the original
   implementation rather than the public upload contract.
6. `src_go/internal/server/` — original-looking log strings and detailed recovered
   dispatch behaviour.
7. `src_go/internal/sim/`, `cmd/gensim/` and `cmd/proxy/` — exact captures, golden
   vectors, metadata and tests that may be redistributed.
8. Root documentation and `battle-server.txt` — provenance statements, original logs
   and redistribution of original output.

## A safer change plan

### Phase 1: provenance and distribution audit

- Record which files were written from decompilation, black-box observation, public
  protocol documentation or ordinary engineering.
- Identify who owns each original binary, decompilation, capture, log and golden file.
- Decide what may be distributed and what must remain private.
- Obtain legal review before deleting evidence or publishing a new licence.

### Phase 2: establish an independent specification

- Write a compact protocol specification using only public wire inputs, outputs,
  state transitions and testable timing/ordering requirements.
- Remove addresses, decompiler names and memory offsets from that specification.
- Have a person who did not inspect the decompilation review the specification and
  identify which requirements are actually observable.

### Phase 3: refactor the distributable Go implementation

- Rewrite comments from the public-contract perspective.
- Remove direct decompiler references from production code.
- Reorganise private state where it does not affect the wire contract.
- Replace distinctive copied diagnostics with project-owned diagnostics.
- Keep compatibility constants and serialization only where required by the protocol.

### Phase 4: validation

- Run protocol tests against the independent specification.
- Keep black-box differential tests only where their inputs and outputs can lawfully
  be used and distributed.
- Re-run the repository's language, build, vet and test checks.
- Have an independent reviewer compare the new source against the original for
  non-essential similarities.

## What should not be done

- Do not simply rename functions, reorder statements or change formatting while
  preserving the same recovered implementation and comments.
- Do not remove evidence of provenance and then claim the implementation was clean-room.
- Do not change protocol bytes solely to make the code look unrelated; that harms
  compatibility and does not answer the legal question.
- Do not assume that a reverse-engineering exception, interoperability exception or
  personal-use rule applies to redistribution.
- Do not distribute the original executable, decompilation, logs or captures until
  their rights and licence status have been checked.

## Bottom line

The parts that most clearly deserve rewriting are the decompiler-derived comments,
offset/function mappings, embedded pseudocode, original diagnostic text, and private
state structures whose shape is justified only by the recovered binary. The public
wire protocol and independently necessary compatibility mechanisms are weaker
candidates for change, although their provenance and accompanying explanations still
need review.

The safest technical direction is not to make the Go code “look different.” It is to
separate a legally reviewed, independently written compatibility specification from
the reverse-engineering evidence, then make the distributable implementation express
only that specification and its own design.
