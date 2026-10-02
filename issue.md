---
title: Declare the plain no-prototype builtins in CORE.pmt and read :unary
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-4850-704d-8c14-c15add6194f2
created: 2026-10-02T17:13:29.367336401Z
updated: 2026-10-02T17:14:10.563361307Z
---

## Prerequisites

- [ ] Read the :lvalue parameter attribute and derive the \$ prototype from it
- [ ] Parse multi sub declarations and select a candidate by arity and argument types

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/decl.go`, `internal/parse/declarations.go`, `internal/parse/coretypes_test.go`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `namedUnary` in `internal/parse/keyword.go:40` (the measured named unaries), the lexer's keyword list (`lexer.Keywords()`), `coreTable` (`internal/parse/declarations.go:39`)

RFC 0001 "Builtins with no prototype": about 20 builtins have no prototype perl reports. CORE.pmt declares their types; `:unary` marks the ones that parse as named unaries; a builtin with no prototype is otherwise a list operator, as an ordinary sub is. This issue declares the plain ones -- including `chomp`/`chop` with `List[Str] @args :lvalue`, `select` as the RFC's multi (`(FileHandle $fh) Str`, `($r, $w, $e, Num $timeout) Int`), and `each` as the RFC's multi (`(Hash %h) List`, `(Array @a) List`), whose candidates must still derive perl's `\[%@]`. `print`/`exec` and kin came with the invocant colon; `grep`/`map`/`sort` are the next issue's. Find the set with perl (keywords whose `prototype("CORE::k")` is undef and which are functions) and the unary subset by the deparse method `keyword.go` records; hold the expected names in the test, not against keyword.go. Declarations with no derived prototype stay out of `coreTable()` (see `aliasTarget`, `internal/parse/use.go:655`), so `TestCoreDeclarationsArePerls` and `TestCoreDerivedPrototypesArePerls` stay green unchanged.

## Steps

- [ ] Measure with perl, then write failing test `TestCoreUnaryBuiltins`: CORE.pmt declares each no-prototype function builtin (names held in the test), and exactly the measured named unaries carry `:unary`
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestCoreUnaryBuiltins$' -count=1 -v`
- [ ] Read `:unary`; add the declarations with their measurement comments; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestCoreSelectAndEachAreMultis`: `select` and `each` each have two candidates with the RFC's signatures, and `each`'s derived prototype is perl's
- [ ] Implement; run test to verify pass; `make test`; commit

## Acceptance criteria

- [ ] CORE.pmt declares the no-prototype builtins and :unary marks the named unaries (`go test ./internal/parse/ -run '^TestCoreUnaryBuiltins$' -count=1 -v | grep -q '^--- PASS: TestCoreUnaryBuiltins'`)
- [ ] select and each are declared as multis (`go test ./internal/parse/ -run '^TestCoreSelectAndEachAreMultis$' -count=1 -v | grep -q '^--- PASS: TestCoreSelectAndEachAreMultis'`)
