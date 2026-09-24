# Every construct, each beside another

One body holding every construct this tier introduces, each adjacent to
another -- and adjacent to tier 02's array, which is what this tier
depends on.

**Tier 03 context.** Introduces nothing of its own; the mixture is the
subject. Depends on 02_variables.

The tier's other cases are one construct each, which makes them
diagnosable: when the `sort` case fails, the construct that failed is
the only one present. That property is also why such cases cannot reach
an adjacency bug -- a parser that handles every construct alone and
mis-handles a pair goes green over the pair.

The constructs sit on consecutive STATEMENTS rather than nested, because
nesting them would need an operator to join them and this tier comes
before the operator tier.

## The whole tier in one body

EVERY DISCRIMINATING CONSTRUCT APPEARS IN BOTH CONTEXTS, which is what
distinguishes this from a list of the tier's constructs. The tier's
subject is not a set of constructs, it is a set of PAIRS whose halves
emit the same op and differ only by the `s` versus `l` flag. A body
holding one half of each asks the parser for one answer where the
construct has two, and a parser that collapsed the two contexts into a
single answer would go green over it.

So `@a` appears in scalar and in list context, `reverse` twice, the
empty-list count idiom twice, `localtime` twice, `caller` twice, and
`map` and `grep` twice each. Measured, that is `padav s`/`padav l`,
`reverse sK/1`/`reverse lK/1`, `aassign sKS`/`aassign lKPS`,
`localtime s`/`localtime l`, `caller s`/`caller l`,
`mapstart sK`/`mapstart lK` and `grepstart sK`/`grepstart lK` -- seven
pairs in one body, with the tier's remaining constructs in scope.

The adjacency with the earlier tier is the first line: `@a` is tier 02's
array, and every statement after it puts that same array in a different
context. Pairing this tier with the one it depends on is the array being
re-used, not a separate statement about arrays.

READING THE OUTPUT. `sort` is a default string sort, so `(3,1,2)` comes
back `1 2 3`. The reverse of that array in scalar context is `213` --
the elements joined with nothing between them, then the string reversed
-- not `2 1 3`, which is the list-context `@revd` on the next line.
`$moved` is 3 and `@kept` is empty from the SAME idiom. `$fields` is 9
because list-context `localtime` returns nine elements; `$stamp` is the
scalar half, a formatted string, and it is COUNTED rather than printed
-- pinning it would make this case depend on the clock and the timezone,
while `@seen` holds two elements whatever the time is. `$who` and
`@frame` are `caller` in the two contexts, and at file scope they return
different AMOUNTS: `$who` is undef and `@frame` is the EMPTY list, which
is the `0` after the `3`.

`@mapped` is the three elements and `$mapcount` is 3; `@kept2` is the
three elements and `$grepcount` is 3 as well, because every element of
`(3, 1, 2)` is true and nothing is filtered. That grep keeps everything
HERE is not a weakening -- the `grep` case is where grep filters, using
an element bound to `$ENV{G}`, and this body's job is adjacency.
Bringing `$ENV` in here would add an opacity it does not otherwise need.
Neither `map` nor `grep` FOLDS despite `@a` being wholly constant, which
the comma operator two statements up does: measured, a block is not
constant-foldable however constant its input.

`..`'s three readings close the body. `@ranged` is the numeric list
range, `@lettered` is the string range whose successor is the magic
increment (`az ba bb`, carrying), and `@window` is the scalar-context
FLIP-FLOP whose `1 2 3 4` selects indices 2 and 3 where both operands
are false.

Fewer ops appear than the constructs suggest. `padrange` fuses
consecutive `my` declarations, and the comma operator in scalar context
erases its own left operands -- so the op set is a UNION across the
tier's cases rather than a property of this one. It happens that this
body does emit all seven pairs, but that is a fact about this program,
not a rule the format requires.

```perl
my @a = (3, 1, 2);
my $count = @a;
my @copy = @a;
my $interp = "@a";
my $last = (4, 5, 6);
my @sorted = sort @a;
my $moved = () = sort @a;
my @kept = (() = sort @a);
my $rev = reverse @a;
my @revd = reverse @a;
my $fields = () = localtime;
my $stamp = localtime;
my $want = wantarray;
my $who = caller;
my @frame = caller;
my @mapped = map { $_ } @a;
my $mapcount = map { $_ } @a;
my @kept2 = grep { $_ } @a;
my $grepcount = grep { $_ } @a;
my @ranged = (1 .. scalar(@a));
my @seed = ("az");
my @lettered = ($seed[0] .. "bb");
my @on = (0, 1, 0, 0, 0, 0);
my @off = (0, 0, 0, 0, 1, 0);
my @idx = (0, 1, 2, 3, 4, 5);
my @window = grep { $on[$_] .. $off[$_] } @idx;
my @seen = ($want, $stamp, $who);
print "$count @copy $interp $last @sorted $moved ", scalar(@kept),
    " $rev @revd $fields ", scalar(@seen), " ", scalar(@frame),
    " @mapped $mapcount @kept2 $grepcount",
    " @ranged @lettered @window\n";
```

```behavior
parses: yes
```

```output
3 3 1 2 3 1 2 6 1 2 3 3 0 213 2 1 3 9 3 0 3 1 2 3 3 1 2 3 1 2 3 az ba bb 1 2 3 4
```
