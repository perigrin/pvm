---
title: A library's own infix operator parses at the level its relations give
state: done
urgency: normal
milestone: type-libraries
blocked_by:
- 01a118fe-8dfe-7c0e-bcbe-06fd8374d1a6
created: 2026-10-08T01:32:14.504542887Z
updated: 2026-10-08T13:46:53.333485705Z
sessions:
- start_sha: cd34146a827f624890247b5c2383db9e2261ebc9
  end_sha: 02fcb424cc451d0c3e6139f95f18eef225f04dc0
  commits: 2
  started_at: 2026-10-08T12:58:09.075211894Z
  ended_at: 2026-10-08T13:46:53.333485705Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-08T12:58:09.075211894Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-08T13:46:53.333485705Z
observed_paths:
- docs/rfcs/0001-type-libraries.md
- internal/lexer/scan.go
- internal/lexer/symbolop_test.go
- internal/parse/coreops_test.go
- internal/parse/declarations.go
- internal/parse/declarations/CORE.pmt
- internal/parse/expr.go
- internal/parse/library_pmt_test.go
- internal/parse/operators.go
- internal/parse/parse.go
- internal/parse/pmt_test.go
- internal/parse/precedence.go
- internal/parse/precrelations_test.go
- internal/parse/refscalar.go
- internal/parse/refscalar_test.go
- internal/parse/refusal.go
- internal/parse/term.go
- internal/parse/use.go
---

## Context

- paths: `internal/parse/declarations.go` (reads a library's operator lines, ~line 237), `internal/parse/expr.go` (`parseInfix`), `internal/parse/testdata/declarations/` (library fixtures)
- commands: `make test` (100% green before every commit; never `--no-verify`); the ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet)$' -count=1`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; perlop's table: `~/dev/perl5/pod/perlop.pod` at the corpus pin (94e5086608), "Operator Precedence and Associativity"

Goal (perigrin, 2026-10-08): every builtin and operator is defined with a signature in CORE.pmt that lets it be parsed, and the parser is derived from CORE.pmt. RFC 0001 "Precedence is a relation between operators": a partial order on each operator line -- `:tighter(OP)`, `:looser(OP)`, `:equiv(OP)`, `:assoc(...)` -- sorted topologically into a total order. The reason precedence is a partial order: a library must place its own infix operator between two existing levels. Today a library's operator line is read and checked, but `parseInfix` never sees it, so nothing a library declares parses.

Ruling (perigrin, 2026-10-08): "XS::Parse::Infix is an influence not a limitation". A library may place an operator at any level its relations give -- between two existing levels (`:tighter(+) :looser(*)`) or `:equiv` any operator's level, including levels XS::Parse::Infix classes no operator at (`&`, `|`, `<<`, `..`). So the refusal 01a118fe-8e34 added in `readLibraryDeclaration` ("`:equiv(&)` is a level XS::Parse::Infix registers no operator at", `coreClassed`/`coreLevelClassed`, `TestLibraryEquivUnclassedLevelRefused`, `TestShippedLibraryPmtRefusesUnclassedLevel`) goes, and RFC 0001's "Fixity and precedence" text saying a library may not `:equiv` into those levels is restated. XS::Parse::Infix's class names stay as one spelling of a level (`:infix(ADD)` is `:equiv(+)`); an unknown class name (e.g. the retired BITAND) is still refused (`TestLibraryCoinedClassRefused` stays). A library's bare `:infix` with no relation is a declaration error once libraries are ordered (it places nothing). The derivation API is `deriveBindingPowers(decls []operatorDecl)` (internal/parse/operators.go).

## Steps

- [ ] Write failing test `TestLibraryOperatorMayUseAnyLevel`: a fixture library's `sub ⊕ :infix :equiv(&) ...` is accepted and binds as `&`; remove the unclassed-level refusal and its tests; restate RFC 0001


- [ ] Write failing test `TestLibraryOperatorParses`: a fixture library declares `sub ⊕ :infix :tighter(+) :looser(*) :assoc(left) (Num $x, Num $y) Num;`; with it in scope, `1 + 2 ⊕ 3 * 4` parses as `1 + (2 ⊕ (3 * 4))`; a library operator spelled `:infix(ADD)` groups as `+` does
- [ ] Write failing test `TestLibraryOperatorErrors`: without the library in scope `⊕` is not an operator; a library relation forming a cycle with CORE.pmt's is a declaration error naming the operators
- [ ] Make the parser take a library's operators into its derived order; `make test`; commit

## Acceptance criteria

- [ ] a library operator may take any level (`go test ./internal/parse/ -run '^TestLibraryOperatorMayUseAnyLevel$' -count=1 -v | grep -q '^--- PASS: TestLibraryOperatorMayUseAnyLevel'`)
- [ ] a library operator between two levels parses there (`go test ./internal/parse/ -run '^TestLibraryOperatorParses$' -count=1 -v | grep -q '^--- PASS: TestLibraryOperatorParses'`)
- [ ] out-of-scope or cyclic library operators are refused (`go test ./internal/parse/ -run '^TestLibraryOperatorErrors$' -count=1 -v | grep -q '^--- PASS: TestLibraryOperatorErrors'`)
