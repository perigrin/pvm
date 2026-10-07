---
title: Derive the parser's keyword shapes from CORE.pmt and remove the hand-kept keyword table
state: done
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-519d-7b42-886e-2222a3076986
- 01a0fd9b-5670-78d0-84be-9b9618e9b081
- 01a1113c-eb6e-7e0e-8554-bcb5231478eb
created: 2026-10-02T17:13:30.827426197Z
updated: 2026-10-07T20:19:21.155656333Z
sessions:
- start_sha: 23e4c45ddfa7065d1eed3976d6e90153c8a3106a
  end_sha: 23e4c45ddfa7065d1eed3976d6e90153c8a3106a
  commits: 0
  started_at: 2026-10-07T20:19:21.097450792Z
  ended_at: 2026-10-07T20:19:21.155656333Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-07T20:19:21.097450792Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-07T20:19:21.155656333Z
---

## Prerequisites

- [ ] split has no prototype but typed positional parameters: typed Perl has no spelling for it
- [ ] Declare grep, map and sort as multis in CORE.pmt, with sort's invocant slot
- [ ] Have internal/infer read builtin and operator types from CORE.pmt and remove internal/types' builtin table

## Context

- paths: `internal/parse/keyword.go` (deleted), `docs/rfcs/0001-type-libraries.md`, `internal/parse/call.go`, `internal/parse/indirect.go`, `internal/parse/declarations.go`, `internal/parse/declarations/CORE.pmt`, `internal/parse/coreshapes_test.go` (new), `docs/specs/perl-parser/04-expressions.md`
- commands: `go test ./internal/parse/ -count=1` (measured on this tree, 2026-10-05: 30s for the whole package, ratchets included, well under the 300s `git-zhi verify` allows a command)
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `namedUnary`, `listOperator`, `niladicParse` (`internal/parse/keyword.go:40`, `:91`, `:195`); their consumers `internal/parse/call.go:75-76`, `:130`, `:185`, `:205-221`, `:560` and `internal/parse/indirect.go:54`; `ShapeOf` (`internal/parse/shape.go:60`)

RFC 0001 "Typed Perl, in `.pmt` only": `internal/parse/keyword.go` folds into CORE.pmt after a test shows the file agrees with the table it replaces, and its measurement comments move with it. For a builtin with a prototype the shape follows from it (`ShapeOf`); for the rest `:unary` says it, and niladic builtins have a `()` prototype. First copy keyword.go's three sets into the new test as golden values and prove CORE.pmt-derived shapes equal them -- any disagreement is a finding to measure against perl, not to paper over. Then the call sites read the derived shapes and the tables go. `internal/parse/keyword_test.go` tests parse behaviour, not the tables, and stays. The criteria check symbols, never the deleted file. This is the last fold: `internal/types/signatures.go` is already gone (the infer issue), and with both tables gone this completes RFC "Typed Perl, in `.pmt` only", so this issue re-marks it.

## Steps

- [ ] Write failing test `TestCoreDerivedShapes` in `internal/parse/coreshapes_test.go`: the named unaries, list operators and niladic builtins derived from CORE.pmt equal the golden sets (copied from keyword.go, with the measurement method noted)
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestCoreDerivedShapes$' -count=1 -v`
- [ ] Implement the derivation; resolve every disagreement by measuring perl; run test to verify pass; `make test`; commit
- [ ] Switch `call.go` and `indirect.go` to the derived shapes; `make test`; commit
- [ ] Move keyword.go's measurement comments into CORE.pmt; delete keyword.go; `make test` (ratchets unchanged); commit
- [ ] Re-mark the RFC heading "Typed Perl, in `.pmt` only" `(*Implemented*, <sha>)`; commit
- [ ] Doc: restate the RFC's "Typed Perl" paragraph on folding the two tables in the present tense (`CORE.pmt` is the builtin table), dropping its count of signatures.go's typed builtins (it said 27; the table had 40); commit
- [ ] Doc: `docs/specs/perl-parser/04-expressions.md` §4.14.2 cites `internal/parse/keyword.go` for `delete` and `exists`; re-measure whether `delete $h{k};` and `exists $h{k};` still parse as `Unknown`, and restate the note against `CORE.pmt`; commit

### Negative Scenarios

- a name CORE.pmt does not declare (`frobnicate`) derives no keyword shape: it parses as an ordinary sub call, never as a named unary or niladic (`go test ./internal/parse/ -run '^TestCoreShapesUnknownNameHasNoShape$' -count=1 -v | grep -q '^--- PASS: TestCoreShapesUnknownNameHasNoShape'`)
- the shape comparison is not tautological: a golden set with one name removed or one CORE.pmt line with a wrong `:unary` is reported as a disagreement naming the builtin (`go test ./internal/parse/ -run '^TestCoreDerivedShapesCatchesMismatch$' -count=1 -v | grep -q '^--- PASS: TestCoreDerivedShapesCatchesMismatch'`)
- parse behaviour does not move: the existing parse tests and the T1/T2 ratchets pass unchanged after the switch (`go test ./internal/parse/ -count=1`)

## Acceptance criteria

- [ ] keyword shapes derived from CORE.pmt equal the measured sets (`go test ./internal/parse/ -run '^TestCoreDerivedShapes$' -count=1 -v | grep -q '^--- PASS: TestCoreDerivedShapes'`)
- [ ] the hand-kept keyword tables are gone (`sh -c '! grep -rqwE "namedUnary|listOperator|niladicParse" --include="*.go" --exclude="*_test.go" internal/parse'`)
- [ ] an undeclared name derives no keyword shape (`go test ./internal/parse/ -run '^TestCoreShapesUnknownNameHasNoShape$' -count=1 -v | grep -q '^--- PASS: TestCoreShapesUnknownNameHasNoShape'`)
- [ ] the shape check reports a disagreement (`go test ./internal/parse/ -run '^TestCoreDerivedShapesCatchesMismatch$' -count=1 -v | grep -q '^--- PASS: TestCoreDerivedShapesCatchesMismatch'`)
- [ ] the parser spec no longer cites the deleted keyword table (`sh -c '! grep -q "keyword\.go" docs/specs/perl-parser/04-expressions.md'`)
- [ ] the RFC no longer counts 27 typed builtins (`sh -c '! grep -q "27 typed" docs/rfcs/0001-type-libraries.md'`)
- [ ] parse behaviour does not move: the existing parse tests and the ratchets pass after the switch (`go test ./internal/parse/ -count=1`)
- [ ] the RFC section on typed Perl is implemented (`grep -q '^### Typed Perl, in .*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
