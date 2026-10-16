---
title: Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one
state: pending
urgency: normal
milestone: core-module-types
blocks:
- 01a1464e-d980-7382-970e-067a5ea3db84
- 01a1464e-da0a-7ea5-9232-1cd68885d4e0
- 01a0fd9b-5700-7514-a143-2caed2d06d8b
- 01a1464e-da9d-7f9e-b2cf-b3d8607907a2
- 01a1464e-dadb-7796-a704-4007150a1bf7
- 01a1464e-db1c-73eb-a3a9-e897d7a56b92
- 01a1464e-de34-7403-95d2-25f4855404b2
- 01a1464e-db4c-78ae-9bd5-84d71ccf5432
- 01a1464e-db76-704f-bd85-023a54d2cf6e
- 01a1464e-dba9-740b-b8b6-8da484195ce3
- 01a1464e-dbd3-7b16-801c-b9a5893d687b
created: 2026-10-16T20:02:14.930089601Z
updated: 2026-10-16T20:04:01.159964523Z
---

## Prerequisites

- none: this issue is first. Every declaration and call-site issue in core-module-types is blocked by it.

## Context

- paths: `internal/perltest/perltest.go` (new), `internal/psc/lib_baseline_test.go` (new), `internal/psc/testdata/lib_baseline.txt` (new), `internal/psc/testdata/lib_true_positives.txt` (new)
- commands: `go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v`
- entrypoints: `internal/psc/check_command.go`, `internal/psc/review_regressions_test.go` (`diagLine`, `isolatePerl`), `internal/psc/inc_roots_test.go` (`runCheck`)

Goal 5 needs psc check's output over perl 5.42's lib measured BEFORE any declaration or call-site change, because once one lands "pre-milestone" can no longer be measured. Capture `PVM_PERL_VERSION=5.42.0 psc check /home/perigrin/.local/share/pvm/versions/5.42.0/lib/perl5/5.42.0` at 8d508426 (the milestone base; its only change from pu is the RFC text), commit it, and add a test that fails on any diagnostic over lib that is not in the baseline, unless the diagnostic is listed in `lib_true_positives.txt` with perl evidence (the `perl -w` line, or a one-liner and its output, showing perl agrees). A diagnostic disappearing is not a failure.

Locating perl: add `internal/perltest`, a small test-helper package with one function that returns perl 5.42.0's path, privlib and archlib (asked of that perl via `$Config{privlibexp}`/`$Config{archlibexp}`), and calls `t.Skip` where perl 5.42.0 or privlib is absent. The ranking, held-to-perl and oracle tests reuse it. No POSIX path assumptions: use `filepath`, and store baseline paths relative to privlib with forward slashes (`filepath.ToSlash`). CI runs Windows and macOS, where this test skips.

Measure the test's runtime and record it in the test's comment; if it exceeds go test's default 10m, say so in the issue before committing rather than raising the timeout silently.

## Steps

- [ ] Write `internal/perltest` with the skip-when-absent helper
- [ ] Write failing unit test `TestLibBaselineRejectsANewDiagnostic`: the comparison reports a diagnostic absent from a two-line baseline
- [ ] Write failing unit test `TestLibBaselineAdmitsAListedTruePositive`: the same diagnostic passes when listed with perl evidence, and fails when listed with an empty evidence field
- [ ] Run to verify failure: `go test ./internal/psc/ -run '^TestLibBaseline' -count=1 -v`
- [ ] Implement the comparison; run to verify pass; commit
- [ ] Capture the baseline at 8d508426 into `internal/psc/testdata/lib_baseline.txt`, first line `# captured at 8d508426 with perl 5.42.0`; commit
- [ ] Write `TestPSCCheckLibBaseline`: runs psc check over privlib with `PVM_PERL_VERSION=5.42.0` and compares; run to verify pass: `go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v`
- [ ] `make test`; commit

## Acceptance Criteria

### Positive Scenarios
- [ ] a diagnostic over lib not in the baseline fails the comparison (`go test ./internal/psc/ -run '^TestLibBaselineRejectsANewDiagnostic$' -count=1 -v | grep -q '^--- PASS: TestLibBaselineRejectsANewDiagnostic'`)
- [ ] a listed true positive with perl evidence is admitted (`go test ./internal/psc/ -run '^TestLibBaselineAdmitsAListedTruePositive$' -count=1 -v | grep -q '^--- PASS: TestLibBaselineAdmitsAListedTruePositive'`)
- [ ] psc check over perl 5.42's lib adds nothing to the baseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] the baseline records the commit it was captured at (`grep -q '^# captured at 8d508426' internal/psc/testdata/lib_baseline.txt`)
