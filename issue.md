---
title: Parse multi sub declarations and select a candidate by arity and argument types
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-4139-773b-8676-e148e90c42bd
blocks:
- 01a0fd9b-4ca3-7d55-82ac-ba125816ee71
- 01a0fd9b-5017-78d0-ae38-331cbe6a9866
created: 2026-10-02T17:13:27.899092724Z
updated: 2026-10-02T17:24:05.20119943Z
---

## Prerequisites

- [ ] Parse typed declarations in .pmt files: typed parameters, defaults and return types

## Context

- paths: `internal/parse/decl.go`, `internal/parse/exports.go`, `internal/parse/pmt_test.go`, `internal/types/declared.go`, `internal/types/select.go` (new), `internal/types/select_test.go` (new)
- commands: `go test ./internal/types/ -count=1`, `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `parseDeclaration`/`declarators` (`internal/parse/decl.go:15-26`), `readSubs` (`internal/parse/exports.go:211`), `Join` (`internal/types/join.go:35`)

RFC 0001 "Multi declarations": `multi sub NAME ...` gives several signatures for one name, and a call site selects one by arity or by argument types that are not subtypes of each other; when it cannot decide, the consumer joins the candidates' results. Bounded polymorphism needs no `multi`: `abs (Num $x) Num` of an `Int` is `meet(join(arguments), declared)`, an `Int`. Selection lives in `internal/types/select.go` and is tested on its own, with `types.Signature` values, so it does not wait on internal/infer (m2) or on CORE.pmt. `multi` is read only in a `.pmt`; in ordinary source the word means what it means today. Selection by context is the `:context` issue's; that issue re-marks this section.

## Steps

- [ ] Write failing test `TestPmtMultiSubKeepsEveryCandidate` in `internal/parse/pmt_test.go`: two `multi sub each` lines (`(Hash %h) List`, `(Array @a) List`) are both kept, in order, under `each`
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestPmtMultiSubKeepsEveryCandidate$' -count=1 -v`
- [ ] Implement `multi sub` in `.pmt` mode; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestSelectByArity` in `internal/types/select_test.go` (select's `(FileHandle $fh) Str` against `($r, $w, $e, Num $timeout) Int`)
- [ ] Run test to verify failure: `go test ./internal/types/ -run '^TestSelectByArity$' -count=1 -v`
- [ ] Implement in `internal/types/select.go` (ABOUTME header); run test to verify pass; `make test`; commit
- [ ] Write failing test `TestSelectByArgumentTypes` (each's `Hash` against `Array`)
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestSelectJoinsWhenUndecided`: an argument of unknown type yields the join of the candidates' returns
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestBoundedResultIsMeetOfJoin`: `abs (Num $x) Num` applied to `Int` gives `Int`
- [ ] Implement `Meet` and the result rule; run test to verify pass; `make test`; commit

### Negative Scenarios

- outside a `.pmt`, `multi` is not a declarator: `sub multi { 1 } multi(1);` in ordinary source parses as it does today and records no candidates (`go test ./internal/parse/ -run '^TestMultiOnlyInPmt$' -count=1 -v | grep -q '^--- PASS: TestMultiOnlyInPmt'`)
- candidates whose argument types are subtypes of each other (`(Num $x) Num` and `(Int $x) Int`) are not chosen by most-specific type: an `Int` argument leaves the call undecided, so the result is the join (RFC: selection uses types that are not subtypes of each other) (`go test ./internal/types/ -run '^TestSelectSubtypeRelatedCandidatesJoin$' -count=1 -v | grep -q '^--- PASS: TestSelectSubtypeRelatedCandidatesJoin'`)
- selection over zero candidates or an empty argument list returns no type without panicking (`go test ./internal/types/ -run '^TestSelectEmptyCandidates$' -count=1 -v | grep -q '^--- PASS: TestSelectEmptyCandidates'`)
- a call whose arity no candidate accepts (`each` with two arguments) -- no type, or the join? The RFC only says an undecided call joins; needs perigrin's ruling (`go test ./internal/types/ -run '^TestSelectNoArityMatch$' -count=1 -v | grep -q '^--- PASS: TestSelectNoArityMatch'`)

## Acceptance criteria

- [ ] a multi sub keeps every candidate for its name (`go test ./internal/parse/ -run '^TestPmtMultiSubKeepsEveryCandidate$' -count=1 -v | grep -q '^--- PASS: TestPmtMultiSubKeepsEveryCandidate'`)
- [ ] a call selects a multi candidate by arity (`go test ./internal/types/ -run '^TestSelectByArity$' -count=1 -v | grep -q '^--- PASS: TestSelectByArity'`)
- [ ] a call selects a multi candidate by argument types (`go test ./internal/types/ -run '^TestSelectByArgumentTypes$' -count=1 -v | grep -q '^--- PASS: TestSelectByArgumentTypes'`)
- [ ] an undecided call joins the candidates (`go test ./internal/types/ -run '^TestSelectJoinsWhenUndecided$' -count=1 -v | grep -q '^--- PASS: TestSelectJoinsWhenUndecided'`)
- [ ] a bounded declaration returns the meet of the joined arguments and the declared type (`go test ./internal/types/ -run '^TestBoundedResultIsMeetOfJoin$' -count=1 -v | grep -q '^--- PASS: TestBoundedResultIsMeetOfJoin'`)
- [ ] multi is not a declarator in ordinary source (`go test ./internal/parse/ -run '^TestMultiOnlyInPmt$' -count=1 -v | grep -q '^--- PASS: TestMultiOnlyInPmt'`)
- [ ] subtype-related candidates do not select by most-specific type (`go test ./internal/types/ -run '^TestSelectSubtypeRelatedCandidatesJoin$' -count=1 -v | grep -q '^--- PASS: TestSelectSubtypeRelatedCandidatesJoin'`)
- [ ] selection with no candidates returns no type (`go test ./internal/types/ -run '^TestSelectEmptyCandidates$' -count=1 -v | grep -q '^--- PASS: TestSelectEmptyCandidates'`)
