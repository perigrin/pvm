---
title: "Test-harness gaps from the fix round: a dying POSIX probe fails, stale exemptions are checked on the real maps, the exemption machinery is one mechanism"
state: in-progress
urgency: normal
milestone: core-module-types
created: 2026-10-21T05:19:06.111971328Z
updated: 2026-10-21T15:14:36.252473055Z
sessions:
- start_sha: 9711a47ad2aaa4e8463f48604119e9b4ed2bf964
  end_sha: ""
  commits: 0
  started_at: 2026-10-21T15:14:36.252473055Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-21T15:14:36.252473055Z
---

crochet:review round 2 of PR #488 at 2db182e6 (2026-10-21). Quality report: scratchpad/review2-core-module-types-paad.md, probes scratchpad/r2probes/, psc builds of head, cf08c0ac and pu in scratchpad/r2bin/. A finding pu gets right is a regression this milestone introduced: fix it. A gap pu shares is out of scope here and goes to the follow-up named in the issue.

Quality I1 (returntypes_test.go:1488-1515 posixScriptLines logs a non-zero exit; a script dying after k of N records passes, on Linux too), I2 (staleexempt.go:23 StaleExemptions runs only on its fixtures, never on platformExempt or heldExempt, calls each sub with no arguments, and counts a perl error as not stale -- 01a15b21's stale step unmet), S2/S11 stale comments, S19 (perltest.go:48-72 duplicates conformance/perl.go:88-132's perl locator). Ponytail round 2: FromPvm superseded by IsReferenceBuild; MissingSubs and ConstantDifferences embed the same lookup (use held_lib.pl's callable); pendingAudit and parameterExempt are two mechanisms for one job; paramTypeAt copies Signature.paramType (export ParamType); one moduleOracleRun helper for the five copies; TestFileStatTakesHandles duplicates the parameter oracle. Keep every test an acceptance criterion names.

## Steps
- [ ] A POSIX probe that exits non-zero fails the test on every OS
- [ ] StaleExemptions checks platformExempt and heldExempt on the OS they name, calling each sub with its oracle case's arguments, and a perl error is a failure
- [ ] Fold the duplicated machinery; keep criterion-named tests

## Acceptance criteria

### Positive Scenarios
- [ ] TestPOSIXProbeDeathFails (`go test ./internal/infer/ -run '^TestPOSIXProbeDeathFails$' -count=1 -v | grep -q '^--- PASS: TestPOSIXProbeDeathFails'`)
- [ ] TestStalePlatformExemptionFails (`go test ./internal/perltest/ -run '^TestStalePlatformExemptionFails$' -count=1 -v | grep -q '^--- PASS: TestStalePlatformExemptionFails'`)
- [ ] TestRealPlatformExemptionsAreChecked (`go test ./internal/infer/ -run '^TestRealPlatformExemptionsAreChecked$' -count=1 -v | grep -q '^--- PASS: TestRealPlatformExemptionsAreChecked'`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] TestStaleCheckPerlErrorFails (`go test ./internal/perltest/ -run '^TestStaleCheckPerlErrorFails$' -count=1 -v | grep -q '^--- PASS: TestStaleCheckPerlErrorFails'`)
