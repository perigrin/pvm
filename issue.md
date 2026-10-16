---
title: my @a = (1) types the array as Int, as legacy did
state: in-progress
urgency: normal
milestone: v0.1
created: 2026-10-13T02:30:22.789635766Z
updated: 2026-10-16T03:49:47.202700894Z
sessions:
- start_sha: 133837eb651f8dcffb4203d072b3701753486869
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T03:49:47.202700894Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T03:49:47.202700894Z
---

## Context

- paths: `internal/infer/lowered.go` (the decl rule), `internal/infer/lowered_vars.go`
- commands: `go test ./internal/infer/ -run '^TestLoweredGoldenRatchet$'`

Found 2026-10-13 by the variables port (01a12a85-31a5), which copied it from legacy so the goldens match: a one-element parenthesised initialiser types the aggregate as that element. `my @a = (1);` gives `sym @a Int`, `my @a = ($x)` gives `@a Str` (goldens `vars/arraylen`, `our_element`, `elem_from_var`). `(1)` lowers to a NumLit with Paren, and the decl rule treats it as the scalar it would be in scalar context. perl 5.42: `my @a = (1); print scalar(@a), ref(\@a)` prints `1 ARRAY` -- the array holds one Int element; it is not an Int. A related accident: `my @b = @a` records element type Array, so `$b[0]` reads Int through the count rule -- right answer, wrong reason. Fix after the cutover (01a0c115-df67) with 01a132ab and 01a132f4; each golden edit gets a `Diagnostics-changed:` trailer.

## Steps

- [ ] Failing test first; an aggregate's initialiser is read in list context, the element type recorded, the aggregate keeping its sigil type; `my @b = @a` copies @a's element type; goldens under trailers; precision must not fall; `make test`; commit

## Acceptance criteria

- [ ] `my @a = (1)` is an Array of Int (`go test ./internal/infer/ -run '^TestParenInitialiserIsListContext$' -count=1 -v | grep -q '^--- PASS: TestParenInitialiserIsListContext'`)
- [ ] `my @b = @a` copies the element type (`go test ./internal/infer/ -run '^TestArrayCopyKeepsElementType$' -count=1 -v | grep -q '^--- PASS: TestArrayCopyKeepsElementType'`)
