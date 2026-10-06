# Loops

`enterloop` and `leaveloop` are tier 05's, claimed there for the bare
block -- a bare `{ ... }` is a loop that runs once. What this tier adds is
`unstack`, the op that distinguishes a loop that iterates from a block
that does not, and `enteriter`/`iter`, which are the `foreach` family.

"for" names two unrelated optrees: C-style `for` emits `enterloop` like a
`while`, `foreach` emits `enteriter`/`iter` and no `enterloop` at all.
Both close with `leaveloop`. `while` and `until` differ by the same one op
as `if` and `unless` -- the test is `and` for one and `or` for the other.
What makes a loop a loop is `enterloop`/`unstack`, not the test.

**Tier 06 control.** Introduces `cond_expr`, `die`, `enteriter`,
`entertry`, `exit`, `goto`, `iter`, `last`, `leavetry`, `next`, `redo`,
`time`, `unstack`. Depends on 05_scoping.

Every case binds a runtime value and branches on it second.
A constant condition ERASES the construct: measured, `if (1) { print "y" }`
emits `pushmark`, `const` and `print` and no branch op at all, and
`while (0) { print "y" }` emits `enter`, `nextstate` and `leave` -- an
empty program. `$ENV{X}` is unset when the harness runs, so `// N` is a
stable operand the optimiser cannot see through.

## The block `while`

Measured, a block `while` emits `enterloop`, tests with tier 04's `and`,
closes each pass with `unstack` and exits through `leaveloop`. The bare
block of tier 05 emits the first and last of those and not `unstack`,
which is why `unstack` is this tier's and the frame ops are not.

`enterloop` names all three jump targets in its own dump --
`enterloop(next->k last->p redo->c)` -- which is why `next`, `last` and
`redo` are claimed in this tier rather than deferred: the loop ops and the
jump ops are one measurement.

```perl
my $n = $ENV{N} // 3;
my $i = 0;
while ($i < $n) { print $i; $i = $i + 1 }
print "\n";
```

```behavior
parses: yes
```

```output
012
```

## `until` is `while` with the other short-circuit

Measured, the only difference from the `while` stream is one position:
`or` where `while` has `and`. Everything else -- `enterloop`, `unstack`,
`leaveloop` -- is identical. So the loop family and the conditional family
share their branch ops entirely, and there is no negation op in either
pair.

The condition is spelled `$i >= $n` rather than `!($i < $n)` for the same
reason the `unless` case does not spell its condition `!$c`: a `not` op
would be tier 04's, would appear in the stream, and would make the case
measure a different construct from the one it names. The `ge` against the
`while` case's `lt` is the source's own inversion, not perl's.

```perl
my $n = $ENV{N} // 3;
my $i = 0;
until ($i >= $n) { print $i; $i = $i + 1 }
print "\n";
```

```behavior
parses: yes
```

```output
012
```

## C-style `for` is a `while`

It emits `enterloop`, not `enteriter`. The init clause runs once before
`enterloop` as an ordinary `padsv_store`; the increment clause is compiled
into the tail of the body, ahead of the `unstack`. From the op stream's
point of view the three clauses are not three things, and the only
structural mark this form leaves that the block `while` does not is an
extra `unstack v*` between the init and `enterloop`.

Measured, `next` targets the increment clause and not the `unstack`, which
is why `next` in a C-style `for` still advances the counter and `next` in
a `while` written the same way does not.

```perl
my $n = $ENV{N} // 3;
for (my $i = 0; $i < $n; $i = $i + 1) { print $i }
print "\n";
```

```behavior
parses: yes
```

```output
012
```

## `foreach` emits `enteriter`/`iter` and no `enterloop`

A different optree from the C-style `for` that shares its keyword.
`enteriter` does what `enterloop` does -- it names the same three jump
targets in its own dump -- and additionally allocates the loop variable's
pad slot with `LVINTRO`, which is why this tier cannot sit before 05.

The per-pass `iter` is the op with no analogue in the `while` family: it
advances the list cursor and pushes the truth of "there was another one"
for the `and` to test, so the loop's condition is not written anywhere in
the source.

The list is built from `$ENV{A}` for the same reason every other case here
binds a runtime value, though the erasure is milder -- measured,
`foreach my $x (1, 2)` does keep its loop ops. The rule is uniform so that
no case in the tier has to be argued about individually.

```perl
my @l = ($ENV{A} // "a", "b", "c");
foreach my $x (@l) { print $x }
print "\n";
```

```behavior
parses: yes
```

