---
title: "no warnings 'syntax' is not honoured: length(@x) is still reported under it"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-16T15:55:16.719573411Z
updated: 2026-10-16T15:55:16.719573411Z
---

## Context

- found by: review R2 (2026-10-16)

psc maps the `uninitialized` and `numeric` warning categories to its diagnostics, but not `syntax`: `length(@x)` ("length() used on @x (did you mean \"scalar(@x)\"?)") is in perl's syntax category, so a bare `no warnings` or `no warnings "syntax"` silences it in perl 5.42 while psc still reports it.

## Steps

- [ ] Measure which psc diagnostics correspond to perl's syntax category; failing test first; `make test`; commit

## Acceptance criteria

- [ ] no warnings syntax silences length(@x) (`go test ./internal/infer/ -run '^TestNoWarningsSyntaxCategory$' -count=1 -v | grep -q '^--- PASS: TestNoWarningsSyntaxCategory'`)
