---
title: "A mixed comparison chain loses its operators: 1 < 2 <= 3 canonicalises as 1 < 2 < 3"
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-10T01:40:31.955610785Z
updated: 2026-10-10T01:47:04.169520676Z
sessions:
- start_sha: 528c1fa24cbd7732a2aa37ec1c0bbfee0828a011
  end_sha: ""
  commits: 0
  started_at: 2026-10-10T01:47:04.169520676Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-10T01:47:04.169520676Z
---

## Context

- paths: `internal/parse/chain.go` (CmpChain; its extend path), `internal/parse/canon.go`
- commands: `make test`; the canon/fidelity ratchets; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-10 while fixing 01a12344: a comparison chain mixing operators of one class loses them. `1 < 2 <= 3` canons to `1 < 2 < 3`, and `$a<$b>$c` to `$a < $b < $c` -- a CmpChain keeps one `Text` for all its links. That changes what the code means, on valid perl (chained comparisons, 5.32+), and the fuzzer cannot see it: the wrong output is itself a fixpoint. Deparse keeps each operator.

## Steps

- [ ] Failing test: each link of a mixed chain keeps its own operator, in the tree and in canon; fix chain.go so each link records its operator; canon writes them; ratchets toward perl only; `make test`; commit

## Acceptance criteria

- [ ] a mixed comparison chain keeps every operator (`go test ./internal/parse/ -run '^TestMixedComparisonChainKeepsItsOperators$' -count=1 -v | grep -q '^--- PASS: TestMixedComparisonChainKeepsItsOperators'`)
