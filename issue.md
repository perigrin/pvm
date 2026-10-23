---
title: Enforce NL's stop-if-any-file-drops rule with a criterion
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-23T14:43:13.147998886Z
updated: 2026-10-23T14:43:13.147998886Z
---

## Context

Raised by a fresh reader (fresh2), m3-oracle-at-scale chain-review round 8 (follow-up).

NL's count criteria assert only `>=` the start counts measured at 054e75e6. Earlier issues may raise those counts first. So a file that drops during NL can be offset by a gain, re-baselined into perlgittfaithful.ratchet, and still pass. Today the stop rule rests on step text and per-fix review.

## Steps

- [ ] Add a git-pinned check that no perlgittfaithful.ratchet row flips 0 to 1 between the last commit of the named-unary issue and HEAD, using the `git log -G` / `git show` pattern

## Acceptance criteria

### Positive Scenarios
- [ ] No faithful row regresses after the named-unary issue (`go test ./internal/parse/ -run '^TestFaithfulRatchetNoRowRegressesAfterNamedUnary$' -count=1 -v | grep -q '^--- PASS: TestFaithfulRatchetNoRowRegressesAfterNamedUnary'`)
