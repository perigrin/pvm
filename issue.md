---
title: Derive the parser's keyword shapes from CORE.pmt and remove the hand-kept keyword table
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-519d-7b42-886e-2222a3076986
blocks:
- 01a0fd9b-5670-78d0-84be-9b9618e9b081
created: 2026-10-02T17:13:30.827426197Z
updated: 2026-10-02T17:24:19.122539271Z
---

## Prerequisites

- [ ] Declare grep, map and sort as multis in CORE.pmt, with sort's invocant slot

## Context

- paths: `internal/parse/keyword.go` (deleted), `internal/parse/call.go`, `internal/parse/indirect.go`, `internal/parse/declarations.go`, `internal/parse/declarations/CORE.pmt`, `internal/parse/coreshapes_test.go` (new)
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `namedUnary`, `listOperator`, `niladicParse` (`internal/parse/keyword.go:40`, `:91`, `:195`); their consumers `internal/parse/call.go:75-76`, `:130`, `:185`, `:205-221`, `:560` and `internal/parse/indirect.go:54`; `ShapeOf` (`internal/parse/shape.go:60`)

RFC 0001 "Typed Perl, in `.pmt` only": `internal/parse/keyword.go` folds into CORE.pmt after a test shows the file agrees with the table it replaces, and its measurement comments move with it. For a builtin with a prototype the shape follows from it (`ShapeOf`); for the rest `:unary` says it, and niladic builtins have a `()` prototype. First copy keyword.go's three sets into the new test as golden values and prove CORE.pmt-derived shapes equal them -- any disagreement is a finding to measure against perl, not to paper over. Then the call sites read the derived shapes and the tables go. `internal/parse/keyword_test.go` tests parse behaviour, not the tables, and stays. The criteria check symbols, never the deleted file.

## Steps

- [ ] Write failing test `TestCoreDerivedShapes` in `internal/parse/coreshapes_test.go`: the named unaries, list operators and niladic builtins derived from CORE.pmt equal the golden sets (copied from keyword.go, with the measurement method noted)
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestCoreDerivedShapes$' -count=1 -v`
- [ ] Implement the derivation; resolve every disagreement by measuring perl; run test to verify pass; `make test`; commit
- [ ] Switch `call.go` and `indirect.go` to the derived shapes; `make test`; commit
- [ ] Move keyword.go's measurement comments into CORE.pmt; delete keyword.go; `make test` (ratchets unchanged); commit

### Negative Scenarios

- a name CORE.pmt does not declare (`frobnicate`) derives no keyword shape: it parses as an ordinary sub call, never as a named unary or niladic (`go test ./internal/parse/ -run '^TestCoreShapesUnknownNameHasNoShape$' -count=1 -v | grep -q '^--- PASS: TestCoreShapesUnknownNameHasNoShape'`)
- the shape comparison is not tautological: a golden set with one name removed or one CORE.pmt line with a wrong `:unary` is reported as a disagreement naming the builtin (`go test ./internal/parse/ -run '^TestCoreDerivedShapesCatchesMismatch$' -count=1 -v | grep -q '^--- PASS: TestCoreDerivedShapesCatchesMismatch'`)
- parse behaviour does not move: the existing parse tests and the T1/T2 ratchets pass unchanged after the switch (`go test ./internal/parse/ -count=1`) (`go test ./internal/parse/ -count=1`)

## Acceptance criteria

- [ ] keyword shapes derived from CORE.pmt equal the measured sets (`go test ./internal/parse/ -run '^TestCoreDerivedShapes$' -count=1 -v | grep -q '^--- PASS: TestCoreDerivedShapes'`)
- [ ] the hand-kept keyword tables are gone (`sh -c '! grep -rqwE "namedUnary|listOperator|niladicParse" --include="*.go" --exclude="*_test.go" internal/parse'`)
- [ ] an undeclared name derives no keyword shape (`go test ./internal/parse/ -run '^TestCoreShapesUnknownNameHasNoShape$' -count=1 -v | grep -q '^--- PASS: TestCoreShapesUnknownNameHasNoShape'`)
- [ ] the shape check reports a disagreement (`go test ./internal/parse/ -run '^TestCoreDerivedShapesCatchesMismatch$' -count=1 -v | grep -q '^--- PASS: TestCoreDerivedShapesCatchesMismatch'`)
