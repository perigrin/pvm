---
title: Chalk reads its precedence table from CORE.pmt's @PRECEDENCE
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-08T00:51:38.259862798Z
updated: 2026-10-08T00:51:38.259862798Z
---

## Context

- paths: `~/dev/chalk/lib/Chalk/Grammar/Perl/PrecedenceTable.pm` (hand-kept, 15 rows, named unary at 4.5 and `not` at 12.5), `~/dev/chalk/lib/Chalk/Bootstrap/Semiring/Precedence.pm`; pvm's `internal/parse/declarations/CORE.pmt`
- commands: Chalk's own test harness (`./prove`); pvm `make test`

Goal (perigrin, 2026-10-08): every builtin and operator is defined with a signature in CORE.pmt that lets it be parsed, and the parser is derived from CORE.pmt. RFC 0001 "Precedence classes are the shared precedence table". Chalk's `PrecedenceTable` becomes a reading of the same rows: Chalk parses CORE.pmt's `@PRECEDENCE` (plain Perl) instead of keeping its own list. Its semiring logic is unchanged; only its data moves. The work lands in the Chalk repository; this issue tracks it for the milestone.

## Steps

- [ ] With perigrin, decide how Chalk locates CORE.pmt (vendored copy, path, or published module)
- [ ] In Chalk: a failing test that PrecedenceTable's rows equal CORE.pmt's; derive them; Chalk's suite green

## Acceptance criteria

- [ ] Chalk's precedence rows equal CORE.pmt's (`sh -c 'test -f ~/dev/chalk/lib/Chalk/Grammar/Perl/PrecedenceTable.pm && ! grep -q "ops => \[" ~/dev/chalk/lib/Chalk/Grammar/Perl/PrecedenceTable.pm'`)
