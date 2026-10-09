---
title: A heredoc with no terminator parses, though perl cannot find the string terminator
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-09T18:54:31.228035786Z
updated: 2026-10-09T18:54:31.228035786Z
---

## Context

- paths: `internal/lexer` (heredoc body scanning), `internal/parse` (refusals)
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-09 by the fuzz work on 01a121ac: a heredoc whose body never appears -- `print <<E;` at end of file, or a body with no terminator line -- parses with no refusal, though perl refuses it ("Can't find string terminator \"E\" anywhere before EOF"). Canon is stable, so the fuzzer does not flag it.

## Steps

- [ ] Measure perl on each form (no body, body without terminator, `<<~`, quoted and bare terminators); failing test first; refuse as perl does, on perl's line; `make test`; commit

## Acceptance criteria

- [ ] a heredoc with no terminator is refused (`go test ./internal/parse/ -run '^TestUnterminatedHeredocIsRefused$' -count=1 -v | grep -q '^--- PASS: TestUnterminatedHeredocIsRefused'`)
