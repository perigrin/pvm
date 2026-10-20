---
title: Declare Exporter, measure it against every module that uses it, and ship it only if resolution holds
state: in-progress
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-d912-7419-8b20-a8e24029b3ec
- 01a1464e-d980-7382-970e-067a5ea3db84
- 01a1464e-db1c-73eb-a3a9-e897d7a56b92
- 01a14683-8375-7810-8a75-b772680603c6
blocks:
- 01a1464e-de34-7403-95d2-25f4855404b2
created: 2026-10-16T20:02:16.04458884Z
updated: 2026-10-20T03:51:41.540596726Z
sessions:
- start_sha: 9e5ae578baacecaaf348477c8982d9720ed83cf0
  end_sha: ""
  commits: 0
  started_at: 2026-10-20T03:51:41.540596726Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-20T03:51:41.540596726Z
---

## Prerequisites

- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one
- [ ] Convert every .tpm to :export and :export_ok, then refuse export lists in a .tpm (01a14683-8375-7810-8a75-b772680603c6)
- [ ] Pin the core-module ranking in a test and record the rule in the RFC
- [ ] Hold Carp.tpm to perl: prototypes, export lists and a return-type oracle

## Context

- paths: `internal/parse/declarations/Exporter.tpm` (new), `internal/parse/heldtoperl_test.go`, `internal/parse/infrastructure_test.go` (new), `internal/infer/returntypes_test.go`
- commands: `go test ./internal/parse/ -run '^TestExporterPrototypes$|^TestExporterExports$|^TestExporterKeepsResolution$' -count=1 -v`, `go test ./internal/infer/ -run '^TestExporterReturnTypes$' -count=1 -v`
- entrypoints: `declaration`/`readDeclaration` (`internal/parse/declarations.go`), the `declarations` fs.FS var a test can swap, RFC 0001 "Core-module libraries", "Infrastructure modules."

Exporter is inherited from by most exporting modules (168 imports from outside its namespace, rank 2), so a declaration that replaces its source could change how those modules' imports resolve. RFC "Infrastructure modules": it is written, then measured; it must keep every module that uses it resolving as before and hold the m1 gates, or it may not ship. `TestExporterKeepsResolution` parses every privlib module that uses Exporter twice, with and without the `.tpm` (swap the `declarations` fs.FS), and requires identical resolved imports, prototypes and signatures; it skips where perl 5.42.0's privlib is absent.

**If the measurement shows a change** that cannot be fixed in the declaration itself: STOP and ask perigrin, giving him the measurement (which modules, which names, before and after). Do not cancel this issue yourself. Only once perigrin rules that Exporter does not ship: cancel this issue with `git zhi issue edit 01a1464e-dd6c-7b5f-b95b-76286ec110a1 --state cancel`, his ruling as the reason, and add the line `Exporter<TAB>01a1464e-dd6c-7b5f-b95b-76286ec110a1<TAB><his reason>` to `internal/parse/testdata/core_module_not_shipped.txt`, which the final gate's `TestCoreModuleDeclarationsCoverTheTop50` reads.

**Ruling 4 (perigrin, 2026-10-16): not shipped is allowed only for Exporter, XSLoader and DynaLoader** (the RFC's "may not ship"). Before cancelling, STOP and ask perigrin, with the measurement, so the cancel reason is his.

m1 gates to run with the `.tpm` in place: `TestParseRatchet`, `TestPerlGitTRatchet`, `TestCanonRatchet`, `TestCanonFidelityRatchet`, `TestT2HoldsAll56`, `TestT2CoreParses`, `TestImportResolutionIsWorthT1Files` (internal/parse), `TestRatchet`, `TestRatchetCorpus`, `TestGoParserExactFloor`, `TestGoParserWrongIsZero` (internal/parseoracle), `TestPSCCheckLibBaseline` (internal/psc). Run each with the command the final-gate issue (01a1464e-de34) gives it: `TestRatchetCorpus` skips without `-parseoracle.corpus`, and the two Go-parser gates skip unless `PARSEORACLE_SHIM` names a built shim, and a skip proves nothing. A ratchet may only be raised. A drop is fixed in the .tpm, or the batch stops and asks perigrin. A drop the `.tpm` cannot fix is the measurement above: stop and ask.

**Ruling 1 (perigrin, 2026-10-16): no export lists in a `.tpm`.** "We don't need the @EXPORT in the .tpm since we have :export and :export_ok." On names a module's own `import` installs: "we should figure out as best we can what is exported (loc, loc_lang, etc.) and mark them as :export or :export_ok accordingly."

## Steps

- [ ] Write failing tests `TestExporterPrototypes` and `TestExporterExports`; run to verify failure
- [ ] Write `declarations/Exporter.tpm` (two `ABOUTME:` lines, measured facts in comments); run to verify pass
- [ ] Write `TestExporterKeepsResolution`; run it: `go test ./internal/parse/ -run '^TestExporterKeepsResolution$' -count=1 -v`
- [ ] Write `TestExporterReturnTypes` (uncallable subs listed by name with their reason); run to verify pass
- [ ] Run the m1 gates above; if anything drops because of the `.tpm` and the `.tpm` cannot fix it, STOP and ask perigrin (above); otherwise `make test`; commit

## Acceptance Criteria

### Positive Scenarios
- [ ] declared prototypes equal perl's prototype() (`go test ./internal/parse/ -run '^TestExporterPrototypes$' -count=1 -v | grep -q '^--- PASS: TestExporterPrototypes'`)
- [ ] declared `:export`/`:export_ok` marks equal what perl's `use Exporter` installs in a caller, bare and by name (`go test ./internal/parse/ -run '^TestExporterExports$' -count=1 -v | grep -q '^--- PASS: TestExporterExports'`)
- [ ] every privlib module using Exporter resolves as it did without the declaration (`go test ./internal/parse/ -run '^TestExporterKeepsResolution$' -count=1 -v | grep -q '^--- PASS: TestExporterKeepsResolution'`)
- [ ] declared return types hold for every value perl returns (`go test ./internal/infer/ -run '^TestExporterReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestExporterReturnTypes'`)

### Negative Scenarios
- [ ] a bare `use Exporter;` does not install `import` in the caller; `use Exporter 'import';` does (measured) (`go test ./internal/parse/ -run '^TestExporterBareUseDoesNotImportImport$' -count=1 -v | grep -q '^--- PASS: TestExporterBareUseDoesNotImportImport'`)
- [ ] the resolution comparison is not vacuous: it compares modules reaching Exporter by `use base`, `use parent`, `@ISA` and `require` alike, and fails if it compares none where privlib is present (`go test ./internal/parse/ -run '^TestExporterKeepsResolutionIsNotVacuous$' -count=1 -v | grep -q '^--- PASS: TestExporterKeepsResolutionIsNotVacuous'`)
