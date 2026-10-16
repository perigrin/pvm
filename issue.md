---
title: "Declare the IO and file-handling exporters: FileHandle, IO::Handle, IO::File, IO::Socket, File::Temp and four more"
state: pending
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-d912-7419-8b20-a8e24029b3ec
- 01a1464e-d980-7382-970e-067a5ea3db84
created: 2026-10-16T20:02:15.542610643Z
updated: 2026-10-16T20:03:47.230419156Z
---

## Prerequisites

- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one
- [ ] Pin the core-module ranking in a test and record the rule in the RFC
- [ ] Hold Carp.tpm to perl: prototypes, export lists and a return-type oracle

## Context

- paths: `internal/parse/declarations/` (one `.tpm` per module below), `internal/parse/heldtoperl_test.go`, `internal/infer/returntypes_test.go`, `internal/infer/testdata/oracle/`
- commands: `go test ./internal/parse/ -run '^TestIOExportersPrototypes$|^TestIOExportersExports$' -count=1 -v`, `go test ./internal/infer/ -run '^TestIOExportersReturnTypes$' -count=1 -v`
- entrypoints: `internal/parse/declarations/Carp.tpm` (the model: measured facts in comments), the held-to-perl helpers from the Carp issue

Modules (from the ranking): FileHandle, IO::Handle, IO::File, IO::Socket, File::Spec::Functions, File::Temp, Archive::Tar, ExtUtils::Manifest, Net::Ping (9).

Write one `.tpm` per module declaring every name in its `@EXPORT`, `@EXPORT_OK` and `%EXPORT_TAGS` with prototype and typed signature, plus the subs and methods its callers use; a declaration is read in place of the module's source, so a name it leaves out is a name the import loses. Each file starts with two `ABOUTME:` lines naming the perl (5.42.0) and module version measured. A sub that never returns is declared `None`. IO::Handle, IO::File, FileHandle and IO::Socket are classes as well as exporters: declare the methods their callers use, beside the exported names. Net::Ping and IO::Socket touch the network and Archive::Tar and File::Temp the filesystem: list what the oracle cannot call.

A `.tpm` changes what the parser reads for every file that uses the module, so run the m1 gates after this batch, not only at the end: `TestParseRatchet`, `TestPerlGitTRatchet`, `TestCanonRatchet`, `TestCanonFidelityRatchet`, `TestT2HoldsAll56`, `TestT2CoreParses`, `TestImportResolutionIsWorthT1Files` (internal/parse), `TestRatchet`, `TestRatchetCorpus`, `TestGoParserExactFloor` (internal/parseoracle), `TestPSCCheckLibBaseline` (internal/psc). A ratchet that moves is updated in the commit that moved it, with the reason.

## Steps

- [ ] Write failing tests `TestIOExportersPrototypes` and `TestIOExportersExports` in `internal/parse/heldtoperl_test.go` over the module list above
- [ ] Run to verify failure: `go test ./internal/parse/ -run '^TestIOExportersPrototypes$|^TestIOExportersExports$' -count=1 -v`
- [ ] Write the `.tpm` files; run to verify pass; commit
- [ ] Write failing test `TestIOExportersReturnTypes` in `internal/infer/returntypes_test.go`, with call cases in `internal/infer/testdata/oracle/` and every uncallable sub listed by name with its reason
- [ ] Run to verify failure: `go test ./internal/infer/ -run '^TestIOExportersReturnTypes$' -count=1 -v`
- [ ] Fix declared return types where perl disagrees; run to verify pass; commit
- [ ] Run the m1 gates above; `make test`; commit

## Acceptance Criteria

### Positive Scenarios
- [ ] declared prototypes equal perl's prototype() (`go test ./internal/parse/ -run '^TestIOExportersPrototypes$' -count=1 -v | grep -q '^--- PASS: TestIOExportersPrototypes'`)
- [ ] declared export lists equal the modules' own (`go test ./internal/parse/ -run '^TestIOExportersExports$' -count=1 -v | grep -q '^--- PASS: TestIOExportersExports'`)
- [ ] declared return types hold for every value perl returns (`go test ./internal/infer/ -run '^TestIOExportersReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestIOExportersReturnTypes'`)
