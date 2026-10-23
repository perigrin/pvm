---
title: Say what a 100% faithful re-emission certifies
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-23T14:43:13.111054843Z
updated: 2026-10-23T14:43:13.111054843Z
---

## Context

Raised by the alignment lens, m3-oracle-at-scale chain-review round 8 (follow-up, low).

Once canon is precedence-aware and the call-paren rule is gone, `Faithful` catches every grouping where the tree disagrees with psc's own precedence model. It cannot see an error that the parser and canon share:
- a builtin misclassified in CORE.tpm or in the derived powers;
- a user sub whose prototype the file can't see (ruled M4).

The plan's §5 calls the metric "a measured claim about parse correctness", which overstates what is measured.

## Steps

- [ ] 61fe's docs say 100% faithful certifies agreement between the tree and psc's precedence model plus the source's paren placement; the model itself is held by the precedence tests and CORE.tpm
- [ ] Optional: a probe that judges each of canon's 84 named-unary builtins against the pinned perl's Deparse,-p grouping

## Acceptance criteria

### Positive Scenarios
- [ ] The plan's M3 section states what re-emission certifies (`grep -q 'precedence model' docs/plans/2026-09-05-parser-conformance-plan.md`)
