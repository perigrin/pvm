---
title: Declare the no-prototype builtins and each's multi in CORE.pmt, and read :unary
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-4850-704d-8c14-c15add6194f2
- 01a0fd9b-4a5b-7e94-ad87-756dfcfe1f4c
blocks:
- 01a0fd9b-519d-7b42-886e-2222a3076986
created: 2026-10-02T17:13:29.367336401Z
updated: 2026-10-03T11:14:17.407922521Z
---

## Prerequisites

- [ ] Parse aliased parameters (\$x, \@a, \%h, \(@list)) and derive \$, \@ and \% from them
- [ ] Parse multi sub declarations and select a candidate by arity and argument types
- [ ] Check CORE.pmt's derived prototypes against perl, with signatures.go's measured rows

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/decl.go`, `internal/parse/derive.go`, `internal/parse/derive_test.go`, `internal/parse/declarations.go`, `internal/parse/coretypes_test.go`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `namedUnary` in `internal/parse/keyword.go:40` (the measured named unaries), the lexer's keyword list (`lexer.Keywords()`), `coreTable` (`internal/parse/declarations.go:39`), `internal/parse/derive.go`, `internal/types/signatures.go:44-84` (the measured rows for `delete`, `exists`, `each`, `chomp`, `chop`, `split`, `defined`, `return`)

RFC 0001 "Builtins with no prototype": about 20 builtins have no prototype perl reports. CORE.pmt declares their types; `:unary` marks the ones that parse as named unaries; a builtin with no prototype is otherwise a list operator, as an ordinary sub is. This issue declares the plain ones -- including `chomp` and `chop` as `sub chomp (List[Str] \(@args = ($_))) Int;` (each element aliased; measured: `chomp(@l, $x)` chomps both, bare `chomp` chomps `$_`), and `select` as the RFC's multi (`(FileHandle $fh) Str`, `($r, $w, $e, Num $timeout) Int`). It also declares `each`, which is prototyped (`\[%@]`, measured), as the RFC's multi, `multi sub each (Hash \%h) List;` and `multi sub each (Array \@a) List;`, and adds the rule that derives one prototype from several multi candidates: candidates of the same arity that differ only in which container one aliased position takes derive `\[...]` with those containers in declaration order, so `each`'s two candidates derive perl's `\[%@]`; `keys`, `values` and the other `\[...]` lines take the same rule in the typing batches. `print`/`exec` and kin came with the invocant colon; `grep`/`map`/`sort` are the next issue's. Find the no-prototype set with perl (keywords whose `prototype("CORE::k")` is undef and which are functions) and the unary subset by the deparse method `keyword.go` records; hold the expected names in the test, not against keyword.go. Copy signatures.go's rows for `delete`, `exists`, `chomp`, `chop`, `split`, `defined`, `return` and `each` into the test as golden values -- never compare against signatures.go, which a later issue deletes. Declarations with no derived prototype stay out of `coreTable()` (see `aliasTarget`, `internal/parse/use.go:655`), so `TestCoreDeclarationsArePerls` and `TestCoreDerivedPrototypesArePerls` stay green. Measured on 5.42, `select()` with no arguments compiles (it returns the selected handle), and neither of the RFC's two candidates accepts zero arguments; under "a call whose arity no candidate accepts fails" that would refuse valid perl, so the declaration needs perigrin's ruling (see the negative scenario) before the parse-time refusal issue relies on it.

## Steps

- [ ] Measure with perl, then write failing test `TestCoreUnaryBuiltins`: CORE.pmt declares each no-prototype function builtin (names held in the test), and exactly the measured named unaries carry `:unary`
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestCoreUnaryBuiltins$' -count=1 -v`
- [ ] Read `:unary`; add the declarations with their measurement comments; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestDerivePrototypeFromMultiCandidates` in `internal/parse/derive_test.go`: `(Hash \%h)` and `(Array \@a)` derive `\[%@]`; candidates differing in more than one position, or in arity, derive no single prototype and are an error where a `:prototype(...)` is required
- [ ] Implement in `internal/parse/derive.go`; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestCoreSelectAndEachAreMultis`: `select` and `each` each have two candidates with the RFC's signatures, and `each`'s derived prototype is perl's
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestCoreNoPrototypeBuiltinsMatchMeasuredSignatures`: the eight golden rows equal their CORE.pmt declarations (for `each`, its candidates)
- [ ] Fix any line that disagrees, measuring which side is right; run test to verify pass; `make test`; commit

### Negative Scenarios

- no-prototype declarations stay out of `coreTable()`: none of the names held in the test appears there, so `aliasTarget` never reads one as `()` (`go test ./internal/parse/ -run '^TestCoreNoPrototypeBuiltinsOutOfCoreTable$' -count=1 -v | grep -q '^--- PASS: TestCoreNoPrototypeBuiltinsOutOfCoreTable'`)
- `:unary` is not inferred: for one measured named unary, `NAME $a, $b` parses as `(NAME $a), $b`, and for one unmarked no-prototype builtin it parses as one call with two arguments (measure both with B::Deparse) (`go test ./internal/parse/ -run '^TestCoreUnaryAndListOperatorParseDiffer$' -count=1 -v | grep -q '^--- PASS: TestCoreUnaryAndListOperatorParseDiffer'`)
- `each` on a scalar (`each $x`), which perl 5.42 forbids (Experimental each on scalar is now forbidden), matches neither `each` candidate (`go test ./internal/parse/ -run '^TestCoreEachRefusesScalar$' -count=1 -v | grep -q '^--- PASS: TestCoreEachRefusesScalar'`)
- the multi-candidate rule refuses what it cannot express: candidates that differ in two aliased positions, or whose union prototype perl would order differently from the declaration, are reported rather than guessed (`go test ./internal/parse/ -run '^TestDeriveMultiCandidatesRefusesInexpressible$' -count=1 -v | grep -q '^--- PASS: TestDeriveMultiCandidatesRefusesInexpressible'`)
- `select()` with no arguments, which perl 5.42 compiles, must match a declaration; the RFC's two candidates do not accept zero arguments, so this needs perigrin's ruling on `select`'s declaration before it becomes a criterion (`go test ./internal/parse/ -run '^TestCoreSelectNoArgumentsAccepted$' -count=1 -v | grep -q '^--- PASS: TestCoreSelectNoArgumentsAccepted'`)

## Acceptance criteria

- [ ] CORE.pmt declares the no-prototype builtins and :unary marks the named unaries (`go test ./internal/parse/ -run '^TestCoreUnaryBuiltins$' -count=1 -v | grep -q '^--- PASS: TestCoreUnaryBuiltins'`)
- [ ] several multi candidates derive one \[...] prototype (`go test ./internal/parse/ -run '^TestDerivePrototypeFromMultiCandidates$' -count=1 -v | grep -q '^--- PASS: TestDerivePrototypeFromMultiCandidates'`)
- [ ] select and each are declared as multis (`go test ./internal/parse/ -run '^TestCoreSelectAndEachAreMultis$' -count=1 -v | grep -q '^--- PASS: TestCoreSelectAndEachAreMultis'`)
- [ ] the no-prototype builtins and each carry the measured signatures (`go test ./internal/parse/ -run '^TestCoreNoPrototypeBuiltinsMatchMeasuredSignatures$' -count=1 -v | grep -q '^--- PASS: TestCoreNoPrototypeBuiltinsMatchMeasuredSignatures'`)
- [ ] no-prototype builtins stay out of the prototype table (`go test ./internal/parse/ -run '^TestCoreNoPrototypeBuiltinsOutOfCoreTable$' -count=1 -v | grep -q '^--- PASS: TestCoreNoPrototypeBuiltinsOutOfCoreTable'`)
- [ ] a :unary builtin and a list-operator builtin parse differently (`go test ./internal/parse/ -run '^TestCoreUnaryAndListOperatorParseDiffer$' -count=1 -v | grep -q '^--- PASS: TestCoreUnaryAndListOperatorParseDiffer'`)
- [ ] each on a scalar matches neither candidate (`go test ./internal/parse/ -run '^TestCoreEachRefusesScalar$' -count=1 -v | grep -q '^--- PASS: TestCoreEachRefusesScalar'`)
- [ ] the multi-candidate rule refuses what it cannot express (`go test ./internal/parse/ -run '^TestDeriveMultiCandidatesRefusesInexpressible$' -count=1 -v | grep -q '^--- PASS: TestDeriveMultiCandidatesRefusesInexpressible'`)
