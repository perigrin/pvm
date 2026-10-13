---
title: "Port the special-form rules onto the lowered tree: MapGrepSort.Op, the eval boundary, and match captures"
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a12a85-32cf-7d76-96f7-b814935374e1
blocks:
- 01a12a85-3417-75eb-bf44-d47cf6546e82
created: 2026-10-11T10:32:14.982514503Z
updated: 2026-10-13T21:46:52.057347158Z
sessions:
- start_sha: a45799887c9831e70d88981c993e14c2f1de3741
  end_sha: ""
  commits: 0
  started_at: 2026-10-13T21:46:52.057347158Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-13T21:46:52.057347158Z
---

## Context

- paths: `internal/infer/lowered.go`, `internal/infer/testdata/lowered/special/` (new)
- commands: `go test ./internal/infer/ -run "^TestLoweredGoldenRatchet$" -count=1 -v`
- entrypoints: `callElementType` (`infer.go:1307`, the `HasPrefix` test at `:1315`), `blockResultType` (`:1422`), `nearestPatternBefore` (`:1726`), `sprintfResultType` (`:1558`)

The last of seven port issues. After it, every fixture under
`testdata/lowered/` matches its golden on the lowered path.

What gets simpler, and what review should check:

- `callElementType`'s `strings.HasPrefix(..., "grep")` (`infer.go:1315`)
  becomes a switch on `MapGrepSort.Op`;
- `nearestPatternBefore` asks the position map for "the last `Match` before
  this offset", which 01a0c113-dfd3 built for exactly this;
- `EvalStr` is the soundness boundary: its result is `Any`. Legacy has no
  rule for it, so a fixture whose golden differs here is a fix, with a
  `Diagnostics-changed:` trailer.

The method is the one "AnalyzeLowered: the entry point, the golden
ratchet, and the declaration and literal rules" set up: fixtures
under `testdata/lowered/special/`, goldens captured from the legacy path, and
`TestLoweredGoldenRatchet` holding `matching.txt` in both directions. Every
rule is REWRITTEN against the lowered vocabulary, not shimmed. No dataflow
IR (§6.1.6): if a rule wants flow, it keeps the approximation it has today.

Rollback: this issue changes only `AnalyzeLowered` and test data. Nothing in `internal/psc` calls it yet, so reverting it changes nothing PSC reports.

## Steps

- [ ] Write fixtures under `testdata/lowered/special/`, drawn from the
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
      ratchet to green. Every `special/` fixture must match. One that cannot
      is a legacy defect or a regression: triage it before moving on. If it
      is a legacy defect, edit its golden in its own commit with a
      `Diagnostics-changed:` trailer naming the cause.
- [ ] Write `TestLoweredSpecialFormsOverBrokenInput`: `my @b = map {;` and
      `my $r = eval {;` (perl rejects both) analyse without a panic, and the
      `Hole` block gives `Any`, not a guessed element type. Run to verify
      failure, implement, run to green, commit.
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
- [ ] Every special fixture matches its legacy golden on the lowered path (`test -n "$(ls internal/infer/testdata/lowered/special/*.pl)" && ! ls internal/infer/testdata/lowered/special/*.pl | sed 's|.*/lowered/||' | grep -vxFf internal/infer/testdata/lowered/matching.txt`)
- [ ] Every fixture in the tree matches its golden on the lowered path (`test "$(ls internal/infer/testdata/lowered/*/*.pl | wc -l)" -gt 0 && test "$(ls internal/infer/testdata/lowered/*/*.pl | wc -l)" -eq "$(grep -c . internal/infer/testdata/lowered/matching.txt)"`)

### Negative Scenarios
- [ ] A map or eval block perl rejects analyses without a panic and yields Any (`go test ./internal/infer/ -run '^TestLoweredSpecialFormsOverBrokenInput$' -count=1 -v | grep -q '^--- PASS: TestLoweredSpecialFormsOverBrokenInput'`)
