---
title: "An if with no else joins its fall-through from the unnarrowed state: the lazy-init guard reports"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-21T21:06:00.307244025Z
updated: 2026-10-21T21:06:00.307244025Z
---

crochet:review round 2 S3, carried untracked through round 3 (V16); destination given by 01a16016-481d.

An `if` with no `else` joins its fall-through path from the unnarrowed state before the condition (internal/infer/flow.go, the if-join), so a guard that fills a variable is lost after the if.

    my $b = undef;
    if (!defined $b) { $b = "D" }
    print "y\n" if $b eq "C";

perl 5.42 -w runs it silently. psc check at pu, 2db182e6, 692a560c and fix/481d all report `left operand of "eq": expected Str, got Undef|Int|Num|Str|NaN|Inf [coercion-mismatch]` (measured 2026-10-21). Pre-existing on pu.

Suggested fix (review 2): pass the state narrowed by the guard's failure as the fall-through state.

## Steps
- [ ] Write the failing test from the repro above
- [ ] Join the fall-through from the guard-failure state

## Acceptance criteria

### Positive Scenarios
- [ ] The lazy-init guard reports nothing (`go test ./internal/infer/ -run '^TestIfWithoutElseJoinsGuardFailure$' -count=1 -v | grep -q '^--- PASS: TestIfWithoutElseJoinsGuardFailure'`)

### Negative Scenarios
- [ ] psc check over perl's lib adds no unlisted line (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
