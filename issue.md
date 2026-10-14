---
title: An unguarded $1 is typed Str, but perl leaves it undef or stale
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-14T02:40:36.434719793Z
updated: 2026-10-14T02:40:36.434719793Z
---

## Context

- paths: `internal/infer/lowered_special.go` (`capture`), `internal/infer/lowered_flow.go` (guard narrowing)
- commands: `go test ./internal/infer/ -run '^TestLoweredGoldenRatchet$'`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-14 by the special-form port (01a12a85-3386), which kept legacy's answer so the goldens match: `$N` is typed Str (or Int for a digits-only group) unconditionally. perl 5.42 says otherwise:

- no match before it: `my $c = $1;` is undef (golden `special/capture_none` says Str)
- a failed match leaves the previous `$1`: `"a1" =~ /(\d)/; "b" =~ /(\d)/; print $1` prints 1
- a match elsewhere can set it before the code runs: `sub f { return $1 } "a1" =~ /(\d)/; print f()` prints 1

So an unguarded capture is `Str|Undef` (or `Int|Undef`), and only a successful match's guarded branch (`if ($s =~ /(\d+)/) { $1 }`) narrows it to the group's type. Fix after the cutover (01a0c115-df67) with the other copied defects; goldens under `Diagnostics-changed:` trailers; precision must not fall.

## Steps

- [ ] Failing tests first; unguarded `$N` is T|Undef; inside a successful match's guard it narrows to T; goldens under trailers; `make test`; commit

## Acceptance criteria

- [ ] an unguarded capture may be undef (`go test ./internal/infer/ -run '^TestUnguardedCaptureMayBeUndef$' -count=1 -v | grep -q '^--- PASS: TestUnguardedCaptureMayBeUndef'`)
- [ ] a capture inside a successful match's guard is its group's type (`go test ./internal/infer/ -run '^TestGuardedCaptureNarrows$' -count=1 -v | grep -q '^--- PASS: TestGuardedCaptureNarrows'`)
