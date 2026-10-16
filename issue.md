---
title: Derive prototypes from typed signatures and typed signatures from prototypes
state: done
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-4139-773b-8676-e148e90c42bd
- 01a0fd9b-43ea-716e-9bb8-dd93ce3bbfcb
blocks:
- 01a0fd9b-4850-704d-8c14-c15add6194f2
- 01a0fd9b-49b2-7ea9-bfa1-ba398baadc3f
created: 2026-10-02T17:13:26.980057702Z
updated: 2026-10-16T18:48:31.181990965Z
sessions:
- start_sha: 9a6ceb6ed82ede814492d9f25a870f300fc55670
  end_sha: 9a6ceb6ed82ede814492d9f25a870f300fc55670
  commits: 0
  started_at: 2026-10-06T06:50:48.819609703Z
  ended_at: 2026-10-06T06:50:48.989766758Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-06T06:50:48.819609703Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-06T06:50:48.989766758Z
---

## Prerequisites

- [ ] Parse typed parameters and defaults in .pmt declarations
- [ ] Enforce the slurpy rule and a final List parameter in .pmt signatures

## Context

- paths: `internal/parse/derive.go` (new), `internal/parse/derive_test.go` (new), `internal/parse/exports.go`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: RFC 0001 table under "A typed signature and a prototype say the same thing"; `declaredSub` (`internal/parse/exports.go:425`); `ShapeOf` (`internal/parse/shape.go:60`) for how a prototype is read today

The RFC's table maps each prototype character to a parameter, measured on 5.42. This issue builds both directions for every row that takes no backslash: a prototype-only declaration gets the coarse typed signature (`sub foo ($$)` reads as `(Scalar $, Scalar $)`), a typed declaration gets its prototype with no `:prototype(...)`, and a declaration with both must agree or the file is in error. Rows: `$` Scalar $; `@`,`%` List @; `+` Array|Hash|Scalar; leading `&` Code &; `*` Glob *; `_` Scalar $ = $_; `;` parameters with defaults other than `die`. RFC 0001 "A required argument defaults to `die`" (perigrin, 2026-10-05): a `die` default runs only when its argument is omitted, so it makes the parameter required (recorded as required by the typed-parameters issue); derivation places no `;` before it, and only a non-`die` default makes a parameter optional, so `(Scalar $x = die)` derives `$` and `(Scalar $x = 1)` derives `;$`. The aliasing rows (`\$`, `\@`, `\%`, `\[...]`) belong to the aliased-parameter issue, which parses the backslash, and to the multi-candidate derivation issue. A derived `@` means no prototype: perl reports `undef` for a sub with none, and the two parse identically (RFC "Builtins that keep their own parse"). A prototype with anything after `@` or `%` (`@$`, measured: "Prototype after '@'") derives no signature: it is the final-`List` error of the previous issue, pointing to `\@$`. Types are finer than prototypes, so prototype to types to prototype round-trips; the reverse loses precision. Expected values live in the test. Re-marking the section waits for the last CORE.pmt typing batch, which holds all 188 derivations to perl.

## Steps

- [ ] Write failing test `TestPmtDeclarationErrorsSurface`: a shipped-style `.pmt` with a malformed typed signature (`sub f (Str);`) read through `coreTable` and through the resolver reports the declaration error (today `moduleFacts.errs` is recorded by `readDeclaration` but discarded at `internal/parse/declarations.go:56` and `internal/parse/module.go:299`; found by the alignment gate on 01a0fd9b-4139-773b-8676-e148e90c42bd); implement so a declaration error is reported, not dropped; `make test`; commit

- [ ] Write failing test `TestDeriveTypesFromPrototype` in `internal/parse/derive_test.go`: one case per unbackslashed table row from prototype to signature
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestDeriveTypesFromPrototype$' -count=1 -v`
- [ ] Implement in `internal/parse/derive.go` (ABOUTME header); run test to verify pass; `make test`; commit
- [ ] Write failing test `TestDerivePrototypeFromTypes`: one case per unbackslashed row from signature to prototype, plus the round trip prototype to types to prototype for each row
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestDeriveDieDefaultIsRequired`: `(Scalar $x = die)` derives `$`; `(Scalar $x = 1)` derives `;$`; `(Scalar $x, Scalar $y = die)` derives `$$`
- [ ] Run test to verify failure, implement, run test to verify pass: `go test ./internal/parse/ -run '^TestDeriveDieDefaultIsRequired$' -count=1 -v`; `make test`; commit
- [ ] Write failing test `TestDerivedAtMeansNoPrototype`: `sub f (List @args);` derives no prototype, recorded as perl records a sub with none
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestPrototypeTypeDisagreementIsError`: `sub f :prototype($$) (Scalar $x);` is an error in the declaration file; `sub f :prototype($) (Str $x);` is not
- [ ] Implement; run test to verify pass; `make test`; commit

