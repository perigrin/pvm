---
title: "Round-3 flow regressions: a ternary's then-arm joins back, frame bookkeeping stays linear, element shadows end at every write, a loop deref no longer reintroduces pu's false positive"
state: in-progress
urgency: normal
milestone: core-module-types
created: 2026-10-21T20:10:34.686320063Z
updated: 2026-10-21T20:10:59.37611203Z
sessions:
- start_sha: 692a560c81fe680ed09244599f0e1bd216c5a075
  end_sha: ""
  commits: 0
  started_at: 2026-10-21T20:10:59.37611203Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-21T20:10:59.37611203Z
---

crochet:review round 3 of PR #488 at 692a560c (2026-10-21). Quality report: scratchpad/review3-core-module-types-paad.md; probes scratchpad/r3probes/ (F-<lens>.md per specialist); rebuild head/r2(2db182e6)/cf(cf08c0ac)/pu psc as the report says. perigrin, 2026-10-21: round 4 runs with Opus workers, and every fix gets a mandatory independent review that runs pu, round 2 and head against perl 5.42 before it merges. A finding pu or round 2 gets right is a regression: fix it without trading it for a new one -- every probe in r2probes/ and r3probes/ that pu or r2 gets right must still be right.

Quality V2 (ops.go:589-604: the then-arm's writes are undone and never joined; `$c ? ($y=undef) : 1; bless $y` silenced -- r2, cf, pu report, perl dies; `!defined $x ? ($x='d') : 0; $x eq 'a'` new false positive; the comment at :576-588 is false), V5 (symbols.go:126-132 recordFrame is O(depth) per update with flow.go:219: 300 nested ifs 24s vs pu 1.5s, a 400-deep writing ternary chain 15s vs 0.1s, nestif400 99s vs 4.9s), V6 (ops.go:239, flow.go:580: element shadows survive ++, `undef $e`, $_ writes in map/grep, foreach over an element or slice, and reference aliases -- pu gets these right), and the report's pre-existing note that round 3's loop-deref join (vars.go:122-129) brings back a false positive pu has and r2/cf had fixed.

## Steps
- [ ] Run every r2probes/ and r3probes/ flow probe on pu, r2, head and this branch against perl; nothing pu or r2 gets right may go wrong
- [ ] Time nested ifs (300) and writing ternary chains (400) against pu: within 2x
- [ ] Correct the false comment

## Acceptance criteria

### Positive Scenarios
- [ ] TestTernaryThenArmJoinsBack (`go test ./internal/infer/ -run '^TestTernaryThenArmJoinsBack$' -count=1 -v | grep -q '^--- PASS: TestTernaryThenArmJoinsBack'`)
- [ ] TestFrameBookkeepingIsLinear (`go test ./internal/infer/ -run '^TestFrameBookkeepingIsLinear$' -count=1 -v | grep -q '^--- PASS: TestFrameBookkeepingIsLinear'`)
- [ ] TestElementShadowEndsAtIncrementAndUndef (`go test ./internal/infer/ -run '^TestElementShadowEndsAtIncrementAndUndef$' -count=1 -v | grep -q '^--- PASS: TestElementShadowEndsAtIncrementAndUndef'`)
- [ ] TestLoopDerefNoFalsePositive (`go test ./internal/infer/ -run '^TestLoopDerefNoFalsePositive$' -count=1 -v | grep -q '^--- PASS: TestLoopDerefNoFalsePositive'`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] TestTernaryGuardNoFalsePositive (`go test ./internal/infer/ -run '^TestTernaryGuardNoFalsePositive$' -count=1 -v | grep -q '^--- PASS: TestTernaryGuardNoFalsePositive'`)
