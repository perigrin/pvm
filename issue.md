---
title: Let a library .pmt use every construct CORE.pmt uses
state: done
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-49b2-7ea9-bfa1-ba398baadc3f
- 01a0fd9b-4850-704d-8c14-c15add6194f2
- 01a0fd9b-5017-78d0-ae38-331cbe6a9866
- 01a0fd9b-5468-7355-a22e-1b16fa5ee354
blocks:
- 01a0fd9b-5700-7514-a143-2caed2d06d8b
created: 2026-10-02T17:13:31.032351231Z
updated: 2026-10-16T18:48:31.927062661Z
sessions:
- start_sha: 946db66bea66035c8d0fb8d36476efd4d05008f8
  end_sha: 946db66bea66035c8d0fb8d36476efd4d05008f8
  commits: 0
  started_at: 2026-10-06T15:14:47.126446535Z
  ended_at: 2026-10-06T15:14:47.293013336Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-06T15:14:47.126446535Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-06T15:14:47.293013336Z
---

## Prerequisites

- [ ] Parse the invocant colon in .pmt signatures and declare print, printf, say, exec and system with it
- [ ] Parse aliased parameters (\$x, \@a, \%h, \&c, \(@list)) and derive \$, \@, \% and \& from them
- [ ] Declare the no-prototype builtins and each's multi in CORE.pmt, and read :unary
- [ ] Declare the forking operators .. and x as multis in CORE.pmt (after `:context(...)` and operator declarations)

## Context

- paths: `internal/parse/module.go`, `internal/parse/decl.go`, `internal/parse/exports.go`, `internal/parse/testdata/declarations/` (new fixture), `internal/parse/library_pmt_test.go` (new), `internal/parse/export_test.go`, `docs/rfcs/0001-type-libraries.md`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: the declaration read path (`internal/parse/module.go:288`), `readModule` (`internal/parse/exports.go:108`), `export_test.go` for exposing an internal entry point to tests

RFC 0001 "One language for every `.pmt`" (perigrin, 2026-10-02): everything CORE.pmt can say, any library's `.pmt` can say -- typed signatures, container types, aliased parameters (`Scalar \$x`, `Array \@a`, `List[Str] \(@args)`), `multi`, `:context`, the invocant colon, `:unary`, operators (`:infix(CLASS)`, `:prefix`, `:postfix`; RFC "Operator declarations": a library declaring an XS::Parse::Infix operator uses the same spelling), declared syntax. CORE.pmt is the declaration file for the interpreter, not a dialect. The invocant colon is declarable for a library because a keyword plugin can build what a sub cannot. The syntax should stay something a second implementation (Chalk) can parse from the RFC alone. A user search path for `.pmt` files is open question 6 and out of scope: feed the fixture through the same read path the embedded declarations use, via `export_test.go`.

From the alignment gate on 01a0fd9b-46c4-7a8c-b8d6-498c5182a47f: declaration errors are kept by the resolver and read with `parse.DeclarationErrors(root)`, but nothing outside tests calls it, so a broken library .pmt is never shown to a user. When library .pmt files become user-writable here, psc must report them: add a step with a test that `psc check` on a file importing a module whose .pmt is in error prints the declaration error.

From the alignment gate on 01a0fd9b-5468-7355-a22e-1b16fa5ee354: `types.Ambiguity` is applied to sub multis but not to an operator's multi candidates (facts.operators), so two ambiguous operator candidates are not an error. A library .pmt that may declare operators needs the same check, with a failing test; and nothing stops a library .pmt from declaring one of the CORE-only coined classes (BITAND, BITOR, SHIFT, RANGE), which RFC 0001 reserves for CORE.pmt.

From the alignment gate on 01a0fd9b-5017-78d0-ae38-331cbe6a9866: `:unary` is not validated -- it is accepted on a signature with two required parameters, `(Scalar $x, Scalar $y)`, though a named unary takes at most one argument. A library .pmt that may write `:unary` needs that check, with a failing test.

## Steps

