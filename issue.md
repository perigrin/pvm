---
title: "CORE.pmt declares perlop's remaining infix operators: ^^, string-bitwise, compound assignment, , and =>"
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-08T00:51:38.207321952Z
updated: 2026-10-08T01:32:22.343906634Z
---

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/precedence.go` (the parser already lexes `^^`, `&.`, `|.`, `~.`, `+=`), `docs/rfcs/0001-type-libraries.md`
- commands: `make test` (100% green before every commit; never `--no-verify`); the ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet)$' -count=1`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; perlop's table: `~/dev/perl5/pod/perlop.pod` at the corpus pin (94e5086608), "Operator Precedence and Associativity"

Goal (perigrin, 2026-10-08): every builtin and operator is defined with a signature in CORE.pmt that lets it be parsed, and the parser is derived from CORE.pmt. RFC 0001 "Precedence is a relation between operators": a partial order on each operator line -- `:tighter(OP)`, `:looser(OP)`, `:equiv(OP)`, `:assoc(...)` -- sorted topologically into a total order. CORE.pmt declares the plain and forking infix operators, but perlop's table names more: `^^` (5.40 logical xor), the string-bitwise `&.` `|.` `^.` and prefix `~.`, the compound assignments `+= -= *= /= .= x= **= %= <<= >>= &= |= ^= &&= ||= //= &.= |.= ^.=`, and `,` `=>` (perlop's `left , =>` row). Declare them with relations (`:equiv(||)` etc.) and, once 01a11920-411a lands, without `:context`.

`,` forks by the result the call site asks for (measured: `my $x = (4, 5)` is 5 with a "Useless use" warning; `(1, (2, 3))` is 3 elements): `multi sub , :infix :looser(=) :assoc(left) (Scalar $l, Scalar $r) Scalar;` and `multi sub , ... (List @l, List @r) List;`. Two rulings needed first (perigrin): the list candidate breaks the final-List rule; `=>` also quotes a bareword left operand, a lexer effect no parameter type states (proposed: `:autoquote`).

## Steps

- [ ] Measure each operator's operand and result types on 5.42
- [ ] Declare `^^`, the string-bitwise operators and the compound assignments, failing tests first; `make test`; commit
- [ ] Bring the final-List exemption and `=>`'s autoquote spelling to perigrin; record in RFC 0001; declare `,` and `=>`, failing tests first; `make test`; commit

## Acceptance criteria

- [ ] every infix operator in perlop's table, and `~.`, has a CORE.pmt line (`go test ./internal/parse/ -run '^TestCoreDeclaresEveryPerlopInfixOperator$' -count=1 -v | grep -q '^--- PASS: TestCoreDeclaresEveryPerlopInfixOperator'`)
- [ ] CORE.pmt declares no operator perlop's table does not name (`go test ./internal/parse/ -run '^TestCoreDeclaresNoUnknownOperator$' -count=1 -v | grep -q '^--- PASS: TestCoreDeclaresNoUnknownOperator'`)
