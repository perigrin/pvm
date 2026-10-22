---
title: A guarded deref in a loop joins its zero-iteration path, and a deref after die narrows nothing
state: pending
urgency: normal
milestone: core-module-types
created: 2026-10-22T13:10:12.368955564Z
updated: 2026-10-22T13:10:12.368955564Z
---

crochet:review round 5 of PR #488 at 7201d0dd (2026-10-22). Quality report: scratchpad/review5-core-module-types-paad.md; probes scratchpad/r5probes/ (RAW-FINDINGS.md), r4/p/, r3probes/, r2probes/. perigrin, 2026-10-22: one narrow round, then close -- fix only these, with Opus, each reviewed against pu, r2 (2db182e6), h3 (692a560c) and h4 (7201d0dd) with perl 5.42 before merge, then a final quality lens over this round's diff alone; every other finding is a follow-up. A probe pu, r2, h3 or h4 gets right must still be right.

Quality A1 (new in round 4, d01a490b): flow.go:474-490 markLoopDeref keys the mark by the guard scope (:480) and loop (:408-415) skips closed scopes, so a guarded deref inside a loop no longer joins the zero-iteration path. r5probes/logicA/lg2.pl, the lazy-init idiom `for (1..$n) { $y->{k}=1 unless defined $y } bless $y`: perl dies; pu and h3 report; r2 and h4 silent (also lg1, lg3, lg4, lg6-lg8). Coverage round 5 (Important, unfiled since round 3's I4 residual): r3probes/logicA/i4m.pl `my $y=undef; eval { die 'x'; my $q=$y->{k} }; bless $y,'Foo';` -- perl dies; pu reports; r2, h3, h4 silent: a deref after a statement that always exits narrows a variable on a path that never reaches it.

## Steps
- [ ] Probe table over the round's probes on pu, r2, h3, h4 and branch
- [ ] Fix at the root in flow.go/vars.go; keep 453e's tests and timing

## Acceptance criteria

### Positive Scenarios
- [ ] TestLoopGuardedDerefJoinsZeroIterations (`go test ./internal/infer/ -run '^TestLoopGuardedDerefJoinsZeroIterations$' -count=1 -v | grep -q '^--- PASS: TestLoopGuardedDerefJoinsZeroIterations'`)
- [ ] TestDerefAfterDieNarrowsNothing (`go test ./internal/infer/ -run '^TestDerefAfterDieNarrowsNothing$' -count=1 -v | grep -q '^--- PASS: TestDerefAfterDieNarrowsNothing'`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] TestLoopDerefNoFalsePositive (`go test ./internal/infer/ -run '^TestLoopDerefNoFalsePositive$' -count=1 -v | grep -q '^--- PASS: TestLoopDerefNoFalsePositive'`)
