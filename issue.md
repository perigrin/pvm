---
title: use 'A', use $x, use;, use {A} and use 5.010 (1) parse, though perl refuses each
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-09T22:56:53.634413304Z
updated: 2026-10-09T22:56:53.634413304Z
---

## Context

- paths: `internal/parse/use.go`, `internal/parse/refusal.go`
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-09 by 01a12285 (which made `use(`/`no(` refuse). perl also refuses, measured on 5.42: `use 'A'`, `use $x`, `use;`, `use {A}`, `use 5.010 (1)`. internal/parse accepts them.

## Steps

- [ ] Measure each (and the `no` spellings); failing test first; refuse through an existing refusal site where one fits; `make test`; commit

## Acceptance criteria

- [ ] a use or no perl refuses is refused (`go test ./internal/parse/ -run '^TestMalformedUseIsRefused$' -count=1 -v | grep -q '^--- PASS: TestMalformedUseIsRefused'`)