- [ ] Write a fixture library `.pmt` (`package My::Lib;`, `our @EXPORT`, a typed sub with a container-typed slurpy, a sub with aliased parameters (`Scalar \$x`, `Array \@a`, `Code \&c` and a `List[Str] \(@args)` list form), a sub with a required parameter defaulting to `die`, a `multi sub` pair, a `:context` multi, an invocant-colon sub, a `:unary` sub, an `:infix(ADD)` operator, a `:keyword(TERMEXPR)` syntax line) and failing test `TestLibraryPmtSpeaksCoreLanguage`: reading it through the declaration path yields every construct, the same as CORE.pmt would
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestLibraryPmtSpeaksCoreLanguage$' -count=1 -v`
- [ ] Remove any CORE-only gate the earlier issues left; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestLibraryPmtOperatorsAndUnary`: the fixture's `:infix(ADD)` operator and `:unary` sub are read with fixity, class and signature, as CORE.pmt's are
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestLibraryPmtAliasedAndContainerParams`: the fixture's aliased parameters derive `\$`, `\@` and `\&`, its `\(@args)` list form is recorded as an aliased list, its `= die` parameter is recorded as required (and derives no `;` before it), and its `List[Str] @args` keeps its container type
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestLibraryPmtImportsCarryTypes`: a source file that `use`s the fixture module sees its exported subs with their prototypes and signatures
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Re-mark the RFC heading "One language for every `.pmt`" `(*Implemented*, <sha>)`; commit

### Negative Scenarios

- a library `.pmt` obeys the same rules as CORE.pmt: an unknown type name, a bare-typed slurpy, a parameter after a `List`, `Array @a` with a flattening sigil, or ambiguous multi candidates in the fixture is an error naming the module (`go test ./internal/parse/ -run '^TestLibraryPmtDeclarationErrors$' -count=1 -v | grep -q '^--- PASS: TestLibraryPmtDeclarationErrors'`)
- `use My::Lib ()` imports neither subs nor declared syntax (RFC: `use M ()` calls no import and brings none) (`go test ./internal/parse/ -run '^TestLibraryPmtEmptyImportBringsNothing$' -count=1 -v | grep -q '^--- PASS: TestLibraryPmtEmptyImportBringsNothing'`)
- a sub declared in the fixture but not in `@EXPORT` is not visible after a bare `use` (`go test ./internal/parse/ -run '^TestLibraryPmtUnexportedSubNotImported$' -count=1 -v | grep -q '^--- PASS: TestLibraryPmtUnexportedSubNotImported'`)

## Acceptance criteria

- [ ] a library .pmt reads every construct CORE.pmt uses (`go test ./internal/parse/ -run '^TestLibraryTpmSpeaksCoreLanguage$' -count=1 -v | grep -q '^--- PASS: TestLibraryTpmSpeaksCoreLanguage'`)
- [ ] imports from a library .pmt carry their signatures (`go test ./internal/parse/ -run '^TestLibraryTpmImportsCarryTypes$' -count=1 -v | grep -q '^--- PASS: TestLibraryTpmImportsCarryTypes'`)
- [ ] the RFC section on one language is implemented (`grep -q '^### One language.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] a library .pmt reports the same declaration errors as CORE.pmt (`go test ./internal/parse/ -run '^TestLibraryTpmDeclarationErrors$' -count=1 -v | grep -q '^--- PASS: TestLibraryTpmDeclarationErrors'`)
- [ ] use with an empty list imports no subs and no syntax (`go test ./internal/parse/ -run '^TestLibraryTpmEmptyImportBringsNothing$' -count=1 -v | grep -q '^--- PASS: TestLibraryTpmEmptyImportBringsNothing'`)
- [ ] an unexported library sub is not imported (`go test ./internal/parse/ -run '^TestLibraryTpmUnexportedSubNotImported$' -count=1 -v | grep -q '^--- PASS: TestLibraryTpmUnexportedSubNotImported'`)
- [ ] a library .pmt declares operators and :unary subs as CORE.pmt does (`go test ./internal/parse/ -run '^TestLibraryTpmOperatorsAndUnary$' -count=1 -v | grep -q '^--- PASS: TestLibraryTpmOperatorsAndUnary'`)
- [ ] a library .pmt declares aliased parameters with Code \&c and the \(@list) form, a die default and container types (`go test ./internal/parse/ -run '^TestLibraryTpmAliasedAndContainerParams$' -count=1 -v | grep -q '^--- PASS: TestLibraryTpmAliasedAndContainerParams'`)
