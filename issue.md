---
title: "Declare the OO-only modules with empty export lists: File::Spec, IO::Select, HTTP::Tiny and five more"
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
created: 2026-10-16T20:02:15.975216591Z
updated: 2026-10-16T21:01:59.912540584Z
---

## Prerequisites

- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one
- [ ] Convert every .tpm to :export and :export_ok, then refuse export lists in a .tpm (01a14683-8375-7810-8a75-b772680603c6)
- [ ] Pin the core-module ranking in a test and record the rule in the RFC
- [ ] Hold Carp.tpm to perl: prototypes, export lists and a return-type oracle

## Context

- paths: `internal/parse/declarations/` (one `.tpm` per module below), `internal/parse/heldtoperl_test.go`, `internal/infer/returntypes_test.go`, `internal/infer/testdata/oracle/`
- commands: `go test ./internal/parse/ -run '^TestOOModulesPrototypes$|^TestOOModulesExports$' -count=1 -v`, `go test ./internal/infer/ -run '^TestOOModulesReturnTypes$' -count=1 -v`
- entrypoints: `internal/parse/declarations/Carp.tpm` (the model: measured facts in comments), the held-to-perl helpers from the Carp issue (01a1464e-db1c)

Modules (from the ranking): File::Spec, File::Spec::Unix, IO::Socket::IP, IO::Select, Term::ReadLine, Tie::Hash, HTTP::Tiny, Locale::Maketext::Simple (8).

Write one `.tpm` per module declaring every name a `use` of it installs in a caller, with prototype and typed signature, plus the subs and methods its callers use; a declaration is read in place of the module's source, so a name it leaves out is a name the import loses. A `.tpm` has no `@EXPORT`, `@EXPORT_OK` or `%EXPORT_TAGS` (the parser refuses them): each exported sub or variable carries `:export` (a bare `use` installs it) or `:export_ok` (it must be asked for), with its tags as attribute arguments, e.g. `sub LOCK_EX :export_ok(flock) () Int;`. What is marked is what perl's `use` installs in a caller, measured for a bare `use M`, for `use M qw(name)` and for `use M ':tag'`, not the module's raw lists. A module with its own `import` is declared the same way: mark what that `import` installs. Each file starts with two `ABOUTME:` lines naming the perl (5.42.0) and module version measured. A sub that never returns is declared `None`. Each has empty export lists (measured 2026-10-16), but a list is not what is declared (ruling 1). Measured 2026-10-16, a bare `use` of File::Spec, File::Spec::Unix, IO::Select, Term::ReadLine, Tie::Hash or HTTP::Tiny installs nothing in the caller: their `.tpm` declares methods and marks nothing `:export`. Locale::Maketext::Simple has its own `import`: a bare `use Locale::Maketext::Simple` installs `loc` and `loc_lang` (measured 2026-10-16), so both are declared `:export`; `use Locale::Maketext::Simple ()` installs neither. IO::Socket::IP inherits IO::Socket's `import`, which re-exports Socket: a bare `use IO::Socket::IP` installs Socket's 170 default subs (`AF_INET`, `inet_aton`, ...; measured 2026-10-16), so they are declared `:export` in `IO/Socket/IP.tpm` with Socket's prototypes. The rule for every module with its own `import`, here and in the other batches (Config, Test::More, IO::Socket and the rest): mark what that `import` installs in a caller, measured for a bare `use`, `use M qw(name)` and `use M ':tag'`, never its raw lists. HTTP::Tiny and IO::Socket::IP touch the network: list them uncallable.

A `.tpm` changes what the parser reads for every file that uses the module, so run the m1 gates after this batch, not only at the end: `TestParseRatchet`, `TestPerlGitTRatchet`, `TestCanonRatchet`, `TestCanonFidelityRatchet`, `TestT2HoldsAll56`, `TestT2CoreParses`, `TestImportResolutionIsWorthT1Files` (internal/parse), `TestRatchet`, `TestRatchetCorpus`, `TestGoParserExactFloor`, `TestGoParserWrongIsZero` (internal/parseoracle), `TestPSCCheckLibBaseline` (internal/psc). Run each with the command the final-gate issue (01a1464e-de34) gives it: `TestRatchetCorpus` skips without `-parseoracle.corpus`, and the two Go-parser gates skip unless `PARSEORACLE_SHIM` names a built shim, and a skip proves nothing. A ratchet may only be raised. A drop is fixed in the .tpm, or the batch stops and asks perigrin.

**Ruling 1 (perigrin, 2026-10-16): no export lists in a `.tpm`.** "We don't need the @EXPORT in the .tpm since we have :export and :export_ok." On names a module's own `import` installs: "we should figure out as best we can what is exported (loc, loc_lang, etc.) and mark them as :export or :export_ok accordingly."

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
- [ ] declared `:export`/`:export_ok` marks equal what perl's `use` installs in a caller, bare, by name and by tag (`go test ./internal/parse/ -run '^TestOOModulesExports$' -count=1 -v | grep -q '^--- PASS: TestOOModulesExports'`)
- [ ] declared return types hold for every value perl returns (`go test ./internal/infer/ -run '^TestOOModulesReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestOOModulesReturnTypes'`)
- [ ] a bare `use Locale::Maketext::Simple;` installs `loc` and `loc_lang`, so a bare `loc(...)` resolves to the declaration (measured) (`go test ./internal/parse/ -run '^TestLocaleMaketextSimpleBareUseInstallsLoc$' -count=1 -v | grep -q '^--- PASS: TestLocaleMaketextSimpleBareUseInstallsLoc'`)
- [ ] a bare `use IO::Socket::IP;` installs Socket's default names, so a bare `inet_aton(...)` resolves (measured) (`go test ./internal/parse/ -run '^TestIOSocketIPBareUseInstallsSocketNames$' -count=1 -v | grep -q '^--- PASS: TestIOSocketIPBareUseInstallsSocketNames'`)

### Negative Scenarios
- [ ] a bare `use File::Spec;` imports nothing, so a bare `catfile(...)` does not resolve to File::Spec's method (`go test ./internal/parse/ -run '^TestFileSpecBareUseImportsNothing$' -count=1 -v | grep -q '^--- PASS: TestFileSpecBareUseImportsNothing'`)
- [ ] HTTP::Tiny and IO::Socket::IP subs that touch the network are in the exemption map with a reason and are never called (`go test ./internal/infer/ -run '^TestOOModulesNetworkSubsExempt$' -count=1 -v | grep -q '^--- PASS: TestOOModulesNetworkSubsExempt'`)
- [ ] Term::ReadLine `readline` is exempt with a reason and never called, since it blocks on a terminal (`go test ./internal/infer/ -run '^TestTermReadLineNeverCalled$' -count=1 -v | grep -q '^--- PASS: TestTermReadLineNeverCalled'`)
- [ ] `use Locale::Maketext::Simple ();` installs neither `loc` nor `loc_lang` (`go test ./internal/parse/ -run '^TestLocaleMaketextSimpleEmptyImportInstallsNothing$' -count=1 -v | grep -q '^--- PASS: TestLocaleMaketextSimpleEmptyImportInstallsNothing'`)
