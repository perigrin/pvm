---
title: Port the variable, subscript and slice rules onto the lowered tree, with identity-keyed annotations
state: done
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a12a85-3100-79ae-8265-aa8b4f39cd85
- 01a0c111-72d3-70fa-88a8-a9a387f04261
blocks:
- 01a12a85-3233-7ec8-9428-747be4ccef1c
created: 2026-10-11T10:32:14.500354268Z
updated: 2026-10-13T02:37:42.519224106Z
sessions:
- start_sha: de501a52304e0902ef40141ecc5d555442b0866b
  end_sha: 25ec742454544ea6db09a7d8ace24e430966243f
  commits: 1
  started_at: 2026-10-13T01:56:52.092817024Z
  ended_at: 2026-10-13T02:37:42.519224106Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-13T01:56:52.092817024Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-13T02:37:42.519224106Z
observed_paths:
- internal/infer/annotations_test.go
- internal/infer/lowered.go
- internal/infer/lowered_test.go
- internal/infer/lowered_vars.go
- internal/infer/testdata/lowered/matching.txt
- internal/infer/testdata/lowered/vars/array_copy.golden
- internal/infer/testdata/lowered/vars/array_copy.pl
- internal/infer/testdata/lowered/vars/array_count.golden
- internal/infer/testdata/lowered/vars/array_count.pl
- internal/infer/testdata/lowered/vars/arraylen.golden
- internal/infer/testdata/lowered/vars/arraylen.pl
- internal/infer/testdata/lowered/vars/arrow_array.golden
- internal/infer/testdata/lowered/vars/arrow_array.pl
- internal/infer/testdata/lowered/vars/arrow_hash.golden
- internal/infer/testdata/lowered/vars/arrow_hash.pl
- internal/infer/testdata/lowered/vars/deref_scalar.golden
- internal/infer/testdata/lowered/vars/deref_scalar.pl
- internal/infer/testdata/lowered/vars/elem_from_var.golden
- internal/infer/testdata/lowered/vars/elem_from_var.pl
- internal/infer/testdata/lowered/vars/element.golden
- internal/infer/testdata/lowered/vars/element.pl
- internal/infer/testdata/lowered/vars/glob.golden
- internal/infer/testdata/lowered/vars/glob.pl
- internal/infer/testdata/lowered/vars/hash_element.golden
- internal/infer/testdata/lowered/vars/hash_element.pl
- internal/infer/testdata/lowered/vars/hash_scalar.golden
- internal/infer/testdata/lowered/vars/hash_scalar.pl
- internal/infer/testdata/lowered/vars/hash_slice.golden
- internal/infer/testdata/lowered/vars/hash_slice.pl
- internal/infer/testdata/lowered/vars/index_int.golden
- internal/infer/testdata/lowered/vars/index_int.pl
- internal/infer/testdata/lowered/vars/index_ref.golden
- internal/infer/testdata/lowered/vars/index_ref.pl
- internal/infer/testdata/lowered/vars/interp.golden
- internal/infer/testdata/lowered/vars/interp.pl
- internal/infer/testdata/lowered/vars/key_ref.golden
- internal/infer/testdata/lowered/vars/key_ref.pl
- internal/infer/testdata/lowered/vars/kv_slice.golden
- internal/infer/testdata/lowered/vars/kv_slice.pl
- internal/infer/testdata/lowered/vars/list_decl_aggregate.golden
- internal/infer/testdata/lowered/vars/list_decl_aggregate.pl
- internal/infer/testdata/lowered/vars/list_slice.golden
- internal/infer/testdata/lowered/vars/list_slice.pl
- internal/infer/testdata/lowered/vars/mixed_element.golden
- internal/infer/testdata/lowered/vars/mixed_element.pl
- internal/infer/testdata/lowered/vars/nested_element.golden
- internal/infer/testdata/lowered/vars/nested_element.pl
- internal/infer/testdata/lowered/vars/our_element.golden
- internal/infer/testdata/lowered/vars/our_element.pl
- internal/infer/testdata/lowered/vars/postfix_deref.golden
- internal/infer/testdata/lowered/vars/postfix_deref.pl
- internal/infer/testdata/lowered/vars/prefix_deref.golden
- internal/infer/testdata/lowered/vars/prefix_deref.pl
- internal/infer/testdata/lowered/vars/scalar.golden
- internal/infer/testdata/lowered/vars/scalar.pl
- internal/infer/testdata/lowered/vars/slice.golden
- internal/infer/testdata/lowered/vars/slice.pl
- internal/infer/testdata/lowered/vars/underscore.golden
- internal/infer/testdata/lowered/vars/underscore.pl
- internal/infer/testdata/lowered/vars/unknown_container.golden
- internal/infer/testdata/lowered/vars/unknown_container.pl
---

## Context

