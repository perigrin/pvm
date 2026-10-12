---
title: "Lower Slice{Kind}: the five-way switch §4.14.1 rule 5 specifies"
state: done
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c110-839f-77fb-8a3d-30fa5a8d6e96
blocks:
- 01a12a85-31a5-769d-88c4-e2add3f8fefc
created: 2026-09-20T23:05:41.330720998Z
updated: 2026-10-12T18:27:52.851017143Z
sessions:
- start_sha: 34b03d318ec7428fe87316e7752b9ef1f1f6164e
  end_sha: 55ee0ff03a881b4415d9e0444d6111d11d76ccc6
  commits: 3
  started_at: 2026-10-12T17:10:07.74618055Z
  ended_at: 2026-10-12T18:27:52.851017143Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-12T17:10:07.74618055Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-12T18:27:52.851017143Z
observed_paths:
- docs/specs/perl-parser/04-expressions.md
- internal/lower/decl.go
- internal/lower/decl_test.go
- internal/lower/expr.go
- internal/lower/id.go
- internal/lower/id_test.go
- internal/lower/node.go
- internal/lower/slice.go
- internal/lower/slice_test.go
- internal/lower/vars_test.go
---

## Context

- paths: `internal/lower/slice.go` (new), `internal/lower/slice_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerSliceKind$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.1 rule 5

§4.14.2 calls `Slice{Kind}` "one five-way switch, and the largest single block
of this list the lowering gets for free". It is carved out of the variables
issue because it is self-contained and the clearest measurable win in the
milestone.

**Why the five kinds must not collapse.** They produce four different result
types (§4.4.7). On perl 5.42.0: `@a[0,2]` gives elements, `@h{qw(x y)}` gives
values, `%a[0,1]` gives index/element pairs, `%h{'x'}` gives key/value pairs.
Collapsing them to one "slice" loses exactly the information PSC needs most.

**The CST already carries the discriminant.** Every slice is `Kind: Index`,
and the pair `(base child's first byte, Index.Text)` selects the kind.
Measured at 24bfac4a; re-measure at execution time:

	@a[0,1]     index "["  over term "@a"     -> SliceArray
	@h{'a','b'} index "{"  over term "@h"     -> SliceHash
	%a[0,1]     index "["  over term "%a"     -> SliceKVArray
	%h{'a'}     index "{"  over term "%h"     -> SliceKVHash
	(1,2)[0]    index "["  over list paren    -> SliceList

A multi-element subscript is one `binary ","` child (`@a[0,1]` gives
`[index "[" [term "@a"] [binary "," [term "0"] [term "1"]]]`). `Slice.Idx`
holds it as lowered; flattening it is not this issue's job.

Nothing is re-derived from source text. The discriminant is two field reads.

**The defect this removes.** The old grammar's `slice_expression`,
`slice_container_variable` and `anonymous_slice_expression` appear nowhere in
`internal/infer`. `my @a=(1,2); my @s=@a[0,1];` gives `@a elem=Int` but
`@s elem=Unknown`. Inference branches two ways on this (KV or not), because
only the element type is at stake. The other three values serve the LSP and
diagnostics.

**Do not follow B::SoN here.** It maps `aslice`, `kvaslice`, `hslice`,
`kvhslice` all to one `Slice` (`FromOptree/OpMap.pm:209-231`), because the
optree it reads has already typed the result. §4.14.1 rule 5: that is a phase
difference. PSC is upstream of it and must not follow.

`SliceList` matters for a second reason: §4.14.1 rule 1 records that
`(1,2)[0] x 3` is an `index` over a `list`, not a parenthesised operand, and
perl agrees (`my @a = ((1,2)[0]) x 3` gives `1 1 1`).

## Steps

- [ ] Write `TestLowerSliceKind`: a five-row table, one per `SliceKind`,
      each asserting the lowered node's `Kind` from the fragments above.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestLowerSliceKind$" -count=1 -v`
- [ ] Implement the five-way switch on `(base sigil, Index.Text)`. Run to
      green. Commit.
- [ ] Write `TestLowerSliceIsNotSubscript`: `@a[0]` lowers to a `Slice`, and
      `$a[0]` to an `Index`. The sigil on the base separates them, not the
      bracket. Run to verify failure, implement, run to green, commit.
- [ ] Write `TestLowerSliceOverBrokenInput`: `@a[;` and `%h{;` (perl rejects
      both) lower to a `Slice` of the right kind whose `Idx` is a `Hole`.
      No panic. Run to verify failure, implement, run to green, commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] All five SliceKind values are produced from their fragments (`go test ./internal/lower/ -run '^TestLowerSliceKind$' -count=1 -v | grep -q '^--- PASS: TestLowerSliceKind'`)
- [ ] A slice and a single-element subscript lower to different kinds (`go test ./internal/lower/ -run '^TestLowerSliceIsNotSubscript$' -count=1 -v | grep -q '^--- PASS: TestLowerSliceIsNotSubscript'`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)

### Negative Scenarios
- [ ] A slice subscript perl rejects lowers to a Hole and the slice keeps its kind (`go test ./internal/lower/ -run '^TestLowerSliceOverBrokenInput$' -count=1 -v | grep -q '^--- PASS: TestLowerSliceOverBrokenInput'`)
