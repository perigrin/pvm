---
title: Parse typed parameters and defaults in .pmt declarations
state: done
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-3fbf-7c6e-90a7-919d8670f758
blocks:
- 01a0fd9b-46c4-7a8c-b8d6-498c5182a47f
- 01a10177-49d9-7915-975e-2a8e3da943bf
created: 2026-10-02T17:13:25.560580634Z
updated: 2026-10-16T18:48:30.886945934Z
sessions:
- start_sha: 337657c599353a2fbf768616a29f3eebc377d0a3
  end_sha: 337657c599353a2fbf768616a29f3eebc377d0a3
  commits: 0
  started_at: 2026-10-05T23:10:52.509379968Z
  ended_at: 2026-10-05T23:10:52.705303477Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-05T23:10:52.509379968Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-05T23:10:52.705303477Z
---

## Prerequisites

- [ ] Add IO, FileHandle and Boolean to the type lattice, with lookup by name

## Context

- paths: `internal/parse/decl.go`, `internal/parse/exports.go`, `internal/parse/module.go`, `internal/parse/declarations.go`, `internal/parse/parse.go`, `internal/lexer/proto.go`, `internal/types/declared.go` (new), `internal/parse/pmt_test.go` (new), `internal/parse/declarations/Moose.pmt`, `internal/parse/declarations/Moose/Role.pmt`, `internal/parse/declarations/Mojolicious/Lite.pmt`, `docs/rfcs/0001-type-libraries.md`, `internal/data/help/types.md`
- commands: `go test ./internal/parse/ -count=1`, `go test ./internal/lexer/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `parseSubDecl` (`internal/parse/decl.go:383`), `declaredSub` (`internal/parse/exports.go:425`), `readModule`/`readSubs` (`exports.go:108`, `:211`), the declaration read path (`internal/parse/module.go:288`, `internal/parse/declarations.go:39`), `parseRoot` (`internal/parse/parse.go:487`), the lexer's prototype-vs-signature switch (`internal/lexer/proto.go:34`)

The first half of RFC 0001 "Typed Perl, in `.pmt` only": a `.pmt` declaration is `sub NAME :attrs (Type $v = default, ...) Type;`, and this issue reads everything up to the closing parenthesis -- the `.pmt` entry point, the lexer's signature mode, typed parameters named by a single lattice name, and defaults. Return types, unions, container types and unknown-name errors are the next issue's. Today `parseSubDecl` reads name, attributes, then the signature as a parenthesised list, and `declaredSub` returns only name and prototype; nothing reads types. Typed Perl is accepted only when the source is a `.pmt`, so the declaration read path needs its own parse entry point (a parser flag set by `module.go:288` and `coreTable`), and `.pm`/`.pl` behaviour must not change: without the signatures feature `sub f (Ref $x) {...}` is a prototype. In typed mode the lexer reads `(` after `sub NAME` as a signature, as `l.signatures` does. Hold the parsed signature as a value from `internal/types` (`types.Signature`/`types.Param` in `internal/types/declared.go`: name, sigil, type, default text; the return type field is filled by the next issue) so later selection and infer code consume it without importing `internal/parse`. Defaults parse on slurpy parameters too: RFC 0001 "A required argument defaults to `die`" (perigrin, 2026-10-05) makes a default on a slurpy a typed-Perl extension (perl's signatures reject it, "A slurpy parameter may not have a default value"), used for `= ($_)` (`print`, `chomp`) and `= die` (`sort`). A `die` default runs only when its argument is omitted, so it makes the parameter required: record it as required on `types.Param`, not as an optional parameter with a default. Before signature mode is switched on, the shipped declarations' three bare prototypes (`sub blessed ($);` at `Moose.pmt:21` and `Moose/Role.pmt:23`, `sub group (&);` at `Mojolicious/Lite.pmt:21`) move to `:prototype(...)`, as perl itself requires once signatures are on, so signature mode never meets a bare prototype. Re-marks no RFC section.

## Steps

- [ ] Move the three bare prototypes in the shipped declarations to `:prototype(...)` (`sub blessed :prototype($);`, `sub group :prototype(&);`); run `go test ./internal/parse/ -count=1` to verify they still read with the same prototypes; `make test`; commit
- [ ] Doc: the Mojolicious::Lite example in `internal/data/help/types.md` and in the RFC's "Declaration files" section shows `sub group (&);`; change both to the shipped spelling, `sub group :prototype(&);`; `go test ./internal/cli/ -run '^TestShowTypesHelp$' -count=1`; commit
- [ ] Write failing test `TestPmtTypedParameters` in `internal/parse/pmt_test.go`: in a `.pmt`, `sub bless :prototype($;$) (Ref $ref, Str $class = __PACKAGE__);` yields name, prototype and two typed params, the second with the default `__PACKAGE__`
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestPmtTypedParameters$' -count=1 -v`
- [ ] Add the `.pmt` parse entry point and lexer signature mode; read parameter types (via `types.FromName`) and defaults into `types.Signature`; record them in `moduleFacts` beside `protos`
- [ ] Run test to verify pass; `make test`; commit
- [ ] Write failing test `TestPmtSlurpyDefaultsParse`: in a `.pmt`, `sub f (List @l = ($_));` gives one slurpy with the default `($_)`, optional; `sub g (List @list = die);` gives one slurpy recorded as required; `sub h (List @list);` gives one slurpy, optional with no default
- [ ] Run test to verify failure, implement, run test to verify pass: `go test ./internal/parse/ -run '^TestPmtSlurpyDefaultsParse$' -count=1 -v`; `make test`; commit
- [ ] Write failing test `TestTypedSyntaxOnlyInPmt`: the same `sub f (Ref $x) { 1 }` parsed as ordinary source keeps `(Ref $x)` as a prototype and records no types
- [ ] Run test to verify pass (fix if not); `make test`; commit
- [ ] Write failing test `TestTypedSyntaxNotEnabledBySignaturesFeature`: in a `.pm` under `use v5.36`, `sub f (Ref $x) { 1 }` records no types
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestTypedSyntaxNotEnabledBySignaturesFeature$' -count=1 -v`
- [ ] Implement; run test to verify pass; `make test`; commit

### Negative Scenarios

- truncated or malformed typed parameters in a `.pmt` are errors, not panics or silent partial records: EOF mid-signature (`sub f (Str $x`), a type with no variable (`sub f (Str);`), a default with no expression (`sub f (Str $x = );`) (`go test ./internal/parse/ -run '^TestPmtMalformedTypedDeclarationIsError$' -count=1 -v | grep -q '^--- PASS: TestPmtMalformedTypedDeclarationIsError'`)
- the signatures feature does not switch typed Perl on outside a `.pmt`: in a `.pm` under `use v5.36`, `sub f (Ref $x) { 1 }` records no types (perl 5.42 refuses it: A signature parameter must start with `$`, `@` or `%`) (`go test ./internal/parse/ -run '^TestTypedSyntaxNotEnabledBySignaturesFeature$' -count=1 -v | grep -q '^--- PASS: TestTypedSyntaxNotEnabledBySignaturesFeature'`)

## Acceptance criteria

- [ ] a .pmt declaration carries typed parameters and defaults (`go test ./internal/parse/ -run '^TestTpmTypedParameters$' -count=1 -v | grep -q '^--- PASS: TestTpmTypedParameters'`)
- [ ] defaults on slurpy parameters parse, and a die default is recorded as required (`go test ./internal/parse/ -run '^TestTpmSlurpyDefaultsParse$' -count=1 -v | grep -q '^--- PASS: TestTpmSlurpyDefaultsParse'`)
- [ ] ordinary source still reads (Ref $x) as a prototype (`go test ./internal/parse/ -run '^TestTypedSyntaxOnlyInTpm$' -count=1 -v | grep -q '^--- PASS: TestTypedSyntaxOnlyInTpm'`)
- [ ] malformed or truncated .pmt parameters are errors, not panics (`go test ./internal/parse/ -run '^TestTpmMalformedTypedDeclarationIsError$' -count=1 -v | grep -q '^--- PASS: TestTpmMalformedTypedDeclarationIsError'`)
- [ ] the signatures feature does not enable typed Perl in ordinary source (`go test ./internal/parse/ -run '^TestTypedSyntaxNotEnabledBySignaturesFeature$' -count=1 -v | grep -q '^--- PASS: TestTypedSyntaxNotEnabledBySignaturesFeature'`)
- [ ] the shipped declarations carry no bare prototype (`sh -c '! grep -rqF -e "sub blessed (\$);" -e "sub group (&);" internal/parse/declarations'`)
- [ ] the help page and the RFC show the shipped prototype spelling (`sh -c '! grep -qF "sub group (&)" internal/data/help/types.md docs/rfcs/0001-type-libraries.md'`)
