---
title: use 5.04 is read as version 5.4, not 5.040, because one helper serves decimal versions and bundle names
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-11T01:59:26.93012205Z
updated: 2026-10-11T01:59:26.93012205Z
---

## Context

- paths: `internal/lexer` (`versionAtLeast`, `noteSignatures`), `internal/parse` (`noteFeatures`)
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-11 by 01a0ec10 (d4b186ca fixed `use 5.036` being read as minor 3). versionAtLeast keeps the first three fraction digits of a decimal version, which is right for `5.036`/`5.042`, but reads `use 5.04` as minor 4 where perl reads 5.040. It cannot switch to perl's full decimal reading as-is: the `:5.36` feature-bundle path calls the same helper with `5.36`, and `:5.10` would then count as 5.100. Measure perl on `use 5.04`, `use 5.4`, `use 5.040`, `use 5.36` (decimal) and the `:5.10` bundle, then give the two callers the reading each needs.

## Steps

- [ ] Measure; failing test first; separate the decimal-version and bundle-name readings; `make test`; commit

## Acceptance criteria

- [ ] a decimal `use VERSION` is read as perl reads it, and a feature bundle name as perl reads it (`go test ./internal/lexer/ -run '^TestVersionReadingsMatchPerl$' -count=1 -v | grep -q '^--- PASS: TestVersionReadingsMatchPerl'`)
