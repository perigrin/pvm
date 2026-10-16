---
title: Select a single fitting-arity candidate despite argument types, and join an unknown context
state: pending
urgency: normal
milestone: core-module-types
created: 2026-10-16T20:02:15.178619007Z
updated: 2026-10-16T20:02:15.178619007Z
---

## Prerequisites

- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one

## Context

- paths: `internal/types/select.go`, `internal/types/select_test.go`
- commands: `go test ./internal/types/ -run '^TestSelect' -count=1 -v`
- entrypoints: `types.Select`, `types.SelectShaped` (`internal/types/select.go`, the `UnknownCtx` check near line 122)

Two rulings issue 01a0fd9b-5700 records as open, settled in `types` before call sites consume `Select`. (1) A parameter type is its coercion target (RFC 0001 "What a parameter type means"), so argument types decide only BETWEEN candidates: a single candidate whose arity fits is selected even when an argument's type does not fit it (`sqrt(Str)`), and the call site reports `coercion-mismatch`; today `Select` returns `Failed` with type Unknown. (2) RFC "Multi declarations": the consumer joins when the call site cannot decide, and `Failed` means ruled out, which psc check would report. A call whose context is `UnknownCtx` is Undecided (join of the candidates), not `Failed`. Where perl decides a question, run perl rather than escalating it.

## Steps

- [ ] Write failing test `TestSelectSingleCandidateTakesMismatchedArgs`: one candidate `(Num $x) Num`, argument `Str`, selects that candidate and returns Num; with two candidates of the same arity, argument types still decide between them
- [ ] Run to verify failure: `go test ./internal/types/ -run '^TestSelectSingleCandidateTakesMismatchedArgs$' -count=1 -v`
- [ ] Implement; run to verify pass; commit
- [ ] Write failing test `TestSelectUnknownContextIsUndecided`: `localtime`'s candidates under `UnknownCtx` give Undecided and the join of their returns
- [ ] Run to verify failure: `go test ./internal/types/ -run '^TestSelectUnknownContextIsUndecided$' -count=1 -v`
- [ ] Implement; run to verify pass; fix any existing select test whose expectation encoded the old behaviour, and say so in the commit
- [ ] `make test`; commit

## Acceptance Criteria

### Positive Scenarios
- [ ] a single fitting-arity candidate is selected despite a mismatched argument type (`go test ./internal/types/ -run '^TestSelectSingleCandidateTakesMismatchedArgs$' -count=1 -v | grep -q '^--- PASS: TestSelectSingleCandidateTakesMismatchedArgs'`)
- [ ] an unknown context joins instead of failing (`go test ./internal/types/ -run '^TestSelectUnknownContextIsUndecided$' -count=1 -v | grep -q '^--- PASS: TestSelectUnknownContextIsUndecided'`)
