---
title: The tier-02 adjacency case omits shift, pop and splice
state: done
urgency: normal
milestone: m3-conformance-corpus
created: 2026-10-23T15:43:24.426165882Z
updated: 2026-10-23T21:08:34.857253847Z
sessions:
- start_sha: 9c3e164290f60fe6c5c1625dbd8e15a2f8a4dc3f
  end_sha: 41ca9f14c8e85a8ad6fce97e8171f817ef9702a7
  commits: 1
  started_at: 2026-10-23T19:10:22.514700061Z
  ended_at: 2026-10-23T19:35:22.146488826Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-23T19:10:22.514700061Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-23T19:35:22.146488826Z
observed_paths:
- conformance/01_literals/README.md
- conformance/11_oo/README.md
- conformance/mdtest/adjacency-01_literals.md
- conformance/mdtest/adjacency-02_variables.md
- conformance/mdtest/adjacency-04_operators.md
- conformance/mdtest/adjacency-07_subroutines.md
- conformance/mdtest/adjacency-10_io.md
- conformance/mdtest/adjacency-11_oo.md
- conformance/mdtest/adjacency-13_opaque.md
- internal/conformance/adjacency_test.go
- internal/conformance/tier02_test.go
- internal/conformance/tier04_test.go
---

## Context

- paths: conformance/mdtest/adjacency-02_variables.md
- entrypoints: conformance/mdtest/adjacency-02_variables.md

Found by the 9d6b worker (m3-conformance-corpus). adjacency-02_variables.md says it holds "every construct this tier introduces", but its body has no `shift`, `pop` or `splice`. The gap with `shift` dates from 6001ca7f, which moved `shift` to tier 02. `pop` and `splice` joined tier 02 in 01a0dde3-9d6b.

## Steps

- Done (498821c3): `shift`, `pop` and `splice` were added to the adjacency case, measured with perl 5.42.0, and its pinned output was updated.
- Done (498821c3): no cases were added, so corpusFloor stays at 270 and corpus.ratchet is unchanged.

## Acceptance criteria

### Positive Scenarios
- [ ] The tier-02 adjacency case uses every op tier 02 introduces (`go test ./internal/conformance/ -run '^TestAdjacencyCoversItsTier$' -count=1 -v | grep -q '^--- PASS: TestAdjacencyCoversItsTier'`)
