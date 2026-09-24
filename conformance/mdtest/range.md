# Three operators spelled `..`

`..` is not one operator in three contexts; it is three operators
wearing one spelling, and CONTEXT chooses. In list context it builds a
list -- by arithmetic over numbers, by perl's magic increment over
strings -- and in scalar context it is a stateful flip-flop that
remembers whether it is on.

**Tier 03 context.** Introduces `range`, `flip` and `flop`. Depends on
02_variables.

Before these cases the corpus claimed none of the three. The only `..`
anywhere was `for my $i (1 .. 2)` in tier 07's adjacency file, which is
loop scaffolding rather than a claim.

ONE `..` EMITS ALL THREE OPS, which defeats the obvious discriminator:
`range`, `flip` and `flop` all appear for a plain list range, because
the optimiser builds the flip-flop machinery even where the context
means it can never be used. So the op set does NOT separate the three
readings and the weight falls on OUTPUT, which separates them cleanly.

Every endpoint below is a runtime value for a PLACEMENT reason, not a
parsing one. Measured, `my @r = (1 .. 3)` emits no range, flip or flop
op while `my @r = (1 .. scalar(@a))` emits all three; the same holds for
string endpoints. Constant folding is an OPTIMISER effect reaching only
the op stream -- `(1 .. 3)` lexes as `Operator("..")` and parses to a
range either way -- but a folded range emits nothing for the lint to
check against this tier's INTRODUCES block. `scalar(@a)` serves rather
than tier 04's `$ENV{X} // <default>` because `dor` is out of this
tier's op budget, and because a scalar-context array is this tier's own
subject.

## The numeric list range

`..` in list context builds the list between its endpoints. The
endpoint is `scalar(@a)` so that the tier's `range` declaration has
something to check.

The negative fact is the lexical half: `..` must lex as ONE operator.
A lexer that read it as two `.` concatenation operators has a different
program, and the source contains a `.` only inside the `..`, which is
what makes the negative reachable rather than vacuous.

```perl
my @a = (1, 2, 3);
my @r = (1 .. scalar(@a));
print "@r\n";
```

```behavior
parses: yes
```

```output
1 2 3
```

```tokens
one operator whose text is ".."
no operator whose text is "."
```

## The string range counts by magic increment

`"az" .. "bb"` is three elements and none of them is a number. It looks
like the numeric range and is a separate claim, because the SUCCESSOR
function is different: a numeric range adds one, a string range applies
the magic increment that also drives `$s++` on a string.

`az` to `ba` is the carry -- the last character wraps from `z` to `a`
and the one before it advances. A parser that read this as an arithmetic
range over numified endpoints would get `0 .. 0` and print a single
zero, since `"az"` numifies to 0.

The COUNT is the second half of the claim. `scalar(@r)` is 3, this
tier's own scalar-context idiom applied to the result -- a range in list
context producing a list whose length is then taken in scalar context,
which is the tier's subject twice over.

```perl
my @seed = ("az");
my @r = ($seed[0] .. "bb");
print "@r ", scalar(@r), "\n";
```

```behavior
parses: yes
```

```output
az ba bb 3
```

```tokens
one operator whose text is ".."
```

## `..` in scalar context is stateful

The third reading, and the one that makes the operator worth three
cases. The other two build lists; this one builds nothing and carries
state between evaluations.

INDICES 2 AND 3 ARE THE CLAIM. At both of them `$on[$_]` is 0 and
`$off[$_]` is 0 -- both operands false -- and the flip-flop selects them
anyway, because it turned on at index 1 and does not turn off until
index 4. No truthiness test can produce that: a parser that read `..`
here as a list range, or as a boolean `or`, or as anything stateless,
selects only indices 1 and 4.

That is why the operands are ARRAY ELEMENTS rather than the `$_ .. $_` a
first draft used. Measured, `grep { $_ .. $_ }` over `(0,1,0,0,1,0)`
prints `1 1` -- and so does `grep { $_ }` over the same list. The
stateful reading and plain truthiness agree, so the case would have
passed against a parser that had never heard of a flip-flop. Two
separate operand arrays are what make the state observable.

The `grep` form is deliberate: the loop spelling of the same claim needs
`enteriter` and `leaveloop`, which are tiers 05 and 06, while
`grepstart` and `grepwhile` are this tier's own.

```perl
my @on = (0, 1, 0, 0, 0, 0);
my @off = (0, 0, 0, 0, 1, 0);
my @i = (0, 1, 2, 3, 4, 5);
my @o = grep { $on[$_] .. $off[$_] } @i;
print "@o\n";
```

```behavior
parses: yes
```

```output
1 2 3 4
```

```tokens
one operator whose text is ".."
```
