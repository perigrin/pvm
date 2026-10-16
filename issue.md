---
title: The arity refusal misses XS subs whose prototypes are not in readable source (blessed, first)
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-06T18:04:16.836570874Z
updated: 2026-10-16T18:48:33.876978767Z
---

## Context

- paths: `internal/parse` (export reading, XS prototype reading -- `xsubpp` prototypes are already read for core modules from the perl.git checkout), `internal/parse/arity.go`
- commands: `make test`; ratchets; perl 5.42 `/home/perigrin/.local/bin/perl`

From the alignment gate on 01a10177-547c-73a9-a302-f25f579a4cb3. The call-arity refusal does not refuse calls perl does when the sub's prototype lives in XS rather than readable Perl source: `use Scalar::Util qw(blessed); blessed($x, 1)` and `use List::Util qw(first); first()` compile-fail in perl ("Too many arguments for Scalar::Util::blessed", "Not enough arguments ...") but pvm accepts them. This is under-refusal (safe), not a wrong refusal.

## Steps

- [ ] Measure the prototypes of the commonly imported XS subs (Scalar::Util, List::Util, POSIX) with `prototype(\&...)`
- [ ] Read them where the XS source or a .tpm states them; failing tests first; ratchets must not regress; `make test`; commit

## Acceptance criteria

- [ ] calls to imported XS subs with a known prototype are refused where perl refuses (`go test ./internal/parse/ -run '^TestArityRefusalReadsXSPrototypes$' -count=1 -v | grep -q '^--- PASS: TestArityRefusalReadsXSPrototypes'`)
