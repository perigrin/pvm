---
title: A library's :prefix and :postfix operators are read but never parsed
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-08T13:39:49.846359833Z
updated: 2026-10-08T13:39:49.846359833Z
---

## Context

- paths: internal/parse/operators.go (`libraryInfix`), internal/parse/term.go, internal/parse/expr.go
- commands: `make test` (green before every commit); perl 5.42 `/home/perigrin/.local/bin/perl`

Found reviewing 01a11923-bb28 (2026-10-08), which wired a library's infix operators into the parser at the level their relations give (`libraryInfix`, `p.infixOp` in internal/parse).

Only infix library operators join the derived order and the parser. A library line `sub ⊖ :prefix :tighter(**) (Num $x) Num;` is read but `⊖ 3` does not parse as that operator.

## Steps

- [ ] Write failing test `TestLibraryPrefixPostfixOperatorsParse`; implement; `make test`; commit

## Acceptance criteria

- [ ] a library's prefix and postfix operators parse at their level (`go test ./internal/parse/ -run '^TestLibraryPrefixPostfixOperatorsParse$' -count=1 -v | grep -q '^--- PASS: TestLibraryPrefixPostfixOperatorsParse'`)
