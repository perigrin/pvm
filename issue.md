---
title: state is always a declarator in internal/parse, but perl reads it as a bareword when its feature is off
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-09T18:18:11.144884563Z
updated: 2026-10-09T18:18:11.144884563Z
---

## Context

- paths: `internal/parse/decl.go` (declarators), the parser's `features` map (`noteFeatures`)
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-09 by the format-panic fuzz work (01a121ac). With the `state` feature off, `state` is an ordinary bareword to perl -- `perl -c` on a file holding just `state` is valid -- but internal/parse always treats it as a declarator; with 01a121ac's `missing_operand` refusal it now refuses instead of building a wrong tree. Neither matches perl. Related: `$ ` and `$\n` at end of file still lex as a Variable, though perl reports a syntax error (01a121ac made a bare `$` at EOF an Error token, not these).

## Steps

- [ ] Measure perl with and without `use feature 'state'` / a version bundle; failing test first; read `state` as perl does per the feature state; measure and fix `$ `/`$\n` at EOF; `make test`; commit

## Acceptance criteria

- [ ] `state` is a declarator only where its feature is on (`go test ./internal/parse/ -run '^TestStateFollowsItsFeature$' -count=1 -v | grep -q '^--- PASS: TestStateFollowsItsFeature'`)
