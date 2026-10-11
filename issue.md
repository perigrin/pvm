---
title: A sub declared in another package is refused as a call from main, where perl reads an indirect method call
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-11T10:20:56.998189268Z
updated: 2026-10-11T10:20:56.998189268Z
---

## Context

- paths: `internal/parse/call.go` (indirect-object reading), `internal/parse` (the package-scoped sub table, 0b13f919)
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-11 reconciling 01a0ee05. With a sub declared in another package -- `package Foo; sub method {1} package main;` -- perl reads `method Pack ("a")` as `'Pack'->method('a')` and `method $obj` as `$obj->method` (indirect object syntax, Deparse-measured). internal/parse now knows `method` is not main's sub (0b13f919), but refuses these (U=1) instead of reading the indirect call.

## Steps

- [ ] Measure the indirect forms with Deparse; failing test first; read them as perl does; ratchets toward perl only; `make test`; commit

## Acceptance criteria

- [ ] a sub named in another package is an indirect method call from main, as perl reads it (`go test ./internal/parse/ -run '^TestIndirectCallToAnotherPackagesSub$' -count=1 -v | grep -q '^--- PASS: TestIndirectCallToAnotherPackagesSub'`)
