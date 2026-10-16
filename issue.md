---
title: any and all are builtins only under their feature; unfeatured they are user subs
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-13T23:22:39.79502639Z
updated: 2026-10-16T18:48:35.697308722Z
---

## Context

- paths: `internal/types/CORE.tpm` (`any`, `all`), `internal/infer/lowered_calls.go` (builtin lookup), `internal/parse/` (feature tracking, as for `fc` in 01a12a7a)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-13 by the calls port (01a12a85-32cf); perigrin ruled "only when featured". CORE.tpm declares `any` and `all`, so `any();` draws an arity diagnostic on both paths. They are builtins only under `use feature 'keyword_any'` / `'keyword_all'`; without the feature perl treats them as user subs: `perl -wce 'any();'` is "syntax OK" and `sub any { 7 } print any()` prints 7. Same shape as the unfeatured `fc` issue (01a12a7a). No test locks in the arity diagnostic today.

## Steps

- [ ] Failing test first; any/all resolve as CORE builtins only under their feature (or a bundle that enables it), otherwise as user subs; `make test`; commit

## Acceptance criteria

- [ ] unfeatured any/all are user subs (`go test ./internal/infer/ -run '^TestAnyAllAreBuiltinOnlyWhenFeatured$' -count=1 -v | grep -q '^--- PASS: TestAnyAllAreBuiltinOnlyWhenFeatured'`)
