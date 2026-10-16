---
title: Parse return types and type expressions in .pmt declarations
state: done
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-4139-773b-8676-e148e90c42bd
blocks:
- 01a0fd9b-43ea-716e-9bb8-dd93ce3bbfcb
- 01a0fd9b-4a5b-7e94-ad87-756dfcfe1f4c
- 01a0fd9b-53a9-703f-82ab-eeb38765f7f1
created: 2026-10-03T11:12:37.336939532Z
updated: 2026-10-16T18:48:32.250542464Z
sessions:
- start_sha: a59be849cf4cbf3cec10b4655a01884d284fb625
  end_sha: a59be849cf4cbf3cec10b4655a01884d284fb625
  commits: 0
  started_at: 2026-10-06T00:12:24.556947251Z
  ended_at: 2026-10-06T00:12:24.61962481Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-06T00:12:24.556947251Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-06T00:12:24.61962481Z
---

## Prerequisites

- [ ] Parse typed parameters and defaults in .pmt declarations

## Context

- paths: `internal/parse/decl.go`, `internal/types/declared.go`, `internal/parse/pmt_test.go`, `docs/rfcs/0001-type-libraries.md`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: the typed-signature reader the previous issue added to `internal/parse/decl.go`; `types.Signature`/`types.Param` (`internal/types/declared.go`); `types.FromName` (`internal/types/names.go`)

The second half of the parse: RFC 0001 "Declaration order: Perl's", "Type names", and the rest of the parse of "Typed Perl, in `.pmt` only". After the closing parenthesis a `.pmt` declaration carries a return type before its `;` (`sub bless :prototype($;$) (Ref $ref, Str $class = __PACKAGE__) Object;`). Type expressions grow from a single name to a union `A|B` (`Str|Undef`) and a container type `Name[Type]` (`List[Str]`) -- container types parse now, only in `.pmt` files, and what they mean is provisional until the paper (record the element type on the param; the lattice stays flat). An unknown name, as a parameter type, a return type, or inside a union or container, is an error naming it. Declaration order is perl's: attributes before the signature; measured on 5.42, `sub f :lvalue ($x) {}` compiles and `sub f ($x) :lvalue {}` dies ("Subroutine attributes must come before the signature"). This issue completes both "Declaration order" and "Type names" (the lattice half came first), so it re-marks them. Under the RFC's convention (perigrin, 2026-10-05, commit 72e01c14), a section is Implemented when its rule is built and tested: the re-mark here covers the rule, and the sections' example declarations (`bless`'s typed line, the operator `sub + :infix (Num $x, Num $y) Num;`) are illustrations that belong to the typing batches and the operator issues, which write them into CORE.pmt and test them against perl.

## Steps

- [ ] Write failing test `TestPmtReturnType` in `internal/parse/pmt_test.go`: `sub bless :prototype($;$) (Ref $ref, Str $class = __PACKAGE__) Object;` records return `Object`
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestPmtReturnType$' -count=1 -v`
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestPmtUnionAndContainerTypes`: `Str|Undef $x` gives the union; `sub f (List[Str] @args);` gives one slurpy param of container `List`, element `Str`
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestPmtUnionAndContainerTypes$' -count=1 -v`
- [ ] Implement `A|B` and `Name[Type]` in type expressions; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestPmtUnknownTypeNameIsError`: `sub f (Strng $x);` in a `.pmt` is reported as an error naming `Strng`
- [ ] Implement via `types.FromName`; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestPmtSignatureBeforeAttributesRefused`: `sub f (Str $x) :lvalue;` in a `.pmt` is refused, as perl refuses it
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Re-mark the RFC headings "Declaration order: Perl's" and "Type names" `(*Implemented*, <sha>)`; commit

### Negative Scenarios

- an unknown RETURN type (`sub f (Str $x) Strng;`) is an error naming `Strng`, as an unknown parameter type is (`go test ./internal/parse/ -run '^TestPmtUnknownReturnTypeIsError$' -count=1 -v | grep -q '^--- PASS: TestPmtUnknownReturnTypeIsError'`)
- a malformed return type is an error, not a panic: a return type with no terminating `;` at EOF, and a dangling union (`sub f (Str $x) Str|;`) (`go test ./internal/parse/ -run '^TestPmtMalformedReturnTypeIsError$' -count=1 -v | grep -q '^--- PASS: TestPmtMalformedReturnTypeIsError'`)
- malformed container types are errors naming the bad part, not panics: `List[Str @args` (unclosed), `List[] @args` (empty), `List[Strng] @args` (unknown element), `Strng[Str] @args` (unknown container) (`go test ./internal/parse/ -run '^TestPmtMalformedContainerTypeRefused$' -count=1 -v | grep -q '^--- PASS: TestPmtMalformedContainerTypeRefused'`)
- container types are typed Perl and so `.pmt`-only: `sub f (List[Str] @args) { 1 }` in ordinary source records no container type (`go test ./internal/parse/ -run '^TestContainerTypeOnlyInPmt$' -count=1 -v | grep -q '^--- PASS: TestContainerTypeOnlyInPmt'`)

## Acceptance criteria

- [ ] a .pmt declaration carries its return type (`go test ./internal/parse/ -run '^TestTpmReturnType$' -count=1 -v | grep -q '^--- PASS: TestTpmReturnType'`)
- [ ] unions and container types such as List[Str] parse in a .pmt (`go test ./internal/parse/ -run '^TestTpmUnionAndContainerTypes$' -count=1 -v | grep -q '^--- PASS: TestTpmUnionAndContainerTypes'`)
- [ ] an unknown type name in a .pmt is an error (`go test ./internal/parse/ -run '^TestTpmUnknownTypeNameIsError$' -count=1 -v | grep -q '^--- PASS: TestTpmUnknownTypeNameIsError'`)
- [ ] a signature before the attributes is refused, as perl refuses it (`go test ./internal/parse/ -run '^TestTpmSignatureBeforeAttributesRefused$' -count=1 -v | grep -q '^--- PASS: TestTpmSignatureBeforeAttributesRefused'`)
- [ ] the RFC section on declaration order is implemented (`grep -q '^### Declaration order.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] the RFC section on type names is implemented (`grep -q '^### Type names.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] an unknown return type in a .pmt is an error (`go test ./internal/parse/ -run '^TestTpmUnknownReturnTypeIsError$' -count=1 -v | grep -q '^--- PASS: TestTpmUnknownReturnTypeIsError'`)
- [ ] a malformed return type is an error, not a panic (`go test ./internal/parse/ -run '^TestTpmMalformedReturnTypeIsError$' -count=1 -v | grep -q '^--- PASS: TestTpmMalformedReturnTypeIsError'`)
- [ ] malformed container types are errors (`go test ./internal/parse/ -run '^TestTpmMalformedContainerTypeRefused$' -count=1 -v | grep -q '^--- PASS: TestTpmMalformedContainerTypeRefused'`)
- [ ] container types are not read in ordinary source (`go test ./internal/parse/ -run '^TestContainerTypeOnlyInTpm$' -count=1 -v | grep -q '^--- PASS: TestContainerTypeOnlyInTpm'`)