### Negative Scenarios

- a prototype character outside the table (`:prototype(Q)`, `:prototype(\x)`) is an error in the declaration file, not a guessed signature (perl 5.42 only warns: Illegal character in prototype) (`go test ./internal/parse/ -run '^TestDeriveRejectsUnknownPrototypeCharacter$' -count=1 -v | grep -q '^--- PASS: TestDeriveRejectsUnknownPrototypeCharacter'`)
- an empty prototype `()` derives zero parameters and is distinct from no prototype, which derives `List @` (absent and empty differ) (`go test ./internal/parse/ -run '^TestDeriveEmptyPrototypeIsNotAbsent$' -count=1 -v | grep -q '^--- PASS: TestDeriveEmptyPrototypeIsNotAbsent'`)
- more disagreements are caught, not only arity: `:prototype($;$) (Scalar $x, Scalar $y)` (the `;` has no default), `:prototype(&@) (Scalar $f, List @l)` (Code against Scalar) (`go test ./internal/parse/ -run '^TestPrototypeTypeDisagreementKinds$' -count=1 -v | grep -q '^--- PASS: TestPrototypeTypeDisagreementKinds'`)
- a required parameter after a defaulted one (`sub f (Str $x = 1, Str $y);`) cannot derive a `;` prototype and is an error, as perl 5.42 dies: Mandatory parameter follows optional parameter (`go test ./internal/parse/ -run '^TestDeriveMandatoryAfterOptionalIsError$' -count=1 -v | grep -q '^--- PASS: TestDeriveMandatoryAfterOptionalIsError'`)
- a prototype with anything after `@` (`:prototype(@$)`) derives no signature and is the final-`List` error pointing to `\@$` (`go test ./internal/parse/ -run '^TestDerivePrototypeAfterAtIsError$' -count=1 -v | grep -q '^--- PASS: TestDerivePrototypeAfterAtIsError'`)

## Acceptance criteria

- [ ] a prototype-only declaration gets the coarse typed signature (`go test ./internal/parse/ -run '^TestDeriveTypesFromPrototype$' -count=1 -v | grep -q '^--- PASS: TestDeriveTypesFromPrototype'`)
- [ ] a typed declaration gets its prototype from its types (`go test ./internal/parse/ -run '^TestDerivePrototypeFromTypes$' -count=1 -v | grep -q '^--- PASS: TestDerivePrototypeFromTypes'`)
- [ ] a die default is required and derives no ; before it (`go test ./internal/parse/ -run '^TestDeriveDieDefaultIsRequired$' -count=1 -v | grep -q '^--- PASS: TestDeriveDieDefaultIsRequired'`)
- [ ] a derived @ means no prototype (`go test ./internal/parse/ -run '^TestDerivedAtMeansNoPrototype$' -count=1 -v | grep -q '^--- PASS: TestDerivedAtMeansNoPrototype'`)
- [ ] a prototype that disagrees with the types is an error (`go test ./internal/parse/ -run '^TestPrototypeTypeDisagreementIsError$' -count=1 -v | grep -q '^--- PASS: TestPrototypeTypeDisagreementIsError'`)
- [ ] an unknown prototype character is an error, not a guess (`go test ./internal/parse/ -run '^TestDeriveRejectsUnknownPrototypeCharacter$' -count=1 -v | grep -q '^--- PASS: TestDeriveRejectsUnknownPrototypeCharacter'`)
- [ ] an empty prototype differs from no prototype (`go test ./internal/parse/ -run '^TestDeriveEmptyPrototypeIsNotAbsent$' -count=1 -v | grep -q '^--- PASS: TestDeriveEmptyPrototypeIsNotAbsent'`)
- [ ] optionality and Code/Scalar disagreements are errors (`go test ./internal/parse/ -run '^TestPrototypeTypeDisagreementKinds$' -count=1 -v | grep -q '^--- PASS: TestPrototypeTypeDisagreementKinds'`)
- [ ] a mandatory parameter after an optional one is an error (`go test ./internal/parse/ -run '^TestDeriveMandatoryAfterOptionalIsError$' -count=1 -v | grep -q '^--- PASS: TestDeriveMandatoryAfterOptionalIsError'`)
- [ ] a prototype with anything after @ derives the final-List error (`go test ./internal/parse/ -run '^TestDerivePrototypeAfterAtIsError$' -count=1 -v | grep -q '^--- PASS: TestDerivePrototypeAfterAtIsError'`)
- [ ] a malformed .pmt declaration's error is reported, not dropped (`go test ./internal/parse/ -run '^TestTpmDeclarationErrorsSurface$' -count=1 -v | grep -q '^--- PASS: TestTpmDeclarationErrorsSurface'`)
