---
title: Fold constant arithmetic so a known Inf or NaN is typed as such
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-14T13:23:06.421431472Z
updated: 2026-10-16T18:48:35.779781348Z
---

## Context

- paths: `internal/infer/lowered_ops.go` (Binary typing), `internal/types/CORE.tpm` (`**` returns `Num|NaN|Inf`)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-14 while fixing 01a135fc. perigrin ruled that a type that only MIGHT be Inf or NaN is checked as finite, and only a value KNOWN to be Inf or NaN is reported. PSC never types a value as known Inf or NaN today: `9**9**9` is `Int|Num|NaN|Inf` like any `**` result, so perl's own warning case goes unreported -- `"ab" x 9**9**9` warns "Non-finite repeat count does nothing" under perl 5.42 -w, as does `"ab" x ((-1)**0.5)`. Folding constant arithmetic (literal operands only, as perl's own constant folding does at compile time) would type such expressions exactly `Inf` or `NaN`, and the ruled rule would then report them.

## Steps

- [ ] Failing test first; fold Binary/Unary over numeric literals to their exact value's type (Int, Num, Inf, NaN), matching perl's compile-time folding; a non-finite repeat count from a folded constant is reported; `make test`; commit

## Acceptance criteria

- [ ] a constant non-finite repeat count is reported (`go test ./internal/infer/ -run '^TestFoldedNonFiniteCountIsReported$' -count=1 -v | grep -q '^--- PASS: TestFoldedNonFiniteCountIsReported'`)
