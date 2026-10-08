---
title: isa may have its own precedence level in perl, above the relational operators
state: in-progress
urgency: normal
milestone: v0.1
created: 2026-10-06T04:34:49.185358746Z
updated: 2026-10-08T09:58:01.48381321Z
sessions:
- start_sha: 5f2311c758d827fdcb132ebd657a21508edaa80a
  end_sha: ""
  commits: 0
  started_at: 2026-10-08T09:58:01.48381321Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-08T09:58:01.48381321Z
---

## Context

- paths: `internal/parse/precedence.go`, `internal/parse/declarations/CORE.pmt` (`isa` declared `:infix(ISA)`, held to precedence.go's level 17 by TestCoreOperatorClassesMatchPrecedence)
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`, `perldoc perlop`

`precedence.go` puts `isa` at level 17 with the relational operators. perlop's precedence table lists `isa` as its own non-associative level, just above the relational operators (found by the alignment gate on 01a0fd9b-53a9-703f-82ab-eeb38765f7f1, not verified against perly.y). If perl binds `isa` tighter than `<`, an expression mixing them parses differently.

Now covered by 01a118fe-8dfe-7c0e-bcbe-06fd8374d1a6 (the parser's precedence derived from CORE.pmt's @PRECEDENCE), whose second criterion is TestIsaPrecedenceMatchesPerl; perlop's table confirms `isa` has its own row, tighter than the relational operators (2026-10-08). Close this when that issue lands.

## Steps

- [ ] Check perly.y's precedence declarations for `isa` against the relational operators, and measure with `perl -MO=Deparse,-p` on `use v5.36; $x isa Foo < 1` and similar
- [ ] If perl gives `isa` its own level: write a failing parse test, move `isa` in precedence.go, renumber CORE.pmt's ISA class to match; ratchets unmoved or improved; `make test`; commit
- [ ] If not: record the measurement in precedence.go's comment and close

## Acceptance criteria

- [ ] isa's precedence matches perl's (`go test ./internal/parse/ -run '^TestIsaPrecedenceMatchesPerl$' -count=1 -v | grep -q '^--- PASS: TestIsaPrecedenceMatchesPerl'`)