- paths: `internal/infer/lowered.go`, `internal/infer/annotations_test.go` (new), `internal/infer/testdata/lowered/vars/` (new)
- commands: `go test ./internal/infer/ -run "^TestLoweredGoldenRatchet$" -count=1 -v`
- entrypoints: `inferElementType` (`infer.go:1262`), `checkElementIndex` (`:1201`), `aggregateElemType` (`:1402`), `preservedElemType` (`:1387`), `lookupNarrowedType` (`:2084`), `sigildName` (`symbols.go:403`)

The third of seven port issues. It ports element typing, index checks,
narrowed-type lookup and the sigil arms onto `Var`, `Deref`, `Index`, `Key`
and `Slice`.

What gets simpler, and what review should check:

- the 8-of-9 `scalar`/`array`/`hash` switch sites that differ only by sigil
  string become one arm reading `Var.Sigil`;
- the three `@_` / `$_[0]` text re-derivations compare `Var{Sigil,Name}`
  fields;
- deref-ness is the lowering's `Arrow || base is Deref` predicate, not the
  old grammar's `container_variable` split.

**Slices change a reported type, on purpose.** Legacy gives
`my @a=(1,2); my @s=@a[0,1];` `@s elem=Unknown`, because no slice grammar
name appears in `internal/infer`. If the port types it `Int`, that fixture's
golden changes, and the commit carries a `Diagnostics-changed:` trailer
saying so.

**The identity tests from 01a0c115-88d1 land here**, now that their fixtures
type on the lowered path:

- `TestAnnotationsKeyedByIdentity`: on
  `my $x = 5; my @y = ($x); my $z = $y[0];`, nodes sharing a start byte get
  distinct annotations. This is the `infer.go:185` overwrite defect,
  asserted directly.
- `TestAnnotationsSurviveEditAbove`: analyse
  `sub a { my $x = 1; } sub b { my $y = 2; }`, insert a statement at the
  top, re-analyse, and assert the annotation keys inside both subs are
  unchanged.

88d1's third test (`ProjectIndex` invalidates per file) is not carried. The
cache at `project.go:62-97` is keyed by path and never invalidated, so
there is no per-file invalidation to preserve; see 01a0c115-a44b for the
`ProjectIndex` flip.

The method is the one "AnalyzeLowered: the entry point, the golden
ratchet, and the declaration and literal rules" set up: fixtures
under `testdata/lowered/vars/`, goldens captured from the legacy path, and
`TestLoweredGoldenRatchet` holding `matching.txt` in both directions. Every
rule is REWRITTEN against the lowered vocabulary, not shimmed. No dataflow
IR (§6.1.6): if a rule wants flow, it keeps the approximation it has today.

Rollback: this issue changes only `AnalyzeLowered` and test data. Nothing in `internal/psc` calls it yet, so reverting it changes nothing PSC reports.

## Steps

- [ ] Write fixtures under `testdata/lowered/vars/`, drawn from the
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
      ratchet to green. Every `vars/` fixture must match. One that cannot
      is a legacy defect or a regression: triage it before moving on. If it
      is a legacy defect, edit its golden in its own commit with a
      `Diagnostics-changed:` trailer naming the cause.
- [ ] Write `TestAnnotationsKeyedByIdentity` and
      `TestAnnotationsSurviveEditAbove` as described above. Run to verify
      failure, implement, run to green, commit.
- [ ] Write `TestLoweredSubscriptOverBrokenInput`: `my @a = (1); my $e = $a[;`
      and `my %h; my $v = $h{;` (perl rejects both) analyse without a panic,
      emit no index diagnostic for the `Hole` subscript, and leave `@a`
      typed. Run to verify failure, implement, run to green, commit.
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
- [ ] Every vars fixture matches its legacy golden on the lowered path (`test -n "$(ls internal/infer/testdata/lowered/vars/*.pl)" && ! ls internal/infer/testdata/lowered/vars/*.pl | sed 's|.*/lowered/||' | grep -vxFf internal/infer/testdata/lowered/matching.txt`)
- [ ] Nodes sharing a start byte get distinct annotations (`go test ./internal/infer/ -run '^TestAnnotationsKeyedByIdentity$' -count=1 -v | grep -q '^--- PASS: TestAnnotationsKeyedByIdentity'`)
- [ ] Annotation keys inside a sub survive an edit above it (`go test ./internal/infer/ -run '^TestAnnotationsSurviveEditAbove$' -count=1 -v | grep -q '^--- PASS: TestAnnotationsSurviveEditAbove'`)

### Negative Scenarios
- [ ] A subscript perl rejects analyses without a panic and draws no index diagnostic (`go test ./internal/infer/ -run '^TestLoweredSubscriptOverBrokenInput$' -count=1 -v | grep -q '^--- PASS: TestLoweredSubscriptOverBrokenInput'`)
