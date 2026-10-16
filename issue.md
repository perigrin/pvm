---
title: Select a single fitting-arity candidate despite argument types, and join an unknown context
state: done
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-d912-7419-8b20-a8e24029b3ec
blocks:
- 01a0fd9b-5700-7514-a143-2caed2d06d8b
- 01a1464e-de34-7403-95d2-25f4855404b2
created: 2026-10-16T20:02:15.178619007Z
updated: 2026-10-16T22:39:38.282088763Z
sessions:
- start_sha: ec99871d43d51af75b3cdcb95b6ddca2a6802abd
  end_sha: 7c74da24f7255401887ced54292a2dbee9bf6434
  commits: 4
  started_at: 2026-10-16T21:46:05.515582804Z
  ended_at: 2026-10-16T22:33:36.599392996Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T21:46:05.515582804Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-16T22:33:36.599392996Z
observed_paths:
- docs/rfcs/0001-type-libraries.md
- internal/lexer/proto.go
- internal/parse/decl.go
- internal/parse/declarations.go
- internal/parse/exportattr_test.go
- internal/parse/exports.go
- internal/parse/parse.go
- internal/parse/ranking_test.go
- internal/parse/testdata/core_module_ranking.txt
- internal/parse/testdata/declarations/My/Exports.tpm
- internal/parse/testdata/declarations/My/Vars.tpm
- internal/parse/typedour_test.go
- internal/perltest/perl542_test.go
- internal/psc/inc_roots_test.go
- internal/psc/lib_baseline_test.go
- internal/types/select.go
- internal/types/select_test.go
---

## Prerequisites

- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one

## Context

- paths: `internal/types/select.go`, `internal/types/select_test.go`
- commands: `go test ./internal/types/ -run '^TestSelect' -count=1 -v`
- entrypoints: `types.Select`, `types.SelectShaped` (`internal/types/select.go`, the `UnknownCtx` check near line 122)

