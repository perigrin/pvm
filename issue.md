---
title: Read :context(...) and select a multi candidate by calling context
state: done
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-4a5b-7e94-ad87-756dfcfe1f4c
blocks:
- 01a0fd9b-519d-7b42-886e-2222a3076986
- 01a0fd9b-5468-7355-a22e-1b16fa5ee354
- 01a10dc1-3499-7e2b-9e6f-d935084437fa
created: 2026-10-02T17:13:28.48300713Z
updated: 2026-10-08T18:40:54.136431018Z
sessions:
- start_sha: 007fbbffeff724026118b1ba246fcf5d177df5f2
  end_sha: 007fbbffeff724026118b1ba246fcf5d177df5f2
  commits: 0
  started_at: 2026-10-06T05:20:04.373014221Z
  ended_at: 2026-10-06T05:20:04.576979619Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-06T05:20:04.373014221Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-06T05:20:04.576979619Z
---

## Prerequisites

- [ ] Parse multi sub declarations and select a candidate by arity and argument types

## Context

- paths: `internal/parse/decl.go`, `internal/parse/pmt_test.go`, `internal/types/declared.go`, `internal/types/select.go`, `internal/types/select_test.go`, `docs/rfcs/0001-type-libraries.md`
- commands: `go test ./internal/types/ -count=1`, `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `internal/types/select.go`; how `:prototype(...)` is read in `declaredSub` (`internal/parse/exports.go:425`)

RFC 0001 "`:context(...)`": a declaration names the contexts it answers for in prototype sigils -- no attribute: every context; `:context($)` scalar; `:context(@)` list; `:context()` void; `:context($@)` scalar or list. As with `:prototype`, absent and empty differ. A call in a context no candidate answers for has no type (`sub sort :context(@) ...`: scalar sort is undefined). Boolean context counts as scalar (measured: `wantarray` reports scalar inside `if (f())`, `!f()`, `f() and ...`). Context selects a declaration; narrowing cannot stand in for it (`my $t = localtime(0)` is a date string). This completes the selection rules, built and tested on their own, so it re-marks both "Multi declarations" and "`:context(...)`" (perigrin, 2026-10-03). Under the RFC's convention (perigrin, 2026-10-05, commit 72e01c14), a section is Implemented when its rule is built and tested: the re-mark here covers the rule, and the sections' example declarations (`each` and `select`, `localtime`, `sort`) are illustrations that belong to the typing batches -- the no-prototype builtins issue, the process-and-time batch and the `grep`/`map`/`sort` issue -- which write them into CORE.pmt and test them against perl. The tests here use literal declarations in the test, not CORE.pmt's lines. Applying the rules at a real call in `infer` is RFC "Call sites", the call-site issue's; refusing at parse time what perl refuses at compile time is the parse-time refusal issue's.

## Steps

- [ ] Write failing test `TestPmtContextAttribute`: `:context($)`, `:context(@)`, `:context()`, `:context($@)` and no attribute are read as five distinct context sets
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestContextAttributeRefused$' -count=1 -v`
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestSelectByContext`: localtime's `:context($) ... Str` and `:context(@) ... List[Int]` select by scalar, list, and boolean (as scalar) context
- [ ] Run test to verify failure: `go test ./internal/types/ -run '^TestSelectByContext$' -count=1 -v`
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestSelectNoContextAnswerHasNoType`: a scalar call to a `:context(@)`-only declaration has no type
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Re-mark the RFC headings "Multi declarations" and "`:context(...)`" `(*Implemented*, <sha>)`; commit

### Negative Scenarios

- malformed `:context` spellings are errors: `:context(%)`, `:context(x)`, an unclosed `:context(` (`go test ./internal/parse/ -run '^TestContextAttributeRefused$' -count=1 -v | grep -q '^--- PASS: TestContextAttributeRefused'`)
- void context is answered only by an absent attribute or `:context()`: a void call to candidates that are only `:context($)` and `:context(@)` has no type (`go test ./internal/types/ -run '^TestSelectVoidContext$' -count=1 -v | grep -q '^--- PASS: TestSelectVoidContext'`)
- boolean context is scalar, so a boolean call to a `:context(@)`-only declaration has no type (it does not fall back to the list candidate) (`go test ./internal/types/ -run '^TestSelectBooleanIsNotList$' -count=1 -v | grep -q '^--- PASS: TestSelectBooleanIsNotList'`)

## Acceptance criteria

- [ ] :context reads every spelling, and absent differs from empty (`go test ./internal/parse/ -run '^TestContextAttributeRefused$' -count=1 -v | grep -q '^--- PASS: TestContextAttributeRefused'`)
- [ ] a call selects a candidate by its calling context (`go test ./internal/types/ -run '^TestSelectByContext$' -count=1 -v | grep -q '^--- PASS: TestSelectByContext'`)
- [ ] a call in a context no candidate answers for has no type (`go test ./internal/types/ -run '^TestSelectNoContextAnswerHasNoType$' -count=1 -v | grep -q '^--- PASS: TestSelectNoContextAnswerHasNoType'`)
- [ ] the RFC section on multi declarations is implemented (`grep -q '^### Multi declarations.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] the RFC section on :context is implemented (`grep -q '^### Context selects by return type (\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] malformed :context spellings are errors (`go test ./internal/parse/ -run '^TestContextAttributeRefused$' -count=1 -v | grep -q '^--- PASS: TestContextAttributeRefused'`)
- [ ] a void call to scalar and list candidates has no type (`go test ./internal/types/ -run '^TestSelectVoidContext$' -count=1 -v | grep -q '^--- PASS: TestSelectVoidContext'`)
- [ ] a boolean call to a list-only declaration has no type (`go test ./internal/types/ -run '^TestSelectBooleanIsNotList$' -count=1 -v | grep -q '^--- PASS: TestSelectBooleanIsNotList'`)

### Criteria superseded (2026-10-08)

01a11920-411a dropped `:context` (perigrin: "multi and return type handle it"): a `:context` attribute is now refused, and a multi selects by return type. This issue's criteria that named the attribute's tests now name their replacements: TestPmtContextAttribute and TestPmtContextAttributeMalformed -> TestContextAttributeRefused; TestSelectVoidContextUnanswered -> TestSelectVoidContext (void context with no Void candidate falls back to scalar); the RFC section `:context(...)` -> "Context selects by return type".
