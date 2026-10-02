---
title: Enforce the slurpy rule and caller-view sigils in .pmt signatures
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-02T17:13:26.249797267Z
updated: 2026-10-02T17:13:26.249797267Z
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

## Acceptance criteria

- [ ] a container type such as List[Str] parses in a .pmt (`go test ./internal/parse/ -run '^TestPmtContainerTypeParses$' -count=1 -v | grep -q '^--- PASS: TestPmtContainerTypeParses'`)
- [ ] a slurpy with a bare element type is refused (`go test ./internal/parse/ -run '^TestPmtBareSlurpyTypeRefused$' -count=1 -v | grep -q '^--- PASS: TestPmtBareSlurpyTypeRefused'`)
- [ ] caller-view sigils allow (Array @, List @) (`go test ./internal/parse/ -run '^TestPmtCallerViewSigils$' -count=1 -v | grep -q '^--- PASS: TestPmtCallerViewSigils'`)
- [ ] the RFC section on slurpies is implemented (`grep -q '^### A slurpy takes no bare.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
