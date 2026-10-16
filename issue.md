---
title: A sub's writes to outer variables are widened everywhere instead of applied at its call sites
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-16T03:37:16.735814648Z
updated: 2026-10-16T03:37:16.735814648Z
---

## Context

- paths: `internal/infer/infer.go` (the written-back loop after a named sub; `widenWritten`), `internal/infer/calls.go` (user calls)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

Found 2026-10-16 deciding PR #486 review I2 (01a14247-d886): a variable a named sub writes is widened to its sigil type everywhere, because PSC does not know where the sub runs. That is sound but loses true positives: `my $p = 5; sub reset_p { $p = undef } reset_p(); print $p + 1` -- perl -w warns "uninitialized"; PSC is silent. Recording each sub's effects on outer variables and applying them at its call sites (and leaving the type alone where the sub is only declared) would restore them without the false positives options (a)/(b) had.

## Steps

- [ ] Failing test first: a call to a sub that sets an outer variable to undef makes it Undef after the call; the declaration alone does not; `make test`; commit

## Acceptance criteria

- [ ] a sub's writes apply at its call sites (`go test ./internal/infer/ -run '^TestSubEffectsApplyAtCallSite$' -count=1 -v | grep -q '^--- PASS: TestSubEffectsApplyAtCallSite'`)
