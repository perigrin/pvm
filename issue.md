---
title: A chained comparison draws a false operand warning, as legacy read it pairwise
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-13T03:45:34.007761338Z
updated: 2026-10-13T03:45:34.007761338Z
---

## Context

- paths: `internal/parse/chain.go` (CmpChain; its extend path), `internal/parse/canon.go`
- commands: `make test`; the canon/fidelity ratchets; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-10 while fixing 01a12344: a comparison chain mixing operators of one class loses them. `1 < 2 <= 3` canons to `1 < 2 < 3`, and `$a<$b>$c` to `$a < $b < $c` -- a CmpChain keeps one `Text` for all its links. That changes what the code means, on valid perl (chained comparisons, 5.32+), and the fuzzer cannot see it: the wrong output is itself a fixpoint. Deparse keeps each operator.

## Steps

- [ ] Failing test: each link of a mixed chain keeps its own operator, in the tree and in canon; fix chain.go so each link records its operator; canon writes them; ratchets toward perl only; `make test`; commit

## Acceptance criteria

- [ ] a mixed comparison chain keeps every operator (`go test ./internal/parse/ -run '^TestMixedComparisonChainKeepsItsOperators$' -count=1 -v | grep -q '^--- PASS: TestMixedComparisonChainKeepsItsOperators'`)
