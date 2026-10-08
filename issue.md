---
title: XS::Parse::Infix's LOW and HIGH classes have no level
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-08T13:39:49.983282202Z
updated: 2026-10-08T13:39:49.983282202Z
---

## Context

- paths: internal/parse/operators.go (`classAnchors`), docs/rfcs/0001-type-libraries.md
- commands: `make test` (green before every commit); perl 5.42 `/home/perigrin/.local/bin/perl`

Found reviewing 01a11923-bb28 (2026-10-08), which wired a library's infix operators into the parser at the level their relations give (`libraryInfix`, `p.infixOp` in internal/parse).

`:infix(LOW)` and `:infix(HIGH)` are XS::Parse::Infix's plugin hooks below `or` and above perl's core operators, but no CORE.pmt operator anchors them, so a library operator classed only LOW or HIGH is "related to no other operator". Decide (perigrin) whether they mean below every core operator / above them, and give them levels in the derived order.

## Steps

- [ ] Write failing test `TestLowHighClassesHaveLevels`; implement; `make test`; commit

## Acceptance criteria

- [ ] LOW and HIGH place an operator below and above the core operators (`go test ./internal/parse/ -run '^TestLowHighClassesHaveLevels$' -count=1 -v | grep -q '^--- PASS: TestLowHighClassesHaveLevels'`)
