---
title: field declarations are not a Decl.Kind inference uses
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-16T01:16:38.286777799Z
updated: 2026-10-16T01:16:38.286777799Z
---

## Context

- found by: the agentic review of PR #486 (plan alignment), 2026-10-16: a deferral with no tracking issue

The lowered vocabulary has `field` (class fields, perl 5.38+) but inference does not type a field from its initialiser or make it visible to methods. No issue tracked it.

## Steps

- [ ] Type a field from its initialiser and scope it to the class's methods; failing test first; `make test`; commit

## Acceptance criteria

- [ ] a class field is typed from its initialiser (`go test ./internal/infer/ -run '^TestFieldDeclIsTyped$' -count=1 -v | grep -q '^--- PASS: TestFieldDeclIsTyped'`)