Two rulings issue 01a0fd9b-5700 records as open, settled in `types` before call sites consume `Select`. (1) A parameter type is its coercion target (RFC 0001 "What a parameter type means"), so argument types decide only BETWEEN candidates: a single candidate whose arity fits is selected even when an argument's type does not fit it (`sqrt(Str)`), and the call site reports `coercion-mismatch`; today `Select` returns `Failed` with type Unknown. (2) RFC "Multi declarations": the consumer joins when the call site cannot decide, and `Failed` means ruled out, which psc check would report. A call whose context is `UnknownCtx` is Undecided (join of the candidates), not `Failed`. (3) Following (1): with two or more candidates whose arity fits and a concrete argument type that fits none of them, no candidate is ruled out, since a type mismatch never rules a call out -- only arity does. `Select` returns Undecided with the join of those candidates' returns, never `Failed`, and the call site reports `coercion-mismatch` for the argument, as it does for one candidate. (This extends ruling (1) to several candidates; it is the planner's reading, not a separate ruling.) Where perl decides a question, run perl rather than escalating it.

**Resolved 2026-10-16 by measurement (perl 5.42.0 is the judge):** the coercion reading applies to plain value parameters only. `sqrt("abc")` and `"a" x *STDOUT` warn and run, so a type mismatch there never rules a candidate out. A backslashed (Alias) container parameter is not a coercion target: `each $x`, `keys $x` and `push $x, 1` are refused at compile time ("Type of arg 1 to each must be hash or array"), so a mismatch in an Alias parameter still rules the candidate out, and TestCoreKeysValuesRefuseScalar and TestCoreEachRefusesScalar stand unchanged.

## Steps

- [ ] Write failing test `TestSelectSingleCandidateTakesMismatchedArgs`: one candidate `(Num $x) Num`, argument `Str`, selects that candidate and returns Num; with two candidates of the same arity, argument types still decide between them
- [ ] Run to verify failure: `go test ./internal/types/ -run '^TestSelectSingleCandidateTakesMismatchedArgs$' -count=1 -v`
- [ ] Implement; run to verify pass; commit
- [ ] Write failing test `TestSelectArgFittingNoSameArityCandidateJoins`: candidates `(Int $x) Int` and `(Code $c) Str`, given an argument typed `FileHandle` (a member of neither parameter type), give Undecided and `Int|Str`, not Failed
- [ ] Implement; run to verify pass; commit
- [ ] Write failing test `TestSelectUnknownContextIsUndecided`: `localtime`'s candidates under `UnknownCtx` give Undecided and the join of their returns
- [ ] Run to verify failure: `go test ./internal/types/ -run '^TestSelectUnknownContextIsUndecided$' -count=1 -v`
- [ ] Implement; run to verify pass; fix any existing select test whose expectation encoded the old behaviour, and say so in the commit
- [ ] `make test`; commit

### Review Findings

Tier-2 gate, 2026-10-16, on 7c74da24 (fresh reviewer subagent; six overlay mutations each caught; perl 5.42 measurements reconfirmed). VERDICT: REOPEN.

1. (in scope, confidence 80) `internal/types/select.go:80` the fallback filters in the wrong order: `Taking(cands, unknown, ...)` applies `answering(ctx)` before `fitsContainers` removes Alias-mismatched candidates, so a mismatched Alias candidate can win the context filter, crowd out a value candidate that would coerce, and then be deleted, leaving Failed. Probe: candidates `sig(List, Hash \%h)` and `sig(Str, Num $x)`, argument `Str`: ListCtx gives Failed, though perl would run the value candidate with a coercion warning. The strict path removes non-fitting candidates before `answering`; the fallback must match. Fix: drop Alias-mismatched candidates (fitsContainers on the real args) before `Taking(..., unknown, ...)`; add the probe as a ListCtx test that fails first.

## Acceptance Criteria

### Positive Scenarios
- [ ] a single fitting-arity candidate is selected despite a mismatched argument type (`go test ./internal/types/ -run '^TestSelectSingleCandidateTakesMismatchedArgs$' -count=1 -v | grep -q '^--- PASS: TestSelectSingleCandidateTakesMismatchedArgs'`)
- [ ] an unknown context joins instead of failing (`go test ./internal/types/ -run '^TestSelectUnknownContextIsUndecided$' -count=1 -v | grep -q '^--- PASS: TestSelectUnknownContextIsUndecided'`)

### Negative Scenarios
- [ ] a single candidate whose arity does not fit is still Failed; mismatched argument types never rescue a wrong arity (`go test ./internal/types/ -run '^TestSelectSingleCandidateWrongArityFails$' -count=1 -v | grep -q '^--- PASS: TestSelectSingleCandidateWrongArityFails'`)
- [ ] two same-arity candidates given a concrete argument type that fits neither are Undecided with the join of their returns, not Failed (`go test ./internal/types/ -run '^TestSelectArgFittingNoSameArityCandidateJoins$' -count=1 -v | grep -q '^--- PASS: TestSelectArgFittingNoSameArityCandidateJoins'`)
- [ ] two same-arity candidates given an argument of Unknown type are Undecided with the join of their returns, not Failed and not the first candidate (`go test ./internal/types/ -run '^TestSelectUnknownArgBetweenCandidatesJoins$' -count=1 -v | grep -q '^--- PASS: TestSelectUnknownArgBetweenCandidatesJoins'`)
- [ ] a single candidate under `UnknownCtx` is selected, not Failed (`go test ./internal/types/ -run '^TestSelectUnknownContextSingleCandidate$' -count=1 -v | grep -q '^--- PASS: TestSelectUnknownContextSingleCandidate'`)
- [ ] `SelectShaped` under `UnknownCtx` is Undecided, as `Select` is (`go test ./internal/types/ -run '^TestSelectShapedUnknownContextIsUndecided$' -count=1 -v | grep -q '^--- PASS: TestSelectShapedUnknownContextIsUndecided'`)
- [ ] an empty candidate list is Failed with type Unknown and does not panic (`go test ./internal/types/ -run '^TestSelectEmptyCandidatesFails$' -count=1 -v | grep -q '^--- PASS: TestSelectEmptyCandidatesFails'`)
