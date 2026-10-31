# Documentation review

Date: 2026-10-05

## Scope and validation

The project Markdown documentation was checked against the current repository
tree, the Go implementation in `src_go/`, the decompilation in `src/`, and
the focused tests:

```text
go test ./internal/config ./internal/replay ./internal/sim -short -count=1
ok
```

The repository language check also passes:

```text
python scripts/check_language.py --strict
files scanned: 195   SPANISH: 0   AMBIGUOUS: 0
```

`AGENTS.md` remains repository guidance and `go1.27.md` remains a copied
external reference. They were not edited.

## Corrections applied

The following documentation issues found during the audit were corrected:

1. Replay upload documentation now distinguishes the original binary's
   no-retry behavior from the Go implementation's single retry.
2. `CAPTURES.md` now describes the settle window in its two actual stages:
   the base settle and the truth-based widening.
3. Local source references now use address-based files under `src/`, while
   upstream/original source names are explicitly labelled as such.
4. Project-document references formerly written as `MEMORIA §N` now use
   `SESSION-JOURNAL.md`.
5. Historical references to the deleted capture runbook are kept only in the
   journal and are identified as historical.
6. Decimal values in the edited performance and capture prose use English
   decimal points.

## Remaining qualifications

These are limits of available evidence rather than language defects:

- The claim that `src_go` is a byte-for-byte reimplementation is demonstrated
  for committed capture scenarios, not every possible input.
- Exact performance figures depend on the machine and measurement session.
- Historical journal conclusions are recorded evidence, not automatically
  reproducible guarantees.
- Some replay-system descriptions rely on upstream source names and cannot be
  fully reconstructed from this repository alone.

## Files left untouched

The pre-existing worktree changes, implementation files, tests, `AGENTS.md`,
and `go1.27.md` were left untouched. The UDP beacon decision remains recorded
as intentional in `INTENTIONAL-DIVERGENCES.md`.
