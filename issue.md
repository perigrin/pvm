---
title: Canon does not know library operators
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-08T13:39:50.249754437Z
updated: 2026-10-16T18:48:35.241802891Z
---

## Context

- paths: internal/parse/canon.go
- commands: `make test` (green before every commit); perl 5.42 `/home/perigrin/.local/bin/perl`

Found reviewing 01a11923-bb28 (2026-10-08), which wired a library's infix operators into the parser at the level their relations give (`libraryInfix`, `p.infixOp` in internal/parse).

Canonical output prints `1 + (2 ⊕ 3 * 4)` with a redundant explicit paren for a library operator, because canon's minimal-parenthesisation consults CORE.tpm's powers only. Make it use the same powers the parse used.

## Steps

- [ ] Write failing test `TestCanonLibraryOperatorMinimalParens`; implement; `make test`; commit

## Acceptance criteria

- [ ] canon parenthesises library operators minimally (`go test ./internal/parse/ -run '^TestCanonLibraryOperatorMinimalParens$' -count=1 -v | grep -q '^--- PASS: TestCanonLibraryOperatorMinimalParens'`)
