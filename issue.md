---
title: Read :context(...) and select a multi candidate by calling context
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-4a5b-7e94-ad87-756dfcfe1f4c
blocks:
- 01a0fd9b-519d-7b42-886e-2222a3076986
- 01a0fd9b-5468-7355-a22e-1b16fa5ee354
- 01a0fd9b-5698-70cf-b6e4-8899a83b0fe0
- 01a10177-4e7b-7d2c-8d2a-2033647ffc4a
- 01a10c6e-988f-7aa6-811d-0e462abc442a
created: 2026-10-02T17:13:28.48300713Z
updated: 2026-10-05T14:20:14.950817876Z
---

## Prerequisites

- [ ] Parse multi sub declarations and select a candidate by arity and argument types

## Context

- paths: `internal/parse/decl.go`, `internal/parse/pmt_test.go`, `internal/types/declared.go`, `internal/types/select.go`, `internal/types/select_test.go`, `docs/rfcs/0001-type-libraries.md`
- commands: `go test ./internal/types/ -count=1`, `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `internal/types/select.go`; how `:prototype(...)` is read in `declaredSub` (`internal/parse/exports.go:425`)

RFC 0001 "`:context(...)`": a declaration names the contexts it answers for in prototype sigils -- no attribute: every context; `:context($)` scalar; `:context(@)` list; `:context()` void; `:context($@)` scalar or list. As with `:prototype`, absent and empty differ. A call in a context no candidate answers for has no type (`sub sort :context(@) ...`: scalar sort is undefined). Boolean context counts as scalar (measured: `wantarray` reports scalar inside `if (f())`, `!f()`, `f() and ...`). Context selects a declaration; narrowing cannot stand in for it (`my $t = localtime(0)` is a date string). This completes the selection rules, built and tested on their own, so it re-marks both "Multi declarations" and "`:context(...)`" (perigrin, 2026-10-03). Applying the rules at a real call in `infer` is RFC "Call sites", the call-site issue's; refusing at parse time what perl refuses at compile time is the parse-time refusal issue's.

## Steps

- [ ] Write failing test `TestPmtContextAttribute`: `:context($)`, `:context(@)`, `:context()`, `:context($@)` and no attribute are read as five distinct context sets
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestPmtContextAttribute$' -count=1 -v`
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestSelectByContext`: localtime's `:context($) ... Str` and `:context(@) ... List[Int]` select by scalar, list, and boolean (as scalar) context
- [ ] Run test to verify failure: `go test ./internal/types/ -run '^TestSelectByContext$' -count=1 -v`
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestSelectNoContextAnswerHasNoType`: a scalar call to a `:context(@)`-only declaration has no type
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Re-mark the RFC headings "Multi declarations" and "`:context(...)`" `(*Implemented*, <sha>)`; commit

### Negative Scenarios

- malformed `:context` spellings are errors: `:context(%)`, `:context(x)`, an unclosed `:context(` (`go test ./internal/parse/ -run '^TestPmtContextAttributeMalformed$' -count=1 -v | grep -q '^--- PASS: TestPmtContextAttributeMalformed'`)
- void context is answered only by an absent attribute or `:context()`: a void call to candidates that are only `:context($)` and `:context(@)` has no type (`go test ./internal/types/ -run '^TestSelectVoidContextUnanswered$' -count=1 -v | grep -q '^--- PASS: TestSelectVoidContextUnanswered'`)
- boolean context is scalar, so a boolean call to a `:context(@)`-only declaration has no type (it does not fall back to the list candidate) (`go test ./internal/types/ -run '^TestSelectBooleanIsNotList$' -count=1 -v | grep -q '^--- PASS: TestSelectBooleanIsNotList'`)

## Acceptance criteria

- [ ] :context reads every spelling, and absent differs from empty (`go test ./internal/parse/ -run '^TestPmtContextAttribute$' -count=1 -v | grep -q '^--- PASS: TestPmtContextAttribute'`)
- [ ] a call selects a candidate by its calling context (`go test ./internal/types/ -run '^TestSelectByContext$' -count=1 -v | grep -q '^--- PASS: TestSelectByContext'`)
- [ ] a call in a context no candidate answers for has no type (`go test ./internal/types/ -run '^TestSelectNoContextAnswerHasNoType$' -count=1 -v | grep -q '^--- PASS: TestSelectNoContextAnswerHasNoType'`)
- [ ] the RFC section on multi declarations is implemented (`grep -q '^### Multi declarations.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] the RFC section on :context is implemented (`grep -q '^### .:context.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] malformed :context spellings are errors (`go test ./internal/parse/ -run '^TestPmtContextAttributeMalformed$' -count=1 -v | grep -q '^--- PASS: TestPmtContextAttributeMalformed'`)
- [ ] a void call to scalar and list candidates has no type (`go test ./internal/types/ -run '^TestSelectVoidContextUnanswered$' -count=1 -v | grep -q '^--- PASS: TestSelectVoidContextUnanswered'`)
- [ ] a boolean call to a list-only declaration has no type (`go test ./internal/types/ -run '^TestSelectBooleanIsNotList$' -count=1 -v | grep -q '^--- PASS: TestSelectBooleanIsNotList'`)
