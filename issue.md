---
title: CORE.pmt states the precedence table as @PRECEDENCE, with perlop's associativity
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-08T00:51:37.959731974Z
updated: 2026-10-08T00:51:37.959731974Z
---

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/operators.go` (`operatorClasses`), `internal/parse/precedence.go`, `docs/rfcs/0001-type-libraries.md`
- commands: `make test` (100% green before every commit; never `--no-verify`); the ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet)$' -count=1`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1`
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl`; perlop's table: `~/dev/perl5/pod/perlop.pod` at the corpus pin (revision 94e5086608), "Operator Precedence and Associativity"

Goal (perigrin, 2026-10-08): every builtin and operator is defined with a signature in CORE.pmt that lets it be parsed, and the parser is derived from CORE.pmt. RFC 0001 "Precedence classes are the shared precedence table". CORE.pmt gains `our @PRECEDENCE`, the rows tightest first with perlop's associativity (right, left, nonassoc, chained, chain/na), each naming the classes at that level. Today the levels live only in Go (`operatorClasses`, perly.y numbers), which a second implementation cannot read, and no class carries associativity.

## Steps

- [ ] Write failing test `TestCorePrecedenceTableIsPerlops`: CORE.pmt's `@PRECEDENCE`, read by the .pmt reader, has perlop's rows in perlop's order and associativity (parse perlop.pod's table at the pin), and every `:infix` class in CORE.pmt names a row
- [ ] Add `@PRECEDENCE` to CORE.pmt and the reader that exposes it (as `coreShapes` is exposed); run test to verify pass; `make test`; commit

## Acceptance criteria

- [ ] CORE.pmt's @PRECEDENCE is perlop's table (`go test ./internal/parse/ -run '^TestCorePrecedenceTableIsPerlops$' -count=1 -v | grep -q '^--- PASS: TestCorePrecedenceTableIsPerlops'`)
