---
title: An argument with no default is typed Void|T; the bare-call default convention goes
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a11937-22da-7250-b922-fabda15009a7
created: 2026-10-08T01:59:02.643468486Z
updated: 2026-10-08T02:09:17.776755463Z
---

## Context

- paths: `internal/parse/declarations/CORE.pmt` (lines 141 caller, 610 reset, 660 send, 704 sleep, 743 srand, 806 umask), `internal/parse/derive.go` (`prototypeFromCandidates`), `docs/rfcs/0001-type-libraries.md` ("A default perl computes is written as the call itself")
- commands: `make test` (100% green before every commit; never `--no-verify`); perl 5.42 `/home/perigrin/.local/bin/perl`

perigrin, 2026-10-08: the bare-call default convention goes. Its six lines -- `sub srand (Int $seed = srand()) Str;` and likewise `caller`, `reset`, `sleep`, `umask` and `send`'s `$to` -- state a self-referential default for what is truly no default: perl computes something no expression states.

An argument that may be absent with no default is typed `Optional[T]`, which is `Void|T` (the arity-0 type and the alias both from 01a11937-22da-7250-b922-fabda15009a7; perigrin, 2026-10-08):

```perl
sub srand (Optional[Int] $seed) Str;
sub send (Glob *socket, Str $msg, Int $flags, Optional[Str] $to) Maybe[Int];
```

No multi: one signature says it, and a multi stays only where the result type differs (measure each; `caller` with and without a level return different fields, possibly the same type). `Void $` spells an absent scalar slot, `Void @` an empty list; `()` stays for a sub that takes nothing at all (`sub break () None;`).

Derivation needs one rule: a parameter whose type includes `Void` is optional, so the existing table puts a `;` before it -- `(Void|Int $seed)` derives `;$`, `send`'s derives `*$$;$`, a lone `(Void $)` `;$`. Measured: `srand(@e)` is `srand(0)` ("0 but true"), not `srand()`: under `;$` the array is one scalar, so it is the Int case, and `srand()` the Void case.

Defaults that ARE values stay: `select()` (the current filehandle) on close/write/print/printf/say, `exec`'s `= $args[0]`, `reverse`'s `= ($_)`.

## Steps

- [ ] Write failing test `TestVoidParamIsOptional`: `(Void|Int $seed)` and a lone `(Void $)` derive `;$`; `send`'s signature derives `*$$;$`; `(Void @)` derives `@`; a Void-including parameter followed by a required non-Void one is a declaration error (negative); `srand()` type-checks with the Void case and `srand("x")` warns as an Int slot does
- [ ] Write failing test `TestCoreHasNoSelfReferentialDefault`: no CORE.pmt parameter defaults to a call of its own sub; the six derive perl's prototypes (`prototype("CORE::srand")` etc.)
- [ ] Measure each return on 5.42; declare the six with `Void|...` (a multi only where the result type differs); remove the convention from RFC 0001 and record `Void|T` as "may be absent, no default" (keep "a default may name a later parameter"); `make test`; commit

## Acceptance criteria

- [ ] a Void-including parameter is optional (`go test ./internal/parse/ -run '^TestVoidParamIsOptional$' -count=1 -v | grep -q '^--- PASS: TestVoidParamIsOptional'`)
- [ ] no CORE.pmt default calls its own sub (`go test ./internal/parse/ -run '^TestCoreHasNoSelfReferentialDefault$' -count=1 -v | grep -q '^--- PASS: TestCoreHasNoSelfReferentialDefault'`)
