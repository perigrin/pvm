---
title: At most 9 library precedence levels fit between two of perl's
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-08T13:39:50.125333168Z
updated: 2026-10-16T18:48:35.051829779Z
---

## Context

- paths: internal/parse/operators.go (library level powers; a `ponytail:` comment marks the ceiling)
- commands: `make test` (green before every commit); perl 5.42 `/home/perigrin/.local/bin/perl`

Found reviewing 01a11923-bb28 (2026-10-08), which wired a library's infix operators into the parser at the level their relations give (`libraryInfix`, `p.infixOp` in internal/parse).

A library level gets a binding power strictly between its CORE.tpm neighbours', which are 10 apart, so a tenth level between the same two neighbours is an error. Either re-derive powers for the whole order including libraries, or widen the spacing.

## Steps

- [ ] Write failing test `TestManyLibraryLevelsBetweenTwo`; implement; `make test`; commit

## Acceptance criteria

- [ ] more than 9 library levels between two perl levels still order correctly (`go test ./internal/parse/ -run '^TestManyLibraryLevelsBetweenTwo$' -count=1 -v | grep -q '^--- PASS: TestManyLibraryLevelsBetweenTwo'`)
