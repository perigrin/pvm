---
title: Declare Exporter, measure it against every module that uses it, and ship it only if resolution holds
state: pending
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-d912-7419-8b20-a8e24029b3ec
- 01a1464e-d980-7382-970e-067a5ea3db84
- 01a1464e-db1c-73eb-a3a9-e897d7a56b92
blocks:
- 01a1464e-de34-7403-95d2-25f4855404b2
created: 2026-10-16T20:02:16.04458884Z
updated: 2026-10-16T20:04:31.704483093Z
---

## Prerequisites

- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one
- [ ] Pin the core-module ranking in a test and record the rule in the RFC
- [ ] Hold Carp.tpm to perl: prototypes, export lists and a return-type oracle

## Context

- paths: `internal/parse/declarations/Exporter.tpm` (new), `internal/parse/heldtoperl_test.go`, `internal/parse/infrastructure_test.go` (new), `internal/infer/returntypes_test.go`
- commands: `go test ./internal/parse/ -run '^TestExporterPrototypes$|^TestExporterExports$|^TestExporterKeepsResolution$' -count=1 -v`, `go test ./internal/infer/ -run '^TestExporterReturnTypes$' -count=1 -v`
- entrypoints: `declaration`/`readDeclaration` (`internal/parse/declarations.go`), the `declarations` fs.FS var a test can swap, RFC 0001 "Core-module libraries", "Infrastructure modules."

Exporter is inherited from by most exporting modules (168 imports from outside its namespace, rank 2), so a declaration that replaces its source could change how those modules' imports resolve. RFC "Infrastructure modules": it is written, then measured; it must keep every module that uses it resolving as before and hold the m1 gates, or it is not shipped. `TestExporterKeepsResolution` parses every privlib module that uses Exporter twice, with and without the `.tpm` (swap the `declarations` fs.FS), and requires identical resolved imports, prototypes and signatures; it skips where perl 5.42.0's privlib is absent.

**If the measurement shows a change** that cannot be fixed in the declaration itself: do not ship the `.tpm`. Cancel this issue with `git zhi issue edit <id> --state cancel`, recording the measurement (which modules, which names, before and after) as the reason, and record Exporter as not shipped, with that reason, in the list the final gate's `TestCoreModuleDeclarationsCoverTheTop50` reads.

m1 gates to run with the `.tpm` in place: `TestParseRatchet`, `TestPerlGitTRatchet`, `TestCanonRatchet`, `TestCanonFidelityRatchet`, `TestT2HoldsAll56`, `TestT2CoreParses`, `TestImportResolutionIsWorthT1Files` (internal/parse), `TestRatchet`, `TestRatchetCorpus`, `TestGoParserExactFloor` (internal/parseoracle), `TestPSCCheckLibBaseline` (internal/psc).

## Steps

- [ ] Write failing tests `TestExporterPrototypes` and `TestExporterExports`; run to verify failure
- [ ] Write `declarations/Exporter.tpm` (two `ABOUTME:` lines, measured facts in comments); run to verify pass
- [ ] Write `TestExporterKeepsResolution`; run it: `go test ./internal/parse/ -run '^TestExporterKeepsResolution$' -count=1 -v`
- [ ] Write `TestExporterReturnTypes` (uncallable subs listed by name with their reason); run to verify pass
- [ ] Run the m1 gates above; if anything moves because of the `.tpm`, follow the cancel rule; otherwise `make test`; commit

## Acceptance Criteria

### Positive Scenarios
- [ ] declared prototypes equal perl's prototype() (`go test ./internal/parse/ -run '^TestExporterPrototypes$' -count=1 -v | grep -q '^--- PASS: TestExporterPrototypes'`)
- [ ] declared export lists equal Exporter's own (`go test ./internal/parse/ -run '^TestExporterExports$' -count=1 -v | grep -q '^--- PASS: TestExporterExports'`)
- [ ] every privlib module using Exporter resolves as it did without the declaration (`go test ./internal/parse/ -run '^TestExporterKeepsResolution$' -count=1 -v | grep -q '^--- PASS: TestExporterKeepsResolution'`)
- [ ] declared return types hold for every value perl returns (`go test ./internal/infer/ -run '^TestExporterReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestExporterReturnTypes'`)
