---
title: A compound assignment narrows to its right side's type, as legacy did
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-13T03:45:34.152292807Z
updated: 2026-10-13T03:45:34.152292807Z
---

## Context

- paths: `internal/infer/lowered_ops.go` (`narrow`, Assign), `internal/types/CORE.pmt` (assignment-operator signatures)
- commands: `go test ./internal/infer/ -run '^TestLoweredGoldenRatchet$'`

Found 2026-10-13 by the operator port (01a12a85-3233), which copied it from legacy so the goldens match: a compound assignment narrows its target to the RIGHT side's type. `my $s = "a"; $s .= 1;` leaves `$s` Int; perl 5.42 gives `s=a1` with POK=1, IOK=0 -- a Str. Likewise `my $t = "ab"; $t x= 3` is `ababab`, a Str. The fix: a compound assignment `op=` narrows to the return type of `op` from CORE.pmt applied to both operands. Fix after the cutover (01a0c115-df67); each golden edit gets a `Diagnostics-changed:` trailer.

## Steps

- [ ] Failing test first; `op=` narrows to `op`'s CORE.pmt return type; goldens under trailers; precision must not fall; `make test`; commit

## Acceptance criteria

- [ ] `$s .= 1` leaves `$s` Str (`go test ./internal/infer/ -run '^TestCompoundAssignNarrowsToOperatorResult$' -count=1 -v | grep -q '^--- PASS: TestCompoundAssignNarrowsToOperatorResult'`)
