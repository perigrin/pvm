# List operators, each asked twice

Five operators that answer a different question depending on what
received the answer. Each appears twice below, once with the `s` flag
and once with `l`, because one half alone asks the parser for one answer
where the construct has two.

**Tier 03 context.** Introduces `reverse`, `sort`, `localtime`,
`mapstart`, `mapwhile`, `grepstart`, `grepwhile`. Depends on
02_variables.

None of these cases is separated from its own other half by an op NAME.
`my $x = reverse @a` emits `reverse sK/1` and `my @b = reverse @a`
emits `reverse lK/1` -- same op, same operand, different answer. The
discrimination is in the flag and in the OUTPUT, which is why the output
blocks carry this topic's weight.

## `reverse`: reversed elements, or a reversed string

In list context `reverse` reverses the elements; in scalar context it
concatenates them and reverses the characters. Same op and same operand
twice: `my @b = reverse @a` emits `reverse lK/1`, `my $s = reverse @a`
emits `reverse sK/1`. The CONTEXT FLAG is the claim; the targ number
beside it is pad allocation and is not stable enough to assert.

The scalar result is `321` and not `3 2 1`: reversing in scalar context
first joins the list with nothing between the elements, then reverses
the resulting string. A parser that treats the two as one operation on
"a value" has no way to say which of these it means.

```perl
my @a = (1, 2, 3);
my @b = reverse @a;
my $s = reverse @a;
print "@b $s\n";
```

```behavior
parses: yes
```

```output
3 2 1 321
```

```tokens
one variable whose text is "@b"
one variable whose text is "$s"
```

## `sort`, and the `() =` count idiom in both contexts

`sort` in list context returns the sorted elements. The second statement
is the count idiom, `my $n = () = sort @a`, which is the only way to
observe the list-versus-scalar distinction on an `aassign`: the inner
assignment to an empty list runs in list context, and the outer
assignment then asks it how many elements it moved --
`aassign sKS` -- the `s` is the claim, and the targ number beside it
is pad allocation rather than something to assert.

The third statement is that same idiom with an ARRAY receiving it, and
it is what makes the second a pair rather than a single measurement:
`my @c = (() = sort @a)` emits an `aassign` too. Same op, same
operand, same spelling of the idiom, and unrelated answers. In scalar
context the empty-list assignment reports how many elements it moved,
which is 3; in list context it yields what it assigned, which is the
empty list, so `@c` has 0 elements. Nothing in the source says which.
Without the third statement the `aassign` here is only ever seen with
`s`, and the claim that context is a flag would be made by a case that
never shows the other value of the flag.

No comparator block appears on purpose: `sort { $a <=> $b } @a` would
drag in a block and the comparison operator, which belong to later
tiers. Default sort is a string sort, which is why the digits come back
in the order they do.

```perl
my @a = (3, 1, 2);
my @b = sort @a;
my $n = () = sort @a;
my @c = (() = sort @a);
print "@b $n ", scalar(@c), "\n";
```

```behavior
parses: yes
```

```output
1 2 3 3 0
```

```tokens
one variable whose text is "$n"
one variable whose text is "@b"
```

## `localtime`: a nine-element list, or one string

The sharpest discriminator in the tier, because its two contexts return
unrelated TYPES. `reverse` and `sort` return rearrangements of the same
data either way. `localtime` in list context is
(sec,min,hour,mday,mon,year,wday,yday,isdst), and in scalar context a
formatted string like `Sun Sep 21 14:03:11 2026`. Nothing about the call
site says which; the assignment target decides.

What this case pins is deliberately NOT the time. `localtime` depends on
the clock and the timezone, so pinning either result verbatim would make
it fail tomorrow and in another zone. Both assertions are facts about
SHAPE: the list has nine elements always, and the scalar string is one
element always. That is exactly the contextual difference, and it is the
part that does not move.

```perl
my $n = () = localtime;
my $s = localtime;
my @c = ($s);
print "$n ", scalar(@c), "\n";
```

```behavior
parses: yes
```

```output
9 1
```

```tokens
one variable whose text is "$n"
one word whose text is "scalar"
```

## `map BLOCK` and `map EXPR`, a fork the optree cannot see

`map` takes either a BLOCK or an EXPRESSION, and the two forms are a
genuine parse fork -- `map BLOCK LIST` takes no comma after the block,
`map EXPR, LIST` requires one. Measured, `my @b = map { $_ } @a` and
`my @b = map($_, @a)` emit the same ops, in the same order, with the
same flags and targs:

    pushmark, pushmark, padav[@a] lM, mapstart lK, mapwhile lK, gvsv[*_]

