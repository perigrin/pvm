---
title: "Declare the variable exporters: Config, Getopt::Long, Test::More, Text::Wrap, ExtUtils::MakeMaker, File::stat"
state: pending
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-dadb-7796-a704-4007150a1bf7
created: 2026-10-16T20:02:15.843622709Z
updated: 2026-10-16T20:03:18.491278168Z
---

## Prerequisites

- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one
- [ ] Pin the core-module ranking in a test and record the rule in the RFC
- [ ] Hold Carp.tpm to perl: prototypes, export lists and a return-type oracle
- [ ] Read typed variables with :export and :export_ok in declaration files

## Context

- paths: `internal/parse/declarations/` (one `.tpm` per module below), `internal/parse/heldtoperl_test.go`, `internal/infer/returntypes_test.go`, `internal/infer/testdata/oracle/`
- commands: `go test ./internal/parse/ -run '^TestVariableExportersPrototypes$|^TestVariableExportersExports$' -count=1 -v`, `go test ./internal/infer/ -run '^TestVariableExportersReturnTypes$' -count=1 -v`
- entrypoints: `internal/parse/declarations/Carp.tpm` (the model: measured facts in comments), the held-to-perl helpers from the Carp issue

Modules (from the ranking): Config, Getopt::Long, Test::More, Text::Wrap, ExtUtils::MakeMaker, File::stat (6).

Write one `.tpm` per module declaring every name in its `@EXPORT`, `@EXPORT_OK` and `%EXPORT_TAGS` with prototype and typed signature, plus the subs and methods its callers use; a declaration is read in place of the module's source, so a name it leaves out is a name the import loses. Each file starts with two `ABOUTME:` lines naming the perl (5.42.0) and module version measured. A sub that never returns is declared `None`. Each exports a variable (measured 2026-10-16): Config `%Config` (`:export`) and subs in `@EXPORT_OK`; Getopt::Long `$REQUIRE_ORDER $PERMUTE $RETURN_IN_ORDER`; Test::More `$TODO`; Text::Wrap `$columns $break $huge`; ExtUtils::MakeMaker `$Verbose $VERSION`; File::stat thirteen `$st_*`. Declare each as `our Type <var> :export;` or `:export_ok`. Config and Test::More have their own `import`; their export lists are what that import brings in. Test::More is used across the T1 corpus, so expect T1 to move.

A `.tpm` changes what the parser reads for every file that uses the module, so run the m1 gates after this batch, not only at the end: `TestParseRatchet`, `TestPerlGitTRatchet`, `TestCanonRatchet`, `TestCanonFidelityRatchet`, `TestT2HoldsAll56`, `TestT2CoreParses`, `TestImportResolutionIsWorthT1Files` (internal/parse), `TestRatchet`, `TestRatchetCorpus`, `TestGoParserExactFloor` (internal/parseoracle), `TestPSCCheckLibBaseline` (internal/psc). A ratchet that moves is updated in the commit that moved it, with the reason.

## Steps

- [ ] Write failing tests `TestVariableExportersPrototypes` and `TestVariableExportersExports` in `internal/parse/heldtoperl_test.go` over the module list above
- [ ] Run to verify failure: `go test ./internal/parse/ -run '^TestVariableExportersPrototypes$|^TestVariableExportersExports$' -count=1 -v`
- [ ] Write the `.tpm` files; run to verify pass; commit
- [ ] Write failing test `TestVariableExportersReturnTypes` in `internal/infer/returntypes_test.go`, with call cases in `internal/infer/testdata/oracle/` and every uncallable sub listed by name with its reason
- [ ] Run to verify failure: `go test ./internal/infer/ -run '^TestVariableExportersReturnTypes$' -count=1 -v`
- [ ] Fix declared return types where perl disagrees; run to verify pass; commit
- [ ] Run the m1 gates above; `make test`; commit

## Acceptance Criteria

### Positive Scenarios
- [ ] declared prototypes equal perl's prototype() (`go test ./internal/parse/ -run '^TestVariableExportersPrototypes$' -count=1 -v | grep -q '^--- PASS: TestVariableExportersPrototypes'`)
- [ ] declared export lists equal the modules' own (`go test ./internal/parse/ -run '^TestVariableExportersExports$' -count=1 -v | grep -q '^--- PASS: TestVariableExportersExports'`)
- [ ] declared return types hold for every value perl returns (`go test ./internal/infer/ -run '^TestVariableExportersReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestVariableExportersReturnTypes'`)
