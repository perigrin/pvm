---
title: while(){} is refused by internal/parse, though perl reads the empty condition as while (1)
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-10T05:33:50.642048534Z
updated: 2026-10-10T05:33:50.642048534Z
---

## Context

- paths: `internal/parse/control.go` (`parseParenCondition`)
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-10 writing TestEmptyForeachListIsRefused: `while(){}` parses to an Unknown in internal/parse, but perl compiles it (`perl -c` OK; Deparse reads an empty while condition as `while (1)`). `until(){}` -- measure.

## Steps

- [ ] Measure `while(){}`, `until(){}`, `while () {}` with Deparse; failing test first; accept as perl does; `make test`; commit

## Acceptance criteria

- [ ] an empty while condition parses (`go test ./internal/parse/ -run '^TestEmptyWhileConditionParses$' -count=1 -v | grep -q '^--- PASS: TestEmptyWhileConditionParses'`)
