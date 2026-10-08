---
title: Chalk reads its precedence from CORE.pmt's operator relations
state: cancelled
urgency: normal
milestone: type-libraries
blocked_by:
- 01a118fe-8d68-71a8-ab06-cc12d7a20589
created: 2026-10-08T00:51:38.259862798Z
updated: 2026-10-08T01:32:14.555859883Z
transitions:
- state: cancelled
  actor: human:git-zhi
  timestamp: 2026-10-08T01:32:14.555859883Z
---

## Context

- paths: `~/dev/chalk/lib/Chalk/Grammar/Perl/PrecedenceTable.pm` (hand-kept, 15 rows, named unary at 4.5 and `not` at 12.5), `~/dev/chalk/lib/Chalk/Bootstrap/Semiring/Precedence.pm`; pvm's `internal/parse/declarations/CORE.pmt`
- commands: Chalk's own test harness (`./prove`); pvm `make test`

Goal (perigrin, 2026-10-08): every builtin and operator is defined with a signature in CORE.pmt that lets it be parsed, and the parser is derived from CORE.pmt. RFC 0001 "Precedence is a relation between operators": a partial order stated on each operator line, Raku-style -- `:tighter(OP)`, `:looser(OP)`, `:equiv(OP)`, `:assoc(left|right|nonassoc|chained|chain_na)` -- from which a total order is derived by topological sort. Chalk's `PrecedenceTable` becomes a reading of CORE.pmt: it derives the same total order from the operator relations (a topological sort Chalk can implement in Perl) instead of keeping its own list; the relations also let a library operator sit between levels without Chalk's fractional workarounds. Its semiring logic is unchanged; only its data moves. The work lands in the Chalk repository; this issue tracks it for the milestone.

## Steps

- [ ] With perigrin, decide how Chalk locates CORE.pmt (vendored copy, path, or published module)
- [ ] In Chalk: a failing test that PrecedenceTable's order equals the one derived from CORE.pmt; derive it; Chalk's suite green

## Acceptance criteria

- [ ] Chalk keeps no hand-written precedence rows (`sh -c 'test -f ~/dev/chalk/lib/Chalk/Grammar/Perl/PrecedenceTable.pm && ! grep -q "ops => \\[" ~/dev/chalk/lib/Chalk/Grammar/Perl/PrecedenceTable.pm'`)
