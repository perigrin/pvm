---
title: Enforce the slurpy rule and caller-view sigils in .pmt signatures
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-4139-773b-8676-e148e90c42bd
blocks:
- 01a0fd9b-4850-704d-8c14-c15add6194f2
- 01a0fd9b-49b2-7ea9-bfa1-ba398baadc3f
created: 2026-10-02T17:13:26.249797267Z
updated: 2026-10-02T17:23:57.373735362Z
---

## Prerequisites

- [ ] Parse typed declarations in .pmt files: typed parameters, defaults and return types

## Context

- paths: `internal/parse/decl.go`, `internal/types/declared.go`, `internal/parse/pmt_test.go`, `docs/rfcs/0001-type-libraries.md`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: the typed-signature reader added to `internal/parse/decl.go` by the previous issue; `types.Param`

RFC 0001 "A slurpy takes no bare element type". `Str @args` is refused (ambiguous container, ambiguous element assertion); a slurpy is untyped or carries an explicit container type, so `List[Str] @args` parses -- container types parse now, only in `.pmt` files, and what they mean is provisional until the paper (record the element type name on the param; the lattice stays flat). In a `.pmt` a parameter's sigil is the caller's view, as a prototype character is: `(Array @, List @)` is allowed although perl's own signatures reject two slurpies, and only a final `List` or `Hash` parameter is slurpy.

## Steps

- [ ] Write failing test `TestPmtContainerTypeParses`: `sub f (List[Str] @args);` in a `.pmt` gives one slurpy param of container `List`, element `Str`
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestPmtContainerTypeParses$' -count=1 -v`
- [ ] Implement `Name[Type]` parsing for parameter types; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestPmtBareSlurpyTypeRefused`: `sub f (Str @args);` and `sub f (Int %h);` are errors; untyped `sub f (@args);` is not
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestPmtCallerViewSigils`: `sub push (Array @, List @) Int;` parses with the first param non-slurpy and the second slurpy
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Re-mark the RFC heading "A slurpy takes no bare element type" `(*Implemented*, <sha>)`; commit

### Negative Scenarios

- a union is still a bare element type: `sub f (Str|Undef @args);` is refused like `Str @args` (`go test ./internal/parse/ -run '^TestPmtBareUnionSlurpyRefused$' -count=1 -v | grep -q '^--- PASS: TestPmtBareUnionSlurpyRefused'`)
- malformed container types are errors naming the bad part, not panics: `List[Str @args` (unclosed), `List[] @args` (empty), `List[Strng] @args` (unknown element), `Strng[Str] @args` (unknown container) (`go test ./internal/parse/ -run '^TestPmtMalformedContainerTypeRefused$' -count=1 -v | grep -q '^--- PASS: TestPmtMalformedContainerTypeRefused'`)
- container types are typed Perl and so `.pmt`-only: `sub f (List[Str] @args) { 1 }` in ordinary source records no container type (`go test ./internal/parse/ -run '^TestContainerTypeOnlyInPmt$' -count=1 -v | grep -q '^--- PASS: TestContainerTypeOnlyInPmt'`)
- a `List` parameter that is not last (`sub f (List @a, Str $x);`, `sub f (List @a, List @b);`) -- the RFC says only the final `List`/`Hash` is slurpy but not whether an earlier one is an error; needs perigrin's ruling before it becomes a criterion (`go test ./internal/parse/ -run '^TestPmtNonFinalListParam$' -count=1 -v | grep -q '^--- PASS: TestPmtNonFinalListParam'`)

## Acceptance criteria

- [ ] a container type such as List[Str] parses in a .pmt (`go test ./internal/parse/ -run '^TestPmtContainerTypeParses$' -count=1 -v | grep -q '^--- PASS: TestPmtContainerTypeParses'`)
- [ ] a slurpy with a bare element type is refused (`go test ./internal/parse/ -run '^TestPmtBareSlurpyTypeRefused$' -count=1 -v | grep -q '^--- PASS: TestPmtBareSlurpyTypeRefused'`)
- [ ] caller-view sigils allow (Array @, List @) (`go test ./internal/parse/ -run '^TestPmtCallerViewSigils$' -count=1 -v | grep -q '^--- PASS: TestPmtCallerViewSigils'`)
- [ ] the RFC section on slurpies is implemented (`grep -q '^### A slurpy takes no bare.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] a slurpy typed with a bare union is refused (`go test ./internal/parse/ -run '^TestPmtBareUnionSlurpyRefused$' -count=1 -v | grep -q '^--- PASS: TestPmtBareUnionSlurpyRefused'`)
- [ ] malformed container types are errors (`go test ./internal/parse/ -run '^TestPmtMalformedContainerTypeRefused$' -count=1 -v | grep -q '^--- PASS: TestPmtMalformedContainerTypeRefused'`)
- [ ] container types are not read in ordinary source (`go test ./internal/parse/ -run '^TestContainerTypeOnlyInPmt$' -count=1 -v | grep -q '^--- PASS: TestContainerTypeOnlyInPmt'`)
