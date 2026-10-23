---
title: A gap row must be excused by its own fold
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-23T14:43:13.293162131Z
updated: 2026-10-23T14:43:13.293162131Z
---

## Context

Raised by a fresh reader (fresh1), m3-oracle-at-scale chain-review round 8 (follow-up).

5f1a keys gaps per statement and accepts a row when "the op which replaced it carries FOLD". In `print (1)+2*3`, perl 5.44 gives `const[IV 1] s` and `const[IV 6] s/FOLD`, so an invented Paren on `1` can be excused by the folded `2*3`.

## Steps

- [ ] Each gap row consumes one distinct absent-or-FOLD op of the mapped kind; an unfolded op of the mapped kind without P, carrying the node's own value, fails the row

## Acceptance criteria

### Positive Scenarios
- [ ] An unrelated fold cannot excuse an invented paren (`go test ./internal/parseoracle/ -run '^TestParenGapRowWithUnfoldedOpFails$' -count=1 -v | grep -q '^--- PASS: TestParenGapRowWithUnfoldedOpFails'`)
