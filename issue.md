---
title: "Close core-module-types: top-50 coverage, precision, lib baseline, exits, m1 gates, RFC re-marked"
state: done
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
- 01a1464e-dba9-740b-b8b6-8da484195ce3
- 01a1464e-dbd3-7b16-801c-b9a5893d687b
- 01a1464e-dc07-754e-9ef4-637df15d51a9
- 01a1464e-dc48-720b-9ab9-fcc7fe47f483
- 01a1464e-dca3-7284-85c3-a06107c16646
- 01a1464e-dd27-7b82-ae71-3e6173375fa7
- 01a1464e-dd6c-7b5f-b95b-76286ec110a1
- 01a1464e-ddb0-720a-93f6-450c767e6e2e
- 01a1464e-dde3-73fe-ab1f-ced7f18c4f9d
- 01a14657-682e-7432-8ea9-828f56cd2f69
- 01a14683-8375-7810-8a75-b772680603c6
- 01a14683-8405-743e-9c90-7bb5ff9cec88
- 01a14683-84ab-7efb-a0c5-b2db8deb8e28
created: 2026-10-16T20:02:16.243834962Z
updated: 2026-10-20T17:30:23.422155411Z
sessions:
- start_sha: 6c1aa61173f56c9b15f32bff3235aae12b2f18f0
  end_sha: cf08c0acdb986b7a59c463c2ed994e0721307294
  commits: 2
  started_at: 2026-10-20T16:43:54.855598868Z
  ended_at: 2026-10-20T17:30:23.422155411Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-20T16:43:54.855598868Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-20T17:30:23.422155411Z
observed_paths:
- docs/contributing/core-module-declarations.md
- docs/rfcs/0001-type-libraries.md
- internal/data/help/types.md
- internal/parse/coverage_test.go
- internal/parse/ranking_test.go
- internal/parse/testdata/core_module_not_shipped.txt
---

## Prerequisites

- [ ] Every other issue in core-module-types done, or cancelled with a recorded reason

## Context

- paths: `internal/parse/coverage_test.go` (new), `internal/parse/testdata/core_module_not_shipped.txt` (new; only `#` comments unless perigrin ruled an infrastructure module out), `docs/rfcs/0001-type-libraries.md`, `internal/data/help/types.md`
- commands: each criterion below, one at a time; `git zhi verify --timeout 900 core-module-types`
- entrypoints: the milestone's resolution command, `internal/parse/testdata/core_module_ranking.txt`

The milestone's closing gate. `TestCoreModuleDeclarationsCoverTheTop50` (the milestone's resolution test) reads the 50 modules from `core_module_ranking.txt` and requires each to have `declarations/<Module/Path>.tpm`, or to be listed in `core_module_not_shipped.txt`; it runs everywhere (no perl). That file's format is pinned: one `Module<TAB>full-uuid<TAB>reason` line per module, `#` comment lines allowed, and the reader accepts only Exporter, XSLoader and DynaLoader (ruling 4); the uuid is the cancelled issue's full id and the reason is perigrin's.

