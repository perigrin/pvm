---
title: "Port the special-form rules onto the lowered tree: MapGrepSort.Op, the eval boundary, and match captures"
state: done
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a12a85-32cf-7d76-96f7-b814935374e1
blocks:
- 01a12a85-3417-75eb-bf44-d47cf6546e82
created: 2026-10-11T10:32:14.982514503Z
updated: 2026-10-14T02:46:26.76369764Z
sessions:
- start_sha: a45799887c9831e70d88981c993e14c2f1de3741
  end_sha: 26f0e9e5343e8d6b9487deedf0564dfef67acab5
  commits: 1
  started_at: 2026-10-13T21:46:52.057347158Z
  ended_at: 2026-10-14T02:46:26.76369764Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-13T21:46:52.057347158Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-14T02:46:26.76369764Z
observed_paths:
- internal/infer/lowered.go
- internal/infer/lowered_ops.go
- internal/infer/lowered_special.go
- internal/infer/lowered_test.go
- internal/infer/lowered_vars.go
- internal/infer/testdata/lowered/matching.txt
- internal/infer/testdata/lowered/special/capture_arith.golden
- internal/infer/testdata/lowered/special/capture_arith.pl
- internal/infer/testdata/lowered/special/capture_digits.golden
- internal/infer/testdata/lowered/special/capture_digits.pl
- internal/infer/testdata/lowered/special/capture_letters.golden
- internal/infer/testdata/lowered/special/capture_letters.pl
- internal/infer/testdata/lowered/special/capture_nearest.golden
- internal/infer/testdata/lowered/special/capture_nearest.pl
- internal/infer/testdata/lowered/special/capture_none.golden
- internal/infer/testdata/lowered/special/capture_none.pl
- internal/infer/testdata/lowered/special/capture_second.golden
- internal/infer/testdata/lowered/special/capture_second.pl
- internal/infer/testdata/lowered/special/do_block.golden
- internal/infer/testdata/lowered/special/do_block.pl
- internal/infer/testdata/lowered/special/do_file.golden
- internal/infer/testdata/lowered/special/do_file.pl
- internal/infer/testdata/lowered/special/eval_block.golden
- internal/infer/testdata/lowered/special/eval_block.pl
- internal/infer/testdata/lowered/special/eval_string.golden
- internal/infer/testdata/lowered/special/eval_string.pl
- internal/infer/testdata/lowered/special/filetest.golden
- internal/infer/testdata/lowered/special/filetest.pl
- internal/infer/testdata/lowered/special/glob_angle.golden
- internal/infer/testdata/lowered/special/glob_angle.pl
- internal/infer/testdata/lowered/special/grep_block.golden
- internal/infer/testdata/lowered/special/grep_block.pl
- internal/infer/testdata/lowered/special/grep_expr.golden
- internal/infer/testdata/lowered/special/grep_expr.pl
- internal/infer/testdata/lowered/special/grep_scalar.golden
- internal/infer/testdata/lowered/special/grep_scalar.pl
- internal/infer/testdata/lowered/special/map_block.golden
- internal/infer/testdata/lowered/special/map_block.pl
- internal/infer/testdata/lowered/special/map_expr.golden
- internal/infer/testdata/lowered/special/map_expr.pl
- internal/infer/testdata/lowered/special/map_last_stmt.golden
- internal/infer/testdata/lowered/special/map_last_stmt.pl
- internal/infer/testdata/lowered/special/map_scalar.golden
- internal/infer/testdata/lowered/special/map_scalar.pl
- internal/infer/testdata/lowered/special/map_str.golden
- internal/infer/testdata/lowered/special/map_str.pl
- internal/infer/testdata/lowered/special/match_bare.golden
- internal/infer/testdata/lowered/special/match_bare.pl
- internal/infer/testdata/lowered/special/match_list.golden
- internal/infer/testdata/lowered/special/match_list.pl
- internal/infer/testdata/lowered/special/match_scalar.golden
- internal/infer/testdata/lowered/special/match_scalar.pl
- internal/infer/testdata/lowered/special/qr.golden
- internal/infer/testdata/lowered/special/qr.pl
- internal/infer/testdata/lowered/special/readline.golden
- internal/infer/testdata/lowered/special/readline.pl
- internal/infer/testdata/lowered/special/ref_arith.golden
- internal/infer/testdata/lowered/special/ref_arith.pl
- internal/infer/testdata/lowered/special/ref_array.golden
- internal/infer/testdata/lowered/special/ref_array.pl
- internal/infer/testdata/lowered/special/ref_code.golden
- internal/infer/testdata/lowered/special/ref_code.pl
- internal/infer/testdata/lowered/special/ref_glob.golden
- internal/infer/testdata/lowered/special/ref_glob.pl
- internal/infer/testdata/lowered/special/ref_hash.golden
- internal/infer/testdata/lowered/special/ref_hash.pl
- internal/infer/testdata/lowered/special/ref_other.golden
- internal/infer/testdata/lowered/special/ref_other.pl
- internal/infer/testdata/lowered/special/ref_paren.golden
- internal/infer/testdata/lowered/special/ref_paren.pl
- internal/infer/testdata/lowered/special/ref_scalar.golden
- internal/infer/testdata/lowered/special/ref_scalar.pl
- internal/infer/testdata/lowered/special/sort_block.golden
- internal/infer/testdata/lowered/special/sort_block.pl
- internal/infer/testdata/lowered/special/sort_list.golden
- internal/infer/testdata/lowered/special/sort_list.pl
- internal/infer/testdata/lowered/special/sort_plain.golden
- internal/infer/testdata/lowered/special/sort_plain.pl
- internal/infer/testdata/lowered/special/subst.golden
- internal/infer/testdata/lowered/special/subst.pl
- internal/infer/testdata/lowered/special/tr.golden
- internal/infer/testdata/lowered/special/tr.pl
- internal/infer/testdata/lowered/special/wantarray.golden
- internal/infer/testdata/lowered/special/wantarray.pl
- internal/infer/testdata/lowered/special/wantarray_top.golden
- internal/infer/testdata/lowered/special/wantarray_top.pl
- internal/lower/positions.go
- internal/lower/positions_test.go
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
- [ ] `Ref` (`\@a`, `\%h`, `\&f`, `\$x`): type it from what the reference is taken TO, as legacy `inferRefgenType` (infer.go) does; `special/` fixtures for each (added 2026-10-13: the calls port found no port claimed Ref)

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
- [ ] Every fixture in the tree matches its golden on the lowered path (`test "$(ls internal/infer/testdata/lowered/*/*.pl | wc -l)" -gt 0 && test "$(ls internal/infer/testdata/lowered/*/*.pl | wc -l)" -eq "$(grep -c '^[^#]' internal/infer/testdata/lowered/matching.txt)"`)

### Negative Scenarios
- [ ] A map or eval block perl rejects analyses without a panic and yields Any (`go test ./internal/infer/ -run '^TestLoweredSpecialFormsOverBrokenInput$' -count=1 -v | grep -q '^--- PASS: TestLoweredSpecialFormsOverBrokenInput'`)
