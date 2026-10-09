---
title: my $0 parses, though perl cannot use a global in my
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-09T23:49:35.181226006Z
updated: 2026-10-09T23:49:35.181226006Z
---

## Context

- paths: `internal/parse/decl.go` (my/our/state declarators)
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-09 by the fuzz work on 01a0ed89: perl refuses `my $0` ("Can't use global $0 in "my""), and likewise other punctuation and special globals in `my`/`state`; internal/parse builds a Declaration. Measure which variables perl refuses in `my`, `our`, `state` (punctuation vars, `$_` under various versions, `@ARGV`, `%ENV`, `$::x`, `$Foo::x`).

## Steps

- [ ] Measure; failing test first; refuse as perl does; `make test`; commit

## Acceptance criteria

- [ ] a global perl refuses in my is refused (`go test ./internal/parse/ -run '^TestMyRefusesAGlobal$' -count=1 -v | grep -q '^--- PASS: TestMyRefusesAGlobal'`)
