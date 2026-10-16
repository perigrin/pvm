---
title: A guard on one variable does not narrow a variable always assigned with it
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-16T15:49:27.38326889Z
updated: 2026-10-16T15:49:27.38326889Z
---

## Context

- found by: R1's re-sweep of perl 5.42's lib (2026-10-16); perigrin accepted the line on PR #486 and asked for this issue
- paths: `internal/infer/flow.go` (guard narrowing, loop joins)

Sys/Syslog.pm:507: `$fail_time` is set whenever `$failed` is, and code guarded by `$failed` reads `$fail_time` -- psc reports `$fail_time` as possibly Undef there; perl never warns. The link between two variables (one set iff the other is) is path-sensitive information PSC does not track. Widening variables assigned in a loop would hide true positives reported today, so the fix is to record such correlations (a guard on one variable narrows the variables assigned in the same branch that set it).

## Steps

- [ ] Failing test first from a minimal Syslog shape; correlate variables assigned together under one condition; no true positive lost (re-sweep perl's lib); `make test`; commit

## Acceptance criteria

- [ ] a guard on one variable narrows a variable always assigned with it (`go test ./internal/infer/ -run '^TestCorrelatedVariablesNarrowTogether$' -count=1 -v | grep -q '^--- PASS: TestCorrelatedVariablesNarrowTogether'`)
