---
title: "Declare the pure-Perl exporters: File::Basename, Symbol, File::Find, File::Path, Errno and five more"
state: pending
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-d912-7419-8b20-a8e24029b3ec
- 01a1464e-d980-7382-970e-067a5ea3db84
created: 2026-10-16T20:02:15.500338342Z
updated: 2026-10-16T20:03:40.694494136Z
---

## Prerequisites

- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one
- [ ] Pin the core-module ranking in a test and record the rule in the RFC
- [ ] Hold Carp.tpm to perl: prototypes, export lists and a return-type oracle

## Context

- paths: `internal/parse/declarations/` (one `.tpm` per module below), `internal/parse/heldtoperl_test.go`, `internal/infer/returntypes_test.go`, `internal/infer/testdata/oracle/`
- commands: `go test ./internal/parse/ -run '^TestPurePerlExportersPrototypes$|^TestPurePerlExportersExports$' -count=1 -v`, `go test ./internal/infer/ -run '^TestPurePerlExportersReturnTypes$' -count=1 -v`
- entrypoints: `internal/parse/declarations/Carp.tpm` (the model: measured facts in comments), the held-to-perl helpers from the Carp issue

Modules (from the ranking): File::Basename, Symbol, File::Find, File::Path, Errno, Text::ParseWords, Class::Struct, File::Copy, Time::Local, Getopt::Std (10; Carp is held to perl by its own issue).

Write one `.tpm` per module declaring every name in its `@EXPORT`, `@EXPORT_OK` and `%EXPORT_TAGS` with prototype and typed signature, plus the subs and methods its callers use; a declaration is read in place of the module's source, so a name it leaves out is a name the import loses. Each file starts with two `ABOUTME:` lines naming the perl (5.42.0) and module version measured. A sub that never returns is declared `None`. Errno's names are constants generated at build (`ENOENT` and the rest, 134 in `@EXPORT_OK`, tag `:POSIX`); measured 2026-10-16, Errno exports no variable -- `%!` is not in its export lists -- so it is here, not with the variable exporters.

A `.tpm` changes what the parser reads for every file that uses the module, so run the m1 gates after this batch, not only at the end: `TestParseRatchet`, `TestPerlGitTRatchet`, `TestCanonRatchet`, `TestCanonFidelityRatchet`, `TestT2HoldsAll56`, `TestT2CoreParses`, `TestImportResolutionIsWorthT1Files` (internal/parse), `TestRatchet`, `TestRatchetCorpus`, `TestGoParserExactFloor` (internal/parseoracle), `TestPSCCheckLibBaseline` (internal/psc). A ratchet that moves is updated in the commit that moved it, with the reason.

## Steps

- [ ] Write failing tests `TestPurePerlExportersPrototypes` and `TestPurePerlExportersExports` in `internal/parse/heldtoperl_test.go` over the module list above
- [ ] Run to verify failure: `go test ./internal/parse/ -run '^TestPurePerlExportersPrototypes$|^TestPurePerlExportersExports$' -count=1 -v`
- [ ] Write the `.tpm` files; run to verify pass; commit
- [ ] Write failing test `TestPurePerlExportersReturnTypes` in `internal/infer/returntypes_test.go`, with call cases in `internal/infer/testdata/oracle/` and every uncallable sub listed by name with its reason
- [ ] Run to verify failure: `go test ./internal/infer/ -run '^TestPurePerlExportersReturnTypes$' -count=1 -v`
- [ ] Fix declared return types where perl disagrees; run to verify pass; commit
- [ ] Run the m1 gates above; `make test`; commit

## Acceptance Criteria

### Positive Scenarios
- [ ] declared prototypes equal perl's prototype() (`go test ./internal/parse/ -run '^TestPurePerlExportersPrototypes$' -count=1 -v | grep -q '^--- PASS: TestPurePerlExportersPrototypes'`)
- [ ] declared export lists equal the modules' own (`go test ./internal/parse/ -run '^TestPurePerlExportersExports$' -count=1 -v | grep -q '^--- PASS: TestPurePerlExportersExports'`)
- [ ] declared return types hold for every value perl returns (`go test ./internal/infer/ -run '^TestPurePerlExportersReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestPurePerlExportersReturnTypes'`)
