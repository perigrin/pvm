---
title: Derive prototypes from typed signatures and typed signatures from prototypes
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-4139-773b-8676-e148e90c42bd
blocks:
- 01a0fd9b-4850-704d-8c14-c15add6194f2
- 01a0fd9b-49b2-7ea9-bfa1-ba398baadc3f
created: 2026-10-02T17:13:26.980057702Z
updated: 2026-10-02T17:14:05.19016088Z
---

## Prerequisites

- [ ] Parse typed declarations in .pmt files: typed parameters, defaults and return types

## Context

- paths: `internal/parse/derive.go` (new), `internal/parse/derive_test.go` (new), `internal/parse/exports.go`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: RFC 0001 table under "A typed signature and a prototype say the same thing"; `declaredSub` (`internal/parse/exports.go:425`); `ShapeOf` (`internal/parse/shape.go:60`) for how a prototype is read today

The RFC's table maps each prototype character to a parameter, measured on 5.42. This issue builds both directions for every row except `\$` (the `:lvalue` issue owns that row): a prototype-only declaration gets the coarse typed signature (`sub foo ($$)` reads as `(Scalar $, Scalar $)`), a typed declaration gets its prototype with no `:prototype(...)`, and a declaration with both must agree or the file is in error. Rows: `$` Scalar $; `@`,`%` List @; `\@`,`\%` Array @, Hash %; `\[$@%]` their union; `+` Array|Hash|Scalar; leading `&` Code &; `*` Glob *; `_` Scalar $ = $_; `;` parameters with defaults. A derived `@` means no prototype: perl reports `undef` for a sub with none, and the two parse identically (RFC "Builtins that keep their own parse"). Types are finer than prototypes, so prototype to types to prototype round-trips; the reverse loses precision. Expected values live in the test. Re-marking the section waits for the CORE.pmt issue, which holds all 188 derivations to perl.

## Steps

- [ ] Write failing test `TestDeriveTypesFromPrototype` in `internal/parse/derive_test.go`: one case per table row (except `\$`) from prototype to signature
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestDeriveTypesFromPrototype$' -count=1 -v`
- [ ] Implement in `internal/parse/derive.go` (ABOUTME header); run test to verify pass; `make test`; commit
- [ ] Write failing test `TestDerivePrototypeFromTypes`: one case per row from signature to prototype, plus the round trip prototype to types to prototype for each row
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestDerivedAtMeansNoPrototype`: `sub f (List @args);` derives no prototype, recorded as perl records a sub with none
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestPrototypeTypeDisagreementIsError`: `sub f :prototype($$) (Scalar $x);` is an error in the declaration file; `sub f :prototype($) (Str $x);` is not
- [ ] Implement; run test to verify pass; `make test`; commit

## Acceptance criteria

- [ ] a prototype-only declaration gets the coarse typed signature (`go test ./internal/parse/ -run '^TestDeriveTypesFromPrototype$' -count=1 -v | grep -q '^--- PASS: TestDeriveTypesFromPrototype'`)
- [ ] a typed declaration gets its prototype from its types (`go test ./internal/parse/ -run '^TestDerivePrototypeFromTypes$' -count=1 -v | grep -q '^--- PASS: TestDerivePrototypeFromTypes'`)
- [ ] a derived @ means no prototype (`go test ./internal/parse/ -run '^TestDerivedAtMeansNoPrototype$' -count=1 -v | grep -q '^--- PASS: TestDerivedAtMeansNoPrototype'`)
- [ ] a prototype that disagrees with the types is an error (`go test ./internal/parse/ -run '^TestPrototypeTypeDisagreementIsError$' -count=1 -v | grep -q '^--- PASS: TestPrototypeTypeDisagreementIsError'`)
