---
title: Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one
state: done
urgency: normal
milestone: core-module-types
blocks:
- 01a1464e-da0a-7ea5-9232-1cd68885d4e0
- 01a0fd9b-5700-7514-a143-2caed2d06d8b
- 01a1464e-da9d-7f9e-b2cf-b3d8607907a2
- 01a1464e-db1c-73eb-a3a9-e897d7a56b92
- 01a1464e-de34-7403-95d2-25f4855404b2
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
created: 2026-10-16T20:02:14.930089601Z
updated: 2026-10-16T22:23:58.098893991Z
sessions:
- start_sha: 2fb0e62ffdc068b717c1b1a211c77e72d09ba51a
  end_sha: ec99871d43d51af75b3cdcb95b6ddca2a6802abd
  commits: 1
  started_at: 2026-10-16T21:09:47.558182701Z
  ended_at: 2026-10-16T21:44:13.200226818Z
- start_sha: a9ce48f0047db49c938b19bdad67387cc27a0bd5
  end_sha: 921f002adc2aa3022e33d7016ef26672b6310406
  commits: 2
  started_at: 2026-10-16T21:55:32.561438668Z
  ended_at: 2026-10-16T22:23:58.098893991Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T21:09:47.558182701Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-16T21:44:13.200226818Z
- state: reopened
  actor: human:git-zhi
  timestamp: 2026-10-16T21:55:32.455543002Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T21:55:32.561438668Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-16T22:23:58.098893991Z
observed_paths:
- docs/rfcs/0001-type-libraries.md
- internal/infer/precision_test.go
- internal/lexer/proto.go
- internal/parse/decl.go
- internal/parse/declarations.go
- internal/parse/exportattr_test.go
- internal/parse/exports.go
- internal/parse/parse.go
- internal/parse/ranking_test.go
- internal/parse/testdata/core_module_ranking.txt
- internal/parse/testdata/declarations/My/Exports.tpm
- internal/parse/testdata/declarations/My/Vars.tpm
- internal/parse/typedour_test.go
- internal/parseoracle/goparser_test.go
- internal/perltest/perl542_test.go
- internal/perltest/perltest.go
- internal/psc/inc_roots_test.go
- internal/psc/lib_baseline_test.go
- internal/psc/testdata/lib_baseline.txt
- internal/psc/testdata/lib_true_positives.txt
---

## Prerequisites

- none: this issue is first. The declaration batches, the Carp helpers (01a1464e-db1c), the select rulings (01a1464e-da0a) and the call-site issue are blocked by it; the ranking (01a1464e-d980) and the typed-variable reader (01a1464e-dadb) are not, since neither changes what psc check reports over lib.

## Context

- paths: `internal/perltest/perltest.go` (new), `internal/psc/lib_baseline_test.go` (new), `internal/psc/testdata/lib_baseline.txt` (new), `internal/psc/testdata/lib_true_positives.txt` (new)
- commands: `go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v`
- entrypoints: `internal/psc/check_command.go`, `internal/psc/review_regressions_test.go` (`diagLine`, `isolatePerl`), `internal/psc/inc_roots_test.go` (`runCheck`)

