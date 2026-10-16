---
title: An array written in a $ slot is checked as Array, not as its count
state: in-progress
urgency: normal
milestone: type-libraries
created: 2026-10-08T04:57:28.837383755Z
updated: 2026-10-16T03:49:41.230062072Z
sessions:
- start_sha: 133837eb651f8dcffb4203d072b3701753486869
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T03:49:41.230062072Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T03:49:41.230062072Z
---

## Context

- paths: `internal/infer/infer.go` (`builtinArgType` and the call-site argument check), `internal/types/narrowing.go` (`NarrowByContext`), `internal/infer/core_test.go`
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found reviewing 01a1193c (2026-10-08). infer reports `my @e; srand(@e)` as an Error, "expects Int, got Array". Perl puts an array written in a `$` prototype slot in scalar context: measured on 5.42, `srand(@e)` is "0 but true" (the count, 0) and `srand(@a)` with two elements is 2. So an Array (or Hash) argument in a non-slurpy `$` slot arrives as its count, an Int, and should be checked as `NarrowByContext(Array, ScalarCtx)` = Int, not as Array. Likely affects every `$`-slot builtin, not only srand.

## Steps

- [ ] Write failing test `TestArrayInScalarSlotIsItsCount`: `my @e; srand(@e)` and `sleep(@a)` have no diagnostics; `my %h; srand(%h)` likewise (measure first); a `Str`-typed `$` slot given `@a` warns as Int-to-Str does (measure)
- [ ] Apply scalar context to an argument written in a non-slurpy `$` slot before checking it; `make test`; commit

## Acceptance criteria

- [ ] an array in a `$` slot is checked as its count (`go test ./internal/infer/ -run '^TestArrayInScalarSlotIsItsCount$' -count=1 -v | grep -q '^--- PASS: TestArrayInScalarSlotIsItsCount'`)
