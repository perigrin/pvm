---
title: "Declare the OO-only modules with empty export lists: File::Spec, IO::Select, HTTP::Tiny and five more"
state: pending
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-d912-7419-8b20-a8e24029b3ec
- 01a1464e-d980-7382-970e-067a5ea3db84
- 01a1464e-db1c-73eb-a3a9-e897d7a56b92
blocks:
- 01a1464e-de34-7403-95d2-25f4855404b2
created: 2026-10-16T20:02:15.975216591Z
updated: 2026-10-16T20:04:28.42213386Z
---

## Prerequisites

- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one
- [ ] Pin the core-module ranking in a test and record the rule in the RFC
- [ ] Hold Carp.tpm to perl: prototypes, export lists and a return-type oracle

## Context

- paths: `internal/parse/declarations/` (one `.tpm` per module below), `internal/parse/heldtoperl_test.go`, `internal/infer/returntypes_test.go`, `internal/infer/testdata/oracle/`
- commands: `go test ./internal/parse/ -run '^TestOOModulesPrototypes$|^TestOOModulesExports$' -count=1 -v`, `go test ./internal/infer/ -run '^TestOOModulesReturnTypes$' -count=1 -v`
- entrypoints: `internal/parse/declarations/Carp.tpm` (the model: measured facts in comments), the held-to-perl helpers from the Carp issue

Modules (from the ranking): File::Spec, File::Spec::Unix, IO::Socket::IP, IO::Select, Term::ReadLine, Tie::Hash, HTTP::Tiny, Locale::Maketext::Simple (8).

Write one `.tpm` per module declaring every name in its `@EXPORT`, `@EXPORT_OK` and `%EXPORT_TAGS` with prototype and typed signature, plus the subs and methods its callers use; a declaration is read in place of the module's source, so a name it leaves out is a name the import loses. Each file starts with two `ABOUTME:` lines naming the perl (5.42.0) and module version measured. A sub that never returns is declared `None`. Each has empty export lists (measured 2026-10-16): the `.tpm` declares methods and exports nothing. Locale::Maketext::Simple has its own `import` that installs `loc`/`loc_lang` into the caller: measure what `use Locale::Maketext::Simple` defines in the caller and declare it, or record in the issue why it cannot be declared. HTTP::Tiny and IO::Socket::IP touch the network: list them uncallable.

A `.tpm` changes what the parser reads for every file that uses the module, so run the m1 gates after this batch, not only at the end: `TestParseRatchet`, `TestPerlGitTRatchet`, `TestCanonRatchet`, `TestCanonFidelityRatchet`, `TestT2HoldsAll56`, `TestT2CoreParses`, `TestImportResolutionIsWorthT1Files` (internal/parse), `TestRatchet`, `TestRatchetCorpus`, `TestGoParserExactFloor` (internal/parseoracle), `TestPSCCheckLibBaseline` (internal/psc). A ratchet that moves is updated in the commit that moved it, with the reason.

## Steps

- [ ] Write failing tests `TestOOModulesPrototypes` and `TestOOModulesExports` in `internal/parse/heldtoperl_test.go` over the module list above
- [ ] Run to verify failure: `go test ./internal/parse/ -run '^TestOOModulesPrototypes$|^TestOOModulesExports$' -count=1 -v`
- [ ] Write the `.tpm` files; run to verify pass; commit
- [ ] Write failing test `TestOOModulesReturnTypes` in `internal/infer/returntypes_test.go`, with call cases in `internal/infer/testdata/oracle/` and every uncallable sub listed by name with its reason
- [ ] Run to verify failure: `go test ./internal/infer/ -run '^TestOOModulesReturnTypes$' -count=1 -v`
- [ ] Fix declared return types where perl disagrees; run to verify pass; commit
- [ ] Run the m1 gates above; `make test`; commit

## Acceptance Criteria

### Positive Scenarios
- [ ] declared prototypes equal perl's prototype() (`go test ./internal/parse/ -run '^TestOOModulesPrototypes$' -count=1 -v | grep -q '^--- PASS: TestOOModulesPrototypes'`)
- [ ] declared export lists equal the modules' own (`go test ./internal/parse/ -run '^TestOOModulesExports$' -count=1 -v | grep -q '^--- PASS: TestOOModulesExports'`)
- [ ] declared return types hold for every value perl returns (`go test ./internal/infer/ -run '^TestOOModulesReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestOOModulesReturnTypes'`)
