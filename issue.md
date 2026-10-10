---
title: Canon is not a fixpoint on *+ ++ (a fuzz find)
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-10T10:55:16.362761646Z
updated: 2026-10-10T10:55:53.865494335Z
---

## Context

- paths: `internal/parse` (glob sigil, unary plus, increment), `internal/parse/testdata/fuzz/FuzzParser/`
- commands: `go test ./internal/parse/ -run '^FuzzParser$' -count=1`; `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-10 by `make fuzz-parser` during the final m1 verify, and filed under perigrin's ruling that a fuzz find is an issue, not a failed milestone. Seed parked at scratchpad/pending-seeds/960a734f11aaaab1: `*+ ++` -- canon is not a fixpoint (4 significant tokens, then 5). perl refuses the input: "Can't modify ref-to-glob cast in postincrement (++)".

## Steps

- [ ] Measure perl on `*+ ++` and its neighbours; failing test first; refuse or read as perl does; commit the seed; `make test`; commit

## Acceptance criteria

- [ ] the seed replays clean (`go test ./internal/parse/ -run '^TestGlobPlusIncrementIsAFixpoint$' -count=1 -v | grep -q '^--- PASS: TestGlobPlusIncrementIsAFixpoint'`)
