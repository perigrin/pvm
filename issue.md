---
title: Remove the WRONG pin when the harness measures the new parser
state: pending
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0a70f-3ccc-77bb-902d-9a0b9d8050a6
created: 2026-09-16T06:00:39.116126082Z
updated: 2026-09-16T06:01:04.813439515Z
---

Two files are pinned as WRONG in the corpus baseline, and the pin comes out
when they do.

    comp/our.t       metric 8
    op/universal.t   metric 16

Both are defects in the TREE-SITTER grammar, which is what the harness still
measures (`run.go:381` calls `parser.New()`). Findings §0.11 has the analysis
and `marker_gaps_test.go` has the regression tests.

## Why they are pinned rather than zero

Commit `a69d8f1c` introduced these and left `TestCorpusBaselineIsIntact`
tripped on purpose, as a signal. It stayed tripped, and a permanently-red
test is one nobody reads -- worse, a red suite cannot show the NEXT
regression. An absolute zero that has never held stops being a gate and
becomes scenery.

The pin names both PATHS rather than a count. A count of 2 would let one
defect be fixed and a different one appear with the gate silent; measured,
that swap passes a count and fails the pin.

## What removes it

Either:

1. **The M1 gate lands** and the harness measures `internal/parse` instead of
   the tree-sitter grammar. This is the expected route, and `comp/our.t`'s
   cause is already fixed there -- the M1 lexer reads `if (/TIE/)` as
   `Quote("/TIE/")`, not division. Verified.
2. **Someone fixes the tree-sitter grammar's two defects** directly. Real
   work on a parser M1-M6 exist to replace, so probably not worth it.

Either way the closing move is the same: delete the pin block in
`ratchet_test.go`, restore `!= 0`, re-baseline in the same commit, and update
findings §0.11.

If only ONE is fixed, replace the pin with the remaining path rather than
loosening it to a count.

## Acceptance

- [ ] The WRONG pin is gone and the gate asserts zero (`go test ./internal/parseoracle/ -run '^TestCorpusBaselineIsIntact$' -count=1 -v | grep -q '^--- PASS: TestCorpusBaselineIsIntact'`)
- [ ] The baseline holds no WRONG rows (`grep -c '^WRONG' internal/parseoracle/testdata/ratchet/baseline.txt | grep -q '^0$'`)
- [ ] Findings §0.11 records the change with the measurement that produced it

Findings: §0.11. Introduced by a69d8f1c; pinned by the commit that closes this issue's parent.
