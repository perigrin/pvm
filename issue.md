---
title: The tier-02 adjacency case omits shift, pop and splice
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-23T15:43:24.426165882Z
updated: 2026-10-23T15:43:24.426165882Z
---

## Context

- paths: conformance/mdtest/adjacency-02_variables.md
- entrypoints: conformance/mdtest/adjacency-02_variables.md

Found by the 9d6b worker (m3-conformance-corpus). adjacency-02_variables.md says it holds "every construct this tier introduces", but its body has no `shift`, `pop` or `splice`. The gap with `shift` dates from 6001ca7f, which moved `shift` to tier 02. `pop` and `splice` joined tier 02 in 01a0dde3-9d6b.

## Steps

- [ ] Add `shift`, `pop` and `splice` to the adjacency case, measured with perl 5.42.0; update its pinned output
- [ ] Raise corpusFloor only if the change adds cases rather than editing one, and regenerate corpus.ratchet

## Acceptance criteria

### Positive Scenarios
- [ ] The tier-02 adjacency case uses every op tier 02 introduces (`go test ./internal/conformance/ -run '^TestAdjacencyCoversItsTier$' -count=1 -v | grep -q '^--- PASS: TestAdjacencyCoversItsTier'`)
