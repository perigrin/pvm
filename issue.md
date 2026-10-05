---
title: Type CORE.pmt's array, list and hash builtins
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-4f8d-71ca-bac4-24eceb1dc4a4
- 01a0fd9b-5017-78d0-ae38-331cbe6a9866
- 01a0fd9b-4ca3-7d55-82ac-ba125816ee71
blocks:
- 01a10c6e-9a95-719a-80e2-2e33037d0096
- 01a0fd9b-5670-78d0-84be-9b9618e9b081
created: 2026-10-03T11:12:38.523502989Z
updated: 2026-10-05T20:29:55.560479114Z
---

## Prerequisites

- [ ] Check CORE.pmt's derived prototypes against perl, with signatures.go's measured rows
- [ ] Derive one prototype from several multi candidates

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/coretypes_test.go`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `TestCoreDerivedPrototypesArePerls` (`internal/parse/coretypes_test.go`, the harness); `perl -MPod::Functions -e 'print "@{$Kinds{ARRAY}} @{$Kinds{LIST}} @{$Kinds{HASH}}"'`; `internal/types/signatures.go:42-43` (the measured `keys` and `values` rows); the multi-candidate prototype rule the multi-candidate derivation issue adds (it follows the `:context(...)` issue, so `:context` is read before this batch); RFC 0001 "`:context(...)`" (the `localtime` spelling the `keys`/`values` multis follow)

Second batch, by perlfunc's categories "Functions for real @ARRAYs", "Functions for list data" and "Functions for real %HASHes" (Pod::Functions kinds `ARRAY`, `LIST`, `HASH`), intersected with CORE.pmt's prototyped lines, excluding `each`, which the no-prototype builtins issue declares as a multi (this batch does not wait for it: `each` keeps its one prototype-only line until that issue replaces it): `all`, `any`, `keys`, `pop`, `push`, `shift`, `splice`, `unshift`, `values`, `join`, `unpack` (measured on 5.42; `push` to `splice` and `join` were typed with the measured rows). `all` and `any` are `sub all (Code \&block, List @list) Boolean;` and `sub any (Code \&block, List @list) Boolean;` (perigrin, 2026-10-05; RFC "The scalar container"), whose derived prototype is perl's `\&@` by the aliased-parameters issue's `\&` row. `keys` and `values` are `\[%@]` and context-dependent: measured, `my $n = keys %h` is 2 for a two-key hash and `my $m = values @a` is 3 for a three-element array, while list context gives the list. So each is declared as `:context` multis, as `localtime` is, over the two containers; the multi-candidate rule groups the four candidates by parameters only, ignoring `:context` and the return type: `multi sub keys :context($) (Hash \%h) Int;`, `multi sub keys :context($) (Array \@a) Int;`, `multi sub keys :context(@) (Hash \%h) List;`, `multi sub keys :context(@) (Array \@a) List;`, and the same for `values`; the derived prototype is still `\[%@]`. Copy their two signatures.go rows (the `List` results) into the test as golden values for the `:context(@)` candidates (never compare against signatures.go, which a later issue deletes). Every type is measured on perl 5.42 and recorded as a comment. Hold the batch's names in the test. Re-marks no RFC section.

## Steps

- [ ] Measure the batch's names; write failing test `TestCoreArrayListHashBuiltinsTyped`: every name in the batch has a typed CORE.pmt line, `all` and `any` with `Code \&block`, `keys` and `values` as four-candidate `:context` multis
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestCoreArrayListHashBuiltinsTyped$' -count=1 -v`
- [ ] Type the lines with their measurement comments; run the harness and the test to verify pass; `make test`; commit
- [ ] Write failing test `TestCoreKeysValuesMatchMeasuredSignatures`: the golden `keys` and `values` rows equal their `:context(@)` candidates, their `:context($)` candidates return `Int`, and each name's candidates derive `\[%@]`
- [ ] Implement any fix, measuring which side is right; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestCoreKeysScalarContextIsCount`: a scalar-context `keys %h` selects the `Int` candidate and a list-context one the `List` candidate
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestCoreKeysScalarContextIsCount$' -count=1 -v`
- [ ] Implement any fix; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestCoreAllAnyDeriveRefAmpAt`: `all` and `any` take `Code \&block, List @list`, return `Boolean`, and derive `\&@`
- [ ] Implement any fix; run test to verify pass; `make test`; commit

### Negative Scenarios

- `keys` and `values` on a scalar (`keys $x`), which perl 5.42 forbids, match neither candidate (`go test ./internal/parse/ -run '^TestCoreKeysValuesRefuseScalar$' -count=1 -v | grep -q '^--- PASS: TestCoreKeysValuesRefuseScalar'`)
- `each` is not typed twice: this batch leaves its line to the multi issue, and the file has exactly one declaration set for `each` (`go test ./internal/parse/ -run '^TestCoreEachDeclaredOnce$' -count=1 -v | grep -q '^--- PASS: TestCoreEachDeclaredOnce'`)

## Acceptance criteria

- [ ] CORE.pmt types every array, list and hash builtin (`go test ./internal/parse/ -run '^TestCoreArrayListHashBuiltinsTyped$' -count=1 -v | grep -q '^--- PASS: TestCoreArrayListHashBuiltinsTyped'`)
- [ ] keys in scalar context is a count and in list context a list (`go test ./internal/parse/ -run '^TestCoreKeysScalarContextIsCount$' -count=1 -v | grep -q '^--- PASS: TestCoreKeysScalarContextIsCount'`)
- [ ] all and any take a Code slot and derive perl prototype (`go test ./internal/parse/ -run '^TestCoreAllAnyDeriveRefAmpAt$' -count=1 -v | grep -q '^--- PASS: TestCoreAllAnyDeriveRefAmpAt'`)
- [ ] keys and values carry the measured signatures and derive perl prototype (`go test ./internal/parse/ -run '^TestCoreKeysValuesMatchMeasuredSignatures$' -count=1 -v | grep -q '^--- PASS: TestCoreKeysValuesMatchMeasuredSignatures'`)
- [ ] every CORE.pmt prototype, derived from its types where typed, is still perl's (`go test ./internal/parse/ -run '^TestCoreDerivedPrototypesArePerls$' -count=1 -v | grep -q '^--- PASS: TestCoreDerivedPrototypesArePerls'`)
- [ ] keys and values on a scalar match no candidate (`go test ./internal/parse/ -run '^TestCoreKeysValuesRefuseScalar$' -count=1 -v | grep -q '^--- PASS: TestCoreKeysValuesRefuseScalar'`)
- [ ] each is declared once (`go test ./internal/parse/ -run '^TestCoreEachDeclaredOnce$' -count=1 -v | grep -q '^--- PASS: TestCoreEachDeclaredOnce'`)
