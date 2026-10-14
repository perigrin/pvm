---
title: An or with one non-guard side narrows its then-branch as if the guard held
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-14T22:25:10.403557041Z
updated: 2026-10-14T22:25:10.403557041Z
---

## Context

- paths: `internal/infer/lowered_flow.go` (`compound`)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

Found 2026-10-14 by the round-2 narrowing fixes (01a13c6a-64ef): when one side of `||`/`or` in a condition is a guard and the other is not, `compound()` keeps only the guard side. So in `if (!defined $ch or $ch !~ /\d/) { ... }` the then-branch narrows `$ch` to Undef, though it is reached whenever `$ch` is defined and has no digit. perl 5.42: `my $ch = "a"; if (!defined $ch or $ch !~ /\d/) { print length($ch) }` prints 1. Narrowing to Undef there can hide true positives (and type `$ch` wrongly inside the branch). An `or` whose other side is not a guard narrows nothing in its then-branch; in its else-branch both sides are false, so the guard's negation still holds.

## Steps

- [ ] Failing test first; `A or B` with B not a guard narrows the then-branch by nothing and the else-branch by not-A; `A and B` symmetric; `make test`; commit

## Acceptance criteria

- [ ] an `or` with one non-guard side does not narrow its then-branch (`go test ./internal/infer/ -run '^TestOrWithNonGuardSideDoesNotNarrowThen$' -count=1 -v | grep -q '^--- PASS: TestOrWithNonGuardSideDoesNotNarrowThen'`)
