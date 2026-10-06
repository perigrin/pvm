---
title: A lexical sub is also recorded in its package and outlives its block
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-06T11:43:34.198292145Z
updated: 2026-10-06T11:43:34.198292145Z
---

## Context

- paths: `internal/parse` (lexical sub table, `p.subKey`, `my`/`our`/`state sub` recording)
- commands: `make test`; ratchets; perl 5.42

From the package-scoping fix for 01a10dc1-3530-762f-9a03-177445629c6b. `my sub` / `state sub` are recorded in the lexical table AND in their package's table, so a qualified `main::g` wrongly finds a `my sub g` (perl: a lexical sub has no package entry). The lexical table also keeps a lexical sub in scope to the end of the file rather than its enclosing block (marked with a `ponytail:` comment).

## Steps

- [ ] Measure with `perl -c`: `my sub g ($) {} main::g(1, 2)`, and a `my sub` used after its block closes
- [ ] Failing tests first; record lexical subs only lexically, scoped to their block; ratchets must not regress; `make test`; commit

## Acceptance criteria

- [ ] a lexical sub is not found by a qualified name and leaves scope with its block (`go test ./internal/parse/ -run '^TestLexicalSubIsNotPackaged$' -count=1 -v | grep -q '^--- PASS: TestLexicalSubIsNotPackaged'`)
