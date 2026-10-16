---
title: "Declare the XS library modules: Encode, Time::HiRes, Compress::Raw::Zlib, SDBM_File"
state: pending
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-d912-7419-8b20-a8e24029b3ec
- 01a1464e-d980-7382-970e-067a5ea3db84
- 01a1464e-db1c-73eb-a3a9-e897d7a56b92
blocks:
- 01a1464e-de34-7403-95d2-25f4855404b2
created: 2026-10-16T20:02:15.635645821Z
updated: 2026-10-16T20:13:05.848817769Z
---

## Prerequisites

- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one
- [ ] Pin the core-module ranking in a test and record the rule in the RFC
- [ ] Hold Carp.tpm to perl: prototypes, export lists and a return-type oracle

## Context

- paths: `internal/parse/declarations/` (one `.tpm` per module below), `internal/parse/heldtoperl_test.go`, `internal/infer/returntypes_test.go`, `internal/infer/testdata/oracle/`
- commands: `go test ./internal/parse/ -run '^TestXSLibrariesPrototypes$|^TestXSLibrariesExports$' -count=1 -v`, `go test ./internal/infer/ -run '^TestXSLibrariesReturnTypes$' -count=1 -v`
- entrypoints: `internal/parse/declarations/Carp.tpm` (the model: measured facts in comments), the held-to-perl helpers from the Carp issue

Modules (from the ranking): Encode, Time::HiRes, Compress::Raw::Zlib, SDBM_File (4).

Write one `.tpm` per module declaring every name in its `@EXPORT`, `@EXPORT_OK` and `%EXPORT_TAGS` with prototype and typed signature, plus the subs and methods its callers use; a declaration is read in place of the module's source, so a name it leaves out is a name the import loses. Each file starts with two `ABOUTME:` lines naming the perl (5.42.0) and module version measured. A sub that never returns is declared `None`. XS: prototypes from perl's `prototype()`. Compress::Raw::Zlib exports 47 constants; SDBM_File is a tie class whose `@EXPORT_OK` holds constants.

A `.tpm` changes what the parser reads for every file that uses the module, so run the m1 gates after this batch, not only at the end: `TestParseRatchet`, `TestPerlGitTRatchet`, `TestCanonRatchet`, `TestCanonFidelityRatchet`, `TestT2HoldsAll56`, `TestT2CoreParses`, `TestImportResolutionIsWorthT1Files` (internal/parse), `TestRatchet`, `TestRatchetCorpus`, `TestGoParserExactFloor` (internal/parseoracle), `TestPSCCheckLibBaseline` (internal/psc). A ratchet that moves is updated in the commit that moved it, with the reason.

## Steps

- [ ] Write failing tests `TestXSLibrariesPrototypes` and `TestXSLibrariesExports` in `internal/parse/heldtoperl_test.go` over the module list above
- [ ] Run to verify failure: `go test ./internal/parse/ -run '^TestXSLibrariesPrototypes$|^TestXSLibrariesExports$' -count=1 -v`
- [ ] Write the `.tpm` files; run to verify pass; commit
- [ ] Write failing test `TestXSLibrariesReturnTypes` in `internal/infer/returntypes_test.go`, with call cases in `internal/infer/testdata/oracle/` and every uncallable sub listed by name with its reason
- [ ] Run to verify failure: `go test ./internal/infer/ -run '^TestXSLibrariesReturnTypes$' -count=1 -v`
- [ ] Fix declared return types where perl disagrees; run to verify pass; commit
- [ ] Run the m1 gates above; `make test`; commit

## Acceptance Criteria

### Positive Scenarios
- [ ] declared prototypes equal perl's prototype() (`go test ./internal/parse/ -run '^TestXSLibrariesPrototypes$' -count=1 -v | grep -q '^--- PASS: TestXSLibrariesPrototypes'`)
- [ ] declared export lists equal the modules' own (`go test ./internal/parse/ -run '^TestXSLibrariesExports$' -count=1 -v | grep -q '^--- PASS: TestXSLibrariesExports'`)
- [ ] declared return types hold for every value perl returns (`go test ./internal/infer/ -run '^TestXSLibrariesReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestXSLibrariesReturnTypes'`)

### Negative Scenarios
- [ ] a bare `use Time::HiRes;` imports nothing (its `@EXPORT` is empty, measured), so `time()` stays CORE's (`go test ./internal/parse/ -run '^TestTimeHiResBareUseKeepsCoreTime$' -count=1 -v | grep -q '^--- PASS: TestTimeHiResBareUseKeepsCoreTime'`)
- [ ] Encode `encode_utf8(undef)` returns a value its declared return type admits (`go test ./internal/infer/ -run '^TestEncodeUndefInputReturnIsDeclared$' -count=1 -v | grep -q '^--- PASS: TestEncodeUndefInputReturnIsDeclared'`)
- [ ] the oracle ties SDBM_File only under the test's temp dir (`go test ./internal/infer/ -run '^TestSDBMFileOracleStaysInScratch$' -count=1 -v | grep -q '^--- PASS: TestSDBMFileOracleStaysInScratch'`)
