---
title: dump is declared with no operand, but perl accepts dump LABEL and dump EXPR
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-08T05:09:38.521054731Z
updated: 2026-10-16T18:48:34.57139992Z
---

## Context

- paths: `internal/parse/declarations/CORE.tpm` (`dump`, and `goto` beside it), `internal/parse/operators.go` (shape derivation)
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found reviewing 01a118fe-8d68 (2026-10-08). CORE.tpm declares `sub dump :equiv(=) () None;` -- no operand -- and perl reports `prototype("CORE::dump")` as `""`, yet 5.42 accepts an operand: `CORE::dump FOO` and `CORE::dump $x` both compile (Deparse: `(CORE::dump FOO)`, `(CORE::dump $x)`), and at the assignment row (`CORE::dump $x = 1` is `CORE::dump ($x = 1)`). perl's `""` here misreports, as `goto`'s undef prototype does. So dump's signature should take an optional label operand like goto's (`Str|Code $label = ""` or `Optional[...]`, measured), and its shape is not a term.

## Steps

- [ ] Measure `dump` with no operand, a bareword label and an expression on 5.42 (`perl -c` and Deparse only; never run dump)
- [ ] Write failing test `TestCoreDumpTakesALabel`; declare the operand; keep its `:equiv(=)` level; `make test`; commit

## Acceptance criteria

- [ ] dump's declaration takes the label operand perl accepts (`go test ./internal/parse/ -run '^TestCoreDumpTakesALabel$' -count=1 -v | grep -q '^--- PASS: TestCoreDumpTakesALabel'`)
