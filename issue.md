---
title: A one-element compound paren list (f(1),) loses its parens in canon
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-09T18:18:11.268233623Z
updated: 2026-10-09T18:18:11.268233623Z
---

## Context

- paths: `internal/parse/canon.go` (paren lists, the one-element trailing-comma case fixed by 01a121ac)
- commands: `make test`; the canon/fidelity ratchets

Found 2026-10-09 by 01a121ac: a one-element paren list whose element is compound, like `my @a = (f(1),);`, loses its parens in canon. Canon stays a fixpoint, so the fuzzer does not flag it, but the source's parens are lost. perl's Deparse keeps a list: `my(@a) = f(1);` -- measure whether canon's output means the same to perl (fidelity) before deciding it is a defect.

## Steps

- [ ] Measure perl Deparse of source vs canon for `(f(1),)`, `(f(1))`, `($x,)`; failing test first if canon changes the meaning; fix; ratchets toward perl only; `make test`; commit

## Acceptance criteria

- [ ] a one-element paren list keeps what perl reads in it (`go test ./internal/parse/ -run '^TestOneElementParenListKeepsItsParens$' -count=1 -v | grep -q '^--- PASS: TestOneElementParenListKeepsItsParens'`)
