---
title: Write one ParenNode pair where Paren and binding power coincide
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-23T14:43:13.25602072Z
updated: 2026-10-23T14:43:13.25602072Z
---

## Context

Raised by a fresh reader (fresh6), m3-oracle-at-scale chain-review round 8 (follow-up).

The parser sets Paren on the `+` in `(1 + 2) * 3`. d519 pins that pair as binding power, not marked as ParenNode. If 75d4 keeps it unmarked, extra-nesting cannot forgive `((1+2))*3`, though perl builds the same tree. If 75d4 writes a Paren pair in addition, every grouped file drops.

## Steps

- [ ] In 75d4, a node that needs both gets ONE pair, marked ParenNode; update d519's fixture

## Acceptance criteria

### Positive Scenarios
- [ ] `((1+2))*3` re-emits faithfully (`go test ./internal/parse/ -run '^TestCanonParenAndBindingShareOnePair$' -count=1 -v | grep -q '^--- PASS: TestCanonParenAndBindingShareOnePair'`)
