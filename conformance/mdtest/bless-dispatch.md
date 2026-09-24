# bless and method dispatch

The classic object system: a reference plus a string, and the four ops
perl uses to find a method on the result.

**Tier 11 oo.** Introduces `anonhash`, `bless`, `emptyavhv`, `isa`,
`method`, `method_named`, `method_super`, `methstart`, `shift`, `stub`,
`tie`, `tied`. Depends on 08_references.

What this tier adds is how the callee is found; the `entersub` and
`leavesub` that surround every call below are tier 07's. A bare
`package` statement emits NO runtime op, so these cases can name
packages without borrowing tier 12's `use`.

Cases print `ref $o` rather than the object itself: printing an object
gives `Foo=HASH(0x55d3...)`, whose address changes every run, where
`ref` gives the stable string the blessing installed. `ref` is tier
08's op.

## bless of an EMPTY anonymous hash

`bless {}, $c` blesses an empty anonymous hash, and the optimiser has a
dedicated op for that case: `emptyavhv`, not `anonhash`. This is the
case that looks simplest and carries the op nobody expects.

Measured, the hash never becomes an `anonhash`:

    6  <0> emptyavhv[t3] s/ANONHASH
    7  <0> padsv[$c:1,3] s
    8  <@> bless sK/2

The populated case below is the same construct with contents in the
braces and emits three different ops. A parser that learns one form
learns nothing about the other.

```perl
my $c = "Foo";
my $o = bless {}, $c;
print ref $o, "\n";
```

```behavior
parses: yes
```

```output
Foo
```

## bless of a POPULATED anonymous hash

`bless { %a }, $c` builds a real anonymous hash first. Measured, with
no `emptyavhv` anywhere:

    d  <0> pushmark s
    e  <0> padhv[%a:1,4] l
    f  <@> anonhash sK*/1
    g  <0> padsv[$c:2,4] s
    h  <@> bless sK/2

The pair with the empty case is the point: one construct in the source
-- an anonymous hash handed to `bless` -- compiles to two unrelated op
streams depending on whether the braces are empty, so the corpus has to
carry both or it measures half of `bless`. `%a` is a tier 02 hash and
`aassign`, `padhv` are tier 02's ops; what this case introduces is
`anonhash` and the `bless` around it.

```perl
my %a = (n => 1);
my $c = "Foo";
my $o = bless { %a }, $c;
print ref $o, "\n";
```

```behavior
parses: yes
```

```output
Foo
```

## A method name written out

`$o->hi`, with the method name written out, resolves the name at
compile time:

    b  <.> method_named[PV "hi"] l
    c  <1> entersub[t3] lKRS/TARG

This is the spelling everyone measures, and measuring only it would
miss the dynamic and SUPER cases, which are different ops for what a
reader calls the same construct.

THE TOKEN FACT, and why the direct spelling needs one as much as the
indirect one does. The indirect case below declares that `new Foo`
holds NO arrow, which is the only place indirect object notation is
visible at all -- both spellings emit the same four ops. But a negative
alone is satisfied VACUOUSLY by a lexer that never emits `->`: one that
folded the arrow into the word beside it, or dropped it as trivia,
passes the indirect claim perfectly while getting every direct call in
this tier wrong. This is the matching positive, and the pair is what
makes either falsifiable. ONE arrow, not "at least one", which is why
this case carries the claim rather than the adjacency body: that body
makes four arrow calls and could only say something vaguer.

```perl
package Foo;
sub hi { return "hi" }
package main;
my $o = bless {}, "Foo";
print $o->hi, "\n";
```

```behavior
parses: yes
```

```output
hi
```

```tokens
one operator whose text is "->"
```

## A method name in a scalar

`$o->$m`, with the name in a variable, is a DIFFERENT op: `method`,
resolved at run time, not `method_named`. The source differs from the
named case by one sigil and the op changes entirely. `method_named`
carries the name as a constant in the op itself; `method` takes it off
the stack, so the `padsv` that supplies it is part of the call rather
than an argument to it:

    d  <0> padsv[$o:3,5] sM
    e  <0> padsv[$m:4,5] s
    f  <.> method lK/1
    g  <1> entersub[t4] lKRS/TARG

```perl
package Foo;
sub hi { return "hi" }
package main;
my $o = bless {}, "Foo";
my $m = "hi";
print $o->$m, "\n";
```

```behavior
parses: yes
```

```output
hi
```

## SUPER:: dispatch

`$o->SUPER::hi()` is a third dispatch op again: `method_super`, which
starts its search in the CURRENT package's `@ISA` rather than in the
invocant's class:

    i  <.> method_super[PV "hi"] l
    j  <1> entersub[t7] lKRS/TARG

`SUPER::` resolves against the package the call is COMPILED in, not the
one the object is blessed into, which is why this case makes the call
at file scope inside `package Derived;` rather than from inside a
method. Written the usual way -- `sub hi { $_[0]->SUPER::hi() }` -- the
op would sit in the sub's own optree, and `perl -MO=Concise,-exec` with
no sub named dumps the main program alone, so nothing would measure it.
The case is shaped by what can be observed, and says so. `our @ISA =
("Base")` is tier 02's `gv`, `rv2av` and `aassign`.

```perl
package Base;
sub hi { return "base" }
package Derived;
our @ISA = ("Base");
my $o = bless {}, "Derived";
print $o->SUPER::hi(), "\n";
```

