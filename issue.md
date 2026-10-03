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
updated: 2026-10-03T11:13:45.622955274Z
---

## Prerequisites

- [ ] Parse return types and type expressions in .pmt declarations

## Context

- paths: `internal/parse/decl.go`, `internal/parse/exports.go`, `internal/parse/pmt_test.go`, `internal/types/declared.go`, `internal/types/select.go` (new), `internal/types/select_test.go` (new)
- commands: `go test ./internal/types/ -count=1`, `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `parseDeclaration`/`declarators` (`internal/parse/decl.go:15-26`), `readSubs` (`internal/parse/exports.go:211`), `Join` (`internal/types/join.go:35`), `IsSubtype` (`internal/types/`)

RFC 0001 "Multi declarations": `multi sub NAME ...` gives several signatures for one name, and selection picks one by arity or by argument types. Rulings of 2026-10-02: **the most specific candidate wins** -- when candidates whose parameters are related by subtyping both accept a call (`(Int $x) Str` beside `(Num $x) Int`, called with an `Int`), the one whose parameters are all subtypes of the other's is selected, as in Raku, CLOS and Julia; candidates with no single most specific one for some argument types (`(Int $a, Num $b)` beside `(Num $a, Int $b)`, called with two `Int`s) are ambiguous, and the declaration is an error. **A call whose arity no candidate accepts fails**: selection reports failure (no candidate and no join), never the join. The join applies only when the call site cannot decide, for example an argument of unknown type. Bounded polymorphism needs no `multi`: `abs (Num $x) Num` of an `Int` is `meet(join(arguments), declared)`, an `Int`. Selection lives in `internal/types/select.go`, tested on its own with `types.Signature` values built in Go (so `each`'s `Hash \%h`/`Array \@a` candidates need no backslash parse here), and does not wait on internal/infer (m2) or on CORE.pmt. Its result distinguishes three outcomes: a selected candidate, an undecided join, and a failure. `multi` is read only in a `.pmt`; in ordinary source the word means what it means today. The parser refusing a call perl refuses at compile time is the parse-time refusal issue's; applying selection at real call sites is the call-site issue's; selection by context is the `:context` issue's, which re-marks this section.

## Steps

- [ ] Write failing test `TestPmtMultiSubKeepsEveryCandidate` in `internal/parse/pmt_test.go`: two `multi sub select` lines (`(FileHandle $fh) Str`, `($r, $w, $e, Num $timeout) Int`) are both kept, in order, under `select`
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestPmtMultiSubKeepsEveryCandidate$' -count=1 -v`
- [ ] Implement `multi sub` in `.pmt` mode; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestSelectByArity` in `internal/types/select_test.go` (select's `(FileHandle $fh) Str` against `($r, $w, $e, Num $timeout) Int`)
- [ ] Run test to verify failure: `go test ./internal/types/ -run '^TestSelectByArity$' -count=1 -v`
- [ ] Implement in `internal/types/select.go` (ABOUTME header); run test to verify pass; `make test`; commit
- [ ] Write failing test `TestSelectByArgumentTypes` (each's `Hash` against `Array`)
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestSelectMostSpecificCandidate`: `(Int $x) Str` beside `(Num $x) Int`, called with an `Int`, selects the `Int` candidate; called with a `Num`, the `Num` one
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestMultiAmbiguousCandidatesIsError` in `internal/parse/pmt_test.go`: a `.pmt` declaring `multi sub f (Int $a, Num $b) Str;` and `multi sub f (Num $a, Int $b) Str;` is an error naming `f`; adding `multi sub f (Int $a, Int $b) Str;` makes it valid
- [ ] Implement the ambiguity check (in `internal/types`, called by the `.pmt` reader); run test to verify pass; `make test`; commit
- [ ] Write failing test `TestSelectNoArityMatch`: `each` called with two arguments reports failure, with no candidate and no joined type
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestSelectJoinsWhenUndecided`: an argument of unknown type yields the join of the candidates' returns
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestBoundedResultIsMeetOfJoin`: `abs (Num $x) Num` applied to `Int` gives `Int`
- [ ] Implement `Meet` and the result rule; run test to verify pass; `make test`; commit

### Negative Scenarios

- outside a `.pmt`, `multi` is not a declarator: `sub multi { 1 } multi(1);` in ordinary source parses as it does today and records no candidates (`go test ./internal/parse/ -run '^TestMultiOnlyInPmt$' -count=1 -v | grep -q '^--- PASS: TestMultiOnlyInPmt'`)
- selection over zero candidates or an empty argument list returns no type without panicking (`go test ./internal/types/ -run '^TestSelectEmptyCandidates$' -count=1 -v | grep -q '^--- PASS: TestSelectEmptyCandidates'`)
- an arity failure is not rescued by an unknown argument type: `each` with two arguments of unknown type still fails rather than joining (`go test ./internal/types/ -run '^TestSelectArityFailureIsNotJoined$' -count=1 -v | grep -q '^--- PASS: TestSelectArityFailureIsNotJoined'`)
- candidates with disjoint parameter types (`(Hash \%h)` and `(Array \@a)`) are not ambiguous, and identical duplicate candidates are an ambiguity error (`go test ./internal/types/ -run '^TestMultiAmbiguityEdgeCases$' -count=1 -v | grep -q '^--- PASS: TestMultiAmbiguityEdgeCases'`)

## Acceptance criteria

- [ ] a multi sub keeps every candidate for its name (`go test ./internal/parse/ -run '^TestPmtMultiSubKeepsEveryCandidate$' -count=1 -v | grep -q '^--- PASS: TestPmtMultiSubKeepsEveryCandidate'`)
- [ ] a call selects a multi candidate by arity (`go test ./internal/types/ -run '^TestSelectByArity$' -count=1 -v | grep -q '^--- PASS: TestSelectByArity'`)
- [ ] a call selects a multi candidate by argument types (`go test ./internal/types/ -run '^TestSelectByArgumentTypes$' -count=1 -v | grep -q '^--- PASS: TestSelectByArgumentTypes'`)
- [ ] the most specific of subtype-related candidates is selected (`go test ./internal/types/ -run '^TestSelectMostSpecificCandidate$' -count=1 -v | grep -q '^--- PASS: TestSelectMostSpecificCandidate'`)
- [ ] candidates with no single most specific one are a declaration error (`go test ./internal/parse/ -run '^TestMultiAmbiguousCandidatesIsError$' -count=1 -v | grep -q '^--- PASS: TestMultiAmbiguousCandidatesIsError'`)
- [ ] a call whose arity no candidate accepts fails, with no join (`go test ./internal/types/ -run '^TestSelectNoArityMatch$' -count=1 -v | grep -q '^--- PASS: TestSelectNoArityMatch'`)
- [ ] an undecided call joins the candidates (`go test ./internal/types/ -run '^TestSelectJoinsWhenUndecided$' -count=1 -v | grep -q '^--- PASS: TestSelectJoinsWhenUndecided'`)
- [ ] a bounded declaration returns the meet of the joined arguments and the declared type (`go test ./internal/types/ -run '^TestBoundedResultIsMeetOfJoin$' -count=1 -v | grep -q '^--- PASS: TestBoundedResultIsMeetOfJoin'`)
- [ ] multi is not a declarator in ordinary source (`go test ./internal/parse/ -run '^TestMultiOnlyInPmt$' -count=1 -v | grep -q '^--- PASS: TestMultiOnlyInPmt'`)
- [ ] selection with no candidates returns no type (`go test ./internal/types/ -run '^TestSelectEmptyCandidates$' -count=1 -v | grep -q '^--- PASS: TestSelectEmptyCandidates'`)
- [ ] an arity failure is not joined even for unknown argument types (`go test ./internal/types/ -run '^TestSelectArityFailureIsNotJoined$' -count=1 -v | grep -q '^--- PASS: TestSelectArityFailureIsNotJoined'`)
- [ ] disjoint candidates are not ambiguous and duplicates are (`go test ./internal/types/ -run '^TestMultiAmbiguityEdgeCases$' -count=1 -v | grep -q '^--- PASS: TestMultiAmbiguityEdgeCases'`)
