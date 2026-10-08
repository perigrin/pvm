---
title: "CORE.pmt declares ++ and -- as prefix and postfix"
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a118fe-8d68-71a8-ab06-cc12d7a20589
created: 2026-10-08T01:32:14.368080588Z
updated: 2026-10-08T09:50:02.518926699Z
---

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/operators.go` (`operatorArity`, `:postfix`)
- commands: `make test` (100% green before every commit; never `--no-verify`); the ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet)$' -count=1`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; perlop's table: `~/dev/perl5/pod/perlop.pod` at the corpus pin (94e5086608), "Operator Precedence and Associativity"

Goal (perigrin, 2026-10-08): every builtin and operator is defined with a signature in CORE.pmt that lets it be parsed, and the parser is derived from CORE.pmt. RFC 0001 "Precedence is a relation between operators": a partial order on each operator line -- `:tighter(OP)`, `:looser(OP)`, `:equiv(OP)`, `:assoc(...)` -- sorted topologically into a total order. perlop's `nonassoc ++ --` row has no CORE.pmt line. Each is `:prefix` and `:postfix`, on an aliased scalar (`\$`): perl refuses `++1` ("Can't modify constant item in preincrement"), which is the existing `ref_scalar_slot` refusal. Measure the result types (magic string increment: `"aa"` becomes `"ab"`; `--` has no string magic).

## Steps

- [ ] Remove `++`/`--` (both fixities) from `undeclaredOperators` in internal/parse/operators.go so their level comes from their CORE.pmt lines; the parser's binding powers must not change (TestParserPrecedenceFromCore)
- [ ] Measure on 5.42; write failing tests; declare `++` and `--` as `:prefix` and `:postfix` with their level; `make test`; commit

## Acceptance criteria

- [ ] ++ and -- are declared prefix and postfix (`go test ./internal/parse/ -run '^TestCoreDeclaresIncDec$' -count=1 -v | grep -q '^--- PASS: TestCoreDeclaresIncDec'`)
- [ ] ++ on a constant is refused as ref_scalar_slot (`go test ./internal/parse/ -run '^TestIncDecRefusesConstant$' -count=1 -v | grep -q '^--- PASS: TestIncDecRefusesConstant'`)
