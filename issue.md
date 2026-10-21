---
title: "Flow regressions from the fix round: guard shadows keep the possible-undef mark, loops join their zero-iteration path, element shadows end at every write, ternaries stay linear"
state: pending
urgency: normal
milestone: core-module-types
created: 2026-10-21T05:19:05.988291198Z
updated: 2026-10-21T05:19:05.988291198Z
---

crochet:review round 2 of PR #488 at 2db182e6 (2026-10-21). Quality report: scratchpad/review2-core-module-types-paad.md, probes scratchpad/r2probes/, psc builds of head, cf08c0ac and pu in scratchpad/r2bin/. A finding pu gets right is a regression this milestone introduced: fix it. A gap pu shares is out of scope here and goes to the follow-up named in the issue.

Quality I3 (flow.go:603,:654: a guard's shadow drops the possible-undef mark; `$b = $c ? blessed($x) : []; if (!ref $b) { $b eq "C" }` reported, pu and perl silent), I4 (flow.go:316,:425,:453, map/grep: loop bodies not joined with the zero-iteration path; `for ... { $y->{k} } bless $y,'Foo'` missed, pu reports -- the pu-wide loop gap for the possible-undef mark is the follow-up, not this issue), I5 (ops.go:228-247, flow.go:453: element shadows survive foreach aliasing, `@r[0,1]=...`, `$r[$i]=...`, `(my $q, @r)=...`; and symbols.go:296-321 clearing overreaches into sibling branches and never-run anon subs), I10 (flow.go:633-688 via ops.go:542-552: right-nested ternary over distinct variables is cubic, 55s at n=1000 vs pu 0.36s), S3 (if without else joins fall-through from the un-narrowed state), S5 (a ternary's else arm starts from the then arm's end state).

## Steps
- [ ] Failing test per finding from its r2probes probe, against perl
- [ ] Fix at the root in flow.go/ops.go/symbols.go
- [ ] Time a 1000-deep nested ternary: within 2x of pu

## Acceptance criteria

### Positive Scenarios
- [ ] TestGuardShadowKeepsPossibleUndef (`go test ./internal/infer/ -run '^TestGuardShadowKeepsPossibleUndef$' -count=1 -v | grep -q '^--- PASS: TestGuardShadowKeepsPossibleUndef'`)
- [ ] TestLoopBodyJoinsZeroIterations (`go test ./internal/infer/ -run '^TestLoopBodyJoinsZeroIterations$' -count=1 -v | grep -q '^--- PASS: TestLoopBodyJoinsZeroIterations'`)
- [ ] TestElementShadowEndsAtEveryWrite (`go test ./internal/infer/ -run '^TestElementShadowEndsAtEveryWrite$' -count=1 -v | grep -q '^--- PASS: TestElementShadowEndsAtEveryWrite'`)
- [ ] TestTernaryArmsStartFromTheCondition (`go test ./internal/infer/ -run '^TestTernaryArmsStartFromTheCondition$' -count=1 -v | grep -q '^--- PASS: TestTernaryArmsStartFromTheCondition'`)
- [ ] TestNestedTernaryIsNotCubic (`go test ./internal/infer/ -run '^TestNestedTernaryIsNotCubic$' -count=1 -v | grep -q '^--- PASS: TestNestedTernaryIsNotCubic'`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] TestShadowClearStaysInItsBranch (`go test ./internal/infer/ -run '^TestShadowClearStaysInItsBranch$' -count=1 -v | grep -q '^--- PASS: TestShadowClearStaysInItsBranch'`)
