---
title: CORE::any { } @a without its feature is left unparsed; perl accepts it
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-06T20:47:05.580229226Z
updated: 2026-10-06T20:47:05.580229226Z
---

## Context

- paths: `internal/parse/call.go` (`featureTakesBlock`, ~624)
- commands: `make test`; ratchets; perl 5.42 `/home/perigrin/.local/bin/perl -MO=Deparse`

From the alignment gate on 01a0fd9b-55cb-70cd-a9da-929190cb581c, pre-existing (broken at 65d2f64b too): `CORE::any { $_ } @a` (and `CORE::all`) WITHOUT the keyword_any/keyword_all feature is accepted by perl 5.42 -- the CORE:: spelling reaches the builtin regardless of the feature -- but pvm leaves it unparsed, because `featureTakesBlock` ignores the `CORE::` spelling.

## Steps

- [ ] Measure `CORE::any { } @a`, `CORE::all { } @a` without their features with Deparse
- [ ] Write a failing parse test; make featureTakesBlock honour the CORE:: spelling; ratchets must not regress; `make test`; commit

## Acceptance criteria

- [ ] CORE::any and CORE::all take their block without the feature (`go test ./internal/parse/ -run '^TestCoreAnyAllTakeBlockWithoutFeature$' -count=1 -v | grep -q '^--- PASS: TestCoreAnyAllTakeBlockWithoutFeature'`)
