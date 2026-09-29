---
title: Indirect object notation with an unknown class, or WORD $var, needs the symbol table the parser does not have
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T05:43:20.616611766Z
updated: 2026-09-29T05:43:20.616611766Z
---

Indirect object notation is read (01a0ddc5-b591) only where the class is a
package this parse KNOWS -- declared by `package`/`class`, or loaded by
`use`/`require`. Two halves of perl's rule are left, both deliberately:

1. `WORD Bareword ...` where WORD is an unknown sub and Bareword is not a known
   package. perl reads it as `'Bareword'->WORD(...)`:

       my $x = new Foo "a";        perl: my $x = 'Foo'->new('a');   ours: refuses

2. `WORD $var ...` as `$var->WORD(...)`, unless WORD is a known sub or starts
   uppercase:

       $foo = doit $object "FOO";  perl: $foo = $object->doit('FOO');   ours: refuses

Both need to know every sub in scope, which perl has from its symbol table and
this parser has only in part: a function imported from a module it could not
read looks unknown. Reading them anyway was measured -- `ok $x, 'n'` became
`$x->ok(...)` and `is Pkg::fill(), 'ok'` became `'Pkg::fill'->is(...)`, and 35
and then 13 T1 files went wrong in canon. A refusal is honest; those were
confident wrong trees.

The route is module resolution good enough to know an import's name set: when
every `use` in a file resolves, an unknown word is genuinely unknown and both
halves can be read.

## Where the fixtures point here

`internal/conformance/conformance_test.go`'s `TestCorpusSkipsAreDocumented` and
`TestRefusalCodeMismatchFails` use `my $x = new Foo "a";` -- half 1 -- naming THIS
issue. When it lands both fail with "no longer refuses; replace this fixture".

## Acceptance criteria

- [ ] `new Foo "a"` with no known package Foo reads as `'Foo'->new('a')` in a file whose imports all resolved (`go test ./internal/parse/ -run '^TestIndirectUnknownClass$' -count=1 -v | grep -q '^--- PASS: TestIndirectUnknownClass'`)
- [ ] REGRESSION GUARD: the canon fidelity ratchet holds -- no unresolved import turns into a method call (`go test ./internal/parse/ -run '^TestCanonFidelityRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonFidelityRatchet'`)