So nothing an op-name lint reads distinguishes the halves, and a parser
that took the comma form for the block form emits an identical optree.
This case therefore asserts at the TOKEN STREAM: the comma IS the fork,
the source is written to contain exactly one (`qw()` builds the list
without any), and the token fact counts it. A parser that admitted a
comma after a block, or required one, changes that count.

What is not identical is not an op. The two optrees differ in the COP
SEQUENCE RANGE beside each lexical: `padav[@a:1,5]` under the block form
against `padav[@a:1,3]` under the expression form. A block is a SCOPE,
and measured, a bare `{ 1; }` standing where the map block stands
advances the counter by exactly the same 2. The trace is in pad
metadata, which the op-name lint does not read. "Byte-identical" is too
strong; "identical in every op" is the measured claim.

The fourth statement is the discriminating pair: `my $n = map { $_ } @a`
emits `mapstart sK` where the two before it emit `mapstart lK`, and in
scalar context the answer is the COUNT of what the list form returns.

The list is opened up by `$ENV{M}` because a CONSTANT argument erases
the construct. `$ENV{M}` is unset when the runner executes, so
`"$ENV{M}a"` is the one-character string `a` -- opaque at compile time,
fixed at run time, and the reason the first element comes back `a`
rather than the `x` that `qw` put there. No arithmetic appears in the
block on purpose: `map { $_ + 1 } @a` emits `add`, which is tier 04's. A
bare `$_` is the smallest body that is still a body.

```perl
my @a = qw(x y z);
$a[0] = "$ENV{M}a";
my @block = map { $_ } @a;
my @expr  = map($_, @a);
my $n = map { $_ } @a;
print "@block @expr $n\n";
```

```behavior
parses: yes
```

```output
a y z a y z 3
```

```tokens
one operator whose text is ","
one variable whose text is "@block"
one variable whose text is "@expr"
```

## `grep BLOCK` and `grep EXPR`, where the arities disagree

The same fork as `map`, measured again for the other keyword and
holding: block form and comma form emit

    pushmark, pushmark, padav[@a] lM, grepstart lK, grepwhile lK, gvsv[*_]

The syntactic difference is a COMMA, which the optree does not record,
so the token stream is the only place the fork is observable. The source
holds exactly one comma and the token fact counts it. As with `map`, the
block form advances each lexical's COP SEQUENCE RANGE by 2, which is pad
metadata rather than an op.

WHERE grep IS SHARPER THAN map, and the reason both cases exist rather
than one: `map` in scalar context returns the count of what it would
have returned in list context, so its two halves report the same NUMBER
in different shapes. `grep` FILTERS, so the count is not the input's --
from three elements it keeps two, the scalar half reports 2 and the list
half produces the two surviving strings. Nothing folds; the
discrimination is between a count and the things counted, and they have
different arities from the input as well as from each other.

`$a[0]` is set from `$ENV{G}` for two reasons at once. A constant
argument erases the construct -- a wholly-constant grep folds and leaves
no `grepstart` to measure -- and grep needs an element that is FALSE, or
it filters nothing and the two contexts stop disagreeing about arity.
`$ENV{G}` is unset when the runner executes, so the element is undef,
which is both opaque at compile time and false at run time.

Written as a bare `$a[0] = $ENV{G}` rather than `"$ENV{G}"`, which is
what the `map` case uses. Measured, the interpolating form of a lone
variable emits `stringify`, and `stringify` is claimed by NO tier -- so
quotes that read as harmless would fail the corpus lint for an op the
READMEs do not yet describe. The plain assignment emits
`aelemfastlex_store` and `multideref`, both tier 02's. No comparison
appears in the block on purpose: `grep { $_ gt "a" } @a` emits `gt`,
which is tier 04's. A bare `$_` tests the element's own truth, which is
the smallest predicate there is.

```perl
my @a = qw(x y z);
$a[0] = $ENV{G};
my @block = grep { $_ } @a;
my @expr  = grep($_, @a);
my $n = grep { $_ } @a;
print "@block @expr $n\n";
```

```behavior
parses: yes
```

```output
y z y z 2
```

```tokens
one operator whose text is ","
one variable whose text is "@block"
one variable whose text is "@expr"
```
