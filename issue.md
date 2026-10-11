---
title: "Node identity: stable ids for internal/lower that survive an edit above them"
state: in-progress
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c10f-37f8-7a43-b3ce-ea194f7defe0
created: 2026-09-20T23:02:46.558435832Z
updated: 2026-10-11T10:36:37.259457461Z
sessions:
- start_sha: 971b903102dc4a335660172f3b1b5ecad5cf3330
  end_sha: ""
  commits: 0
  started_at: 2026-10-11T10:36:37.259457461Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-11T10:36:37.259457461Z
---

## Context

- paths: `internal/lower/id.go` (new), `internal/lower/id_test.go` (new)
- commands: `go test ./internal/lower/ -run "^TestIDStableUnderEditAbove$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/06-incremental-lsp.md` §6.1.6 item 4, §6.4.3

This is the highest-risk item in the milestone and it has no dependencies, so
it starts first. §6.1.6 item 4 is binding: **facts are keyed by node identity,
never by byte offset.** rust-analyzer rejects offset keys by name, and §6.4.3
records the same defect in PSC today: `infer.go:126` builds a
`map[uint32]types.Type` keyed on `StartByte`, so every edit invalidates the
whole map, and the write at `infer.go:185` lets a parent written after its
child at the same offset overwrite the child.

**The test that separates a real scheme from a fake one**: insert one
statement at the TOP of a two-sub file, re-parse, re-assign ids, and assert
that no id inside either sub changed. Byte offset fails it. Preorder index
fails it. Both are the obvious implementations, so the test exists to reject
them.

The shape §6.1.6 item 4 names is `crates/span/src/ast_id.rs`: an enumeration
index by kind, name and disambiguator. Ids are allocated per enclosing
declaration, so the counter inside `sub b` restarts and does not see what
happened above it. Rule 6 of §4.14.1 constrains the meaning: identity is per
SYNTAX SITE. One id for one `sub { $i }`, even though three loop iterations
create three cells. A table keyed on a captured value keys on
`(node id, capture instance)`, not on node id alone. That is a
scope-resolution concern and no cell node exists in §4.14, so nothing here
models it, but the id must not be defined in a way that forecloses it.

The allocator reads `internal/parse` trees, and those trees are often broken.
At 24bfac4a an Unknown can even have children: `if ($x) {` gives
`[unknown refusal=unclosed_opener [block]]`. The allocator must assign ids
over such a tree and must not panic.

Scope: `internal/lower/id.go` only. No node types, no lowering rules. It ships
the allocator and its invariant tests.

## Steps

- [ ] Write `TestIDStableUnderEditAbove`: parse `sub a { my $x = 1; } sub b { my $y = 2; }`,
      allocate ids, record them; parse `my $z = 0;` followed by the same
      text, allocate again, and assert every id inside `sub a` and `sub b`
      is unchanged.
- [ ] Run to verify failure (the package does not exist yet):
      `go test ./internal/lower/ -run "^TestIDStableUnderEditAbove$" -count=1 -v`
- [ ] Implement the allocator in `internal/lower/id.go`: per-declaration
      enumeration by kind, name and disambiguator, in the manner of
      `ast_id.rs`. Run to green. Commit.
- [ ] Write `TestIDDistinguishesSameSpanNodes`: parse
      `my $x = 5; my @y = ($x); my $z = $y[0];`, assert that nodes sharing a
      start byte get DIFFERENT ids, and that a map keyed by id holds one
      entry per node. Run to verify failure, implement, run to green, commit.
- [ ] Write `TestIDDistinguishesSiblingsWithSameShape`: two identical
      statements in one sub (`my $a = 1; my $a = 1;`) get different ids.
      Run to verify failure, implement, run to green, commit.
- [ ] Write `TestIDOverBrokenTree`: allocate ids over `if ($x) {` and
      `sub a { my $x = ; } sub b { 1 }`. Assert no panic, an id for every
      node including the Unknown, and that the ids inside `sub b` match
      those from a clean `sub b { 1 }` at the same position. Run to verify
      failure, implement, run to green, commit.
- [ ] Run the package suite: `go test ./internal/lower/ -count=1`
- [ ] Commit.

## Acceptance Criteria

### Positive Scenarios
- [ ] An edit above a sub leaves every id inside it unchanged (`go test ./internal/lower/ -run '^TestIDStableUnderEditAbove$' -count=1 -v | grep -q '^--- PASS: TestIDStableUnderEditAbove'`)
- [ ] Nodes that share a start byte get distinct ids (`go test ./internal/lower/ -run '^TestIDDistinguishesSameSpanNodes$' -count=1 -v | grep -q '^--- PASS: TestIDDistinguishesSameSpanNodes'`)
- [ ] Structurally identical siblings get distinct ids (`go test ./internal/lower/ -run '^TestIDDistinguishesSiblingsWithSameShape$' -count=1 -v | grep -q '^--- PASS: TestIDDistinguishesSiblingsWithSameShape'`)
- [ ] The package is green (`go test ./internal/lower/ -count=1`)

### Negative Scenarios
- [ ] A broken tree, an Unknown with children included, gets ids without a panic, and the broken sub does not disturb the ids of the sub after it (`go test ./internal/lower/ -run '^TestIDOverBrokenTree$' -count=1 -v | grep -q '^--- PASS: TestIDOverBrokenTree'`)
