---
title: Truly-no-default calls are their own candidate; the bare-call default convention goes
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a11937-22da-7250-b922-fabda15009a7
created: 2026-10-08T01:59:02.643468486Z
updated: 2026-10-08T02:04:27.867279621Z
---

## Context

- paths: `internal/parse/declarations/CORE.pmt` (lines 141 caller, 610 reset, 660 send, 704 sleep, 743 srand, 806 umask), `internal/parse/derive.go` (`prototypeFromCandidates`), `docs/rfcs/0001-type-libraries.md` ("A default perl computes is written as the call itself")
- commands: `make test` (100% green before every commit; never `--no-verify`); perl 5.42 `/home/perigrin/.local/bin/perl`

perigrin, 2026-10-08: the bare-call default convention goes ("I don't think we need" it). Its six lines -- `sub srand (Int $seed = srand()) Str;` and likewise `caller`, `reset`, `sleep`, `umask` and `send`'s `$to` -- state a self-referential default for what is truly no default: perl computes something no expression states. A truly-no-default case is its own candidate instead:

```perl
multi sub srand (Void $) Str;      # perl picks the seed (Perl_seed)
multi sub srand (Int $seed) Str;
```

Measured: `srand(@e)` is `srand(0)` ("0 but true"), not `srand()`, since the prototype is `;$` -- so for these six the empty case is only "no argument in that scalar slot". Spelling (perigrin, 2026-10-08): `Void $` for an absent scalar slot, `Void @` for an empty list -- `multi sub srand (Void $) Str;`, `multi sub send (Glob *socket, Str $msg, Int $flags, Void $to) Int|Undef;`. `()` stays for a sub that takes nothing at all (`sub break () None;`). A Void parameter is optional -- its slot can only be empty -- so the existing table derives it: the sigil gives the character, a `;` goes before it. A lone `(Void $)` derives `;$` (perigrin, 2026-10-08), so `f(1)` parses and is then a type error (Int is not Void). `(Void @)` derives `@`.

Prototype derivation must keep perl's prototypes (`;$` for five, `*$$;$` for send): candidates whose prototypes are equal once `;` is ignored derive the one with the earliest `;` -- `(Void $)`'s `;$` with `(Int $seed)`'s `$` gives `;$`; send's `*$$;$` with `*$$$` gives `*$$;$`. Today `prototypeFromCandidates` derives none when they differ.

## Steps

- [ ] Write failing test `TestVoidParamIsOptional`: a lone `(Void $)` derives `;$` and a lone `(Void @)` `@`; `(Void $)` with `(Int $seed)` derives `;$`; send's candidates derive `*$$;$`; candidates differing other than by `;` still derive none (negative); a Void parameter followed by a non-Void one is a declaration error (negative); `srand()` selects the `Void $` candidate and `srand(@e)` the `Int` one
- [ ] Write failing test `TestCoreHasNoSelfReferentialDefault`: no CORE.pmt parameter defaults to a call of its own sub, and the six are multis whose empty-or-shorter candidate stands for perl's own choice; their derived prototypes equal perl's (`prototype("CORE::srand")` etc.)
- [ ] Measure each candidate's return on 5.42 (`caller` with and without a level differ in fields); declare; remove the convention from RFC 0001 (keep "a default may name a later parameter"); `make test`; commit

## Acceptance criteria

- [ ] a Void parameter derives an optional slot, and only trails (`go test ./internal/parse/ -run '^TestVoidParamIsOptional$' -count=1 -v | grep -q '^--- PASS: TestVoidParamIsOptional'`)
- [ ] no CORE.pmt default calls its own sub (`go test ./internal/parse/ -run '^TestCoreHasNoSelfReferentialDefault$' -count=1 -v | grep -q '^--- PASS: TestCoreHasNoSelfReferentialDefault'`)
