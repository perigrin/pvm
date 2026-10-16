---
title: "Close core-module-types: top-50 coverage, precision, lib baseline, exits, m1 gates, RFC re-marked"
state: pending
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-d912-7419-8b20-a8e24029b3ec
- 01a1464e-d980-7382-970e-067a5ea3db84
- 01a1464e-da0a-7ea5-9232-1cd68885d4e0
- 01a0fd9b-5700-7514-a143-2caed2d06d8b
- 01a1464e-da9d-7f9e-b2cf-b3d8607907a2
- 01a1464e-dadb-7796-a704-4007150a1bf7
- 01a1464e-db1c-73eb-a3a9-e897d7a56b92
- 01a1464e-db4c-78ae-9bd5-84d71ccf5432
- 01a1464e-db76-704f-bd85-023a54d2cf6e
created: 2026-10-16T20:02:16.243834962Z
updated: 2026-10-16T20:03:50.951528888Z
---

## Prerequisites

- [ ] Every other issue in core-module-types done, or cancelled with a recorded reason

## Context

- paths: `internal/parse/coverage_test.go` (new), `internal/parse/testdata/core_module_not_shipped.txt` (new; empty unless an infrastructure issue was cancelled), `docs/rfcs/0001-type-libraries.md`, `internal/data/help/types.md`
- commands: each criterion below, one at a time; `git zhi verify --timeout 900 core-module-types`
- entrypoints: the milestone's resolution command, `internal/parse/testdata/core_module_ranking.txt`

The milestone's closing gate. `TestCoreModuleDeclarationsCoverTheTop50` (the milestone's resolution test) reads the 50 modules from `core_module_ranking.txt` and requires each to have `declarations/<Module/Path>.tpm`, or to be listed in `core_module_not_shipped.txt` with the cancelled issue's id and measured reason; it runs everywhere (no perl). Then hold every gate the milestone touches, by name, and re-mark the RFC section "Core-module libraries" `(*Implemented*, <shas>)` ("Call sites" is re-marked by 01a0fd9b-5700). Update the "Declaration Files" section of `internal/data/help/types.md` to say which core modules ship declarations.

## Steps

- [ ] Write `TestCoreModuleDeclarationsCoverTheTop50`; run to verify pass: `go test ./internal/parse/ -run '^TestCoreModuleDeclarationsCoverTheTop50$' -count=1 -v`
- [ ] Run each criterion below; fix anything red at its root, in the issue that owns it, reopening that issue
- [ ] Re-mark "Core-module libraries" `(*Implemented*, <shas>)`; update types.md; `go test ./internal/cli/ -run '^TestShowTypesHelp$' -count=1`
- [ ] `make test`; commit

## Acceptance Criteria

### Positive Scenarios
- [ ] every top-50 module has a declaration or a recorded reason (`go test ./internal/parse/ -run '^TestCoreModuleDeclarationsCoverTheTop50$' -count=1 -v | grep -q '^--- PASS: TestCoreModuleDeclarationsCoverTheTop50'`)
- [ ] inference precision holds at least 65/76 with WRONG 0 (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)
- [ ] psc check over perl 5.42's lib adds nothing to the baseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] bare croak and confess are exits (`go test ./internal/infer/ -run '^TestExitsAndGuardsByCoreAndCarp$' -count=1 -v | grep -q '^--- PASS: TestExitsAndGuardsByCoreAndCarp'`)
- [ ] qualified croak and confess are exits (`go test ./internal/infer/ -run '^TestQualifiedCroakConfessAreExits$' -count=1 -v | grep -q '^--- PASS: TestQualifiedCroakConfessAreExits'`)
- [ ] m1 gate: T1 parse ratchet (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet'`)
- [ ] m1 gate: perl.git t/ ratchet (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] m1 gate: canon ratchet (`go test ./internal/parse/ -run '^TestCanonRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonRatchet'`)
- [ ] m1 gate: canon fidelity ratchet (`go test ./internal/parse/ -run '^TestCanonFidelityRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonFidelityRatchet'`)
- [ ] m1 gate: T2 holds all 56 (`go test ./internal/parse/ -run '^TestT2HoldsAll56$' -count=1 -v | grep -q '^--- PASS: TestT2HoldsAll56'`)
- [ ] m1 gate: T2 core parses (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)
- [ ] m1 gate: import resolution is worth T1 files (`go test ./internal/parse/ -run '^TestImportResolutionIsWorthT1Files$' -count=1 -v | grep -q '^--- PASS: TestImportResolutionIsWorthT1Files'`)
- [ ] m1 gate: parse oracle ratchet (`go test ./internal/parseoracle/ -run '^TestRatchet$' -count=1 -v | grep -q '^--- PASS: TestRatchet'`)
- [ ] m1 gate: parse oracle corpus ratchet (`go test ./internal/parseoracle/ -run '^TestRatchetCorpus$' -count=1 -v | grep -q '^--- PASS: TestRatchetCorpus'`)
- [ ] m1 gate: Go parser exact floor (`go test ./internal/parseoracle/ -run '^TestGoParserExactFloor$' -count=1 -v | grep -q '^--- PASS: TestGoParserExactFloor'`)
- [ ] the RFC section on call sites is implemented (`grep -q '^### Call sites (\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] the RFC section on core-module libraries is implemented (`grep -q '^### Core-module libraries (\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
