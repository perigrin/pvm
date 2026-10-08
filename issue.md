---
title: CORE.pmt declares every operator perlop names
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-08T00:51:38.207321952Z
updated: 2026-10-08T00:51:38.207321952Z
---

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `docs/rfcs/0001-type-libraries.md`
- commands: `make test` (100% green before every commit; never `--no-verify`); the ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet)$' -count=1`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1`
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl`; perlop's table: `~/dev/perl5/pod/perlop.pod` at the corpus pin (revision 94e5086608), "Operator Precedence and Associativity"

Goal (perigrin, 2026-10-08): every builtin and operator is defined with a signature in CORE.pmt that lets it be parsed, and the parser is derived from CORE.pmt. RFC 0001 "Precedence classes are the shared precedence table". CORE.pmt declares the plain and forking infix and prefix operators, but perlop's table names more: `^^` (5.40 logical xor), the string-bitwise `&.` `|.` `^.` `~.`, the compound assignments `+= -= *= /= .= x= **= %= <<= >>= &= |= ^= &&= ||= //= &.= |.= ^.=`, `++` `--`, `,` `=>`, `?:`, `->`. The first three groups are ordinary `:infix`/`:prefix` lines in existing rows; `++`/`--` need `:postfix` and `:prefix`; `,`/`=>`, `?:` and `->` have shapes no current declaration form states -- decide each (perigrin) before declaring.

## Steps

- [ ] Measure each operator's operand and result types on 5.42; declare `^^`, the string-bitwise operators and the compound assignments, failing tests first
- [ ] Declare `++`/`--` as `:prefix` and `:postfix`, failing tests first
- [ ] Bring `,`/`=>`, `?:` and `->` to perigrin with a proposed form for each; declare once ruled
- [ ] `make test`; commit after each group

## Acceptance criteria

- [ ] every infix and prefix operator in perlop's table has a CORE.pmt line (`go test ./internal/parse/ -run '^TestCoreDeclaresEveryPerlopOperator$' -count=1 -v | grep -q '^--- PASS: TestCoreDeclaresEveryPerlopOperator'`)
