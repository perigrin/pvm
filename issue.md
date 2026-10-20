---
title: "Fix the three narrowing regressions: a deref keeps a blessed object, possible-undef joins across branches, an element guard ends at assignment"
state: done
urgency: normal
milestone: core-module-types
created: 2026-10-20T19:18:42.264199971Z
updated: 2026-10-20T20:41:31.706736762Z
sessions:
- start_sha: cf08c0acdb986b7a59c463c2ed994e0721307294
  end_sha: dee5725d764cf502cf58ded7dcdf4778cb5781d1
  commits: 3
  started_at: 2026-10-20T19:19:19.932107544Z
  ended_at: 2026-10-20T20:41:31.706736762Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-20T19:19:19.932107544Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-20T20:41:31.706736762Z
observed_paths:
- .github/workflows/ci.yml
- docs/contributing/core-module-declarations.md
- docs/rfcs/0001-type-libraries.md
- internal/data/help/types.md
- internal/infer/calls.go
- internal/infer/flow.go
- internal/infer/infer.go
- internal/infer/narrow_lib_test.go
- internal/infer/narrowing_regressions_test.go
- internal/infer/ops.go
- internal/infer/returntypes_test.go
- internal/infer/symbols.go
- internal/infer/vars.go
- internal/parse/declarations/IO/Socket/IP.tpm
- internal/parse/heldtoperl_test.go
- internal/parse/holdcoverage_test.go
- internal/perltest/batches.go
- internal/perltest/fixture.go
- internal/perltest/fixture_test.go
- internal/perltest/perl542_test.go
- internal/perltest/perltest.go
- internal/psc/lib_baseline_test.go
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
