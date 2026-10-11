---
title: Swap the annotation key from StartByte to node identity
state: cancelled
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a0c10e-c81e-7565-88fb-d973f18ed667
blocks:
- 01a0c115-a44b-7393-a828-ad29844b20b4
created: 2026-09-20T23:10:09.105095184Z
updated: 2026-10-11T10:19:36.38583883Z
transitions:
- state: cancelled
  actor: human:git-zhi
  timestamp: 2026-10-11T10:19:36.38583883Z
---

## Cancelled (2026-10-11), perigrin's ruling

"Cancel: keyed by id from the start." Superseded: the lowered path is keyed by
node identity from its first issue, so no StartByte-keyed annotation is ever
swapped; the legacy path is deleted by 01a0c115-df67. Found by the chain review
of 2026-10-11: as filed this issue could not be done, since ids exist only on
the lowered tree, which first appears in the cutover it was ordered before.

## Context

- paths: `internal/infer/infer.go` (lines 57-60, 184, and the annotation reads), `internal/infer/project.go:30`, `internal/psc/lsp.go`, `internal/infer/annotations_test.go` (new)
- commands: `go test ./internal/infer/ -run "^TestAnnotationsKeyedByIdentity$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/06-incremental-lsp.md` §6.4.3; §6.1.6 item 4

§6.4.3 names this defect precisely: PSC's annotation map is keyed by `StartByte`
(`infer.go:57-60`, `map[uint32]types.Type`). **Byte-keyed annotations are
invalidated wholesale by any edit, because every key after the edit shifts.**
rust-analyzer's contributor guide rejects this key by name and states the
invariant: "syntax tree is a value type ... Using the tree as a store for
semantic info is convenient in traditional compilers, but doesn't work nicely in
the IDE."

There is a second, sharper defect at `infer.go:184`, and it is a correctness bug
rather than a performance one. The write is:

	if typ != types.Unknown { annotations[node.StartByte()] = typ }

so a PARENT written after its CHILD at the same offset overwrites the child.
Measured on `my $x = 5; my @y = ($x); my $z = $y[0];`, byte 0 carries four
nodes and the `number` inside `(1..5)` reads back as its parent's type.

This issue swaps the key to the node identity from `internal/lower/id.go`. It is
parallel to the lowering-rule slices and joins them at the cutover.

**The second site is `internal/infer/project.go:30`** — the same byte key in the
cross-file cache. §6.4.3's exception applies there and must be preserved:
`ProjectIndex` results are cached per file path and **must survive a keystroke
in an unrelated buffer.** Invalidate a `ProjectIndex` entry only when THAT
file's content changes. Do not let the key swap quietly widen that invalidation.

**Explicitly out of scope, per §6.4.3: do NOT try to incrementally update the
annotation map.** Re-run `Analyze` on the whole tree after each parse.
Justification, measured: `Analyze` is a bottom-up walk over an in-memory tree
with no I/O, 0.5-2 ms for 40,000 nodes, against a parse of 6-500 ms on the same
files. It fits the 10 ms p95 budget with room. If it ever stops fitting, the fix
is to make `Analyze` incremental per-sub — a much later project — not to patch a
byte-keyed map. An implementation here that adds a `DamageSet` or any
invalidation machinery is out of scope.

The public signature of `Analyze`/`AnalyzeWithOptions` changes with the key
type, which is why the two `internal/psc` callers (`lsp.go:53`,
`check_command.go:94`) are named above: `lsp.go` also reads the map by byte for
`DefinitionAtByte` and hover, and that read goes through the position map from
the position-map issue rather than through the annotation key.

## Steps

- [ ] Write `TestAnnotationsKeyedByIdentity`: analyse
      `my $x = 5; my @y = ($x); my $z = $y[0];` and assert the map holds a
      distinct entry per node rather than collapsing the four nodes at byte 0 —
      the `infer.go:184` defect, asserted directly.
- [ ] Run to verify failure: `go test ./internal/infer/ -run "^TestAnnotationsKeyedByIdentity$" -count=1 -v`
- [ ] Swap the key type. Run to green. Commit.
- [ ] Write `TestAnnotationsSurviveEditAbove`: analyse a two-sub file, insert a
      statement at the top, re-analyse, and assert the annotations inside both
      subs are keyed the same as before — the edit did not shift them.
      Run to verify failure, implement, run to green, commit.
- [ ] Write `TestProjectIndexInvalidatesPerFile`: a `ProjectIndex` entry for
      file A survives a change to file B and is dropped by a change to A —
      §6.4.3's exception, preserved across the key swap.
      Run to verify failure, implement, run to green, commit.
- [ ] Run the infer suite: `go test ./internal/infer/ -count=1`
- [ ] Run the psc suite, since the callers see the signature change:
      `go test ./internal/psc/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] Nodes sharing a start byte get distinct annotations (`go test ./internal/infer/ -run "^TestAnnotationsKeyedByIdentity$" -count=1 -v | grep -q "^--- PASS: TestAnnotationsKeyedByIdentity"`)
- [ ] Annotation keys survive an edit above the annotated nodes (`go test ./internal/infer/ -run "^TestAnnotationsSurviveEditAbove$" -count=1 -v | grep -q "^--- PASS: TestAnnotationsSurviveEditAbove"`)
- [ ] ProjectIndex invalidation stays per-file (`go test ./internal/infer/ -run "^TestProjectIndexInvalidatesPerFile$" -count=1 -v | grep -q "^--- PASS: TestProjectIndexInvalidatesPerFile"`)
- [ ] The infer package is green (`go test ./internal/infer/ -count=1`)
- [ ] The psc package still builds and passes against the new signature (`go test ./internal/psc/ -count=1`)
