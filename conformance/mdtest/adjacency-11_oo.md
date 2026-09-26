# Every construct, each beside another

One body holding every construct this tier introduces, each adjacent to
another -- and the one file in the corpus written because the bug it
catches was already known.

**Tier 11 oo.** Introduces nothing of its own; the mixture is the
subject. Depends on 08_references, and `bless` is the whole of the
dependency: the blessed hash below is a tier 08 reference plus a
string. Pairing with tier 10 instead would assert nothing -- this tier
uses no file handle.

MEASURED against our parser, and this is the case the whole
adjacency-file design was written for. Before the fix:

    class Foo { ADJUST { 1 } }                   parses, 0 Unknowns
    class Foo { ADJUST { 1 } method m { 2 } }    1 Unknown, swallowing both

`ADJUST` alone parsed -- that is the ADJUST case in `class.md`, and it
was green throughout. `ADJUST` followed by anything -- `method`,
`field`, or a second `ADJUST` -- did not. A corpus of one construct per
file goes green over this by CONSTRUCTION, because every construct in
such a corpus is measured alone and every construct alone passed.

The cause was that `ADJUST` sat in neither phaser table, so its `{` was
classified as an anonymous hash rather than a block; the statement then
wanted a `;` that valid perl does not write, and the declaration after
it was swallowed. Adding `ADJUST` beside `BEGIN` in both tables closed
it, and this case is green in both orders.

## The whole tier in one body

A `class` with a `field`, an `ADJUST` and a `method` in one body; a
classic `bless` of an empty anon hash; a compile-time method call and a
dynamic one on the result; an infix `isa` and a `can` chain; and a
`tie` with the `tied` that asks about it. The class side and the
classic side are adjacent to each other as well, because a parser that
switches modes on `class` has to switch back.

This case carried a `refuses:`/`refusal:` pair while the ADJUST gap was
open. Measured then, the Unknown started at the `ADJUST` keyword and ran
to the end of the `method` after it: the expression parser read the
ADJUST block as a subscripted bareword, then found the `method` before
the terminator, which is the `trailing_tokens` site. Both lines are gone
now that it parses -- a refusal record kept past its bug hides the
regression it was written to catch.

`$yes` and `$code` are the argument-extent slice's two constructs
(issue 01a0c730), adjacent to the dispatch they are not. `$b isa Bar`
is an INFIX OPERATOR: measured, it emits `<2> isa` with no `entersub`
and no `method_named` at all, so it sits here beside four method calls
that emit both, and a parser that read it as a fifth is caught by the
mixture and by nothing else. `$b->can("hi")->($b)` is TWO calls through
one chain and only the first is a method call; measured, two `entersub`
under one `method_named`, standing next to the plain `$b->hi` and the
dynamic `$b->$name` it must not be confused with.

`tie` and `tied` are here for a reason the other constructs are not:
they need `Bar` to grow a second ROLE. The same package is a plain
blessed class -- `bless {}, "Bar"`, reached by `$b->hi` -- and a tie
implementation, because `TIESCALAR` and `FETCH` live in it. A parser
that treats a package as having one kind sees a body in which the same
name is both, which no one-construct file can present. `tie my $t,
"Bar", "arg"` declares `$t` IN the argument list, standing next to `my
$b = bless {}, "Bar"`, so the two spellings of "make a variable hold a
Bar" are adjacent and only one of them is an assignment. `tied($t)` is
a NAMED UNARY where `tie` is a list operator -- measured, `<1> tied
sK/1` against `<@> tie vK/2`: one child and no mark against a mark and
two.

`FETCH` returns `fetched` and not `tied`, which reads more naturally
and is the trap: the word would otherwise appear in the output where a
reader would take it for the keyword.

`Foo2` is `ref($c)` then `$c->m`, which ADJUST raised from 1 to 2;
`Barbarbar` is `ref($b)` then the same method reached two ways;
`fetched` is `$t` read through the tie, which is `FETCH` reached with
no call written anywhere, and `y` is `tied($t)` answering. Nothing
prints a raw object: `Bar=HASH(0x...)` carries an address that changes
every run, so `ref` is what this asserts on.

```perl
use feature 'class', 'isa';
no warnings 'experimental::class';
class Foo {
    field $x = 1;
    ADJUST { $x = 2 }
    method m { $x }
}
package Bar;
sub hi { return "bar" }
sub TIESCALAR { return bless {}, "Bar" }
sub FETCH { return "fetched" }
package main;
my $c = Foo->new;
my $b = bless {}, "Bar";
my $name = "hi";
my $yes = $b isa Bar;
my $code = $b->can("hi")->($b);
tie my $t, "Bar", "arg";
my $is = tied($t) ? "y" : "n";
print ref($c), $c->m, ref($b), $b->hi, $b->$name, $yes, $code, $t, $is, "\n";
```

```behavior
parses: yes
```

```output
Foo2Barbarbar1barfetchedy
```
