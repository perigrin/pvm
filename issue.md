---
title: The postfix call and subscripts ( [ { are placed by relations kept in Go, not CORE.pmt
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-08T09:50:05.713946929Z
updated: 2026-10-16T18:48:34.732051133Z
---

## Context

- paths: `internal/parse/operators.go` (`undeclaredOperators`: the `(`, `[`, `{` postfix entries), `internal/parse/declarations/CORE.tpm`, `docs/rfcs/0001-type-libraries.md`
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found reviewing 01a118fe-8dfe (2026-10-08). The parser's precedence is derived from CORE.tpm's relations, but the postfix call `(` and the subscripts `[` and `{` -- perly.y's tightest operators, below the terms -- are placed by relations kept in Go (`undeclaredOperators`), because perlop's table has no row for them and no issue gives them a CORE.tpm line. The goal is every operator defined in CORE.tpm; these need either a declaration form (a postfix circumfix: `f(...)`, `$a[...]`, `$h{...}`) or perigrin's ruling that they stay parser-side as grammar rather than operators.

## Steps

- [ ] Bring perigrin a proposed form (or the case for keeping them parser-side); record the ruling in RFC 0001
- [ ] If declared: failing test first; move them out of `undeclaredOperators`; the parser's binding powers must not change; `make test`; commit

## Acceptance criteria

- [ ] the call and subscripts are either declared in CORE.tpm or recorded in RFC 0001 as grammar (`go test ./internal/parse/ -run '^TestCallAndSubscriptsSettled$' -count=1 -v | grep -q '^--- PASS: TestCallAndSubscriptsSettled'`)
