---
title: Parse return types and type expressions in .pmt declarations
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-03T11:12:37.336939532Z
updated: 2026-10-03T11:12:37.336939532Z
---

## Prerequisites

- [ ] Parse typed parameters and defaults in .pmt declarations

## Context

- paths: `internal/parse/decl.go`, `internal/types/declared.go`, `internal/parse/pmt_test.go`, `internal/parse/declarations/Moose.pmt`, `internal/parse/declarations/Moose/Role.pmt`, `internal/parse/declarations/Mojolicious/Lite.pmt`, `docs/rfcs/0001-type-libraries.md`, `internal/data/help/types.md`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: the typed-signature reader the previous issue added to `internal/parse/decl.go`; `types.Signature`/`types.Param` (`internal/types/declared.go`); `types.FromName` (`internal/types/names.go`)

The second half of the parse: RFC 0001 "Declaration order: Perl's", "Type names", and the rest of the parse of "Typed Perl, in `.pmt` only". After the closing parenthesis a `.pmt` declaration carries a return type before its `;` (`sub bless :prototype($;$) (Ref $ref, Str $class = __PACKAGE__) Object;`). Type expressions grow from a single name to a union `A|B` (`Str|Undef`) and a container type `Name[Type]` (`List[Str]`) -- container types parse now, only in `.pmt` files, and what they mean is provisional until the paper (record the element type on the param; the lattice stays flat). An unknown name, as a parameter type, a return type, or inside a union or container, is an error naming it. Declaration order is perl's: attributes before the signature; measured on 5.42, `sub f :lvalue ($x) {}` compiles and `sub f ($x) :lvalue {}` dies ("Subroutine attributes must come before the signature"). Once signature mode reads `(` as a signature, the shipped bare prototypes (`Moose.pmt:21`, `Moose/Role.pmt:23`, `Mojolicious/Lite.pmt:21`) move to `:prototype(...)`, as perl itself requires once signatures are on. This issue completes both "Declaration order" and "Type names" (the lattice half came first), so it re-marks them.

## Steps

- [ ] Write failing test `TestPmtReturnType` in `internal/parse/pmt_test.go`: `sub bless :prototype($;$) (Ref $ref, Str $class = __PACKAGE__) Object;` records return `Object`
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestPmtReturnType$' -count=1 -v`
- [ ] Implement; move the three bare prototypes in the shipped declarations to `:prototype(...)`; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestPmtUnionAndContainerTypes`: `Str|Undef $x` gives the union; `sub f (List[Str] @args);` gives one slurpy param of container `List`, element `Str`
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestPmtUnionAndContainerTypes$' -count=1 -v`
- [ ] Implement `A|B` and `Name[Type]` in type expressions; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestPmtUnknownTypeNameIsError`: `sub f (Strng $x);` in a `.pmt` is reported as an error naming `Strng`
- [ ] Implement via `types.FromName`; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestPmtSignatureBeforeAttributesRefused`: `sub f (Str $x) :lvalue;` in a `.pmt` is refused, as perl refuses it
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Re-mark the RFC headings "Declaration order: Perl's" and "Type names" `(*Implemented*, <sha>)`; commit
- [ ] Doc: the Mojolicious::Lite example in `internal/data/help/types.md` and in the RFC's "Declaration files" section shows `sub group (&);`; change both to the shipped spelling, `sub group :prototype(&);`; `go test ./internal/cli/ -run '^TestShowTypesHelp$' -count=1`; commit

### Negative Scenarios

- an unknown RETURN type (`sub f (Str $x) Strng;`) is an error naming `Strng`, as an unknown parameter type is (`go test ./internal/parse/ -run '^TestPmtUnknownReturnTypeIsError$' -count=1 -v | grep -q '^--- PASS: TestPmtUnknownReturnTypeIsError'`)
- a malformed return type is an error, not a panic: a return type with no terminating `;` at EOF, and a dangling union (`sub f (Str $x) Str|;`) (`go test ./internal/parse/ -run '^TestPmtMalformedReturnTypeIsError$' -count=1 -v | grep -q '^--- PASS: TestPmtMalformedReturnTypeIsError'`)
- malformed container types are errors naming the bad part, not panics: `List[Str @args` (unclosed), `List[] @args` (empty), `List[Strng] @args` (unknown element), `Strng[Str] @args` (unknown container) (`go test ./internal/parse/ -run '^TestPmtMalformedContainerTypeRefused$' -count=1 -v | grep -q '^--- PASS: TestPmtMalformedContainerTypeRefused'`)
- container types are typed Perl and so `.pmt`-only: `sub f (List[Str] @args) { 1 }` in ordinary source records no container type (`go test ./internal/parse/ -run '^TestContainerTypeOnlyInPmt$' -count=1 -v | grep -q '^--- PASS: TestContainerTypeOnlyInPmt'`)

## Acceptance criteria

- [ ] a .pmt declaration carries its return type (`go test ./internal/parse/ -run '^TestPmtReturnType$' -count=1 -v | grep -q '^--- PASS: TestPmtReturnType'`)
- [ ] unions and container types such as List[Str] parse in a .pmt (`go test ./internal/parse/ -run '^TestPmtUnionAndContainerTypes$' -count=1 -v | grep -q '^--- PASS: TestPmtUnionAndContainerTypes'`)
- [ ] an unknown type name in a .pmt is an error (`go test ./internal/parse/ -run '^TestPmtUnknownTypeNameIsError$' -count=1 -v | grep -q '^--- PASS: TestPmtUnknownTypeNameIsError'`)
- [ ] a signature before the attributes is refused, as perl refuses it (`go test ./internal/parse/ -run '^TestPmtSignatureBeforeAttributesRefused$' -count=1 -v | grep -q '^--- PASS: TestPmtSignatureBeforeAttributesRefused'`)
- [ ] the RFC section on declaration order is implemented (`grep -q '^### Declaration order.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] the RFC section on type names is implemented (`grep -q '^### Type names.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] an unknown return type in a .pmt is an error (`go test ./internal/parse/ -run '^TestPmtUnknownReturnTypeIsError$' -count=1 -v | grep -q '^--- PASS: TestPmtUnknownReturnTypeIsError'`)
- [ ] a malformed return type is an error, not a panic (`go test ./internal/parse/ -run '^TestPmtMalformedReturnTypeIsError$' -count=1 -v | grep -q '^--- PASS: TestPmtMalformedReturnTypeIsError'`)
- [ ] malformed container types are errors (`go test ./internal/parse/ -run '^TestPmtMalformedContainerTypeRefused$' -count=1 -v | grep -q '^--- PASS: TestPmtMalformedContainerTypeRefused'`)
- [ ] container types are not read in ordinary source (`go test ./internal/parse/ -run '^TestContainerTypeOnlyInPmt$' -count=1 -v | grep -q '^--- PASS: TestContainerTypeOnlyInPmt'`)
- [ ] the help page and the RFC show the shipped prototype spelling (`sh -c '! grep -qF "sub group (&)" internal/data/help/types.md docs/rfcs/0001-type-libraries.md'`)