```output
abc
```

## `do BLOCK while COND` is a post-test loop

The body runs before the condition is ever evaluated, and this is not
`while` with the parts moved. Measured with N unset, `unstack` is the only
loop op in the stream: there is no `enterloop` and no `leaveloop`
anywhere, exactly as the postfix modifier measures. So `next` and `last`
have no frame to address here either -- perl builds a plain scope and a
backward jump.

And the ORDER is the construct. The body precedes the test, where the
block `while` puts the test first and jumps over the body. A parser that
desugars this to `while (COND) { BLOCK }` emits the same op NAMES in a
different order, which is why this case pins behaviour and not an op list.

The behavioural pin is the discriminating half: N is unset, so the
condition is false on its first evaluation and the body still runs once. A
`while` loop with that same false condition prints NOTHING -- measured, it
compiles to an empty program. One line of output where the desugared
spelling produces none.

The token facts count. This source holds exactly one `do` and exactly one
`while`, so a lexer that split `do` off as a separate statement, or that
read the trailing `while` as opening a second loop, would produce a count
this case refuses.

```perl
my $i = $ENV{N} // 0;
do { print "once" } while $i > 0;
print "\n";
```

```behavior
parses: yes
```

```output
once
```

```tokens
one word whose text is "do"
one word whose text is "while"
```

## The postfix loop modifier builds no loop frame

`EXPR while COND` emits `enter`/`leave` and `unstack` but NO `enterloop`.
This is the exception that proves postfix `if`: the conditional case
measures that postfix and block `if` are the same construct to perl, and
the loop modifiers are where that reasoning fails. Same modifier syntax,
different machinery.

A postfix loop has no block, so there is nothing for `next` and `last` to
target, so perl does not build the loop frame that names those targets.
The consequence is observable rather than cosmetic: `next` inside a
postfix `while` does not address that loop, because that loop has no
`enterloop` to address.

The body is `$i = $i - 1` rather than `$i--` because `postdec` is an op no
tier here claims, and the case's subject is the loop, not the decrement.
The lint refused the shorter spelling and was right to.

```perl
my $i = $ENV{N} // 3;
print $i = $i - 1 while $i > 0;
print "\n";
```

```behavior
parses: yes
```

```output
210
```

## The postfix `foreach` iterates a LIST, and the list is not the body

`EXPR foreach LIST` is the modifier whose trailing operand is a list rather
than a condition, and it is the only modifier form that still builds the
`enteriter`/`iter` machinery. Measured, its op stream differs from the
block `foreach` above by exactly the ops of `$_` against `my $x` -- `gvsv`
where the block form has `padsv` -- so unlike the postfix `while` directly
above, which builds no loop frame at all, this one IS the block form's
optree under a different spelling.

WHICH HALF IS THE BODY is what this case exists to pin, and the output is
what pins it. `EXPR foreach LIST` runs EXPR once per element of LIST.

The discrimination is in the ASYMMETRY of the source, which is why `@l`
holds three DIFFERENT elements and the body accumulates them in order.
Only the correct reading can print `abc`: a reader that took the halves
the other way round iterates the one-element list `$s .= $_` and runs `@l`
as its body, which appends nothing and prints the empty string. A case
whose two halves were interchangeable would pass either way round, and
three ordered characters is what forbids that.

Our parser read it the swapped way while scoring ZERO refusals. The tree
was right and the EMISSION was inverted -- `$s += $_ foreach 1..3`
re-emitted as `foreach ($s += $_) 1 .. 3`, the keyword moved to the front,
the body parenthesised as though it were a condition, the `;` dropped.
That is not a parse of anything, and no refusal count could see it: 251
lines of perl.git `t/` are this shape and every one of them scored clean.
Issue 01a0dfc7.

The body is `$s .= $_` rather than `print $_` because `print $_` is a
DIFFERENT open defect -- the parenless `print $VAR` that cannot yet be
told from `print FILEHANDLE LIST` -- and a case refusing for that reason
would measure the filehandle slot rather than the modifier. `.=` compiles
to tier 01's `multiconcat`, not to a `concat` no tier here claims.

The token fact counts, and the count is the falsifying half. This source
holds exactly ONE `foreach`, so a reader that split the statement in two at
the keyword -- the body as one statement and a second loop word for the
list -- produces a count this refuses. There is no negative fact for
`for`: a negative must name a spelling REACHABLE from the source, and this
source contains no `for`, so the claim could never fail and would assert
nothing.

