---
title: "A niladic builtin called with arguments is not refused: time(1) parses as time()(1)"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-06T21:21:10.899653387Z
updated: 2026-10-06T21:21:10.899653387Z
---

## Context

- paths: `internal/parse/call.go` (the niladic path), `internal/parse/arity.go` (`call_arity`)
- commands: `make test`; ratchets; perl 5.42 `/home/perigrin/.local/bin/perl`

Found while building the 55cb gate fixes. A niladic builtin given arguments is not refused: `time(1)` canons as `time()(1)`, and a featured `__SUB__(1)` as `__SUB__()(1)`, while perl 5.42 refuses both at compile time (syntax error). The niladic path returns before the arity check (01a10177-547c's call_arity) runs.

## Steps

- [ ] Measure `time(1)`, `wantarray(1)`, `__SUB__(1)` under current_sub, `getppid(1)` and the other `()`-prototype builtins with `perl -c`
- [ ] Write a failing test; run the arity check on the niladic path (or refuse there with call_arity); ratchets must not regress; `make test`; commit

## Acceptance criteria

- [ ] a niladic builtin called with arguments is refused (`go test ./internal/parse/ -run '^TestNiladicBuiltinWithArgumentsRefused$' -count=1 -v | grep -q '^--- PASS: TestNiladicBuiltinWithArgumentsRefused'`)
