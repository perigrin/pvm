---
title: The parser's operator precedence and associativity come from CORE.pmt's relations
state: done
urgency: normal
milestone: type-libraries
blocked_by:
- 01a118fe-8d68-71a8-ab06-cc12d7a20589
- 01a118fe-8e34-7b3e-af7b-e426488966d0
blocks:
- 01a11923-bb28-76fc-bce5-078b86077ca4
created: 2026-10-08T00:51:38.110268863Z
updated: 2026-10-08T09:57:17.230117872Z
sessions:
- start_sha: b2750dd0677c83a05291e2da5bc3246872c9edbd
  end_sha: 5f2311c758d827fdcb132ebd657a21508edaa80a
  commits: 1
  started_at: 2026-10-08T09:14:10.203628911Z
  ended_at: 2026-10-08T09:57:17.230117872Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-08T09:14:10.203628911Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-08T09:57:17.230117872Z
observed_paths:
- docs/rfcs/0001-type-libraries.md
- internal/parse/call.go
- internal/parse/canon.go
- internal/parse/chain.go
- internal/parse/coreops_test.go
- internal/parse/decl.go
- internal/parse/declarations.go
- internal/parse/declarations/CORE.pmt
- internal/parse/expr.go
- internal/parse/expr_test.go
- internal/parse/operators.go
- internal/parse/parse.go
- internal/parse/precedence.go
- internal/parse/precrelations_test.go
---

## Context

- paths: `internal/parse/precedence.go`, `internal/parse/operators.go` (`operatorClasses`, `coreOnlyClasses`), the Pratt binding powers in `internal/parse`
- commands: `make test` (100% green before every commit; never `--no-verify`); the ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet)$' -count=1`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; perlop's table: `~/dev/perl5/pod/perlop.pod` at the corpus pin (94e5086608), "Operator Precedence and Associativity"

Goal (perigrin, 2026-10-08): every builtin and operator is defined with a signature in CORE.pmt that lets it be parsed, and the parser is derived from CORE.pmt. RFC 0001 "Precedence is a relation between operators": a partial order stated on each operator line, Raku-style -- `:tighter(OP)`, `:looser(OP)`, `:equiv(OP)`, `:assoc(left|right|nonassoc|chained|chain_na)` -- from which a total order is derived by topological sort. With CORE.pmt's relations read and sorted (the previous issue), the parser's binding powers and associativity are derived from them, as its keyword shapes now are. The hand-kept table goes. This fixes `isa`, which precedence.go places with the relational operators while perlop gives it its own tighter row (01a10f7e).

## Steps

- [ ] Write failing test `TestParserPrecedenceFromCore`: the binding power and associativity the parser uses for each operator equal the order derived from CORE.pmt
- [ ] Write failing test `TestIsaPrecedenceMatchesPerl`: `isa` binds tighter than `<` (measure with `-MO=Deparse,-p`)
- [ ] Derive them; resolve every disagreement by measuring perl; delete the hand-kept table; ratchets must not regress (`isa` may improve, checked against perl); `make test`; commit

## Acceptance criteria

- [ ] the parser's precedence is CORE.pmt's (`go test ./internal/parse/ -run '^TestParserPrecedenceFromCore$' -count=1 -v | grep -q '^--- PASS: TestParserPrecedenceFromCore'`)
- [ ] isa binds tighter than the relational operators, as perlop says (`go test ./internal/parse/ -run '^TestIsaPrecedenceMatchesPerl$' -count=1 -v | grep -q '^--- PASS: TestIsaPrecedenceMatchesPerl'`)