Goal 5 needs psc check's output over perl 5.42's lib measured BEFORE any declaration or call-site change, because once one lands "pre-milestone" can no longer be measured. Capture `PVM_PERL_VERSION=5.42.0 psc check /home/perigrin/.local/share/pvm/versions/5.42.0/lib/perl5/5.42.0` at 8d508426 (the milestone base; its only change from pu is the RFC text), commit it, and add a test that fails on any diagnostic over lib that is not in the baseline, unless the diagnostic is listed in `lib_true_positives.txt` with perl evidence: the line `perl -w` prints for that file (or for a minimal file reproducing the diagnostic's construct), copied verbatim, showing perl itself warns or dies there. Nothing else counts as evidence: not a reading of the source, not a perldoc quote. A diagnostic disappearing is not a failure.

Locating perl: add `internal/perltest`, a small test-helper package with one function that returns perl 5.42.0's path, privlib and archlib (asked of that perl via `$Config{privlibexp}`/`$Config{archlibexp}`), and calls `t.Skip` where perl 5.42.0 or privlib is absent. The ranking, held-to-perl and oracle tests reuse it. The ranking issue (01a1464e-d980) no longer waits on this one; whichever of the two lands first writes `internal/perltest`, and the other reuses it. No POSIX path assumptions: use `filepath`, and store baseline paths relative to privlib with forward slashes (`filepath.ToSlash`). CI runs Windows and macOS, where this test skips.

**Two gates are red or loose before any change, and are fixed here, first.** (1) `TestGoParserWrongIsZero` (internal/parseoracle) fails at 8d508426 with the shim built (measured 2026-10-16): `comp/proto.t`, pinned there as the one allowed WRONG, no longer scores WRONG, and the test says to delete the pin and restore the bare zero check in the commit that earned it. Every batch now runs this gate, so it must be green before the first `.tpm` lands. (2) `const precisionFloor` in `internal/infer/precision_test.go` is 62, but `TestInferencePrecision` measures 65 exact of 76 at 8d508426 (measured 2026-10-16); raise it to 65 so the batches cannot lose what is held today, and correct the comment's "62 of the 76".

Measure the test's runtime and record it in the test's comment; if it exceeds go test's default 10m, say so in the issue before committing rather than raising the timeout silently.

## Steps

- [ ] Raise `const precisionFloor` in `internal/infer/precision_test.go` from 62 to 65 and update its comment; `go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v`; commit
- [ ] Delete `TestGoParserWrongIsZero`'s `comp/proto.t` pin and restore the bare zero check, as its failure message says; run it with the final-gate issue's command for it; commit
- [ ] Write `internal/perltest` with the skip-when-absent helper (or reuse it, if the ranking issue landed first)
- [ ] Write failing unit test `TestLibBaselineRejectsANewDiagnostic`: the comparison reports a diagnostic absent from a two-line baseline
- [ ] Write failing unit test `TestLibBaselineAdmitsAListedTruePositive`: the same diagnostic passes when listed with perl evidence, and fails when listed with an empty evidence field
- [ ] Run to verify failure: `go test ./internal/psc/ -run '^TestLibBaseline' -count=1 -v`
- [ ] Implement the comparison; run to verify pass; commit
- [ ] Capture the baseline at 8d508426 into `internal/psc/testdata/lib_baseline.txt`, first line `# captured at 8d508426 with perl 5.42.0`; commit
- [ ] Write `TestPSCCheckLibBaseline`: runs psc check over privlib with `PVM_PERL_VERSION=5.42.0` and compares; run to verify pass: `go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v`
- [ ] `make test`; commit

### Review Findings

Tier-2 gate, 2026-10-16, on ec99871d (fresh reviewer subagent; lenses applied inline). VERDICT: REOPEN.

1. `internal/psc/lib_baseline_test.go:23` libDiagShape drops refusal diagnostics (underscore codes such as `[call_arity]`, `[unclosed_opener]`) and `info` severity; newLibDiagnostics skips non-matching lines, so a parser regression that makes psc refuse a lib file never fails the gate. Fix: accept `[a-z_-]+` and `info`, and fail on any unindented privlib-prefixed line that does not parse; add a unit test with a `[call_arity]` line. (confidence 95)
2. `TestPSCCheckLibBaseline` passes when psc check did not finish: runCheck discards the error from Execute, and a WalkDir abort or stat failure yields few or no lines. Fix: require the error to be nil or exactly "diagnostics found", or require at least one baseline entry still reported. (confidence 70)
3. The not-installed-in-pvm branch of Perl542 (GetVersionInfo error, then skip) has no direct test. Fix: stub perl.GetVersionInfo to return an error and assert Perl542 skips. (confidence 65)

## Acceptance Criteria

### Positive Scenarios
- [ ] a diagnostic over lib not in the baseline fails the comparison (`go test ./internal/psc/ -run '^TestLibBaselineRejectsANewDiagnostic$' -count=1 -v | grep -q '^--- PASS: TestLibBaselineRejectsANewDiagnostic'`)
- [ ] a listed true positive with perl evidence is admitted (`go test ./internal/psc/ -run '^TestLibBaselineAdmitsAListedTruePositive$' -count=1 -v | grep -q '^--- PASS: TestLibBaselineAdmitsAListedTruePositive'`)
- [ ] psc check over perl 5.42's lib adds nothing to the baseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] the baseline records the commit it was captured at (`grep -q '^# captured at 8d508426' internal/psc/testdata/lib_baseline.txt`)
- [ ] the inference precision floor is raised to what is held today (`grep -q 'const precisionFloor = 65' internal/infer/precision_test.go`)
- [ ] the Go parser scores no file WRONG, with the stale `comp/proto.t` pin gone (`O=$HOME/.cache/pvm/oracleshim; test -f $O/t/test.pl || PARSEORACLE_SHIM=$O PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^TestRatchetCorpus$' -parseoracle.corpus -count=1 -timeout 40m >/dev/null; PARSEORACLE_SHIM=$O PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parseoracle/ -run '^TestGoParserWrongIsZero$' -count=1 -timeout 40m -v | grep -q '^--- PASS: TestGoParserWrongIsZero'`)

### Negative Scenarios
- [ ] the perl locator skips, not fails, where perl 5.42.0 is not installed (`go test ./internal/perltest/ -run '^TestPerl542SkipsWhenAbsent$' -count=1 -v | grep -q '^--- PASS: TestPerl542SkipsWhenAbsent'`)
- [ ] the perl locator skips, not fails, where perl 5.42.0 is present but its privlib directory is missing (`go test ./internal/perltest/ -run '^TestPerl542SkipsWhenPrivlibMissing$' -count=1 -v | grep -q '^--- PASS: TestPerl542SkipsWhenPrivlibMissing'`)
- [ ] a diagnostic whose path uses Windows backslash separators matches its forward-slash, privlib-relative baseline entry (`go test ./internal/psc/ -run '^TestLibBaselineNormalizesWindowsSeparators$' -count=1 -v | grep -q '^--- PASS: TestLibBaselineNormalizesWindowsSeparators'`)
- [ ] a baseline diagnostic that no longer appears is not a failure (`go test ./internal/psc/ -run '^TestLibBaselineAllowsADisappearedDiagnostic$' -count=1 -v | grep -q '^--- PASS: TestLibBaselineAllowsADisappearedDiagnostic'`)
- [ ] a malformed baseline line is an error naming the line, not silently skipped or matched; `#` comment lines are not diagnostics (`go test ./internal/psc/ -run '^TestLibBaselineRejectsAMalformedLine$' -count=1 -v | grep -q '^--- PASS: TestLibBaselineRejectsAMalformedLine'`)
- [ ] a true-positive entry whose evidence is not `perl -w` output -- no line of perl's `... at <file> line <n>.` shape, e.g. prose -- is rejected (`go test ./internal/psc/ -run '^TestLibTruePositiveRequiresPerlWOutput$' -count=1 -v | grep -q '^--- PASS: TestLibTruePositiveRequiresPerlWOutput'`)
- [ ] a true-positive entry admits only the exact diagnostic it lists, not the same message at another file or line (`go test ./internal/psc/ -run '^TestLibTruePositiveMatchesExactDiagnostic$' -count=1 -v | grep -q '^--- PASS: TestLibTruePositiveMatchesExactDiagnostic'`)
