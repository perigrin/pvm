---
title: Declare POSIX
state: pending
urgency: normal
milestone: core-module-types
created: 2026-10-16T20:02:15.687306804Z
updated: 2026-10-16T20:02:15.687306804Z
---

## Prerequisites

- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one
- [ ] Pin the core-module ranking in a test and record the rule in the RFC
- [ ] Hold Carp.tpm to perl: prototypes, export lists and a return-type oracle

## Context

- paths: `internal/parse/declarations/` (one `.tpm` per module below), `internal/parse/heldtoperl_test.go`, `internal/infer/returntypes_test.go`, `internal/infer/testdata/oracle/`
- commands: `go test ./internal/parse/ -run '^TestPOSIXPrototypes$|^TestPOSIXExports$' -count=1 -v`, `go test ./internal/infer/ -run '^TestPOSIXReturnTypes$' -count=1 -v`
- entrypoints: `internal/parse/declarations/Carp.tpm` (the model: measured facts in comments), the held-to-perl helpers from the Carp issue

Modules (from the ranking): POSIX (1).

Write one `.tpm` per module declaring every name in its `@EXPORT`, `@EXPORT_OK` and `%EXPORT_TAGS` with prototype and typed signature, plus the subs and methods its callers use; a declaration is read in place of the module's source, so a name it leaves out is a name the import loses. Each file starts with two `ABOUTME:` lines naming the perl (5.42.0) and module version measured. A sub that never returns is declared `None`. POSIX is large: measured 2026-10-16 on 5.42.0, after `import` it has 587 names in `@EXPORT`, 179 in `@EXPORT_OK` and 35 tags, built in `import` rather than at load. Prototypes from perl's `prototype()`. `POSIX::_exit` is declared `None` and is never called by the oracle.

A `.tpm` changes what the parser reads for every file that uses the module, so run the m1 gates after this batch, not only at the end: `TestParseRatchet`, `TestPerlGitTRatchet`, `TestCanonRatchet`, `TestCanonFidelityRatchet`, `TestT2HoldsAll56`, `TestT2CoreParses`, `TestImportResolutionIsWorthT1Files` (internal/parse), `TestRatchet`, `TestRatchetCorpus`, `TestGoParserExactFloor` (internal/parseoracle), `TestPSCCheckLibBaseline` (internal/psc). A ratchet that moves is updated in the commit that moved it, with the reason.

## Steps

- [ ] Write failing tests `TestPOSIXPrototypes` and `TestPOSIXExports` in `internal/parse/heldtoperl_test.go` over the module list above
- [ ] Run to verify failure: `go test ./internal/parse/ -run '^TestPOSIXPrototypes$|^TestPOSIXExports$' -count=1 -v`
- [ ] Write the `.tpm` files; run to verify pass; commit
- [ ] Write failing test `TestPOSIXReturnTypes` in `internal/infer/returntypes_test.go`, with call cases in `internal/infer/testdata/oracle/` and every uncallable sub listed by name with its reason
- [ ] Run to verify failure: `go test ./internal/infer/ -run '^TestPOSIXReturnTypes$' -count=1 -v`
- [ ] Fix declared return types where perl disagrees; run to verify pass; commit
- [ ] Run the m1 gates above; `make test`; commit

## Acceptance Criteria

### Positive Scenarios
- [ ] declared prototypes equal perl's prototype() (`go test ./internal/parse/ -run '^TestPOSIXPrototypes$' -count=1 -v | grep -q '^--- PASS: TestPOSIXPrototypes'`)
- [ ] declared export lists equal the modules' own (`go test ./internal/parse/ -run '^TestPOSIXExports$' -count=1 -v | grep -q '^--- PASS: TestPOSIXExports'`)
- [ ] declared return types hold for every value perl returns (`go test ./internal/infer/ -run '^TestPOSIXReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestPOSIXReturnTypes'`)
