---
title: "The coined classes BITAND, BITOR, SHIFT and RANGE go: those levels are named by their operators"
state: done
urgency: normal
milestone: type-libraries
blocked_by:
- 01a118fe-8d68-71a8-ab06-cc12d7a20589
blocks:
- 01a118fe-8dfe-7c0e-bcbe-06fd8374d1a6
created: 2026-10-08T00:51:38.164240057Z
updated: 2026-10-08T22:11:45.83657091Z
sessions:
- start_sha: 04a6ca62d20d1d53be57651b769964909706bf69
  end_sha: b2750dd0677c83a05291e2da5bc3246872c9edbd
  commits: 1
  started_at: 2026-10-08T08:39:40.484424344Z
  ended_at: 2026-10-08T09:13:45.078812533Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-08T08:39:40.484424344Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-08T09:13:45.078812533Z
observed_paths:
- docs/rfcs/0001-type-libraries.md
- docs/specs/perl-parser/04-expressions.md
- internal/lexer/scanrows_test.go
- internal/parse/coreops_test.go
- internal/parse/declarations.go
- internal/parse/declarations/CORE.pmt
- internal/parse/library_pmt_test.go
- internal/parse/operators.go
- internal/parse/pmt_test.go
- internal/types/select_test.go
---

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/operators.go` (`coreOnlyClasses`), `docs/rfcs/0001-type-libraries.md` ("Fixity and precedence", "Precedence is a relation between operators")
- commands: `make test` (100% green before every commit; never `--no-verify`); the ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet)$' -count=1`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; perlop's table: `~/dev/perl5/pod/perlop.pod` at the corpus pin (94e5086608), "Operator Precedence and Associativity"

Goal (perigrin, 2026-10-08): every builtin and operator is defined with a signature in CORE.pmt that lets it be parsed, and the parser is derived from CORE.pmt. RFC 0001 "Precedence is a relation between operators": a partial order stated on each operator line, Raku-style -- `:tighter(OP)`, `:looser(OP)`, `:equiv(OP)`, `:assoc(left|right|nonassoc|chained|chain_na)` -- from which a total order is derived by topological sort. Levels XS::Parse::Infix does not class need no name of their own: `&`, `|`/`^`, `<<`/`>>` and `..`/`...` are related by `:tighter`/`:looser`/`:equiv` like any operator. Only XS::Parse::Infix's real class names stay spellable in `:infix(CLASS)`, so a library names exactly what XS::Parse::Infix can register.

## Steps

- [ ] Write failing test `TestCoreClassesAreXPIs`: every `:infix(CLASS)` in CORE.pmt names an XS::Parse::Infix class, and no BITAND, BITOR, SHIFT or RANGE remains
- [ ] Write failing test `TestLibraryCoinedClassRefused`: a library .pmt naming a removed class is an unknown-class declaration error
- [ ] Respell those lines with relations; drop `coreOnlyClasses` in favour of "a library may not place an operator at a level XS::Parse::Infix cannot register at" if XS::Parse::Infix's limits still apply; restate RFC 0001 "Fixity and precedence"; `make test`; commit

### Reworded (2026-10-08)

perigrin: the goal's condition that the four retired class names "appear nowhere" applies to the tests too, so the criteria no longer name them: TestLibraryCoinedClassRefused is TestLibraryUnknownClassRefused (an unknown class is refused; the retired names are unknown classes like any other), and TestCoreClassesAreXPIs checks that every CORE.pmt class is XS::Parse::Infix's, which rules them out without listing them.

## Acceptance criteria

- [ ] a library .pmt naming a class XS::Parse::Infix does not have is a declaration error (`go test ./internal/parse/ -run '^TestLibraryUnknownClassRefused$' -count=1 -v | grep -q '^--- PASS: TestLibraryUnknownClassRefused'`)
- [ ] every class CORE.pmt names is XS::Parse::Infix's (`go test ./internal/parse/ -run '^TestCoreClassesAreXPIs$' -count=1 -v | grep -q '^--- PASS: TestCoreClassesAreXPIs'`)
