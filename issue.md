---
title: A library's own infix operator parses at the level its relations give
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a118fe-8dfe-7c0e-bcbe-06fd8374d1a6
created: 2026-10-08T01:32:14.504542887Z
updated: 2026-10-08T01:32:28.07820372Z
---

## Context

- paths: `internal/parse/declarations.go` (reads a library's operator lines, ~line 237), `internal/parse/expr.go` (`parseInfix`), `internal/parse/testdata/declarations/` (library fixtures)
- commands: `make test` (100% green before every commit; never `--no-verify`); the ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet)$' -count=1`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; perlop's table: `~/dev/perl5/pod/perlop.pod` at the corpus pin (94e5086608), "Operator Precedence and Associativity"

Goal (perigrin, 2026-10-08): every builtin and operator is defined with a signature in CORE.pmt that lets it be parsed, and the parser is derived from CORE.pmt. RFC 0001 "Precedence is a relation between operators": a partial order on each operator line -- `:tighter(OP)`, `:looser(OP)`, `:equiv(OP)`, `:assoc(...)` -- sorted topologically into a total order. The reason precedence is a partial order: a library must place its own infix operator between two existing levels. Today a library's operator line is read and checked, but `parseInfix` never sees it, so nothing a library declares parses.

## Steps

- [ ] Write failing test `TestLibraryOperatorParses`: a fixture library declares `sub ⊕ :infix :tighter(+) :looser(*) :assoc(left) (Num $x, Num $y) Num;`; with it in scope, `1 + 2 ⊕ 3 * 4` parses as `1 + (2 ⊕ (3 * 4))`; a library operator spelled `:infix(ADD)` groups as `+` does
- [ ] Write failing test `TestLibraryOperatorErrors`: without the library in scope `⊕` is not an operator; a library relation forming a cycle with CORE.pmt's is a declaration error naming the operators
- [ ] Make the parser take a library's operators into its derived order; `make test`; commit

## Acceptance criteria

- [ ] a library operator between two levels parses there (`go test ./internal/parse/ -run '^TestLibraryOperatorParses$' -count=1 -v | grep -q '^--- PASS: TestLibraryOperatorParses'`)
- [ ] out-of-scope or cyclic library operators are refused (`go test ./internal/parse/ -run '^TestLibraryOperatorErrors$' -count=1 -v | grep -q '^--- PASS: TestLibraryOperatorErrors'`)
