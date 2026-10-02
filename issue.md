---
title: "Parse typed declarations in .pmt files: typed parameters, defaults and return types"
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-3fbf-7c6e-90a7-919d8670f758
blocks:
- 01a0fd9b-43ea-716e-9bb8-dd93ce3bbfcb
- 01a0fd9b-46c4-7a8c-b8d6-498c5182a47f
- 01a0fd9b-4a5b-7e94-ad87-756dfcfe1f4c
- 01a0fd9b-53a9-703f-82ab-eeb38765f7f1
created: 2026-10-02T17:13:25.560580634Z
updated: 2026-10-02T17:22:49.175841531Z
---

## Prerequisites

- [ ] Add IO, FileHandle and Boolean to the type lattice, with lookup by name

## Context

- paths: `internal/parse/decl.go`, `internal/parse/exports.go`, `internal/parse/module.go`, `internal/parse/declarations.go`, `internal/parse/parse.go`, `internal/lexer/proto.go`, `internal/types/declared.go` (new), `internal/parse/pmt_test.go` (new), `internal/parse/declarations/Moose.pmt`, `internal/parse/declarations/Moose/Role.pmt`, `internal/parse/declarations/Mojolicious/Lite.pmt`, `docs/rfcs/0001-type-libraries.md`, `internal/data/help/types.md`
- commands: `go test ./internal/parse/ -count=1`, `go test ./internal/lexer/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `parseSubDecl` (`internal/parse/decl.go:383`), `declaredSub` (`internal/parse/exports.go:425`), `readModule`/`readSubs` (`exports.go:108`, `:211`), the declaration read path (`internal/parse/module.go:288`, `internal/parse/declarations.go:39`), `parseRoot` (`internal/parse/parse.go:487`), the lexer's prototype-vs-signature switch (`internal/lexer/proto.go:34`)

RFC 0001 "Declaration order: Perl's", "Type names", and the parse half of "Typed Perl, in `.pmt` only". A `.pmt` declaration is `sub NAME :attrs (Type $v = default, ...) Type;`. Today `parseSubDecl` reads name, attributes, then the signature as a parenthesised list, and `declaredSub` returns only name and prototype; nothing reads types. Typed Perl is accepted only when the source is a `.pmt`, so the declaration read path needs its own parse entry point (a parser flag set by `module.go:288` and `coreTable`), and `.pm`/`.pl` behaviour must not change: without the signatures feature `sub f (Ref $x) {...}` is a prototype. In typed mode the lexer must read `(` after `sub NAME` as a signature (as `l.signatures` does), so the three bare-prototype lines in the shipped declarations (`Moose.pmt:21`, `Moose/Role.pmt:23`, `Mojolicious/Lite.pmt:21`) move to `:prototype(...)`, as perl itself requires once signatures are on. Hold the parsed signature as a value from `internal/types` (`types.Signature`/`types.Param` in `internal/types/declared.go`: name, sigil, type, default text, return type) so later selection and infer code consume it without importing `internal/parse`. Container types (`Name[...]`) are the next issue's. Measured on 5.42: `sub f :lvalue ($x) {}` compiles, `sub f ($x) :lvalue {}` dies ("Subroutine attributes must come before the signature"). If this overruns one session, split defaults and return types into their own issue.

## Steps

- [ ] Write failing test `TestPmtTypedDeclaration` in `internal/parse/pmt_test.go`: in a `.pmt`, `sub bless :prototype($;$) (Ref $ref, Str $class = __PACKAGE__) Object;` yields name, prototype, two typed params (second with a default) and return `Object`; `Str|Undef $x` gives the union
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestPmtTypedDeclaration$' -count=1 -v`
- [ ] Add the `.pmt` parse entry point and lexer signature mode; read parameter types, defaults and return type into `types.Signature`; record them in `moduleFacts` beside `protos`; move the three bare prototypes to `:prototype(...)`
- [ ] Run test to verify pass; `make test`; commit
- [ ] Write failing test `TestPmtUnknownTypeNameIsError`: `sub f (Strng $x);` in a `.pmt` is reported as an error naming `Strng`
- [ ] Implement via `types.FromName`; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestPmtSignatureBeforeAttributesRefused`: `sub f (Str $x) :lvalue;` in a `.pmt` is refused, as perl refuses it
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestTypedSyntaxOnlyInPmt`: the same `sub f (Ref $x) { 1 }` parsed as ordinary source keeps `(Ref $x)` as a prototype and records no types
- [ ] Run test to verify pass (fix if not); `make test`; commit
- [ ] Re-mark the RFC headings "Declaration order: Perl's" and "Type names" `(*Implemented*, <sha>)`; commit
- [ ] Doc: the Mojolicious::Lite example in `internal/data/help/types.md` and in the RFC's "Declaration files" section shows `sub group (&);`; change both to the shipped spelling, `sub group :prototype(&);`; `go test ./internal/cli/ -run '^TestShowTypesHelp$' -count=1`; commit

## Acceptance criteria

- [ ] a .pmt declaration carries typed parameters, defaults and a return type (`go test ./internal/parse/ -run '^TestPmtTypedDeclaration$' -count=1 -v | grep -q '^--- PASS: TestPmtTypedDeclaration'`)
- [ ] an unknown type name in a .pmt is an error (`go test ./internal/parse/ -run '^TestPmtUnknownTypeNameIsError$' -count=1 -v | grep -q '^--- PASS: TestPmtUnknownTypeNameIsError'`)
- [ ] a signature before the attributes is refused, as perl refuses it (`go test ./internal/parse/ -run '^TestPmtSignatureBeforeAttributesRefused$' -count=1 -v | grep -q '^--- PASS: TestPmtSignatureBeforeAttributesRefused'`)
- [ ] ordinary source still reads (Ref $x) as a prototype (`go test ./internal/parse/ -run '^TestTypedSyntaxOnlyInPmt$' -count=1 -v | grep -q '^--- PASS: TestTypedSyntaxOnlyInPmt'`)
- [ ] the RFC section on declaration order is implemented (`grep -q '^### Declaration order.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] the RFC section on type names is implemented (`grep -q '^### Type names.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] the help page and the RFC show the shipped prototype spelling (`sh -c '! grep -qF "sub group (&)" internal/data/help/types.md docs/rfcs/0001-type-libraries.md'`)
