---
title: Parse the inner { of cpan_ceph_tie_runtime.t's `map { { $_->get_columns } } @x` as perl does
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-24T13:37:12.332914405Z
updated: 2026-10-24T13:37:12.332914405Z
---

Found by the narrowing worker (01a16c1b-a587) on 2026-10-24. perigrin ruled that a `{` psc misreads as a block is a parser defect, to be fixed at the root.

## Context

- paths: `internal/parse/` (where the parser decides whether a `{` at statement start or in a block body opens a block)
- entrypoints: the parser's choice between a block and an anonymous hash at a `{`

T1 file cpan_ceph_tie_runtime.t contains `map { { $_->get_columns } } @x`: a map block whose body is `{ $_->get_columns }`. psc parses the inner `{` as a bare block, so canon writes `{...;}`, and perl rejects that with `syntax error near ";}"`. The file still counts as faithful, because the tree itself says Block.

Measure toke.c's guess with perl 5.44 (Concise) on this line and on reductions of it, then make the parser agree.

## Acceptance criteria

### Positive Scenarios

- `go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 | grep -q '^ok'` (a probe row for `map { { $_->get_columns } } @x` with perl's tree, whose canon compiles under perl)
