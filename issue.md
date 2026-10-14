---
title: Move internal/infer's own tests onto the lowered path, and assert the precision floor there
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a12a85-3386-7e72-8d74-38bf69a9c7d7
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-10-11T10:32:15.127376918Z
updated: 2026-10-14T02:46:41.07481046Z
sessions:
- start_sha: 26f0e9e5343e8d6b9487deedf0564dfef67acab5
  end_sha: ""
  commits: 0
  started_at: 2026-10-14T02:46:41.07481046Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-14T02:46:41.07481046Z
---

## Context

- paths: `internal/infer/infer_test.go` (the helpers at `:22-54` and `findNodeType`), `internal/infer/symbols_test.go`, `internal/infer/core_test.go`, `internal/infer/mutation_test.go`, `internal/infer/precision_test.go`, `test/e2e/psc_check_test.go` (`TestTypeInferenceAccuracy`)
- commands: `go test ./internal/infer/ -run "^TestInferencePrecision$" -count=1 -v`
- entrypoints: `internal/infer/infer_test.go:22` (`analyzeSource`), `:33` (`analyzeSourceWithOptions`), `:43` (`analyzeSourceFull`), `:54` (`findNodeType`); `internal/infer/precision_test.go:134`

At 24bfac4a `internal/infer` has 246 tests (`grep -c '^func Test'` over its
eight `_test.go` files). Most build a tree-sitter tree and call `Analyze`:
three of the test files import `internal/parser` (`infer_test.go`,
`symbols_test.go`, `precision_test.go`), and `core_test.go` and
`mutation_test.go` reach it through `infer_test.go`'s helpers. When
01a0c115-df67 deletes the legacy walkers, every one of them breaks. They are
also the strongest differential this milestone has: they encode what PSC is
supposed to say. So they move to the lowered path BEFORE either caller
flips, not after.

Most of the move is four helpers. `analyzeSource`, `analyzeSourceWithOptions`
and `analyzeSourceFull` switch to `parse.Parse`, `lower.Lower` and
`AnalyzeLowered`. `findNodeType` searches a byte-keyed map by source
offset; it switches to asking the position map for the node at each offset.
`symbols_test.go` calls `CollectDeclarations` on a tree-sitter tree directly
and moves to the lowered symbol pass. `project_test.go` exercises
`ProjectIndex`, which flips with the batch checker (01a0c115-a44b), so it
moves there.

**One caller outside `internal/infer`.** At e18b59c8,
`test/e2e/psc_check_test.go` imports `internal/parser` for
`TestTypeInferenceAccuracy` (`:110`), which calls `parser.New()` and legacy
`infer.Analyze` directly. No other issue owns it, and 01a0c115-df67 deletes
`internal/parser`, so it moves here with the rest of the suite: parse with
`parse.Parse`, lower, run `AnalyzeLowered`, and find each fragment's type
through the position map.

A test that fails after the move is a finding, not noise: either a port
missed a rule (fix the port) or the legacy expectation was a legacy defect
(change the expectation in its own commit with a `Diagnostics-changed:`
trailer, as for a golden). Do not delete a test to get green. If one tests
something that exists only in the tree-sitter tree, delete it in its own
commit that says so.

**The precision gate moves too, and gains a floor.** `TestInferencePrecision`
runs the legacy `parser.New()`/`Analyze` path (`precision_test.go:141-144`)
and asserts only that WRONG is zero (`:218`). It never asserts the exact
count. Measured at 24bfac4a, against 76 observed values: exact 62 (81.6%),
wider 12, WRONG 0, no answer 2. Goal (6) is "PSC does not get worse", so the
floor is 62: add an assertion that exact is at least 62, keep WRONG at zero,
and run it on the lowered path. It skips when perl is absent; CI installs
perl (`.github/workflows/ci.yml:48`), and a skip prints no `exact:` line, so
the floor criterion below fails rather than passing vacuously.

## Steps

- [ ] Add the floor to `TestInferencePrecision` on the legacy path first:
      `exact >= 62`. Run it and confirm it passes at 62, then temporarily
      set the floor to 63 and confirm it fails. Commit the 62 floor.
- [ ] Switch `TestInferencePrecision` to `parse.Parse`, `lower.Lower` and
      `AnalyzeLowered`. Run it. If exact drops below 62 or WRONG rises,
      fix the port, not the floor.
- [ ] Switch the three `analyzeSource*` helpers and `findNodeType`. Run the
      suite: `go test ./internal/infer/ -count=1`
- [ ] Triage every failure as above, one commit per cause.
- [ ] Switch `symbols_test.go` to the lowered symbol pass. Run, triage,
      commit.
- [ ] Write `TestPrecisionFloorRejectsADrop`: factor the floor into one
      function the precision test calls, and assert it rejects an exact
      count of 61 and accepts 62. Run to verify failure, implement, run to
      green, commit.
- [ ] Switch `TestTypeInferenceAccuracy` in `test/e2e/psc_check_test.go`
      to the lowered path. Run it:
      `go test ./test/e2e/ -run "^TestTypeInferenceAccuracy$" -count=1 -v`.
      Triage, commit.
- [ ] Confirm the only test file left importing `internal/parser`, in
      `internal/infer` or `test/e2e`, is `legacy_capture_test.go`.
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
- [ ] The precision gate passes on the lowered path with WRONG at zero and the exact floor asserted (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)
- [ ] The lowered path gets at least 62 of 76 observed values exactly right (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | awk '/^--- PASS: TestInferencePrecision/ {p = 1} /^  exact:/ {n = $2} END {exit !(p && n >= 62)}'`)
- [ ] No infer or e2e test reaches the tree-sitter tree except the legacy capture tool (`! grep -l 'pvm/internal/parser"' internal/infer/*_test.go test/e2e/*_test.go | grep -v 'legacy_capture_test.go'`)
- [ ] The e2e inference accuracy test runs on the lowered path (`go test ./test/e2e/ -run '^TestTypeInferenceAccuracy$' -count=1 -v | grep -q '^--- PASS: TestTypeInferenceAccuracy'`)
- [ ] The suite was moved, not deleted: it still holds at least the 246 tests it had (`test "$(grep -h '^func Test' internal/infer/*_test.go | wc -l)" -ge 246`)

### Negative Scenarios
- [ ] The floor is real: the check the precision test uses rejects an exact count of 61 (`go test ./internal/infer/ -run '^TestPrecisionFloorRejectsADrop$' -count=1 -v | grep -q '^--- PASS: TestPrecisionFloorRejectsADrop'`)
