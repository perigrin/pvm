---
title: The parser's operator precedence and associativity come from CORE.pmt's @PRECEDENCE
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a118fe-8d68-71a8-ab06-cc12d7a20589
created: 2026-10-08T00:51:38.110268863Z
updated: 2026-10-08T00:51:56.08379807Z
---

## Context

- paths: `internal/parse/precedence.go`, `internal/parse/operators.go` (`operatorClasses`, `coreOnlyClasses`), the Pratt binding powers in `internal/parse`
- commands: `make test` (100% green before every commit; never `--no-verify`); the ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet)$' -count=1`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1`
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl`; perlop's table: `~/dev/perl5/pod/perlop.pod` at the corpus pin (revision 94e5086608), "Operator Precedence and Associativity"

Goal (perigrin, 2026-10-08): every builtin and operator is defined with a signature in CORE.pmt that lets it be parsed, and the parser is derived from CORE.pmt. RFC 0001 "Precedence classes are the shared precedence table". With `@PRECEDENCE` in CORE.pmt, the parser's binding powers and associativity are derived from it, as its keyword shapes now are (55cb). The Go class map and any hand-kept level numbers go. This also fixes `isa`, which precedence.go places with the relational operators while perlop gives it its own tighter row (01a10f7e).

## Steps

- [ ] Write failing test `TestParserPrecedenceFromCore`: the binding power and associativity the parser uses for each infix operator equal what `@PRECEDENCE` gives (golden = today's table, with `isa` corrected per perlop)
- [ ] Derive them; resolve every disagreement by measuring perl (`-MO=Deparse,-p`); delete the hand-kept table; ratchets must not regress -- `isa` may improve, checked against perl; `make test`; commit

## Acceptance criteria

- [ ] the parser's precedence is CORE.pmt's (`go test ./internal/parse/ -run '^TestParserPrecedenceFromCore$' -count=1 -v | grep -q '^--- PASS: TestParserPrecedenceFromCore'`)
- [ ] isa binds tighter than the relational operators, as perlop says (`go test ./internal/parse/ -run '^TestIsaPrecedenceMatchesPerl$' -count=1 -v | grep -q '^--- PASS: TestIsaPrecedenceMatchesPerl'`)
