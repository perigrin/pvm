---
title: "Flow facts through loops and method-argument checks: gaps pu shares"
state: pending
urgency: normal
milestone: v0.1
blocks:
- 01a16049-0596-72c5-b8df-0a6837e2a777
created: 2026-10-21T05:19:23.69581367Z
updated: 2026-10-22T00:06:57.153827242Z
---

Gaps crochet:review round 2 (PR #488, 2026-10-21) found that pu shares, so they are not regressions of core-module-types:

- Loops: the possible-undef mark and plain assignments inside a loop body are never joined with the zero-iteration path (flow.go:316, :425, :453, map/grep blocks) -- pu misses these too.
- `no strict 'refs'`: a deref retypes a string (vars.go:111-119), flag-only (round-1 I1(c), round-2 S6).
- Method calls: arguments to a method are never checked (calls.go:750), so declared method parameters are latent (round-2 S18).

- From the 01a16016-453e review (2026-10-21): `exits()` (flow.go:1022) knows only `return`, so a ternary arm or an if branch that is `next` or `last` is not an exit (t06, u06, u07 in scratchpad/r4/p/); pu shares it.

- From the 01a16016-453e confirm review (2026-10-22): `our @r` declared inside a guard on a lexical `my @r` does not hide the guard's `$r[0]` shadow (scratchpad/r4/p/s07); perl dies (`@main::r` is empty); every build, pu included, is silent.

## Steps
- [ ] Join a loop body's end state with the state before the loop, for every flow fact
- [ ] Check method arguments against declared method parameters
- [ ] psc check over perl 5.42's lib adds no unlisted line

## Acceptance criteria

### Positive Scenarios
- [ ] A possible-undef mark set in a loop survives its zero-iteration path (`go test ./internal/infer/ -run '^TestLoopKeepsPossibleUndefMark$' -count=1 -v | grep -q '^--- PASS: TestLoopKeepsPossibleUndefMark'`)
- [ ] A method call's arguments are checked (`go test ./internal/infer/ -run '^TestMethodArgumentsChecked$' -count=1 -v | grep -q '^--- PASS: TestMethodArgumentsChecked'`)

### Negative Scenarios
- [ ] psc check over perl's lib adds no unlisted line (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
