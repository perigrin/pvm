---
title: CORE.pmt states each operator's precedence as a relation to another, with perlop's associativity
state: pending
urgency: normal
milestone: type-libraries
blocks:
- 01a118fe-8dfe-7c0e-bcbe-06fd8374d1a6
- 01a118fe-8e34-7b3e-af7b-e426488966d0
- 01a118fe-8e93-74b8-b0d3-5d447b2613eb
created: 2026-10-08T00:51:37.959731974Z
updated: 2026-10-08T01:11:09.615037995Z
---

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/decl.go` (attribute reading), `internal/parse/operators.go` (`operatorClasses`), `docs/rfcs/0001-type-libraries.md`
- commands: `make test` (100% green before every commit; never `--no-verify`); the ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet)$' -count=1`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; perlop's table: `~/dev/perl5/pod/perlop.pod` at the corpus pin (94e5086608), "Operator Precedence and Associativity"

Goal (perigrin, 2026-10-08): every builtin and operator is defined with a signature in CORE.pmt that lets it be parsed, and the parser is derived from CORE.pmt. RFC 0001 "Precedence is a relation between operators": a partial order stated on each operator line, Raku-style -- `:tighter(OP)`, `:looser(OP)`, `:equiv(OP)`, `:assoc(left|right|nonassoc|chained|chain_na)` -- from which a total order is derived by topological sort. Today an operator line names an XS::Parse::Infix class (`:infix(MUL)`) whose level lives only in Go (`operatorClasses`, perly.y numbers), and no line carries associativity. This issue makes CORE.pmt state the relations and associativity, reads them, and derives the total order; switching the parser to it is the next issue.

## Steps

- [ ] Write failing test `TestCorePrecedenceRelationsAreParsed`: the .pmt reader reads `:tighter`, `:looser`, `:equiv` and `:assoc` on an operator line; `:infix(CLASS)` reads as `:equiv` of that class's anchor operator
- [ ] Write failing test `TestCoreDerivedPrecedenceIsPerlops`: the total order and associativity derived from CORE.pmt's relations (topological sort) equal perlop's table at the pin, row for row
- [ ] Write failing tests for the negative cases: a cycle (`:tighter(*)` on `+` while `*` is `:tighter(+)`) and an operator related to nothing are declaration errors naming the operators
- [ ] Add the relations to CORE.pmt's operator lines; implement the reading and the sort; `make test`; commit

## Acceptance criteria

- [ ] precedence relations are read from a .pmt (`go test ./internal/parse/ -run '^TestCorePrecedenceRelationsAreParsed$' -count=1 -v | grep -q '^--- PASS: TestCorePrecedenceRelationsAreParsed'`)
- [ ] CORE.pmt's relations derive perlop's table (`go test ./internal/parse/ -run '^TestCoreDerivedPrecedenceIsPerlops$' -count=1 -v | grep -q '^--- PASS: TestCoreDerivedPrecedenceIsPerlops'`)
- [ ] a cycle or an unrelated operator is a declaration error (`go test ./internal/parse/ -run '^TestPrecedenceRelationErrors$' -count=1 -v | grep -q '^--- PASS: TestPrecedenceRelationErrors'`)
