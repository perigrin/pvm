---
title: The three parse ratchet headers are derived on write and skipped on read, so each can drift silently
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-26T13:50:43.493701239Z
updated: 2026-10-09T04:48:37.141255104Z
sessions:
- start_sha: 98d028499fdad30e12a1e2e54cc03ddb455833f4
  end_sha: ""
  commits: 0
  started_at: 2026-10-09T04:48:37.141255104Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T04:48:37.141255104Z
---

The three parse-side ratchet headers are derived on write and skipped on read,
so each can drift from the case lines beside it and nothing fails. The corpus
ratchet had the same defect and is fixed; these three are the same code shape
in a file a worker holds right now.

## Measured, on the corpus ratchet

Before its fix, the committed baseline's header was edited to

    # 999 cases, 1 clean, 998 refusing.

and `TestCorpusRatchet` PASSED, in 0.04s. The header is the one number every
report of that corpus quotes and nothing verified it.

Root cause: `internal/conformance/ratchet.go:261`'s `parseRatchetFile` skips
every line beginning `#`, and `renderRatchet` writes the summary as a `#`
line. Derived on write, ignored on read.

Fixed by `TestRatchetHeaderAgreesWithItsBody`, which counts the body and
compares. RED was confirmed against the corrupted header before the fix
landed.

## The same shape, three more files

`internal/parse/ratchet_test.go:278`:

    if line == "" || strings.HasPrefix(line, "#") {

That reader serves `t1.ratchet`, `t1canon.ratchet` and `t1fidelity.ratchet`,
whose headers carry the figures this milestone has been quoting all session:

    # 986 files, 523 clean (53.0%), 2188 total.
    # 986 files, 684 clean (69.4%), 302 total.
    # 986 files, 360 clean (36.5%), 626 total.

Each includes a PERCENTAGE, which the corpus header does not, so the check has
one more field to agree on and one more way to drift.

## Why it is worth doing rather than noting

A baseline that recalls rather than recomputes is right when written and wrong
the moment the thing it counts changes, silently in both directions. That is
the same failure as `t1canon.ratchet` recording 46 canon-broken files among
its 302 disagreeing as though that were the shape of the world -- the
instrument measuring the symptom as the baseline -- which is what `01a0dde4`
is about.

Two instruments in this repo have now been found reporting a number nobody
checked. The third through fifth are this issue.

## Scope

1. A header-agrees-with-body test for each of the three parse ratchets,
   modelled on `TestRatchetHeaderAgreesWithItsBody` in
   `internal/conformance/ratchet_test.go`. Confirm RED by corrupting each
   header before landing the check -- a test that has never failed has not
   been tested.
2. The percentage is derived from two counts in the same line, so it can
   disagree with its own numerator and denominator. Check it too.
3. Consider whether the reader should reject an unrecognised `#` line that
   LOOKS like a summary, rather than only checking the one it finds. Out of
   scope if it complicates the reader; say which.

## Context

- paths: `internal/parse/ratchet_test.go` (the reader at 278, and whichever
  functions render the three headers), `internal/conformance/ratchet_test.go`
  for the model
- commands: `go test ./internal/parse/ -count=1`,
  `go test ./internal/conformance/ -count=1`
- NOTE: `internal/parse/ratchet_test.go` is held by the worker on `01a0dde4`
  as of 2026-09-26. Do not start until that lands.

## Acceptance criteria

- [ ] Each parse ratchet's header is checked against its body (`go test ./internal/parse/ -run 'HeaderAgrees' -count=1 -v | grep -cq '^--- PASS'`)
- [ ] The percentage is checked, not just the counts (`go test ./internal/parse/ -run 'HeaderAgrees' -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the parse ratchets hold (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the corpus ratchet holds (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
