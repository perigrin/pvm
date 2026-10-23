---
title: "Tier 04_operators: the first tier where a parse can be WRONG rather than Unknown"
state: done
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c361-7479-7275-ace5-cd9a6dd2bd6c
blocks:
- 01a0c432-fbd5-7125-82fa-b52abb3ac7bc
created: 2026-09-21T09:52:30.999321803Z
updated: 2026-10-23T19:18:29.656211894Z
sessions:
- start_sha: 71979369e31e3d89136535fe4d6b05a22095e26b
  end_sha: dc1bea2c2baed5af0cf71ddf16e379ee18855d30
  commits: 1
  started_at: 2026-09-21T21:11:14.092040239Z
  ended_at: 2026-09-21T21:36:34.399829125Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T21:11:14.092040239Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T21:36:34.399829125Z
observed_paths:
- conformance/04_operators/00_adjacency.t
- conformance/04_operators/05_logical.t
- conformance/04_operators/README.md
- conformance/11_oo/00_adjacency.t
- conformance/11_oo/03_method_named.t
- conformance/11_oo/README.md
- internal/conformance/tier04_test.go
- internal/conformance/tier11_test.go
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

Criteria updated 2026-10-23: e8ee9001 replaced the tier-keyed tests with topic files; each remapped command now names the test that holds the same property (old -> new: `TestCorpus/04_operators` -> `TestCorpus` over the tier-04 topics (adjacency-04_operators.md, arithmetic.md, bitwise.md, comparison.md, compound-assignment.md, named-operators.md, unary.md)). Left unchanged: `TestTierOperatorsAdjacency`, because no current test holds that property.

Criteria updated 2026-10-23 (agent:ccx-adj, issue 01a1696e): the deleted per-tier adjacency tests are replaced by one generic test, `TestAdjacencyCoversItsTier`, whose per-tier subtest checks that the tier's adjacency topic emits every op its README's INTRODUCES list names, that each case in it is multi-statement, and that it emits an op of the declared prerequisite. Old -> new: `TestTierOperatorsAdjacency` -> `TestAdjacencyCoversItsTier/04_operators`. These commands pass once branch ccx/adj merges.

## Acceptance criteria

- [ ] Each of the 11 nonassoc levels is measured against perl rather than inferred from the binding-power table (`go test ./internal/conformance/ -run '^TestTierOperatorsNonassocMeasured$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsNonassocMeasured'`)
- [ ] Comparison chaining is measured: 3 > 2 > 1 is 1 under 5.42 cmpchain (`go test ./internal/conformance/ -run '^TestTierOperatorsCmpchain$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsCmpchain'`)
- [ ] The 31 adjacent precedence pairs are fixtures whose two groupings produce different observable output, discriminated by effect order (`go test ./internal/conformance/ -run '^TestTierOperatorsAdjacentPairs$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsAdjacentPairs'`)
- [ ] Every file in 04_operators/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierOperatorsPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsPerlValidated'`)
- [ ] The dependency lint passes over 04_operators/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierOperatorsLint$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsLint'`)
- [ ] The tier's adjacency topic puts every construct 04_operators introduces in one body, each adjacent to another, pairing with the declared prerequisite rather than tier N-1 (`go test ./internal/conformance/ -run '^TestAdjacencyCoversItsTier$/^04_operators$' -count=1 -v | grep -q '^    --- PASS: TestAdjacencyCoversItsTier/04_operators'`)
- [ ] 04_operators/README.md states what the tier introduces, what it depends on, and why check 2 places it here, in the form the dependency lint reads (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierOperatorsRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsRefusalsCited'`)
- [ ] The tier's own files are green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$/^(adjacency-04_operators|arithmetic|bitwise|comparison|compound-assignment|named-operators|unary)\.md$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
