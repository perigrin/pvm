---
title: CORE.pmt types all and any without their feature gate
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-06T11:51:25.623320084Z
updated: 2026-10-16T18:48:33.294168722Z
---

## Context

- paths: `internal/parse/declarations/CORE.tpm` (all, any), the feature-gated builtins (fc, say, isa, ...), RFC 0001
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

From the alignment gate on 01a10177-4e7b-7d2c-8d2a-2033647ffc4a. `all` and `any` exist only under `use feature qw(keyword_all keyword_any)` (5.42); without it, `all {1} @x` calls an undefined sub "all". CORE.tpm types them unconditionally and records no gate, as it does for other feature-gated builtins (fc's parse is gated by the parser, not by CORE.tpm).

## Steps

- [ ] Measure which CORE.tpm builtins are feature-gated on 5.42, and how the parser gates each today
- [ ] Decide (with perigrin if the RFC is silent) how CORE.tpm states a feature gate -- e.g. an attribute such as `:feature(keyword_all)` -- with a failing test that `all` outside the feature is not the builtin; `make test`; commit

## Acceptance criteria

- [ ] a feature-gated builtin is the builtin only under its feature (`go test ./internal/parse/ -run '^TestCoreFeatureGatedBuiltinsAreGated$' -count=1 -v | grep -q '^--- PASS: TestCoreFeatureGatedBuiltinsAreGated'`)
