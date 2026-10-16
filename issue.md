---
title: Declare XSLoader, measure it against every module that uses it, and ship it only if resolution holds
state: pending
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-d912-7419-8b20-a8e24029b3ec
- 01a1464e-d980-7382-970e-067a5ea3db84
- 01a1464e-db1c-73eb-a3a9-e897d7a56b92
- 01a14683-8375-7810-8a75-b772680603c6
blocks:
- 01a1464e-de34-7403-95d2-25f4855404b2
created: 2026-10-16T20:02:16.112056153Z
updated: 2026-10-16T21:02:02.26840431Z
---

## Prerequisites

- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one
- [ ] Convert every .tpm to :export and :export_ok, then refuse export lists in a .tpm (01a14683-8375-7810-8a75-b772680603c6)
- [ ] Pin the core-module ranking in a test and record the rule in the RFC
- [ ] Hold Carp.tpm to perl: prototypes, export lists and a return-type oracle

## Context

- paths: `internal/parse/declarations/XSLoader.tpm` (new), `internal/parse/heldtoperl_test.go`, `internal/parse/infrastructure_test.go` (new), `internal/infer/returntypes_test.go`
- commands: `go test ./internal/parse/ -run '^TestXSLoaderPrototypes$|^TestXSLoaderExports$|^TestXSLoaderKeepsResolution$' -count=1 -v`, `go test ./internal/infer/ -run '^TestXSLoaderReturnTypes$' -count=1 -v`
- entrypoints: `declaration`/`readDeclaration` (`internal/parse/declarations.go`), the `declarations` fs.FS var a test can swap, RFC 0001 "Core-module libraries", "Infrastructure modules."

XSLoader is called by every XS module at load (48 imports, rank 6), so a declaration that replaces its source could change how those modules' imports resolve. RFC "Infrastructure modules": it is written, then measured; it must keep every module that uses it resolving as before and hold the m1 gates, or it may not ship. `TestXSLoaderKeepsResolution` parses every privlib module that uses XSLoader twice, with and without the `.tpm` (swap the `declarations` fs.FS), and requires identical resolved imports, prototypes and signatures; it skips where perl 5.42.0's privlib is absent.

**If the measurement shows a change** that cannot be fixed in the declaration itself: STOP and ask perigrin, giving him the measurement (which modules, which names, before and after). Do not cancel this issue yourself. Only once perigrin rules that XSLoader does not ship: cancel this issue with `git zhi issue edit 01a1464e-ddb0-720a-93f6-450c767e6e2e --state cancel`, his ruling as the reason, and add the line `XSLoader<TAB>01a1464e-ddb0-720a-93f6-450c767e6e2e<TAB><his reason>` to `internal/parse/testdata/core_module_not_shipped.txt`, which the final gate's `TestCoreModuleDeclarationsCoverTheTop50` reads.

**Ruling 4 (perigrin, 2026-10-16): not shipped is allowed only for Exporter, XSLoader and DynaLoader** (the RFC's "may not ship"). Before cancelling, STOP and ask perigrin, with the measurement, so the cancel reason is his.

m1 gates to run with the `.tpm` in place: `TestParseRatchet`, `TestPerlGitTRatchet`, `TestCanonRatchet`, `TestCanonFidelityRatchet`, `TestT2HoldsAll56`, `TestT2CoreParses`, `TestImportResolutionIsWorthT1Files` (internal/parse), `TestRatchet`, `TestRatchetCorpus`, `TestGoParserExactFloor`, `TestGoParserWrongIsZero` (internal/parseoracle), `TestPSCCheckLibBaseline` (internal/psc). Run each with the command the final-gate issue (01a1464e-de34) gives it: `TestRatchetCorpus` skips without `-parseoracle.corpus`, and the two Go-parser gates skip unless `PARSEORACLE_SHIM` names a built shim, and a skip proves nothing. A ratchet may only be raised. A drop is fixed in the .tpm, or the batch stops and asks perigrin. A drop the `.tpm` cannot fix is the measurement above: stop and ask.

**Ruling 1 (perigrin, 2026-10-16): no export lists in a `.tpm`.** "We don't need the @EXPORT in the .tpm since we have :export and :export_ok." On names a module's own `import` installs: "we should figure out as best we can what is exported (loc, loc_lang, etc.) and mark them as :export or :export_ok accordingly."

## Steps

- [ ] Write failing tests `TestXSLoaderPrototypes` and `TestXSLoaderExports`; run to verify failure
- [ ] Write `declarations/XSLoader.tpm` (two `ABOUTME:` lines, measured facts in comments); run to verify pass
- [ ] Write `TestXSLoaderKeepsResolution`; run it: `go test ./internal/parse/ -run '^TestXSLoaderKeepsResolution$' -count=1 -v`
- [ ] Write `TestXSLoaderReturnTypes` (uncallable subs listed by name with their reason); run to verify pass
- [ ] Run the m1 gates above; if anything drops because of the `.tpm` and the `.tpm` cannot fix it, STOP and ask perigrin (above); otherwise `make test`; commit

## Acceptance Criteria

### Positive Scenarios
- [ ] declared prototypes equal perl's prototype() (`go test ./internal/parse/ -run '^TestXSLoaderPrototypes$' -count=1 -v | grep -q '^--- PASS: TestXSLoaderPrototypes'`)
- [ ] declared `:export`/`:export_ok` marks equal what perl's `use XSLoader` installs in a caller, bare and by name (`go test ./internal/parse/ -run '^TestXSLoaderExports$' -count=1 -v | grep -q '^--- PASS: TestXSLoaderExports'`)
- [ ] every privlib module using XSLoader resolves as it did without the declaration (`go test ./internal/parse/ -run '^TestXSLoaderKeepsResolution$' -count=1 -v | grep -q '^--- PASS: TestXSLoaderKeepsResolution'`)
- [ ] declared return types hold for every value perl returns (`go test ./internal/infer/ -run '^TestXSLoaderReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestXSLoaderReturnTypes'`)

### Negative Scenarios
- [ ] a bare `use XSLoader;` imports nothing (empty export lists, measured), so a bare `load()` does not resolve to `XSLoader::load` (`go test ./internal/parse/ -run '^TestXSLoaderBareUseImportsNothing$' -count=1 -v | grep -q '^--- PASS: TestXSLoaderBareUseImportsNothing'`)
- [ ] the resolution comparison is not vacuous: it fails if it compares no module where privlib is present (`go test ./internal/parse/ -run '^TestXSLoaderKeepsResolutionIsNotVacuous$' -count=1 -v | grep -q '^--- PASS: TestXSLoaderKeepsResolutionIsNotVacuous'`)
