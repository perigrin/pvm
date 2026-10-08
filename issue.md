---
title: "CORE.pmt declares perlop's remaining infix operators: ^^, string-bitwise, compound assignment, , and =>"
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a11920-411a-7cea-b78f-c1568aba6d25
created: 2026-10-08T00:51:38.207321952Z
updated: 2026-10-08T13:17:34.493128133Z
---

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/precedence.go` (the parser already lexes `^^`, `&.`, `|.`, `~.`, `+=`), `docs/rfcs/0001-type-libraries.md`
- commands: `make test` (100% green before every commit; never `--no-verify`); the ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet)$' -count=1`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; perlop's table: `~/dev/perl5/pod/perlop.pod` at the corpus pin (94e5086608), "Operator Precedence and Associativity"

Goal (perigrin, 2026-10-08): every builtin and operator is defined with a signature in CORE.pmt that lets it be parsed, and the parser is derived from CORE.pmt. RFC 0001 "Precedence is a relation between operators": a partial order on each operator line -- `:tighter(OP)`, `:looser(OP)`, `:equiv(OP)`, `:assoc(...)` -- sorted topologically into a total order. CORE.pmt declares the plain and forking infix operators, but perlop's table names more: `^^` (5.40 logical xor), the string-bitwise `&.` `|.` `^.` and prefix `~.`, the compound assignments `+= -= *= /= .= x= **= %= <<= >>= &= |= ^= &&= ||= //= &.= |.= ^.=`, and `,` `=>` (perlop's `left , =>` row). Declare them with relations (`:equiv(||)` etc.) and, once 01a11920-411a lands, without `:context`.

Smartmatch `~~` is also undeclared, and perlop's table lists it in the `==` row (non-chaining, `cmpClasses`); measured on 5.42, `1 ~~ 1` is true. Declare it too (measure its operand and result types; note any deprecation warning 5.42 emits).

Rulings (perigrin, 2026-10-08):

- **`,` is the one exception to the final-List rule.** "The rule was put in for list flattening ... there is absolutely no way to distinguish (@a, @b) from (@a) in Perl currently": the comma IS the flattening, so the rule about flattened arguments does not bind it. Measured on 5.42: `(@a, 3)` with `@a = (1, 2)` is 3 elements (the left operand flattens too); `my $x = (4, 5)` is 5 with a "Useless use" warning; `(1, (2, 3))` is 3 elements. So `,` is `multi sub , :infix ... (List @l, List @r) List;` (list context: append) and `(Scalar $l, Scalar $r) Scalar` (scalar context: evaluate left, return right). The exception is `,` (and `=>`, the same operator) only -- not operators in general: the reader refuses a non-final List parameter everywhere else, and a test says so.
- **`:autoquote` is a parameter attribute**, not the operator's: `=>` quotes the bareword on its left, so `sub => :infix :equiv(,) (Str $lhs :autoquote, ...)`. The lexer reads it from CORE.pmt to know which operand to quote.
- `,`'s relations: between rightward list operators and `=` (perlop's `left , =>` row); replace `shapeRelations`' list-operator `:looser(=)` with `:looser(,)` once `,` is declared, and remove `,`/`=>` from `undeclaredOperators`.

Context is no longer an attribute: a multi selects by return type in the calling context (01a11920-411a).

## Steps

- [ ] Write failing test `TestUndeclaredOperatorsAreOnlyTernaryArrowAndSubscripts`: after this issue (and D2, 01a11923-baa0, which removes `++`/`--`), `undeclaredOperators` places only `?`, `->`, `(`, `[`, `{`, and the compound assignments are CORE.pmt lines
- [ ] Measure each operator's operand and result types on 5.42
- [ ] Declare `^^`, the string-bitwise operators and the compound assignments, failing tests first; `make test`; commit
- [ ] Remove each operator this issue declares from `undeclaredOperators` in internal/parse/operators.go (`^^`, `&.`, `|.`, `^.`, `~.`, `~~`, `,`, `=>` and the compound assignments (`compoundAssignments()`)), so its level comes from its CORE.pmt line; the parser's binding powers must not change (TestParserPrecedenceFromCore)
- [ ] Record both rulings in RFC 0001; write failing tests `TestCommaIsTheFinalListException` (`,`/`=>` may take `List @l, List @r`; any other operator or sub with a non-final List parameter is still refused) and `TestFatCommaAutoquotes` (`foo => 1` quotes `foo` because `$lhs` is `:autoquote`; the lexer reads it from CORE.pmt); declare `,` and `=>`; `make test`; commit

## Acceptance criteria

- [ ] `,` alone breaks the final-List rule (`go test ./internal/parse/ -run '^TestCommaIsTheFinalListException$' -count=1 -v | grep -q '^--- PASS: TestCommaIsTheFinalListException'`)
- [ ] `=>` quotes through its parameter's :autoquote (`go test ./internal/parse/ -run '^TestFatCommaAutoquotes$' -count=1 -v | grep -q '^--- PASS: TestFatCommaAutoquotes'`)
- [ ] undeclaredOperators keeps only `?`, `->`, `(`, `[` and `{`, and compoundAssignments is gone (`go test ./internal/parse/ -run '^TestUndeclaredOperatorsAreOnlyTernaryArrowAndSubscripts$' -count=1 -v | grep -q '^--- PASS: TestUndeclaredOperatorsAreOnlyTernaryArrowAndSubscripts'`)
- [ ] every infix operator in perlop's table, and `~.`, has a CORE.pmt line (`go test ./internal/parse/ -run '^TestCoreDeclaresEveryPerlopInfixOperator$' -count=1 -v | grep -q '^--- PASS: TestCoreDeclaresEveryPerlopInfixOperator'`)
- [ ] CORE.pmt declares no operator perlop's table does not name (`go test ./internal/parse/ -run '^TestCoreDeclaresNoUnknownOperator$' -count=1 -v | grep -q '^--- PASS: TestCoreDeclaresNoUnknownOperator'`)
