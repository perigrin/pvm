---
title: A method-call argument leaves the arguments before it checked, and a list-returning call is not possible-undef in a list slot
state: pending
urgency: normal
milestone: core-module-types
created: 2026-10-22T13:10:12.410335626Z
updated: 2026-10-22T13:10:12.410335626Z
---

crochet:review round 5 of PR #488 at 7201d0dd (2026-10-22). Quality report: scratchpad/review5-core-module-types-paad.md; probes scratchpad/r5probes/ (RAW-FINDINGS.md), r4/p/, r3probes/, r2probes/. perigrin, 2026-10-22: one narrow round, then close -- fix only these, with Opus, each reviewed against pu, r2 (2db182e6), h3 (692a560c) and h4 (7201d0dd) with perl 5.42 before merge, then a final quality lens over this round's diff alone; every other finding is a follow-up. A probe pu, r2, h3 or h4 gets right must still be right.

Quality B1 (new in round 4, aca86f5d): calls.go:184-186 flattens *lower.MethodCall and :130-149 then runs no checkArg at all, so `fmod(undef, $o->one)` and `fmod("abc", $o->one)` (logicB/f101, f102) are silent where r2 and h3 report and perl warns. Quality B2 (new in round 4, aca86f5d; under --warn-possible-undef but broad): calls.go:522 forcedUndef (:375) wraps every excuse including the slurpy flatten and a List always carries the Undef bit, so `print Dumper($y)` and `print shuffle(@r)` report "expects Str, got List" where r2 and h3 are silent and perl is clean (r5probes/orch/dumper.pl, shuffle.pl, logicB/w7.pl). The user-sub and coderef forms of B1 are 01a16075 (out of scope).

## Steps
- [ ] Check each argument before the first flattening one against its parameter
- [ ] A flattened list into a slurpy is not forced possible-undef
- [ ] Probe table on pu, r2, h3, h4 and branch, default and every flag

## Acceptance criteria

### Positive Scenarios
- [ ] TestArgumentsBeforeAMethodCallAreChecked (`go test ./internal/infer/ -run '^TestArgumentsBeforeAMethodCallAreChecked$' -count=1 -v | grep -q '^--- PASS: TestArgumentsBeforeAMethodCallAreChecked'`)
- [ ] TestListIntoSlurpyNotPossibleUndef (`go test ./internal/infer/ -run '^TestListIntoSlurpyNotPossibleUndef$' -count=1 -v | grep -q '^--- PASS: TestListIntoSlurpyNotPossibleUndef'`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] TestCheckArgPossibleUndefUnderFlag (`go test ./internal/infer/ -run '^TestCheckArgPossibleUndefUnderFlag$' -count=1 -v | grep -q '^--- PASS: TestCheckArgPossibleUndefUnderFlag'`)