```behavior
parses: yes
```

```output
base
```

## Indirect object notation: `new Foo`

`new Foo` and `Foo->new` compile to the SAME op stream. Indirect object
notation is a lexing problem, not a compilation one. Both spellings
emit:

    3  <0> pushmark s
    4  <$> const[PV "Foo"] sM/BARE
    5  <.> method_named[PV "new"] s
    6  <1> entersub[t2] sKRS/TARG

Nothing downstream of the lexer can tell them apart, so a corpus that
only checks behaviour or the optree cannot see the difference at all --
the same shape as the `5e-1` case in tier 01, and the reason the format
carries token facts. This is the tier's hard marker, `indirect-new`.

```perl
package Foo;
sub new { my $c = shift; return bless {}, $c }
package main;
my $o = new Foo;
print ref $o, "\n";
```

```behavior
parses: yes
```

```output
Foo
```

```tokens
one word whose text is "bless"
```

## `isa` is an infix operator, not dispatch

`$o isa Foo` emits a binary `isa` op with no call machinery at all --
no `entersub`, no `method_named` -- where `$o->isa("Foo")` emits both:

    my $infix  = $o isa Foo;      a  <2> isa sK/2
    my $method = $o->isa("Foo");  h  <.> method_named[PV "isa"] s
                                  i  <1> entersub[t5] sKRS/TARG,STRICT

One word, two unrelated parses, and only one of them is this tier's
dispatch machinery. 29 of T1's 986 files use the word and the corpus
named neither spelling; a tier that named only the method spelling
would have named the half that is NOT distinctive.

IT IS FEATURE-GATED, which is the second half of its parse and the
reason this case opens with `use v5.36`. Measured, without the feature
the infix spelling is not a weaker parse but a SYNTAX ERROR, so a lexer
cannot decide what `isa` is from the token alone. Measured under 5.42.0
it is no longer experimental: `use v5.36` with `$o isa Foo` emits no
warning, where `class` still needs a `no warnings` line.

THE FALSE CASE IS WHAT MAKES THE OUTPUT FALSIFYING. `$yes` is 1 and
`$no` is the EMPTY STRING -- perl's false, which interpolates to
nothing -- so the two run together as `1` and the `|` is what proves a
second value was printed at all. A parser that read `isa` as always
true prints `11|`; one that read it as always false prints `|`. Both
wrong readings are one byte from the right one.

`bless {}, $ENV{X} // "Foo"` rather than `Foo->new`, for two reasons:
the `//` keeps the class name a RUNTIME value so nothing folds, and
`Foo->new` would put an ARROW in a source whose whole point is that it
has none. What the facts cannot say is how many `isa`s there are --
there are TWO, and the grammar admits only `one` and `no`.

```perl
use v5.36;
package Foo;
package Bar;
package main;
my $o = bless {}, $ENV{X} // "Foo";
my $yes = $o isa Foo;
my $no = $o isa Bar;
print "$yes$no|\n";
```

```behavior
parses: yes
```

```output
1|
```

```tokens
one word whose text is "main"
one string literal whose text is "\"Foo\""
```

## The `->can(...)->()` chain is two calls

`$o->can("hi")->($o)` is TWO CALLS THROUGH ONE CHAIN, and only the
first is a method call; the second arrow dereferences a code ref.
Measured:

    b  <0> pushmark s
    c  <0> padsv[$o:3,4] sM      <- the ARGUMENT of the second call
    d  <0> pushmark s
    e  <0> padsv[$o:3,4] sM      <- the INVOCANT of the first
    f  <$> const[PV "hi"] sM
    g  <.> method_named[PV "can"] s
    h  <1> entersub[t4] sKRS/TARG    <- the method call
    i  <1> entersub[t5] lKS/TARG     <- the CODE DEREF

TWO `entersub`, ONE `method_named`. That asymmetry is the whole
construct. A parser that read BOTH arrows as method calls emits two
`method_named`; one that read the second arrow as part of the first
call emits one `entersub`. Both are ordinary-looking op streams and
both are wrong. The flags say the same thing again -- `h` is `sKRS`, a
resolved method call, and `i` is `lKS` with no `R`. 132 of T1's 986
files -- 13% -- use `can`, and the construct introduces no op of its
own, which is why it is worth a case rather than a list entry.

THE INVOCANT IS PASSED EXPLICITLY, and that is not stylistic. `can`
returns a bare CODE ref and calling a code ref passes no receiver, so
`$o->can("hi")->()` calls `hi` with an EMPTY `@_`. That is the semantic
content of "only the first is a method call".

The body returns `"called"` and not `"hi"`, which is what makes the
count of ONE string true: written the obvious way the source holds two
strings spelled `"hi"` and the fact reads `got 2, want 1`. Found by the
runner, and recorded because the obvious spelling is the broken one.
What the facts cannot say is that there are exactly TWO arrows, the
construct's defining count; the grammar admits only `one` and `no`.

```perl
package Foo;
sub hi { return "called" }
package main;
my $o = bless {}, $ENV{X} // "Foo";
print $o->can("hi")->($o), "\n";
```

```behavior
parses: yes
```

```output
called
```

```tokens
one string literal whose text is "\"hi\""
one word whose text is "can"
```
