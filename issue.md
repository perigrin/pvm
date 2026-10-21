---
title: "Test-harness gaps round 3 found: StaleExemptions fails closed, held exemptions are stale-checked, no criterion passes without testing"
state: done
urgency: normal
milestone: core-module-types
created: 2026-10-21T20:10:35.216500331Z
updated: 2026-10-21T21:43:55.845593938Z
sessions:
- start_sha: a893c4c7dbd73e99ee996512595b346f7c0d3243
  end_sha: 84130609b9dfb7fb2f5a8d8ab1e6ef31f8857cca
  commits: 1
  started_at: 2026-10-21T20:24:17.847073473Z
  ended_at: 2026-10-21T21:43:55.845593938Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-21T20:24:17.847073473Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-21T21:43:55.845593938Z
observed_paths:
- internal/infer/returntypes_test.go
- internal/parse/exportattr_test.go
- internal/parse/heldtoperl_test.go
- internal/parse/pin_test.go
- internal/perltest/fixture.go
- internal/perltest/staleexempt.go
- internal/perltest/staleexempt_test.go
---

crochet:review round 3 of PR #488 at 692a560c (2026-10-21). Quality report: scratchpad/review3-core-module-types-paad.md; probes scratchpad/r3probes/ (F-<lens>.md per specialist); rebuild head/r2(2db182e6)/cf(cf08c0ac)/pu psc as the report says. perigrin, 2026-10-21: round 4 runs with Opus workers, and every fix gets a mandatory independent review that runs pu, round 2 and head against perl 5.42 before it merges. A finding pu or round 2 gets right is a regression: fix it without trading it for a new one -- every probe in r2probes/ and r3probes/ that pu or r2 gets right must still be right.

Quality V8 (staleexempt.go:42-75 fails open: a missing call case dies inside the eval and counts as dies; methods are called as functions; only reasons containing "dies" are checked), V9 (TestPOSIXProbeDeathFails returntypes_test.go:1555 fails on a machine without 5.42 and passes on any child failure), V14 (heldExempt is never stale-checked, heldtoperl_test.go:53-111), V20 (several acceptance criteria pass without testing anything -- list them from the report and give each a positive control), coverage F6 (doneAuditIssues omits 01a15ac0-6cfa).

## Steps
- [ ] A missing call case, a method, and any reason wording are all checked; an exemption that no longer differs fails
- [ ] Give every vacuous criterion a positive control

## Acceptance criteria

### Positive Scenarios
- [ ] TestStaleCheckMissingCaseFails (`go test ./internal/perltest/ -run '^TestStaleCheckMissingCaseFails$' -count=1 -v | grep -q '^--- PASS: TestStaleCheckMissingCaseFails'`)
- [ ] TestStaleCheckCallsMethods (`go test ./internal/perltest/ -run '^TestStaleCheckCallsMethods$' -count=1 -v | grep -q '^--- PASS: TestStaleCheckCallsMethods'`)
- [ ] TestHeldExemptionsAreChecked (`go test ./internal/parse/ -run '^TestHeldExemptionsAreChecked$' -count=1 -v | grep -q '^--- PASS: TestHeldExemptionsAreChecked'`)
- [ ] TestPOSIXProbeDeathFails (`go test ./internal/infer/ -run '^TestPOSIXProbeDeathFails$' -count=1 -v | grep -q '^--- PASS: TestPOSIXProbeDeathFails'`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] TestPendingAuditHasNoDoneIssue (`go test ./internal/infer/ -run '^TestPendingAuditHasNoDoneIssue$' -count=1 -v | grep -q '^--- PASS: TestPendingAuditHasNoDoneIssue'`)
