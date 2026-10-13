---
title: "Port the statement and flow rules onto the lowered tree: conditionals, guards, loops, returns, use"
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a12a85-304b-77f1-9bda-e6b73c317305
blocks:
- 01a12a85-31a5-769d-88c4-e2add3f8fefc
created: 2026-10-11T10:32:14.336031794Z
updated: 2026-10-13T00:37:12.207504849Z
sessions:
- start_sha: 0908d613150f31c001ca72445c97207e13a34b47
  end_sha: ""
  commits: 0
  started_at: 2026-10-13T00:37:12.207504849Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-13T00:37:12.207504849Z
---

## Context

- paths: `internal/infer/lowered.go`, `internal/infer/testdata/lowered/flow/` (new)
- commands: `go test ./internal/infer/ -run "^TestLoweredGoldenRatchet$" -count=1 -v`
- entrypoints: `walkConditionalStatement` (`infer.go:2750`), `extractGuardPattern` (`:2283`), `extractNegatedGuard` (`:2469`), `walkLoopStatement` (`:3107`), `collectExplicitReturns` (`:3294`), `collectImplicitReturn` (`:3337`), `blockAlwaysExits` (`:3495`)

The second of seven port issues. It ports the rules that walk statements: guard narrowing in
conditionals (`defined`, `ref`, `isa`, and user-defined guards), the
negated-guard and branch-join logic, loops, explicit and implicit returns,
and `use` statements.

Two things the lowered tree changes here:

- **`unless` is a branch swap, not a `Not`** (§4.14.5, and the statements
  issue's lowering). A guard under `unless` narrows the ELSE branch. The
  legacy rule reaches the same answer by negating; check each guard fixture
  for both spellings.
- **`blockAlwaysExits` detects `return`/`die`/`exit` by text.** `LoopEx.Op`
  makes `return` a field read. `last`/`next` are needed for correct branch
  joins too; adding them changes what PSC reports, so it is NOT part of this
  port. If a fixture shows the gap, record it and file a follow-up.

The method is the one "AnalyzeLowered: the entry point, the golden
ratchet, and the declaration and literal rules" set up: fixtures
under `testdata/lowered/flow/`, goldens captured from the legacy path, and
`TestLoweredGoldenRatchet` holding `matching.txt` in both directions. Every
rule is REWRITTEN against the lowered vocabulary, not shimmed. No dataflow
IR (§6.1.6): if a rule wants flow, it keeps the approximation it has today.

Rollback: this issue changes only `AnalyzeLowered` and test data. Nothing in `internal/psc` calls it yet, so reverting it changes nothing PSC reports.

## Steps

- [ ] Write fixtures under `testdata/lowered/flow/`, drawn from the
      existing cases in `infer_test.go` where they exist, using only
      constructs whose rules are already ported plus this slice's.
- [ ] Capture their goldens from the legacy path:
      `go test ./internal/infer/ -run "^TestCaptureLegacyGolden$" -infer.capture -count=1`.
      Commit the fixtures and goldens.
- [ ] Run the ratchet and confirm the new fixtures do not match yet:
      `go test ./internal/infer/ -run "^TestLoweredGoldenRatchet$" -count=1 -v`
- [ ] Port this slice's rules into `AnalyzeLowered`. Commit as each rule
      group goes green rather than once at the end.
- [ ] Update `matching.txt` with `-infer.ratchet.update` and re-run the
      ratchet to green. Every `flow/` fixture must match. One that cannot
      is a legacy defect or a regression: triage it before moving on. If it
      is a legacy defect, edit its golden in its own commit with a
      `Diagnostics-changed:` trailer naming the cause.
- [ ] Write `TestLoweredFlowOverBrokenInput`: `if (;) { my $y = 1; }` and
      `my $x; if ($x) {` (perl rejects both) analyse without a panic, emit
      no diagnostic inside a `Hole`, and still type `$y` in the intact
      block. Run to verify failure, implement, run to green, commit.
- [ ] Run the infer suite: `go test ./internal/infer/ -count=1`
- [ ] Commit.

### Whole-suite criteria removed (2026-10-11)

perigrin, on a milestone verify that would have run for 9+ hours: "Why are
we doing 73 full test suite runs?!?" Criteria that only re-ran a whole suite
or package (`go test ./...`, `make test`, the whole conformance or parse
package) restated "make test passed", which gates every commit, while CI
gates every PR -- and `git zhi verify` re-ran each copy for every done issue.
They are removed; this issue's named tests and ratchet checks remain.

## Acceptance Criteria

### Positive Scenarios
- [ ] The lowered path holds the golden ratchet in both directions (`go test ./internal/infer/ -run '^TestLoweredGoldenRatchet$' -count=1 -v | grep -q '^--- PASS: TestLoweredGoldenRatchet'`)
- [ ] Every flow fixture matches its legacy golden on the lowered path (`test -n "$(ls internal/infer/testdata/lowered/flow/*.pl)" && ! ls internal/infer/testdata/lowered/flow/*.pl | sed 's|.*/lowered/||' | grep -vxFf internal/infer/testdata/lowered/matching.txt`)

### Negative Scenarios
- [ ] A conditional perl rejects analyses without a panic, says nothing about the Hole, and still types the intact block (`go test ./internal/infer/ -run '^TestLoweredFlowOverBrokenInput$' -count=1 -v | grep -q '^--- PASS: TestLoweredFlowOverBrokenInput'`)
