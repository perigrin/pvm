---
title: Let a library .pmt use every construct CORE.pmt uses
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-49b2-7ea9-bfa1-ba398baadc3f
created: 2026-10-02T17:13:31.032351231Z
updated: 2026-10-02T17:14:24.842069226Z
---

## Prerequisites

- [ ] Parse the invocant colon in .pmt signatures and declare print, printf, say, exec and system with it
- [ ] Read :context(...) and select a multi candidate by calling context

## Context

- paths: `internal/parse/module.go`, `internal/parse/exports.go`, `internal/parse/testdata/declarations/` (new fixture), `internal/parse/library_pmt_test.go` (new), `internal/parse/export_test.go`, `docs/rfcs/0001-type-libraries.md`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: the declaration read path (`internal/parse/module.go:288`), `readModule` (`internal/parse/exports.go:108`), `export_test.go` for exposing an internal entry point to tests

RFC 0001 "One language for every `.pmt`" (perigrin, 2026-10-02): everything CORE.pmt can say, any library's `.pmt` can say -- typed signatures, `multi`, `:context`, the invocant colon, declared syntax. CORE.pmt is the declaration file for the interpreter, not a dialect. The invocant colon is declarable for a library because a keyword plugin can build what a sub cannot. The syntax should stay something a second implementation (Chalk) can parse from the RFC alone. A user search path for `.pmt` files is open question 6 and out of scope: feed the fixture through the same read path the embedded declarations use, via `export_test.go`.

## Steps

- [ ] Write a fixture library `.pmt` (`package My::Lib;`, `our @EXPORT`, a typed sub, a `multi sub` pair, a `:context` multi, an invocant-colon sub, a `:keyword(TERMEXPR)` syntax line) and failing test `TestLibraryPmtSpeaksCoreLanguage`: reading it through the declaration path yields every construct, the same as CORE.pmt would
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestLibraryPmtSpeaksCoreLanguage$' -count=1 -v`
- [ ] Remove any CORE-only gate the earlier issues left; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestLibraryPmtImportsCarryTypes`: a source file that `use`s the fixture module sees its exported subs with their prototypes and signatures
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Re-mark the RFC heading "One language for every `.pmt`" `(*Implemented*, <sha>)`; commit

## Acceptance criteria

- [ ] a library .pmt reads every construct CORE.pmt uses (`go test ./internal/parse/ -run '^TestLibraryPmtSpeaksCoreLanguage$' -count=1 -v | grep -q '^--- PASS: TestLibraryPmtSpeaksCoreLanguage'`)
- [ ] imports from a library .pmt carry their signatures (`go test ./internal/parse/ -run '^TestLibraryPmtImportsCarryTypes$' -count=1 -v | grep -q '^--- PASS: TestLibraryPmtImportsCarryTypes'`)
- [ ] the RFC section on one language is implemented (`grep -q '^### One language.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