```perl
my @l = ($ENV{A} // "a", "b", "c");
my $s = "";
$s .= $_ foreach @l;
print "$s\n";
```

```behavior
parses: yes
```

```output
abc
```

```tokens
one word whose text is "foreach"
```

## `continue BLOCK` runs on `next` and not on `last`

There is NO `continue` op. The block is part of the loop it follows, and
the whole construct is two facts already in this tier: the `next` target on
`enterloop` points AT the continue block, and the `last` target points past
it. Measured, and the dump says it in one line:

    d  <{> enterloop(next->12 last->1c redo->e) v
    ...
    12     <0> pushmark s              <- the continue block starts HERE
    13     <0> padav[@s:3,10] lRM
    ...
    17     <0> unstack v
    1c <2> leaveloop vKP/2             <- and `last` lands HERE

So `next` reaches the continue block and `last` jumps over it. That is
the whole semantics, and it needs no op this tier does not already claim --
which is the contrast with `given`/`when`, whose five ops are unclaimed.

The OUTPUT is the discriminating half, and it discriminates in both
directions at once. `c2` with no `b2` is the pass `next` skipped, and the
continue block still ran. No `c4` at all is the pass `last` left, and the
continue block did not. A reader that attaches the block correctly but runs
it on every exit prints a `c4` this refuses; one that treats it as a second
bare block after the loop runs it ONCE, after the loop is over, and prints a
single trailing `c4` for the same reason. Three `c` entries for four passes
is what only the real semantics produces.

`while` is not the only form that takes one. Measured on 5.42.0, `until`
and the LIST form of `foreach` take one too, and a BARE block takes one --
a bare block is a loop that runs once, so it runs its continue once:

    { push @s,"b" } continue { push @s,"c" }          ->  b c

The C-style head is the exception, and perl REJECTS it rather than
accepting it with different semantics:

    $ perl -e 'for (my $i=0; $i<3; $i++) { } continue { }'
    syntax error at -e line 1, near "} continue "

The token facts count because the word is not always this construct. Bare
`continue;` is a DIFFERENT statement form -- the jump out of a `when` block,
which perl reports as `Can't "continue" outside a when block` -- so a reader
that took the word alone as a loop clause would read the wrong one. This
source holds exactly one `continue`, and it is followed by a block.

```perl
my $n = $ENV{N} // 5;
my $i = 0;
my @s;
while ($i < $n) {
    $i = $i + 1;
    next if $i == 2;
    last if $i == 4;
    push @s, "b$i";
} continue {
    push @s, "c$i";
}
print "@s\n";
```

```behavior
parses: yes
```

```output
b1 c1 c2 b3 c3
```

```tokens
one word whose text is "continue"
one word whose text is "next"
one word whose text is "last"
```

## The seen-hash idiom

`$seen{$k}++` yields the element's value from before the store. A reader that reads the element where the value is used -- after the store -- skips every key. chalk's T2 type-narrowing pass is exactly this loop, and it narrowed nothing. Found by B::SoN translating chalk's lib/ and running chalk's own suite
against the emitted Perl.

```perl
my %seen;
my $out = "";
for my $k (qw(a a b a c)) {
    next if $seen{$k}++;
    $out .= $k;
}
print "$out\n";
```

```behavior
parses: yes
```

```output
abc
```

## A post-increment of undef yields 0

`pp_postinc` sets its result to 0 when the old value is not defined. A post-decrement does not, and a string comes back as it was. Found by B::SoN translating chalk's lib/ and running chalk's own suite
against the emitted Perl.

```perl
my ($x, $y, %h);
my $p = $x++;
my $q = $y--;
my $r = $h{k}++;
my $s = "aa";
my $t = $s++;
print join(",", map { defined $_ ? "[$_]" : "undef" } $p, $q, $r, $t), "\n";
```

```behavior
parses: yes
```

```output
[0],undef,[0],[aa]
```

## Named scalars in a map, grep or foreach list

Each list element is aliased to `$_` or the loop variable, so perl flags the operand as a modifiable lvalue. It is still read for its value. Found by B::SoN translating chalk's lib/ and running chalk's own suite
against the emitted Perl.

```perl
my $p = 1 + time() * 0;
my $q = 2 + time() * 0;
my $o = "";
for my $v ($p, $q) { $o .= "[$v]" }
print join(",", map { "[$_]" } $p, $q), " ", join(",", grep { $_ > 1 } $p, $q), " $o\n";
```

```behavior
parses: yes
```

```output
[1],[2] 2 [1][2]
```
