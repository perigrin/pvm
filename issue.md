---
title: T2 core round-trip ratchet and the lexer fuzz target
state: pending
urgency: normal
milestone: m0-lexer-round-trip
created: 2026-09-14T15:42:44.686499906Z
updated: 2026-09-14T15:42:56.532288435Z
---

M0's headline metric: 100% lossless round-trip over the T2 core corpus —
perl5 `base cmd comp opbasic class`, 56 files, verified present.

Plus the fuzz target, because 56 files is a sample and the invariants must
hold on ANY input, including half-typed buffers an LSP will see.

## Why a ratchet and not a count

The parse-fidelity harness already has this shape and it was hard-won:
`internal/parseoracle/ratchet.go` freezes a per-file baseline and fails on
movement in EITHER direction, so an improvement has to be recorded
deliberately rather than silently absorbed. Reuse that machinery; do not
write a second one.

An improvement failing the build sounds perverse and is not: perl-lsp
shipped a ratchet whose baseline contained the line `0`, and nothing
noticed. A baseline that can only get better is a baseline nobody reads.

## The fuzz target

Spec §7.6.2 gives the seeds and all four invariants. Target is 1M execs with
zero crashers. Invariant 4 (forward progress) is structural, from the first
issue in this milestone, because `go test -fuzz` hangs rather than failing
on an infinite loop.

## Acceptance

- [ ] All 56 T2 core files round-trip byte-for-byte (`go test ./internal/lexer/ -run TestT2CoreRoundTrips -count=1`)
- [ ] A committed baseline records per-file status and fails on movement in either direction (`go test ./internal/lexer/ -run TestLexerRatchet -count=1`)
- [ ] The fuzz target runs 1M execs with zero crashers (`go test ./internal/lexer/ -run TestFuzzSeedCorpus -count=1`)
- [ ] All four §7.6.2 invariants hold across the seed corpus (`go test ./internal/lexer/ -run TestFuzzInvariants -count=1`)

Plan: M0 headline metric. Spec: §7.6.2.
