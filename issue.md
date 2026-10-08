---
title: Operators name their @PRECEDENCE row; the coined classes BITAND, BITOR, SHIFT and RANGE go
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a118fe-8d68-71a8-ab06-cc12d7a20589
created: 2026-10-08T00:51:38.164240057Z
updated: 2026-10-08T00:51:57.20901962Z
---

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/operators.go` (`coreOnlyClasses`), `docs/rfcs/0001-type-libraries.md` ("Fixity and precedence")
- commands: `make test` (100% green before every commit; never `--no-verify`); the ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet)$' -count=1`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1`
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl`; perlop's table: `~/dev/perl5/pod/perlop.pod` at the corpus pin (revision 94e5086608), "Operator Precedence and Associativity"

perigrin, 2026-10-08 (option 3): with the levels in CORE.pmt's `@PRECEDENCE`, a level XS::Parse::Infix does not class needs no coined class name. The core-only rows are named by a representative operator, e.g. `:infix(..)`, `:infix(&)`; only XS::Parse::Infix's real class names stay spellable, so a library .pmt can name exactly what XS::Parse::Infix can register. Spelling to confirm in the RFC.

## Steps

- [ ] Record the spelling in RFC 0001 "Fixity and precedence"; write failing test `TestCoreClassesAreXPIsOrRows`: every `:infix(...)` in CORE.pmt names an XS::Parse::Infix class or an operator in a `@PRECEDENCE` row, and no BITAND, BITOR, SHIFT or RANGE remains
- [ ] Respell the core-only lines; drop `coreOnlyClasses` in favour of "a library may not name a row XS::Parse::Infix cannot register at"; `make test`; commit

## Acceptance criteria

- [ ] no coined class names remain (`go test ./internal/parse/ -run '^TestCoreClassesAreXPIsOrRows$' -count=1 -v | grep -q '^--- PASS: TestCoreClassesAreXPIsOrRows'`)
