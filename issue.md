---
title: reverse's scalar candidate does not declare its $_ default
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-08T01:28:39.358534484Z
updated: 2026-10-08T01:52:32.719953544Z
---

## Context

- paths: `internal/parse/declarations/CORE.pmt` (`reverse`, lines ~617-621), `docs/rfcs/0001-type-libraries.md` (defaults; the slurpy rule)
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Measured on 5.42: with `$_ = "xyz"`, `my $t = reverse` is `"zyx"` -- bare `reverse` in scalar context reverses `$_` (perlfunc documents it). CORE.pmt's scalar candidate, `(List @list) Str`, does not say so. A list parameter may take a default -- RFC 0001's typed-Perl extension (perl's signatures refuse it), confirmed by perigrin 2026-10-08, already used by `chomp`/`chop` (`\(@args = ($_))`) and `print` (`= ($_)`) -- so the scalar candidate is `(List @list = ($_)) Str`. The list candidate keeps no default: bare `reverse` in list context is the empty list (measure to confirm).

## Steps

- [ ] Measure bare `reverse` in list context on 5.42
- [ ] Write failing test `TestCoreReverseScalarDefaultsToTopic`: the scalar candidate's list parameter defaults to `($_)`; declare it; `make test`; commit

## Acceptance criteria

- [ ] reverse's scalar candidate states its $_ default (`go test ./internal/parse/ -run '^TestCoreReverseScalarDefaultsToTopic$' -count=1 -v | grep -q '^--- PASS: TestCoreReverseScalarDefaultsToTopic'`)
