---
title: Parse the invocant colon in .pmt signatures and declare print, printf, say, exec and system with it
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-3fbf-7c6e-90a7-919d8670f758
created: 2026-10-02T17:13:27.729943188Z
updated: 2026-10-02T17:14:01.369044362Z
---

## Prerequisites

- [ ] Add IO, FileHandle and Boolean to the type lattice, with lookup by name
- [ ] Enforce the slurpy rule and caller-view sigils in .pmt signatures
- [ ] Derive prototypes from typed signatures and typed signatures from prototypes

## Context

- paths: `internal/parse/decl.go`, `internal/parse/derive.go`, `internal/types/declared.go`, `internal/parse/declarations.go`, `internal/parse/declarations/CORE.pmt`, `internal/parse/pmt_test.go`, `internal/parse/coredecl_test.go`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: the typed-signature reader in `internal/parse/decl.go`; `coreTable` (`internal/parse/declarations.go:39`); `aliasTarget` (`internal/parse/use.go:655`)

RFC 0001 "Builtins that keep their own parse": a leading slot taken with no comma after it (`print STDERR LIST`, `exec {'/bin/sh'} @args`) is written with Raku's invocant colon, `sub print (FileHandle $fh = select(): List[Str] @args = ($_)) Boolean;`. A signature with an invocant colon derives no prototype, matching perl's `undef` for `print`, `printf`, `say`, `exec`, `system` (measured). This issue parses the colon, records the invocant on `types.Signature`, and declares those five in CORE.pmt with types; `sort` is the multi issue's. The parser's existing parse of these builtins does not change. `coreTable()` must keep holding only builtins with a prototype: `aliasTarget` treats presence there as a known prototype and would read `\&CORE::print` as `()`, and `TestCoreDeclarationsArePerls` (`internal/parse/coretable_test.go:20`) must stay green unchanged, so typed declarations with no derived prototype are recorded beside it, not in it. Measured per the RFC: the default handle is the selected one, and `print` with no arguments prints `$_`.

## Steps

- [ ] Write failing test `TestPmtInvocantColonParses`: `sub exec (Str $program: List[Str] @args) Boolean;` gives an invocant param and one slurpy; the default form `FileHandle $fh = select():` parses
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestPmtInvocantColonParses$' -count=1 -v`
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestInvocantColonDerivesNoPrototype`: that declaration derives no prototype
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Measure each form with perl, then write failing test `TestCoreDeclaresInvocantBuiltins`: CORE.pmt declares typed `print`, `printf`, `say`, `exec`, `system` with an invocant; none appears in `coreTable()`
- [ ] Add the declarations (with the measurement comments) and an accessor for typed core declarations; run test to verify pass; `make test`; commit

## Acceptance criteria

- [ ] the invocant colon parses in a .pmt signature (`go test ./internal/parse/ -run '^TestPmtInvocantColonParses$' -count=1 -v | grep -q '^--- PASS: TestPmtInvocantColonParses'`)
- [ ] a signature with an invocant colon derives no prototype (`go test ./internal/parse/ -run '^TestInvocantColonDerivesNoPrototype$' -count=1 -v | grep -q '^--- PASS: TestInvocantColonDerivesNoPrototype'`)
- [ ] CORE.pmt declares print, printf, say, exec and system with an invocant (`go test ./internal/parse/ -run '^TestCoreDeclaresInvocantBuiltins$' -count=1 -v | grep -q '^--- PASS: TestCoreDeclaresInvocantBuiltins'`)
