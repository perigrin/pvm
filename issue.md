---
title: map/grep's brace reads an empty {} as a block, though perl reads map {}, @a as a hash
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-09T08:17:49.841107029Z
updated: 2026-10-09T08:17:49.841107029Z
---

## Context

- paths: `internal/lexer/intuit.go` (`intuitCurly`, the map/grep/sort brace), `internal/parse/parse.go` (`braceOpensAnonHash`), spec §4.9.2's table (docs/specs/perl-parser/)
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found by 01a11f08-71bb (an empty `{}` at statement position is an anonymous hash), 2026-10-09. `intuitCurly` still answers block for an empty `{}` after `map`/`grep`/`sort`, while spec §4.9.2's table says `map {}, @a` is a hash; its comment says it applies the same rule as `braceOpensAnonHash`, which is now only mostly true. Measured on 5.42: `map {}, @a` compiles an anonymous hash (emptyavhv).

## Steps

- [ ] Measure perl on `map {}, @a`, `map {} @a`, `grep {}, @a`, `sort {} @a` and the `{ }` / comment forms
- [ ] Failing test first; make intuitCurly agree with perl (and correct its comment); ratchets toward perl only; `make test`; commit

## Acceptance criteria

- [ ] map/grep's empty brace is read as perl reads it (`go test ./internal/lexer/ -run '^TestIntuitCurlyEmptyBrace$' -count=1 -v | grep -q '^--- PASS: TestIntuitCurlyEmptyBrace'`)
