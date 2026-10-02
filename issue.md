---
title: Type CORE.pmt's prototyped builtins and hold every derived prototype to perl
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-4850-704d-8c14-c15add6194f2
blocks:
- 01a0fd9b-5670-78d0-84be-9b9618e9b081
created: 2026-10-02T17:13:29.229586875Z
updated: 2026-10-02T17:22:58.830173835Z
---

## Prerequisites

- [ ] Read the :lvalue parameter attribute and derive the \$ prototype from it

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/declarations.go`, `internal/parse/coretypes_test.go` (new), `internal/parse/export_test.go`, `docs/rfcs/0001-type-libraries.md`, `internal/data/help/types.md`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `internal/parse/declarations/CORE.pmt` (188 `sub NAME :prototype(...);` lines), `coreTable` (`internal/parse/declarations.go:39`), `TestCoreDeclarationsArePerls` (`internal/parse/coretable_test.go:20`) for asking perl via `conformance.PerlPath()`, `internal/types/signatures.go:30-115` for the 27 typed builtins and their measurement comments

Every one of CORE.pmt's 188 prototyped builtins gets a typed signature (RFC "Typed Perl, in `.pmt` only": `sub bless :prototype($;$) (Ref $ref, Str $class = __PACKAGE__) Object;`). The 27 builtins `internal/types/signatures.go` types take its argument and return types, and its measurement comments move above their CORE.pmt lines. Copy those 27 rows as golden values into the new test -- never compare against signatures.go, which a later issue deletes, and `git-zhi verify` re-runs this issue's criteria on the final tree. Element types the paper has not defined are written as plain `List`. The derived prototype of every line is held to `prototype("CORE::name")` for all 188, which finishes RFC "A typed signature and a prototype say the same thing" (the `\$` row came with `:lvalue`). Keep `coreTable()` and `TestCoreDeclarationsArePerls` green.

## Steps

- [ ] Write failing test `TestCoreDerivedPrototypesArePerls` in `internal/parse/coretypes_test.go`: every CORE.pmt line's prototype derived from its TYPES equals perl's (skip only when no perl, as `TestCoreDeclarationsArePerls` does)
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestCoreDerivedPrototypesArePerls$' -count=1 -v`
- [ ] Type the 188 lines, measuring with perl where the type is not obvious; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestCoreTypesMatchMeasuredSignatures`: the 27 golden rows (arity, argument and return types) equal their CORE.pmt declarations
- [ ] Move signatures.go's measurement comments into CORE.pmt; fix any line that disagrees, measuring which side is right; run test to verify pass; `make test`; commit
- [ ] Re-mark the RFC heading "A typed signature and a prototype say the same thing" `(*Implemented*, <sha>)`; commit
- [ ] Doc: `internal/data/help/types.md` says the shipped files "use only the plain-Perl part"; replace that with a sentence saying `CORE.pmt` carries typed declarations, and show its typed `bless` line; in the RFC's "CORE.pmt: the builtin table" section, show typed lines in place of the bare-prototype ones; `go test ./internal/cli/ -run '^TestShowTypesHelp$' -count=1`; commit

## Acceptance criteria

- [ ] every prototype derived from CORE.pmt types is perl's (`go test ./internal/parse/ -run '^TestCoreDerivedPrototypesArePerls$' -count=1 -v | grep -q '^--- PASS: TestCoreDerivedPrototypesArePerls'`)
- [ ] CORE.pmt carries the measured types of the builtins internal/infer types today (`go test ./internal/parse/ -run '^TestCoreTypesMatchMeasuredSignatures$' -count=1 -v | grep -q '^--- PASS: TestCoreTypesMatchMeasuredSignatures'`)
- [ ] the RFC section on typed signatures and prototypes is implemented (`grep -q '^### A typed signature and a prototype.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] the help page shows a typed CORE.pmt declaration and no longer calls the shipped files untyped (`sh -c '! grep -q "plain-Perl part" internal/data/help/types.md && grep -qF "sub bless" internal/data/help/types.md'`)
