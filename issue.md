---
title: Split dropped-paren rows by psc node kind, not by following token
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-23T14:43:13.221389845Z
updated: 2026-10-23T14:43:13.221389845Z
---

## Context

Raised by a fresh reader (fresh4), m3-oracle-at-scale chain-review round 8 (follow-up).

The 43 causes measured in the dropped-paren class are one family of fixes: psc not setting Paren. The token that happens to follow cuts them into pieces. If 01a167da-3323 splits, its siblings will be arbitrary.

## Steps

- [ ] When 3323 splits, group rows by the psc node kind under the paren, or accept the token-based cut and say so

## Acceptance criteria

### Positive Scenarios
- [ ] The dropped-paren split groups by node kind (`grep -qi 'node kind' docs/specs/perl-parser/07-conformance.md`)
