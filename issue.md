---
title: "AnalyzeLowered: the entry point, the golden ratchet, and the declaration and literal rules"
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c113-dfd3-76a4-8b61-ae61b324bcb4
- 01a0c112-8a65-7d5c-94ec-69859b2928d3
blocks:
- 01a12a85-3100-79ae-8265-aa8b4f39cd85
created: 2026-10-11T10:32:14.155064204Z
updated: 2026-10-12T23:53:36.807153314Z
sessions:
- start_sha: 2f38962960ad8ceb79c510b01a862647f0fbacaa
  end_sha: ""
  commits: 0
  started_at: 2026-10-12T23:53:36.807153314Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-12T23:53:36.807153314Z
---

## Context

- paths: `internal/infer/lowered.go` (new), `internal/infer/lowered_test.go` (new), `internal/infer/golden_test.go` (new, permanent), `internal/infer/legacy_capture_test.go` (new, deleted by 01a0c115-df67), `internal/infer/testdata/lowered/` (new: fixtures, `.golden` files, `matching.txt`), `internal/infer/symbols.go`
- commands: `go test ./internal/infer/ -run "^TestLoweredGoldenRatchet$" -count=1 -v`
- entrypoints: `internal/infer/infer.go:122` (`AnalyzeWithOptions`), `internal/infer/symbols.go:255` (`CollectDeclarations`)

The first of seven issues that port `internal/infer`'s rules onto the lowered
tree, one per lowering slice. Together they replace what 01a0c115-a44b used
to hold in one issue: a rewrite of about 5,200 non-test lines (`infer.go`
4,459, `symbols.go` 451, `project.go` 135, `diagnostics.go` 161 at 24bfac4a).

This issue adds `infer.AnalyzeLowered` ALONGSIDE `Analyze`. Nothing calls it
from `internal/psc` yet, and nothing legacy is deleted.

**The key is node identity from birth.** `AnalyzeLowered` returns
`map[lower.ID]types.Type`, not `map[uint32]types.Type`. It is a new function,
so it needs no key swap. That replaces 01a0c115-88d1, which could not be done
as filed: `Analyze` walks the tree-sitter tree, which has no ids. The
identity tests 88d1 specified move to the literals-and-variables port
issues, where their fixtures first type. Diagnostics still carry byte spans
(`Diagnostic.StartByte`/`EndByte`, `diagnostics.go:40`); they come from the
position map (01a0c113-dfd3), which is why this issue runs after it.

**The golden ratchet is how every port proves itself, and how goal (6) is
kept.** Each fixture `testdata/lowered/<slice>/<name>.pl` has a
`<name>.golden` captured from the LEGACY path: its diagnostics
(start, end, code, message) and its symbol table (name, type), sorted.
`TestLoweredGoldenRatchet` runs `lower.Lower` and `AnalyzeLowered` on every
fixture, and compares the set of fixtures that match their golden against
`testdata/lowered/matching.txt`. It fails in BOTH directions, as the parse
ratchets do: a listed fixture that stops matching is a regression, and an
unlisted one that starts matching must be added. `-infer.ratchet.update`
rewrites the list.

Why golden files, not a live differential: a test that calls legacy
`Analyze` turns permanently red when 01a0c115-df67 deletes it, and
`git zhi verify` re-runs every done issue's criteria at the end of the
milestone. The golden files stay as the guard after the legacy path is
gone. The capture tool (`TestCaptureLegacyGolden`, run only with
`-infer.capture`) is the one place that calls legacy `Analyze`. No criterion
names it, and df67 deletes it.

**A golden that changes must be explained.** A fixture's first capture adds
its `.golden`. Any later commit that MODIFIES a `.golden` changes what PSC
reports, and goal (6) requires the commit to explain it. Convention: such a
commit carries a `Diagnostics-changed:` trailer per changed file, naming the
cause (for example "legacy Range rule matched `.` before `..`;
`findOperatorText` is gone"). The milestone umbrella checks the trailer.
(perigrin, 2026-10-11: decided. The repo merges with merge commits, not
squashes, so the trailer survives on the commit that carries it.)

**This slice's rules: declarations and literals together.** Neither is
observable without the other: `my $x = 5;` types `$x` only when both the
declaration rule and the `NumLit` rule exist. Port `CollectDeclarations`
(`symbols.go:255`) onto `Decl` and sub declarations, and the literal typing
onto `NumLit`, `StrLit`, `QwList`, `Bareword` and `Hole`.

**A `Hole` is typed `Any` and never inspected** (§4.14.2, Hole row).

Rollback: this issue adds code and test data only. Reverting it removes
`AnalyzeLowered` and changes nothing PSC reports.

## Steps

- [ ] Write `TestLoweredGoldenRatchet` and its helper in `golden_test.go`.
      It fails with no fixtures present.
- [ ] Run to verify failure: `go test ./internal/infer/ -run "^TestLoweredGoldenRatchet$" -count=1 -v`
- [ ] Write `TestCaptureLegacyGolden` in `legacy_capture_test.go`, skipped
      unless `-infer.capture` is passed.
- [ ] Write fixtures under `testdata/lowered/decl/`: `my`, `our`, `local`,
      `state` declarations, the list form `my ($a, $b) = ...`, named and
      anonymous subs, packages; and under `testdata/lowered/literals/`:
      numbers, strings, `qw`, a backtick string, a fat-comma bareword. Draw
      them from the existing cases in `infer_test.go` and `symbols_test.go`
      where those exist. Use only declarations and literals.
- [ ] Capture: `go test ./internal/infer/ -run "^TestCaptureLegacyGolden$" -infer.capture -count=1`.
      Commit the fixtures and goldens.
- [ ] Add `AnalyzeLowered` in `internal/infer/lowered.go`, dispatching on the
      lowered kinds, with every kind no port handles yet typed `Any` and no
      diagnostic. Port the declaration and literal rules. Run the ratchet
      with `-infer.ratchet.update` to record the fixtures that match, then
      without it to green. Every `decl/` and `literals/` fixture must match;
      if one cannot, it is a legacy defect or a regression: triage it, and
      if it is a legacy defect, edit its golden in a commit with a
      `Diagnostics-changed:` trailer.
- [ ] Write `TestHoleIsTypedAnyAndNeverInspected`: lower `my $x = (;` and
      an unterminated `"abc` (perl rejects both), run `AnalyzeLowered`, and assert the
      `Hole` is annotated `Any`, no diagnostic points inside it, and nothing
      panics. Run to verify failure, implement, run to green, commit.
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
- [ ] Every declaration fixture matches its legacy golden on the lowered path (`test -n "$(ls internal/infer/testdata/lowered/decl/*.pl)" && ! ls internal/infer/testdata/lowered/decl/*.pl | sed 's|.*/lowered/||' | grep -vxFf internal/infer/testdata/lowered/matching.txt`)
- [ ] Every literal fixture matches its legacy golden on the lowered path (`test -n "$(ls internal/infer/testdata/lowered/literals/*.pl)" && ! ls internal/infer/testdata/lowered/literals/*.pl | sed 's|.*/lowered/||' | grep -vxFf internal/infer/testdata/lowered/matching.txt`)

### Negative Scenarios
- [ ] A Hole is typed Any, nothing inspects it, and analysis of input perl rejects does not panic (`go test ./internal/infer/ -run '^TestHoleIsTypedAnyAndNeverInspected$' -count=1 -v | grep -q '^--- PASS: TestHoleIsTypedAnyAndNeverInspected'`)
