---
title: Canon re-spaces a hex float literal so it no longer compiles
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-24T05:41:33.815026386Z
updated: 2026-10-24T05:41:33.815026386Z
---

Found by the 75d4 worker (01a1664d-75d4, exec/75d4 0f1bacff) on 2026-10-24, while running Deparse on whole files. The defect predates 75d4; base canon does the same.

## Context

- paths: `internal/parse/canon.go` (how canon writes numeric literals)
- entrypoints: canon writing a hexadecimal floating-point literal

Canon re-spaces hexadecimal floating-point literals such as `0x1.8p1`, and the result doesn't compile. op/hexfp.t and op/sprintf2.t compare faithful token by token, but their canon fails `perl -c`. Canon should write a numeric literal exactly as the source spelled it.

## Acceptance criteria

### Positive Scenarios

- `go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 | grep -q '^ok'` (probe rows for `0x1.8p1`, `0x.8p0` and `0x1p-2`, each faithful with canon compiling under perl 5.44)
