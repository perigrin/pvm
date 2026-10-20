---
title: "Fix the three narrowing regressions: a deref keeps a blessed object, possible-undef joins across branches, an element guard ends at assignment"
state: pending
urgency: normal
milestone: core-module-types
created: 2026-10-20T19:18:42.264199971Z
updated: 2026-10-20T19:18:42.264199971Z
---

crochet:review of PR #488, round 1 (2026-10-21). Quality lens I1, I3, I6 (verified regressions against pu; report scratchpad review-core-module-types-paad.md). I1 internal/infer/vars.go:98-105 narrowDerefBase overwrites an Object with HashRef and leaks out of &&/?: arms. I3 symbols.go:60, ops.go:265, calls.go:178-188: Symbol.PossibleUndef is not saved/restored/joined across branches and is lost through copies and ternaries; a definite Undef in an else branch is silenced. I6 flow.go:954-976, infer.go:1277, vars.go:59-63: the $r[0] guard shadow survives assignment, shift and splice.

Parameter rule (RFC 0001 "What a parameter type means", perigrin 2026-10-02): a parameter type is the type the operation coerces its argument to. A parameter is wrong when its type is not what the operation coerces to or preserves (a lone reference carp/warn/croak pass through; NaN/Inf a C double keeps), or it admits a value perl has no coercion for and dies on. A finding that only shows perl coercing a value (sum("abc"), encode with a ref) is withdrawn, not fixed. Every fix is measured with /home/perigrin/.local/bin/perl; never bend a type to quiet the lib baseline.

## Steps
- [ ] Write each failing test from the verifier's probe first
- [ ] Fix at the root shared by all callers
- [ ] Lib baseline and precision unchanged

## Acceptance criteria

### Positive Scenarios
- [ ] TestDerefKeepsBlessedObject (`go test ./internal/infer/ -run '^TestDerefKeepsBlessedObject$' -count=1 -v | grep -q '^--- PASS: TestDerefKeepsBlessedObject'`)
- [ ] TestDerefNarrowingStaysInItsArm (`go test ./internal/infer/ -run '^TestDerefNarrowingStaysInItsArm$' -count=1 -v | grep -q '^--- PASS: TestDerefNarrowingStaysInItsArm'`)
- [ ] TestPossibleUndefJoinsAcrossBranches (`go test ./internal/infer/ -run '^TestPossibleUndefJoinsAcrossBranches$' -count=1 -v | grep -q '^--- PASS: TestPossibleUndefJoinsAcrossBranches'`)
- [ ] TestPossibleUndefFollowsCopies (`go test ./internal/infer/ -run '^TestPossibleUndefFollowsCopies$' -count=1 -v | grep -q '^--- PASS: TestPossibleUndefFollowsCopies'`)
- [ ] TestElementGuardEndsAtAssignment (`go test ./internal/infer/ -run '^TestElementGuardEndsAtAssignment$' -count=1 -v | grep -q '^--- PASS: TestElementGuardEndsAtAssignment'`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] TestElementGuardEndsAtShift (`go test ./internal/infer/ -run '^TestElementGuardEndsAtShift$' -count=1 -v | grep -q '^--- PASS: TestElementGuardEndsAtShift'`)
- [ ] TestDefiniteUndefInElseIsReported (`go test ./internal/infer/ -run '^TestDefiniteUndefInElseIsReported$' -count=1 -v | grep -q '^--- PASS: TestDefiniteUndefInElseIsReported'`)
