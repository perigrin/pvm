---
title: Parse typed parameters and defaults in .pmt declarations
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-3fbf-7c6e-90a7-919d8670f758
blocks:
- 01a0fd9b-46c4-7a8c-b8d6-498c5182a47f
created: 2026-10-02T17:13:25.560580634Z
updated: 2026-10-03T11:15:29.500129742Z
---

## Prerequisites

- [ ] Add IO, FileHandle and Boolean to the type lattice, with lookup by name

## Context

- paths: `internal/parse/decl.go`, `internal/parse/exports.go`, `internal/parse/module.go`, `internal/parse/declarations.go`, `internal/parse/parse.go`, `internal/lexer/proto.go`, `internal/types/declared.go` (new), `internal/parse/pmt_test.go` (new)
- commands: `go test ./internal/parse/ -count=1`, `go test ./internal/lexer/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `parseSubDecl` (`internal/parse/decl.go:383`), `declaredSub` (`internal/parse/exports.go:425`), `readModule`/`readSubs` (`exports.go:108`, `:211`), the declaration read path (`internal/parse/module.go:288`, `internal/parse/declarations.go:39`), `parseRoot` (`internal/parse/parse.go:487`), the lexer's prototype-vs-signature switch (`internal/lexer/proto.go:34`)

The first half of RFC 0001 "Typed Perl, in `.pmt` only": a `.pmt` declaration is `sub NAME :attrs (Type $v = default, ...) Type;`, and this issue reads everything up to the closing parenthesis -- the `.pmt` entry point, the lexer's signature mode, typed parameters named by a single lattice name, and defaults. Return types, unions, container types and unknown-name errors are the next issue's. Today `parseSubDecl` reads name, attributes, then the signature as a parenthesised list, and `declaredSub` returns only name and prototype; nothing reads types. Typed Perl is accepted only when the source is a `.pmt`, so the declaration read path needs its own parse entry point (a parser flag set by `module.go:288` and `coreTable`), and `.pm`/`.pl` behaviour must not change: without the signatures feature `sub f (Ref $x) {...}` is a prototype. In typed mode the lexer reads `(` after `sub NAME` as a signature, as `l.signatures` does. Hold the parsed signature as a value from `internal/types` (`types.Signature`/`types.Param` in `internal/types/declared.go`: name, sigil, type, default text; the return type field is filled by the next issue) so later selection and infer code consume it without importing `internal/parse`. Ordering hazard: the shipped declarations still carry three bare prototypes (`Moose.pmt:21`, `Moose/Role.pmt:23`, `Mojolicious/Lite.pmt:21`), which the next issue moves to `:prototype(...)`; `make test` must stay green here, so if signature mode makes those lines unreadable, stop and raise it with perigrin rather than special-casing `(&)`. Re-marks no RFC section.

## Steps

- [ ] Write failing test `TestPmtTypedParameters` in `internal/parse/pmt_test.go`: in a `.pmt`, `sub bless :prototype($;$) (Ref $ref, Str $class = __PACKAGE__);` yields name, prototype and two typed params, the second with the default `__PACKAGE__`
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestPmtTypedParameters$' -count=1 -v`
- [ ] Add the `.pmt` parse entry point and lexer signature mode; read parameter types (via `types.FromName`) and defaults into `types.Signature`; record them in `moduleFacts` beside `protos`
- [ ] Run test to verify pass; `make test`; commit
- [ ] Write failing test `TestTypedSyntaxOnlyInPmt`: the same `sub f (Ref $x) { 1 }` parsed as ordinary source keeps `(Ref $x)` as a prototype and records no types
- [ ] Run test to verify pass (fix if not); `make test`; commit
- [ ] Write failing test `TestTypedSyntaxNotEnabledBySignaturesFeature`: in a `.pm` under `use v5.36`, `sub f (Ref $x) { 1 }` records no types
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestTypedSyntaxNotEnabledBySignaturesFeature$' -count=1 -v`
- [ ] Implement; run test to verify pass; `make test`; commit

### Negative Scenarios

- truncated or malformed typed parameters in a `.pmt` are errors, not panics or silent partial records: EOF mid-signature (`sub f (Str $x`), a type with no variable (`sub f (Str);`), a default with no expression (`sub f (Str $x = );`) (`go test ./internal/parse/ -run '^TestPmtMalformedTypedDeclarationIsError$' -count=1 -v | grep -q '^--- PASS: TestPmtMalformedTypedDeclarationIsError'`)
- the signatures feature does not switch typed Perl on outside a `.pmt`: in a `.pm` under `use v5.36`, `sub f (Ref $x) { 1 }` records no types (perl 5.42 refuses it: A signature parameter must start with `$`, `@` or `%`) (`go test ./internal/parse/ -run '^TestTypedSyntaxNotEnabledBySignaturesFeature$' -count=1 -v | grep -q '^--- PASS: TestTypedSyntaxNotEnabledBySignaturesFeature'`)

## Acceptance criteria

- [ ] a .pmt declaration carries typed parameters and defaults (`go test ./internal/parse/ -run '^TestPmtTypedParameters$' -count=1 -v | grep -q '^--- PASS: TestPmtTypedParameters'`)
- [ ] ordinary source still reads (Ref $x) as a prototype (`go test ./internal/parse/ -run '^TestTypedSyntaxOnlyInPmt$' -count=1 -v | grep -q '^--- PASS: TestTypedSyntaxOnlyInPmt'`)
- [ ] malformed or truncated .pmt parameters are errors, not panics (`go test ./internal/parse/ -run '^TestPmtMalformedTypedDeclarationIsError$' -count=1 -v | grep -q '^--- PASS: TestPmtMalformedTypedDeclarationIsError'`)
- [ ] the signatures feature does not enable typed Perl in ordinary source (`go test ./internal/parse/ -run '^TestTypedSyntaxNotEnabledBySignaturesFeature$' -count=1 -v | grep -q '^--- PASS: TestTypedSyntaxNotEnabledBySignaturesFeature'`)