**Ruling 4 (perigrin, 2026-10-16): not shipped is allowed only for Exporter, XSLoader and DynaLoader** (the RFC's "may not ship"). Before cancelling, STOP and ask perigrin, with the measurement, so the cancel reason is his.

**Running the parse-oracle gates.** `TestRatchetCorpus` skips unless given `-parseoracle.corpus`, and `TestGoParserExactFloor` / `TestGoParserWrongIsZero` skip unless `PARSEORACLE_SHIM` names a built shim; a skip prints `--- SKIP`, so a criterion grepping for `--- PASS` fails rather than passing vacuously. The criteria below spell what `.github/workflows/parse-fidelity.yml` ("Run the fidelity ratchet") does: `PERL5_CORPUS` at the perl5 checkout pinned in `internal/parseoracle/testdata/corpus.pin` (`~/dev/perl5` here), `PARSEORACLE_TIMEOUT=8m`, `PARSEORACLE_REQUIRE_CORPUS=1`, and `PARSEORACLE_SHIM` at a fixed directory (`$HOME/.cache/pvm/oracleshim`). The sweep builds the shim there when it holds no `t/test.pl`, as CI's cache miss does, so the Go-parser criteria first run the sweep if the shim is missing. Measured 2026-10-16 at 8d508426: the sweep took 510s and passed; with the shim built, `TestGoParserExactFloor` passed (47 of 56 exact) and `TestGoParserWrongIsZero` FAILED, because the pinned WRONG `comp/proto.t` no longer scores WRONG (the test says: delete the pin). The baseline issue (01a1464e-d912) fixes that first. A shim built for another perl or corpus revision is stale: delete the directory after a re-pin.

**CI.** Open the PR for `feat/core-module-types` against `pu` and require its checks green: `gh pr checks <branch>` exits 0 when every check passed, 1 when one failed or no PR exists, and 8 while one is pending (measured 2026-10-16 on PRs 484 and 487, and `gh help` for 8). Then hold every gate the milestone touches, by name, and re-mark the RFC section "Core-module libraries" `(*Implemented*, <shas>)` ("Call sites" is re-marked by 01a0fd9b-5700). Update the "Declaration Files" section of `internal/data/help/types.md` to say which core modules ship declarations.

## Steps

- [ ] Write `TestCoreModuleDeclarationsCoverTheTop50`; run to verify pass: `go test ./internal/parse/ -run '^TestCoreModuleDeclarationsCoverTheTop50$' -count=1 -v`
- [ ] Run each criterion below; fix anything red at its root, in the issue that owns it, reopening that issue
- [ ] Re-mark "Core-module libraries" `(*Implemented*, <shas>)`; update types.md; `go test ./internal/cli/ -run '^TestShowTypesHelp$' -count=1`
- [ ] `make test`; commit
- [ ] Push `feat/core-module-types` and open the PR against `pu` (`gh pr create --head feat/core-module-types --base pu`); wait for its checks to finish and fix any red at its root

## Acceptance Criteria

### Positive Scenarios
- [ ] every top-50 module has a declaration or a recorded reason (`go test ./internal/parse/ -run '^TestCoreModuleDeclarationsCoverTheTop50$' -count=1 -v | grep -q '^--- PASS: TestCoreModuleDeclarationsCoverTheTop50'`)
- [ ] inference precision holds at least 65/76 with WRONG 0 (`const precisionFloor = 65`, raised by 01a1464e-d912) (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)
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
- [ ] m1 gate: parse oracle corpus ratchet (`PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_TIMEOUT=8m PARSEORACLE_REQUIRE_CORPUS=1 go test ./internal/parseoracle/ -run '^TestRatchetCorpus$' -parseoracle.corpus -count=1 -timeout 40m -v | grep -q '^--- PASS: TestRatchetCorpus'`)
- [ ] m1 gate: Go parser exact floor (`O=$HOME/.cache/pvm/oracleshim; test -f $O/t/test.pl || PARSEORACLE_SHIM=$O PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^TestRatchetCorpus$' -parseoracle.corpus -count=1 -timeout 40m >/dev/null; PARSEORACLE_SHIM=$O PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parseoracle/ -run '^TestGoParserExactFloor$' -count=1 -timeout 40m -v | grep -q '^--- PASS: TestGoParserExactFloor'`)
- [ ] m1 gate: the Go parser scores no file WRONG (`O=$HOME/.cache/pvm/oracleshim; test -f $O/t/test.pl || PARSEORACLE_SHIM=$O PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^TestRatchetCorpus$' -parseoracle.corpus -count=1 -timeout 40m >/dev/null; PARSEORACLE_SHIM=$O PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parseoracle/ -run '^TestGoParserWrongIsZero$' -count=1 -timeout 40m -v | grep -q '^--- PASS: TestGoParserWrongIsZero'`)
- [ ] the RFC section on call sites is implemented (`grep -q '^### Call sites (\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] the RFC section on core-module libraries is implemented (`grep -q '^### Core-module libraries (\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] the milestone's PR has every CI check green (`gh pr checks feat/core-module-types`)

### Negative Scenarios
- [ ] the coverage check fails a top-50 module with neither a `.tpm` nor a not-shipped entry (`go test ./internal/parse/ -run '^TestCoverageRejectsAnUndeclaredModule$' -count=1 -v | grep -q '^--- PASS: TestCoverageRejectsAnUndeclaredModule'`)
- [ ] the coverage check fails a not-shipped entry that is not `Module<TAB>full-uuid<TAB>reason`: a missing or empty reason, or an id that is a prefix rather than a full uuid (`go test ./internal/parse/ -run '^TestCoverageRejectsANotShippedEntryWithoutReason$' -count=1 -v | grep -q '^--- PASS: TestCoverageRejectsANotShippedEntryWithoutReason'`)
- [ ] `#` comment lines in the not-shipped file are not entries (`go test ./internal/parse/ -run '^TestCoverageNotShippedSkipsComments$' -count=1 -v | grep -q '^--- PASS: TestCoverageNotShippedSkipsComments'`)
- [ ] the coverage check fails a module that both ships a `.tpm` and is listed not shipped (`go test ./internal/parse/ -run '^TestCoverageRejectsShippedAndNotShipped$' -count=1 -v | grep -q '^--- PASS: TestCoverageRejectsShippedAndNotShipped'`)
- [ ] the coverage check fails a not-shipped entry for any module but Exporter, XSLoader and DynaLoader, ranked (POSIX) or not (ruling 4) (`go test ./internal/parse/ -run '^TestCoverageNotShippedOnlyInfrastructure$' -count=1 -v | grep -q '^--- PASS: TestCoverageNotShippedOnlyInfrastructure'`)
- [ ] module names map to paths with forward slashes and keep underscores: `File::Spec::Functions` to `File/Spec/Functions.tpm`, `SDBM_File` to `SDBM_File.tpm` (`go test ./internal/parse/ -run '^TestCoverageMapsModuleNamesToPaths$' -count=1 -v | grep -q '^--- PASS: TestCoverageMapsModuleNamesToPaths'`)
