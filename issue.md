---
title: "Lower Slice{Kind}: the five-way switch §4.14.1 rule 5 specifies"
state: pending
urgency: normal
milestone: m2-lower-the-tree
created: 2026-09-20T23:05:41.330720998Z
updated: 2026-09-20T23:05:41.330720998Z
---

## Context

- paths: `internal/lower/slice.go` (new), `internal/lower/slice_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestLowerSliceKind$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.1 rule 5

§4.14.2 calls `Slice{Kind}` "one five-way switch, and the largest single block
of this list the lowering gets for free". It is carved out of the variables
issue for that reason: it is self-contained, and it is the clearest measurable
win in the milestone.

**Why the five kinds must not collapse.** They produce four different result
types (§4.4.7). Re-measured on perl 5.42.0: `@a[0,2]` gives elements,
`@h{qw(x y)}` gives values, `%a[0,1]` gives index/element pairs, `%h{'x'}` gives
key/value pairs. Collapsing them to one "slice" loses exactly the information
PSC needs most.

**The CST already carries the discriminant.** Every slice is `Kind: Index` —
the sole `Kind: Index` construction site is `parseSubscript` in
`internal/parse/expr.go` — and the pair `(base child's first byte, Index.Text)`
selects the kind. Verified at HEAD:

	@a[0,1]     index "["  over term "@a"     -> SliceArray
	@h{'a','b'} index "{"  over term "@h"     -> SliceHash
	%a[0,1]     index "["  over term "%a"     -> SliceKVArray
	%h{'a'}     index "{"  over term "%h"     -> SliceKVHash
	(1,2)[0]    index "["  over a list        -> SliceList

Nothing has to be re-derived from source text; the discriminant is two field
reads.

**The defect this removes, measured.** The old grammar's `slice_expression`,
`slice_container_variable` and `anonymous_slice_expression` appear NOWHERE in
`internal/infer`, and the consequence is measurable: `my @a=(1,2); my @s=@a[0,1];`
gives `@a elem=Int` but `@s elem=Unknown`. Inference branches only two ways on
this today (KV or not, because only the element type is at stake) and does not
branch at all yet. The other three values serve the LSP and diagnostics.

**Do not follow B::SoN here.** It maps `aslice`, `kvaslice`, `hslice`, `kvhslice`
all to one `Slice` (`FromOptree/OpMap.pm:209-231`) and `aelem`, `helem`,
`multideref`, `gelem` all to one `Subscript` (`:205-208`), because the optree it
reads has already typed the result. §4.14.1 rule 5 is explicit: that is a phase
difference, not a disagreement — PSC is upstream of it and must not follow.

The `SliceList` case matters for a second reason: §4.14.1 rule 1 records that
`(1,2)[0] x 3` is an `index` over a `list`, not a parenthesised operand, and
perl agrees (`my @a = ((1,2)[0]) x 3` gives `1 1 1`). Getting `SliceList` right
here keeps the Paren rule honest in the operators issue.

## Steps

- [ ] Write `TestLowerSliceKind`: a five-row table, one per `SliceKind`, each
      asserting the lowered node's `Kind` field from the fragments above.
- [ ] Run to verify failure: `go test ./internal/lower/ -run "^TestLowerSliceKind$" -count=1 -v`
- [ ] Implement the five-way switch on `(base sigil, Index.Text)`. Run to green.
      Commit.
- [ ] Write `TestLowerSliceIsNotSubscript`: `@a[0]` lowers to a `Slice`, and
      `$a[0]` lowers to an `Index` — the sigil on the base is what separates
      them, not the bracket. Run to verify failure, implement, run to green,
      commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] All five SliceKind values are produced from their fragments (`go test ./internal/lower/ -run "^TestLowerSliceKind$" -count=1 -v | grep -q "^--- PASS: TestLowerSliceKind"`)
- [ ] A slice and a single-element subscript lower to different kinds (`go test ./internal/lower/ -run "^TestLowerSliceIsNotSubscript$" -count=1 -v | grep -q "^--- PASS: TestLowerSliceIsNotSubscript"`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)
