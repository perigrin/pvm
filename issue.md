---
title: "The coined classes BITAND, BITOR, SHIFT and RANGE go: those levels are named by their operators"
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a118fe-8d68-71a8-ab06-cc12d7a20589
blocks:
- 01a118fe-8e5f-717a-b046-6020db95c724
created: 2026-10-08T00:51:38.164240057Z
updated: 2026-10-08T01:11:15.903471516Z
---

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/operators.go` (`coreOnlyClasses`), `docs/rfcs/0001-type-libraries.md` ("Fixity and precedence", "Precedence is a relation between operators")
- commands: `make test` (100% green before every commit; never `--no-verify`); the ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet)$' -count=1`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; perlop's table: `~/dev/perl5/pod/perlop.pod` at the corpus pin (94e5086608), "Operator Precedence and Associativity"

Goal (perigrin, 2026-10-08): every builtin and operator is defined with a signature in CORE.pmt that lets it be parsed, and the parser is derived from CORE.pmt. RFC 0001 "Precedence is a relation between operators": a partial order stated on each operator line, Raku-style -- `:tighter(OP)`, `:looser(OP)`, `:equiv(OP)`, `:assoc(left|right|nonassoc|chained|chain_na)` -- from which a total order is derived by topological sort. Levels XS::Parse::Infix does not class need no name of their own: `&`, `|`/`^`, `<<`/`>>` and `..`/`...` are related by `:tighter`/`:looser`/`:equiv` like any operator. Only XS::Parse::Infix's real class names stay spellable in `:infix(CLASS)`, so a library names exactly what XS::Parse::Infix can register.

## Steps

- [ ] Write failing test `TestCoreClassesAreXPIs`: every `:infix(CLASS)` in CORE.pmt names an XS::Parse::Infix class, and no BITAND, BITOR, SHIFT or RANGE remains
- [ ] Respell those lines with relations; drop `coreOnlyClasses` in favour of "a library may not place an operator at a level XS::Parse::Infix cannot register at" if XS::Parse::Infix's limits still apply; restate RFC 0001 "Fixity and precedence"; `make test`; commit

## Acceptance criteria

- [ ] no coined class names remain (`go test ./internal/parse/ -run '^TestCoreClassesAreXPIs$' -count=1 -v | grep -q '^--- PASS: TestCoreClassesAreXPIs'`)
