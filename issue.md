---
title: Unfeatured fc $x is refused, where perl reads the indirect call $x->fc
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-11T10:20:57.183467353Z
updated: 2026-10-11T10:20:57.183467353Z
---

## Context

- paths: `internal/parse` (feature-gated builtins; the `fc` arity from 01a0df71)
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-11 reconciling 01a0df71. Without `use feature 'fc'` (or a bundle that enables it), `ord fc $char` is not the builtin to perl: `fc $char` is indirect-object syntax for `$char->fc` (Deparse-measured). internal/parse refuses it (U=1): no wrong tree, but perl's reading is not reached.

## Steps

- [ ] Measure unfeatured `fc EXPR` and `fc($x)` with Deparse; failing test first; read as perl does; `make test`; commit

## Acceptance criteria

- [ ] unfeatured `fc $x` is read as perl reads it (`go test ./internal/parse/ -run '^TestUnfeaturedFcIsIndirect$' -count=1 -v | grep -q '^--- PASS: TestUnfeaturedFcIsIndirect'`)
