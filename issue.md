---
title: "Tier 04_operators: the first tier where a parse can be WRONG rather than Unknown"
state: pending
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c361-7479-7275-ace5-cd9a6dd2bd6c
created: 2026-09-21T09:52:30.999321803Z
updated: 2026-09-21T13:40:21.693174426Z
---

**Introduces:** arithmetic, string, comparison, logical operators,
precedence, associativity, the and/&& cliff

**The first tier where a parse can be WRONG rather than Unknown.**
Everything before this either parses or refuses; here it can parse into the
wrong shape and say nothing.

## What makes this tier different

The parser is Pratt/precedence-climbing over 32 levels, and `canon.go` reads
the `infix[]`/`prefix[]` tables at TEN sites. One wrong binding power
mis-groups going in and un-mis-groups coming out, so a round-trip check
cannot see it. This tier is what can.

Two hazards to cover rather than assume:

- **Nonassoc is checked separately.** Binding power STOPS recursion at equal
  power but does not REJECT. 11 of the 32 levels are `%nonassoc`, so
  `$a < $b < $c` must be measured against perl rather than inferred.
- **Comparison chaining** (`chain.go`, 5.32+ `cmpchain`) makes `3 > 2 > 1`
  evaluate to 1. A real 5.42 behaviour and it belongs here.

## The grouping question, deferred here on purpose

Whether this tier needs internal grouping by precedence level was left open
by the spec and deferred to this issue. Decide once the files exist and the
directory is actually unreadable -- a grouping invented before the files
exist will be the wrong grouping.

## Acceptance

- [ ] Every file in 04_operators/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierOperatorsPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsPerlValidated'`)
- [ ] The dependency lint passes over 04_operators/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierOperatorsLint$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsLint'`)
- [ ] 04_operators/README.md states what the tier introduces, what it depends on, and why check 2 places it here (`test -s conformance/04_operators/README.md`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierOperatorsRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsRefusalsCited'`)
- [ ] The whole corpus is green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
