---
title: A sub signature's parens are marked Paren where perl sets no OPf_PARENS
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-24T16:47:56.986645254Z
updated: 2026-10-24T16:47:56.986645254Z
---

Found by the independent review of 01a16602-5f1a (exec/5f1a, ed4f27c9) on 2026-10-24. The defect predates that branch.

## Context

- paths: `internal/parse/` (signature parsing), plus the parser's SyntaxParen flag that 5f1a introduced
- entrypoints: a sub signature's parens

psc puts a Paren on a signature's parens, but perl records no OPf_PARENS there. For `sub g ($x){}`, psc maps the Paren to ex-rv2sv, and perl's op carries no P. `($x,$y)` maps to ex-list with no P either. This is the same syntax-pair class that 5f1a split into SyntaxParen for `->(`, `not (` and list slices, but signatures were not among those three. Canon is still faithful.

Fix: the signature's pair takes SyntaxParen (takeSyntaxPair), and a parens fixture holds it to perl 5.44 Concise.

## Acceptance criteria

### Positive Scenarios

- [ ] A signature's parens carry no Paren, and perl agrees, op by op (`go test ./internal/parseoracle/ -run '^TestParenFixturesAgreeWithPerl$' -count=1 -v | grep -q '^--- PASS'`)
