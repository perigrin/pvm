---
title: Truly-no-default calls are their own candidate; the bare-call default convention goes
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a11937-22da-7250-b922-fabda15009a7
created: 2026-10-08T01:59:02.643468486Z
updated: 2026-10-08T02:01:37.643260735Z
---

## Context

- paths: `internal/parse/declarations/CORE.pmt` (lines 141 caller, 610 reset, 660 send, 704 sleep, 743 srand, 806 umask), `internal/parse/derive.go` (`prototypeFromCandidates`), `docs/rfcs/0001-type-libraries.md` ("A default perl computes is written as the call itself")
- commands: `make test` (100% green before every commit; never `--no-verify`); perl 5.42 `/home/perigrin/.local/bin/perl`

perigrin, 2026-10-08: the bare-call default convention goes ("I don't think we need" it). Its six lines -- `sub srand (Int $seed = srand()) Str;` and likewise `caller`, `reset`, `sleep`, `umask` and `send`'s `$to` -- state a self-referential default for what is truly no default: perl computes something no expression states. A truly-no-default case is its own candidate instead:

```perl
multi sub srand () Str;            # perl picks the seed (Perl_seed)
multi sub srand (Int $seed) Str;
```

Measured: `srand(@e)` is `srand(0)` ("0 but true"), not `srand()`, since the prototype is `;$` -- so for these six the empty case is only "no argument written", which `()` states (as `sub break () None;` does). `(Void @)` is strictly equivalent (perigrin, 2026-10-08): the prototype is derived from the candidates and governs how arguments are read before selection, so selection sees the same values either way -- `srand(@e)` is one scalar under the derived `;$` and matches neither; `f(@empty)` under a derived `@` flattens to nothing and matches both. A `Void @` parameter therefore derives no prototype character, as `()` has none. Either spelling is accepted; 01a11937-22da-7250-b922-fabda15009a7 adds `Void`. Defaults that ARE values stay: `select()` (the current filehandle) on close/write/print/printf/say, `exec`'s `= $args[0]`.

Prototype derivation must keep perl's prototypes (`;$` for five, `*$$;$` for send): candidates that are prefixes of one another derive a `;` where the shortest ends -- today `prototypeFromCandidates` derives none when arities differ.

## Steps

- [ ] Write failing test `TestCandidatePrefixesDeriveSemicolon`: candidates `()` and `($)` derive `;$`; `(*$$)` and `(*$$$)` derive `*$$;$`; candidates that differ other than by a trailing extension still derive none (negative)
- [ ] Write failing test `TestVoidListEquivalentToEmpty`: a `(Void @)` candidate derives the same prototype as `()` (no character), so `(Void @)` with `(Int $seed)` derives `;$`; and selection picks it for exactly the calls `()` is picked for
- [ ] Write failing test `TestCoreHasNoSelfReferentialDefault`: no CORE.pmt parameter defaults to a call of its own sub, and the six are multis whose empty-or-shorter candidate stands for perl's own choice; their derived prototypes equal perl's (`prototype("CORE::srand")` etc.)
- [ ] Measure each candidate's return on 5.42 (`caller` with and without a level differ in fields); declare; remove the convention from RFC 0001 (keep "a default may name a later parameter"); `make test`; commit

## Acceptance criteria

- [ ] (Void @) and () are equivalent (`go test ./internal/parse/ -run '^TestVoidListEquivalentToEmpty$' -count=1 -v | grep -q '^--- PASS: TestVoidListEquivalentToEmpty'`)
- [ ] prefix candidates derive a ; prototype (`go test ./internal/parse/ -run '^TestCandidatePrefixesDeriveSemicolon$' -count=1 -v | grep -q '^--- PASS: TestCandidatePrefixesDeriveSemicolon'`)
- [ ] no CORE.pmt default calls its own sub (`go test ./internal/parse/ -run '^TestCoreHasNoSelfReferentialDefault$' -count=1 -v | grep -q '^--- PASS: TestCoreHasNoSelfReferentialDefault'`)
