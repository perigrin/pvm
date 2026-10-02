---
title: Type calls to imported library subs from their .pmt declarations and select multis at call sites
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-02T17:13:31.136589197Z
updated: 2026-10-02T17:13:31.136589197Z
---

## Prerequisites

- [ ] Have internal/infer read builtin and operator types from CORE.pmt and remove internal/types' builtin table
- [ ] Let a library .pmt use every construct CORE.pmt uses
- [ ] Retire the tree-sitter vocabulary from internal/infer (m2-lower-the-tree, 01a0c115-df67-7d69-8883-73d40b0ae15d)

## Context

- paths: `internal/infer/infer.go`, `internal/infer/declared_test.go` (new), `test/e2e/psc_check_test.go`, `test/e2e/testdata/`
- commands: `go test ./internal/infer/ -count=1`, `go test ./test/e2e/ -run '^TestPSCCheck' -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `internal/infer/infer.go` call-site typing (after m2 it walks `internal/parse` nodes), `internal/types/select.go`, `parse.Imports` (`internal/parse/exports.go:49`)

The milestone body's "internal/infer reads the declarations": once m2 has moved infer onto `internal/parse`, a call to a sub imported from a module with a `.pmt` is typed from that declaration, and a call to a multi -- builtin or library -- selects its candidate at the call site by arity, argument types, context or operand shape through `internal/types/select.go`, joining when undecided. This changes what `psc check` reports wherever a declaration is more precise, which is why it waits for m2's behaviour-preserving cutover. It re-marks no RFC section (outside the RFC gate, inside the milestone). Use a shipped declaration (e.g. `Moose.pmt`) or extend one with measured types for the e2e case.

## Steps

- [ ] Write failing test `TestInferTypesCallFromDeclaration` in `internal/infer/declared_test.go`: a call to an imported sub declared with a return type in a `.pmt` is typed with it
- [ ] Run test to verify failure: `go test ./internal/infer/ -run '^TestInferTypesCallFromDeclaration$' -count=1 -v`
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestInferSelectsMultiAtCallSite`: `localtime(0)` in scalar and list context, `each %h` against `each @a`, and an undecided call joining
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing e2e test `TestPSCCheckUsesDeclarations` in `test/e2e/psc_check_test.go`: `psc check` on a fixture reports a coercion mismatch that only the declaration's types reveal
- [ ] Run test to verify failure: `go test ./test/e2e/ -run '^TestPSCCheckUsesDeclarations$' -count=1 -v`
- [ ] Make it pass; `make test`; commit

## Acceptance criteria

- [ ] infer types a call from its .pmt declaration (`go test ./internal/infer/ -run '^TestInferTypesCallFromDeclaration$' -count=1 -v | grep -q '^--- PASS: TestInferTypesCallFromDeclaration'`)
- [ ] infer selects a multi candidate at the call site (`go test ./internal/infer/ -run '^TestInferSelectsMultiAtCallSite$' -count=1 -v | grep -q '^--- PASS: TestInferSelectsMultiAtCallSite'`)
- [ ] psc check reports what a declaration reveals (`go test ./test/e2e/ -run '^TestPSCCheckUsesDeclarations$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckUsesDeclarations'`)
