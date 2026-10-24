---
title: Canon drops the ; after an anonymous-sub statement
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-24T04:11:53.56548767Z
updated: 2026-10-24T04:11:53.56548767Z
---

Found by the independent review of NL-c (01a1692f-bf96, exec/nlc 3de59a88) on 2026-10-24. The defect predates NL-c; 71f565f4's canon.go does the same.

## Context

- paths: `internal/parse/canon.go`
- entrypoints: canon writing the `;` that ends a statement whose expression ends in `}`

Canon drops the `;` after an anonymous-sub statement. `sub { 1 }; 2;` re-emits as `sub {1;} 2;`, which perl rejects as a syntax error. Faithful rejects it correctly, so any file with this shape counts as unfaithful because of a canon bug. Check other expression statements that end in `}`: `do { 1 }; 2;`, `+{ a => 1 }; 2;`, `\{ }; 2;`.

## Acceptance criteria

### Positive Scenarios

- `go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 | grep -q '^ok'` (faithful probe rows for `sub { 1 }; 2;`, `sub { 1; };`, `do { 1 }; 2;` and `+{ a => 1 }; 2;`)
