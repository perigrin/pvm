---
title: reverse's scalar candidate does not declare its $_ default
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-08T01:28:39.358534484Z
updated: 2026-10-08T01:28:39.358534484Z
---

## Context

- paths: `internal/parse/declarations/CORE.pmt` (`reverse`, lines ~617-621), `docs/rfcs/0001-type-libraries.md` (defaults; the slurpy rule)
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Measured on 5.42: with `$_ = "xyz"`, `my $t = reverse` is `"zyx"` -- bare `reverse` in scalar context reverses `$_` (perlfunc documents it). CORE.pmt's scalar candidate, `(List @list) Str`, does not say so, and no spelling yet lets a slurpy parameter default to `$_`. Needs perigrin's ruling: may a slurpy parameter take a default, e.g. `(List @list = $_) Str`?

## Steps

- [ ] Bring the spelling to perigrin; record it in RFC 0001
- [ ] Write failing test `TestCoreReverseScalarDefaultsToTopic` against the ruled spelling; declare it; `make test`; commit

## Acceptance criteria

- [ ] reverse's scalar candidate states its $_ default (`go test ./internal/parse/ -run '^TestCoreReverseScalarDefaultsToTopic$' -count=1 -v | grep -q '^--- PASS: TestCoreReverseScalarDefaultsToTopic'`)
