---
title: Parse the invocant colon in .pmt signatures and declare print, printf, say, exec and system with it
state: done
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-3fbf-7c6e-90a7-919d8670f758
- 01a0fd9b-43ea-716e-9bb8-dd93ce3bbfcb
- 01a0fd9b-46c4-7a8c-b8d6-498c5182a47f
blocks:
- 01a0fd9b-519d-7b42-886e-2222a3076986
- 01a0fd9b-5698-70cf-b6e4-8899a83b0fe0
created: 2026-10-02T17:13:27.729943188Z
updated: 2026-10-16T18:48:31.449268939Z
sessions:
- start_sha: f1adc965eb2a39250bb3225b5cc5834c23ebb915
  end_sha: f1adc965eb2a39250bb3225b5cc5834c23ebb915
  commits: 0
  started_at: 2026-10-06T09:14:12.600058766Z
  ended_at: 2026-10-06T09:14:12.804356666Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-06T09:14:12.600058766Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-06T09:14:12.804356666Z
---

## Prerequisites

- [ ] Add IO, FileHandle and Boolean to the type lattice, with lookup by name
- [ ] Enforce the slurpy rule and a final List parameter in .pmt signatures
- [ ] Derive prototypes from typed signatures and typed signatures from prototypes

## Context

- paths: `internal/parse/decl.go`, `internal/parse/derive.go`, `internal/types/declared.go`, `internal/parse/declarations.go`, `internal/parse/declarations/CORE.pmt`, `internal/parse/pmt_test.go`, `internal/parse/coredecl_test.go`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: the typed-signature reader in `internal/parse/decl.go`; `coreTable` (`internal/parse/declarations.go:39`); `aliasTarget` (`internal/parse/use.go:655`); `internal/types/signatures.go:82-83` (the measured `print` and `say` rows)

RFC 0001 "Builtins that keep their own parse": a leading slot taken with no comma after it (`print STDERR LIST`, `exec {'/bin/sh'} @args`) is written with Raku's invocant colon, `sub print (FileHandle $fh = select(): List[Str] @args = ($_)) Boolean;`. A signature with an invocant colon derives no prototype, matching perl's `undef` for `print`, `printf`, `say`, `exec`, `system` (measured). This issue parses the colon, records the invocant on `types.Signature`, and declares those five in CORE.pmt with types; `sort` is the multi issue's. The parser's existing parse of these builtins does not change. `coreTable()` must keep holding only builtins with a prototype: `aliasTarget` treats presence there as a known prototype and would read `\&CORE::print` as `()`, and `TestCoreDeclarationsArePerls` (`internal/parse/coretable_test.go:20`) must stay green unchanged, so typed declarations with no derived prototype are recorded beside it, not in it. Measured per the RFC: the default handle is the selected one, and `print` with no arguments prints `$_`. `print` and `say` are two of the twelve builtins `internal/types/signatures.go` types that have no prototype; copy their two rows (arity, argument and return types) into the test as golden values -- never compare against signatures.go, which a later issue deletes.

From the alignment gate on 01a0fd9b-4a5b-7e94-ad87-756dfcfe1f4c: without the invocant colon, `print`'s optional `FileHandle $fh = select()` binds positionally, so `types.Select` returns Failed for `print(Str, Str)` and `print(ArrayRef)`. The invocant slot must be outside positional binding in selection, with a test that `print(Str, Str)` selects.

## Steps

- [ ] Write failing test `TestPmtInvocantColonParses`: `sub exec (Str $program: List[Str] @args) Boolean;` gives an invocant param and one slurpy; the default form `FileHandle $fh = select():` parses
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestPmtInvocantColonParses$' -count=1 -v`
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestInvocantColonDerivesNoPrototype`: that declaration derives no prototype
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Measure each form with perl, then write failing test `TestCoreDeclaresInvocantBuiltins`: CORE.pmt declares typed `print`, `printf`, `say`, `exec`, `system` with an invocant; none appears in `coreTable()`
- [ ] Add the declarations (with the measurement comments) and an accessor for typed core declarations; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestCoreInvocantBuiltinsMatchMeasuredSignatures`: the golden `print` and `say` rows equal their CORE.pmt declarations; fix any line that disagrees, measuring which side is right
- [ ] Run test to verify pass; `make test`; commit

### Negative Scenarios

- the colon marks only a leading slot: a colon after the second parameter (`(Str $a, Str $b: List @l)`), two colons, or a colon with nothing before it (`(: List @l)`) is refused (`go test ./internal/parse/ -run '^TestPmtInvocantColonMisplacedRefused$' -count=1 -v | grep -q '^--- PASS: TestPmtInvocantColonMisplacedRefused'`)
- an invocant signature with an explicit prototype (`sub f :prototype($@) (Str $p: List @a);`) is a disagreement error, since the colon derives no prototype (`go test ./internal/parse/ -run '^TestInvocantColonWithPrototypeIsError$' -count=1 -v | grep -q '^--- PASS: TestInvocantColonWithPrototypeIsError'`)
- the parser's own parse of `print STDERR LIST`, `print {$fh} LIST`, bare `print` and `exec {"/bin/sh"} @args` is unchanged: compare against golden trees held in the test (`go test ./internal/parse/ -run '^TestInvocantBuiltinParseUnchanged$' -count=1 -v | grep -q '^--- PASS: TestInvocantBuiltinParseUnchanged'`)

## Acceptance criteria

- [ ] the invocant colon parses in a .pmt signature (`go test ./internal/parse/ -run '^TestTpmInvocantColonParses$' -count=1 -v | grep -q '^--- PASS: TestTpmInvocantColonParses'`)
- [ ] a signature with an invocant colon derives no prototype (`go test ./internal/parse/ -run '^TestInvocantColonDerivesNoPrototype$' -count=1 -v | grep -q '^--- PASS: TestInvocantColonDerivesNoPrototype'`)
- [ ] CORE.pmt declares print, printf, say, exec and system with an invocant (`go test ./internal/parse/ -run '^TestCoreDeclaresInvocantBuiltins$' -count=1 -v | grep -q '^--- PASS: TestCoreDeclaresInvocantBuiltins'`)
- [ ] an invocant colon anywhere but after the first parameter is refused (`go test ./internal/parse/ -run '^TestTpmInvocantColonMisplacedRefused$' -count=1 -v | grep -q '^--- PASS: TestTpmInvocantColonMisplacedRefused'`)
- [ ] an invocant colon with an explicit prototype is an error (`go test ./internal/parse/ -run '^TestInvocantColonWithPrototypeIsError$' -count=1 -v | grep -q '^--- PASS: TestInvocantColonWithPrototypeIsError'`)
- [ ] declaring the invocant builtins leaves their parse unchanged (`go test ./internal/parse/ -run '^TestInvocantBuiltinParseUnchanged$' -count=1 -v | grep -q '^--- PASS: TestInvocantBuiltinParseUnchanged'`)
- [ ] CORE.pmt's print and say carry the measured signatures (`go test ./internal/parse/ -run '^TestCoreInvocantBuiltinsMatchMeasuredSignatures$' -count=1 -v | grep -q '^--- PASS: TestCoreInvocantBuiltinsMatchMeasuredSignatures'`)
