---
title: Convert every .tpm to :export and :export_ok, then refuse export lists in a .tpm
state: pending
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-dadb-7796-a704-4007150a1bf7
blocks:
- 01a1464e-db1c-73eb-a3a9-e897d7a56b92
- 01a1464e-db4c-78ae-9bd5-84d71ccf5432
- 01a1464e-db76-704f-bd85-023a54d2cf6e
- 01a1464e-dba9-740b-b8b6-8da484195ce3
- 01a1464e-dbd3-7b16-801c-b9a5893d687b
- 01a1464e-dc07-754e-9ef4-637df15d51a9
- 01a14683-84ab-7efb-a0c5-b2db8deb8e28
- 01a1464e-dc48-720b-9ab9-fcc7fe47f483
- 01a1464e-dca3-7284-85c3-a06107c16646
- 01a1464e-dd27-7b82-ae71-3e6173375fa7
- 01a1464e-dd6c-7b5f-b95b-76286ec110a1
- 01a1464e-ddb0-720a-93f6-450c767e6e2e
- 01a1464e-dde3-73fe-ab1f-ced7f18c4f9d
created: 2026-10-16T20:59:46.420731313Z
updated: 2026-10-16T21:02:03.256742696Z
---

## Prerequisites

- [ ] Read typed variables with :export and :export_ok in declaration files (01a1464e-dadb)

## Context

- paths: `internal/parse/declarations/Carp.tpm`, `internal/parse/declarations/Moose.tpm`, `internal/parse/declarations/Moose/Role.tpm`, `internal/parse/declarations/Mojolicious/Lite.tpm`, `internal/parse/declarations/Future/AsyncAwait.tpm`, `internal/parse/testdata/declarations/My/Lib.tpm`, `internal/parse/exports.go`, `internal/parse/declarations.go`, `internal/parse/exportattr_test.go`, `docs/rfcs/0001-type-libraries.md` ("Declaration files"), `internal/data/help/types.md` ("Declaration Files" example)
- commands: `go test ./internal/parse/ -run '^TestShippedDeclarationsKeepTheirImports$|^TestDeclarationExportList' -count=1 -v`
- entrypoints: `readExportTags` and `moduleFacts.exports`/`defaults` (`exports.go`), `readDeclarationWith` (`declarations.go`), the attribute reader from 01a1464e-dadb

**Ruling 1 (perigrin, 2026-10-16): no export lists in a `.tpm`.** "We don't need the @EXPORT in the .tpm since we have :export and :export_ok." On names a module's own `import` installs: "we should figure out as best we can what is exported (loc, loc_lang, etc.) and mark them as :export or :export_ok accordingly."

**Ruling 2 (perigrin, 2026-10-16): migrate, then remove.** Every existing `.tpm` is converted to `:export`/`:export_ok` attributes, and the parser then refuses `our @EXPORT` / `@EXPORT_OK` / `%EXPORT_TAGS` in a `.tpm` as an error.

Every shipped `.tpm` declares its exports with `our @EXPORT` / `@EXPORT_OK` today: Carp, Moose, Moose::Role, Mojolicious::Lite, and the test fixture `My/Lib.tpm` (Future::AsyncAwait declares keywords and has no list; check it stays as it reads now). Convert each to `:export`/`:export_ok` on its subs, then make the parser refuse `our @EXPORT`, `our @EXPORT_OK` and `our %EXPORT_TAGS` in a `.tpm` as an error naming the file and line. A module's own source keeps its lists: the refusal is for `.tpm` files only. The conversion must not change what any `use` of these modules imports, except where perl says the list was wrong: measured 2026-10-16, `use Carp qw(verbose)` installs nothing in the caller (`verbose` is in Carp's `@EXPORT_OK` only as an `@EXPORT_FAIL` flag), so Carp.tpm marks no `verbose`, and that one change is recorded in `TestShippedDeclarationsKeepTheirImports`' table with the measurement. Moose, Moose::Role and Mojolicious::Lite are not core; keep their current names (measured when each was written) and mark them `:export`. Update the RFC's "Declaration files" example and `types.md`'s example to the attribute form.

