---
title: <> and <<>> lower to the same Readline, though perl reads them differently
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-12T23:47:10.838770288Z
updated: 2026-10-12T23:47:10.838770288Z
---

## Context

- paths: `internal/lower/special.go` (the `<...>` term: `Readline{Handle, Magic}`), `internal/lower/node.go`
- commands: `go test ./internal/lower/`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-12 by the position-map issue (01a0c113-dfd3). `<>` and `<<>>` both lower to `Readline{Magic: true}` with a nil Handle, so the lowered tree cannot tell them apart. perl does: Deparse reads `<>` as `readline ARGV` and keeps `<<>>` as itself, because `<<>>` opens each @ARGV file with three-argument open (no magic on filenames like `ls|`). A consumer reasoning about what a loop reads, or a lint warning about `<>`, needs the difference.

## Steps

- [ ] Measure with Deparse; failing test first; a field or Magic value that distinguishes `<<>>` (e.g. `Double bool`, or Magic as an enum); `make test`; commit

## Acceptance criteria

- [ ] `<>` and `<<>>` lower to distinguishable Readlines (`go test ./internal/lower/ -run '^TestLowerArgvReadlineKinds$' -count=1 -v | grep -q '^--- PASS: TestLowerArgvReadlineKinds'`)
