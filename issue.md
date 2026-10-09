---
title: map/grep's brace reads an empty {} as a block, though perl reads map {}, @a as a hash
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-09T08:17:49.841107029Z
updated: 2026-10-09T09:25:33.130979422Z
sessions:
- start_sha: dd8abfe77cbc8881e7d098ea6f530d2fd51e4ed7
  end_sha: cfa7b6880d8cf587db4d4537ba1fa256261b1d64
  commits: 3
  started_at: 2026-10-09T08:35:32.152444656Z
  ended_at: 2026-10-09T09:25:33.130979422Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T08:35:32.152444656Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-09T09:25:33.130979422Z
observed_paths:
- docs/specs/perl-parser/00-findings.md
- internal/lexer/intuit.go
- internal/lexer/intuit_test.go
- internal/lexer/keyword.go
- internal/lexer/lexer.go
- internal/parse/barewordinvocant_test.go
- internal/parse/call.go
- internal/parse/lastindexblock_test.go
- internal/parse/mapbrace_test.go
- internal/parse/postfixhashslice_test.go
- internal/parse/sites.go
- internal/parse/smartmatchaggregate_test.go
- internal/parse/testdata/t1fidelity.ratchet
- internal/parse/use.go
- internal/parseoracle/ratchet_test.go
- internal/parseoracle/testdata/ratchet/baseline.txt
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
