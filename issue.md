---
title: Known non-numeric string content is lost through a variable, so perl's 'isn't numeric' goes unreported
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-14T21:56:19.741674618Z
updated: 2026-10-14T21:56:19.741674618Z
---

## Context

- paths: `internal/infer/lowered_ops.go` (`lossyCoercion`), `internal/infer/symbols.go`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

Found 2026-10-14 by cutover 1's round-2 re-measure. Under perigrin's lossy-coercion ruling, a Str of UNKNOWN content in a numeric slot is silent by default and a KNOWN non-numeric literal is reported. Content is known only at the literal itself, so a string that reaches the slot through a variable or element loses it: range.t:73 `@array = ('a'..'e'); $array[0] != 5` and taint_mode.t:340 `0 + $empty` (`$empty = ""`) -- perl -w warns "isn't numeric" on both, PSC is silent by default. Carrying "known non-numeric" (and "known numeric") through assignment and element types would restore these without reversing the ruling.

## Steps

- [ ] Failing test first; a variable or element assigned only non-numeric literals is known non-numeric; a mixed or computed one stays unknown; `make test`; commit

## Acceptance criteria

- [ ] a known non-numeric string reached through a variable is reported by default (`go test ./internal/infer/ -run '^TestKnownNonNumericThroughVariable$' -count=1 -v | grep -q '^--- PASS: TestKnownNonNumericThroughVariable'`)