This issue blocks the Carp helpers (01a1464e-db1c), every declaration batch, the guide and the call-site issue (01a0fd9b-5700, whose fixture uses `:export_ok`): ruling 2 puts it ahead of all of them.

A `.tpm` changes what the parser reads for every file that uses the module, so run the m1 gates after the conversion: `TestParseRatchet`, `TestPerlGitTRatchet`, `TestCanonRatchet`, `TestCanonFidelityRatchet`, `TestT2HoldsAll56`, `TestT2CoreParses`, `TestImportResolutionIsWorthT1Files` (internal/parse), `TestRatchet`, `TestRatchetCorpus`, `TestGoParserExactFloor`, `TestGoParserWrongIsZero` (internal/parseoracle), `TestPSCCheckLibBaseline` (internal/psc). Run each with the command the final-gate issue (01a1464e-de34) gives it: `TestRatchetCorpus` skips without `-parseoracle.corpus`, and the two Go-parser gates skip unless `PARSEORACLE_SHIM` names a built shim, and a skip proves nothing. A ratchet may only be raised. A drop is fixed in the .tpm, or the batch stops and asks perigrin.

## Steps

- [ ] Write `TestShippedDeclarationsKeepTheirImports`: a literal table of what a bare `use M` and `use M qw(name)` import for each shipped `.tpm` and the fixture, as the parser reads them today (with Carp's `verbose` row expecting the measured nothing); run to verify it fails only on `verbose`
- [ ] Convert each `.tpm` to `:export`/`:export_ok`; run to verify pass; commit
- [ ] Write failing tests `TestDeclarationExportListRefused` (`our @EXPORT` in a fixture `.tpm`), `TestDeclarationExportOkListRefused` and `TestDeclarationExportTagsRefused`, their fixtures inline or under `t.TempDir()`, never in `testdata/declarations/`
- [ ] Run to verify failure: `go test ./internal/parse/ -run '^TestDeclarationExport' -count=1 -v`
- [ ] Implement the refusal; run to verify pass; commit
- [ ] Update the RFC's "Declaration files" example and the `types.md` example; `go test ./internal/cli/ -run '^TestShowTypesHelp$' -count=1`; commit
- [ ] Run the m1 gates above; `make test`; commit

## Acceptance Criteria

### Positive Scenarios
- [ ] every shipped `.tpm` imports what it did, except Carp's measured `verbose` (`go test ./internal/parse/ -run '^TestShippedDeclarationsKeepTheirImports$' -count=1 -v | grep -q '^--- PASS: TestShippedDeclarationsKeepTheirImports'`)
- [ ] `our @EXPORT` in a `.tpm` is an error naming the file and line (`go test ./internal/parse/ -run '^TestDeclarationExportListRefused$' -count=1 -v | grep -q '^--- PASS: TestDeclarationExportListRefused'`)
- [ ] `our @EXPORT_OK` in a `.tpm` is an error (`go test ./internal/parse/ -run '^TestDeclarationExportOkListRefused$' -count=1 -v | grep -q '^--- PASS: TestDeclarationExportOkListRefused'`)
- [ ] `our %EXPORT_TAGS` in a `.tpm` is an error (`go test ./internal/parse/ -run '^TestDeclarationExportTagsRefused$' -count=1 -v | grep -q '^--- PASS: TestDeclarationExportTagsRefused'`)
- [ ] the RFC's "Declaration files" example uses attributes (`! grep -q 'our @EXPORT = qw(any get post' docs/rfcs/0001-type-libraries.md`)

### Negative Scenarios
- [ ] no shipped or fixture `.tpm` keeps an export list (`! grep -rqE '^[[:space:]]*our[[:space:]]+[@%]EXPORT' internal/parse/declarations internal/parse/testdata/declarations`)
- [ ] a module's own source with `our @EXPORT` is still read for its exports, not refused (`go test ./internal/parse/ -run '^TestModuleSourceExportListStillRead$' -count=1 -v | grep -q '^--- PASS: TestModuleSourceExportListStillRead'`)
- [ ] `use Carp qw(verbose)` imports nothing from the declaration (measured: perl installs nothing) (`go test ./internal/parse/ -run '^TestCarpVerboseImportsNothing$' -count=1 -v | grep -q '^--- PASS: TestCarpVerboseImportsNothing'`)
