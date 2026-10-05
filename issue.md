---
title: Derive one prototype from several multi candidates
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-4850-704d-8c14-c15add6194f2
created: 2026-10-05T20:28:48.152828461Z
updated: 2026-10-05T20:30:10.297377852Z
---

## Prerequisites

- [ ] Parse aliased parameters (\$x, \@a, \%h, \&c, \(@list)) and derive \$, \@, \% and \& from them
- [ ] Read :context(...) and select a multi candidate by calling context

## Context

- paths: `internal/parse/derive.go`, `internal/parse/derive_test.go`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `internal/parse/derive.go` (the single-declaration derivation and the aliasing rows); the `multi sub` candidate list the multi issue keeps per name, and the context set the `:context(...)` issue records on each candidate (this issue needs it to show that `:context` never splits a group, so it follows that issue, which itself follows the multi issue)

A prototyped builtin declared as a `multi` still has exactly one prototype, so its candidates must derive one. This issue adds that rule to `internal/parse/derive.go`, on its own and tested with literal declarations, so the CORE.pmt issues that declare multis (`each`; `keys` and `values`; `localtime` and `gmtime`) only use it. The rule groups candidates by their PARAMETERS only, ignoring `:context` and the return type: candidates whose parameters are identical are one group (`localtime`'s `:context($) ... Str` and `:context(@) ... List[Int]` over `(Int $time = time)` derive one `;$`); groups of the same arity that differ only in which container one aliased position takes derive `\[...]` with those containers in declaration order. So `each`'s `(Hash \%h)` and `(Array \@a)` derive `\[%@]`, and `keys`' four candidates -- `:context($)` returning `Int` and `:context(@)` returning `List`, each over `Hash \%h` and `Array \@a` -- collapse to two parameter groups and derive `\[%@]` as well (measured on 5.42: `prototype("CORE::each")`, `prototype("CORE::keys")` and `prototype("CORE::values")` are all `\[%@]`; `localtime` and `gmtime` are `;$`). Candidates that differ in more than one position, or in arity, derive no single prototype: that is an error where a prototype is required, reported, never guessed. Expected values live in the test. Re-marks no RFC section: the rule is part of "A typed signature and a prototype say the same thing", which the last CORE.pmt typing batch re-marks.

## Steps

- [ ] Write failing test `TestDerivePrototypeFromMultiCandidates` in `internal/parse/derive_test.go`: `multi sub each (Hash \%h) List;` and `multi sub each (Array \@a) List;` derive `\[%@]`
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestDerivePrototypeFromMultiCandidates$' -count=1 -v`
- [ ] Implement the rule in `internal/parse/derive.go`; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestDeriveMultiCandidatesGroupByParameters`: `keys`' four candidates (`:context($)` returning `Int` and `:context(@)` returning `List`, over `Hash \%h` and `Array \@a`) derive `\[%@]`; two `localtime`-style candidates with identical parameters `(Int $time = time)` and different `:context` and return types derive `;$`
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestDeriveMultiCandidatesGroupByParameters$' -count=1 -v`
- [ ] Group by parameters before comparing positions; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestDeriveMultiCandidatesRefusesInexpressible`, then implement the refusal; run test to verify pass; `make test`; commit

### Negative Scenarios

- the rule refuses what it cannot express: candidates that differ in two aliased positions, candidates of different arity, or a union perl would order differently from the declaration are reported as an error naming the sub, not guessed (`go test ./internal/parse/ -run '^TestDeriveMultiCandidatesRefusesInexpressible$' -count=1 -v | grep -q '^--- PASS: TestDeriveMultiCandidatesRefusesInexpressible'`)
- return type and `:context` never split a group: two candidates with the same parameters but different return types and no `:context` derive one prototype, and do not count as two positions differing (`go test ./internal/parse/ -run '^TestDeriveMultiCandidatesIgnoreReturnType$' -count=1 -v | grep -q '^--- PASS: TestDeriveMultiCandidatesIgnoreReturnType'`)
- a single candidate derives exactly what the plain derivation gives it: `multi sub f (Hash \%h) List;` alone derives `\%`, not `\[%]` (`go test ./internal/parse/ -run '^TestDeriveSingleMultiCandidateIsPlain$' -count=1 -v | grep -q '^--- PASS: TestDeriveSingleMultiCandidateIsPlain'`)

## Acceptance criteria

- [ ] each's two candidates derive \[%@] (`go test ./internal/parse/ -run '^TestDerivePrototypeFromMultiCandidates$' -count=1 -v | grep -q '^--- PASS: TestDerivePrototypeFromMultiCandidates'`)
- [ ] candidates group by parameters only, so keys' four derive \[%@] and localtime's two derive ;$ (`go test ./internal/parse/ -run '^TestDeriveMultiCandidatesGroupByParameters$' -count=1 -v | grep -q '^--- PASS: TestDeriveMultiCandidatesGroupByParameters'`)
- [ ] the multi-candidate rule refuses what it cannot express (`go test ./internal/parse/ -run '^TestDeriveMultiCandidatesRefusesInexpressible$' -count=1 -v | grep -q '^--- PASS: TestDeriveMultiCandidatesRefusesInexpressible'`)
- [ ] return types do not split a parameter group (`go test ./internal/parse/ -run '^TestDeriveMultiCandidatesIgnoreReturnType$' -count=1 -v | grep -q '^--- PASS: TestDeriveMultiCandidatesIgnoreReturnType'`)
- [ ] a single multi candidate derives the plain prototype (`go test ./internal/parse/ -run '^TestDeriveSingleMultiCandidateIsPlain$' -count=1 -v | grep -q '^--- PASS: TestDeriveSingleMultiCandidateIsPlain'`)
