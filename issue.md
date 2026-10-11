---
title: "&2 and &== 2 are accepted by perl, which names the sub by one character; internal/parse refuses or misreads them"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-11T01:56:35.187635694Z
updated: 2026-10-11T01:56:35.187635694Z
---

## Context

- paths: `internal/parse/term.go` (`ampTakesPunct`, the `&` name rule from 01a0ef41)
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-11 by 01a0ef41. perl's tokenizer takes the character after a nameless `&` as the sub's name, and that includes a digit and the first character of a multi-character operator. Measured on 5.42 with Deparse: `&2;` and `& 2;` compile as `&2;` (a call to a sub named 2); `&== 2;` compiles as `&= = 2;` (a call to `=`, then assignment). internal/parse refuses `&2`/`& 2` and does not split `&==`/`&=>`.

## Steps

- [ ] Measure each with Deparse; failing test first; read them as perl does; `make test`; commit

## Acceptance criteria

- [ ] `&` followed by a digit or a multi-character operator is read as perl reads it (`go test ./internal/parse/ -run '^TestAmpersandTakesOneCharacterName$' -count=1 -v | grep -q '^--- PASS: TestAmpersandTakesOneCharacterName'`)
