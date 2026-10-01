# The aggregate-argument operators

`push`, `unshift`, `shift`, `values` and `each` share an ARGUMENT RULE nothing
else in the corpus has: the first argument is the aggregate ITSELF, not
an expression to be flattened.

**Tier 02 variables.** Introduces `aassign`, `aelem`, `aelemfast`,
`aelemfast_lex`, `aelemfastlex_store`, `aslice`, `av2arylen`, `delete`,
`each`, `gv`, `gvsv`, `helem`, `hslice`, `multideref`, `padav`, `padhv`,
`push`, `rv2av`, `rv2hv`, `sassign`, `shift`, `unshift`, `values`. Depends on
01_literals.

Measured, `push @a, @tail` emits `padav[@a] lRM` and `padav[@tail] l` --
the SAME op with different flags, the container slot against the
flattened slot. A parser that flattens the first slot builds a tree perl
does not build while printing something plausible, which is why these
cases pin an ELEMENT as well as a count.

They also emit ops of their own, which is the opposite of what the
`exists`/`delete` cases would lead a reader to expect, so the contrast
is worth stating. `keys` is the trap: `scalar(keys %h)` compiles `keys`
away to the flag `sM/KEYS`, but measured, `my @k = keys %h` in LIST
context emits a `keys` op just as `values` does. The flag is a property
of the CONTEXT, not of the keyword, so nothing below writes `keys` at
all rather than quietly adding an op the tier has never claimed.

## `push`: the first argument is an array, not an expression

Every other list operator in the corpus takes expressions and
flattens them. `push` does not: its first argument slot holds the array
ITSELF, and only the arguments after the comma are flattened into it. A
parser that treats `push @a, @tail` as both arrays flattened builds a
tree perl does not build, while producing output plausible enough that
nothing behavioural would notice.

`lRM` is the aggregate slot -- lvalue, ref-modify, the array as a
container. `l` is the flattened slot. The flags are the whole
distinction, and this case pins THREE consequences of them, because each
alone is reachable by a wrong tree. `@tail` still holds TWO elements:
the flattened slot is read, not consumed, so a parser that MOVED the
elements rather than copying them fails here and nowhere else. `@a`
holds THREE, which is the count a flattening parser would also reach --
so the count alone proves nothing, and it is printed only to make the
third line legible. `$a[2]` is 30, the LAST element of `@tail`, which
pins the order the flattening put them in.

There is no constant-folding trap here, and the tier's `$ENV{X}` idiom
would be wrong to copy: measured, `push @a, 3` with everything constant
still emits `push`, because an array is a runtime container and the
optimiser has nothing to fold it into. (`//` is tier 04's `dor`, so the
idiom is out of budget here regardless.)

```perl
my @a = (10);
my @tail = (20, 30);
push @a, @tail;
print scalar(@tail), "\n";
print scalar(@a), "\n";
print $a[2], "\n";
```

```behavior
parses: yes
```

```output
2
3
30
```

```tokens
one word whose text is "push"
```

## `unshift`: the same rule at the other end

This case exists because the rule and the OP are separable
claims. A parser can hard-code `push`'s aggregate first slot as a
special case for the one keyword it was tested on, and `unshift` is the
second keyword that rule has to cover. T1 uses it in 15 files.

Measured, the flags are `push`'s exactly over a different op: `lRM` on
the container, `l` on the flattened list, and `unshift` where `push` had
`push`. So the tier claims a second op and no new rule.

The ELEMENT is what separates the two behaviourally. `@head` still holds
ONE element -- the flattened slot read, not consumed, which is the
`push` case's first line mirrored. `@a` holds three, the count a
flattening parser also reaches. `$a[0]` is 10, the value that arrived
from `@head`, and that is the line separating this operator from `push`:
under `push` the same three-element result would have 20 there. A parser
that got the operator right and the END wrong passes a count check and
fails this.

The negative token fact is the falsifying half: NO word spelled `shift`.
A lexer that read `unshift` as `un` followed by `shift` -- or that
longest-matched the keyword table wrongly -- would produce a `shift`
here and fail. `shift` is a different op, the next case's, so it is also
the spelling this case must not accidentally contain.

```perl
my @a = (20, 30);
my @head = (10);
unshift @a, @head;
print scalar(@head), "\n";
print scalar(@a), "\n";
print $a[0], "\n";
```

```behavior
parses: yes
```

```output
1
3
10
```

```tokens
one word whose text is "unshift"
no word whose text is "shift"
```

## `shift`: the rule again, and the array gives up an element

`shift @a` takes the aggregate itself as `push` and `unshift` do, and is
the first of the three to CONSUME: it removes `@a`'s first element and
returns it. Measured, its first slot is `padav[@a] lRM`, the container
flags again, and the op is `shift sK/1`. A parser that flattened the
slot would hand `shift` a list it cannot remove from.

The value and the count separate it from a read: `$x` is 10, the element
that was first, and `@a` is left with two. `my $self = shift` in a method
body (tier 11) is the same op with the array left implicit.

```perl
my @a = (10, 20, 30);
my $x = shift @a;
print $x, "\n";
print scalar(@a), "\n";
```

```behavior
parses: yes
```

```output
10
2
```

## `values` takes the container, not a flattened list

`values %h` is not `values(%h)` with the hash flattened into a
list of key-value pairs. The hash is the argument, whole, and the
operator reads its values out of the container. A parser that flattens
`%h` first hands `values` a flat list of scalars where perl hands it one
hash -- and for a one-key hash the printed answer would still look
right, which is why this case pins the op and a token fact rather than
only the output.

Measured, the hash reaches the operator as `padhv`, the container op,
not as an aassign'd list: `padhv[%h:1,3] lRM` then `values[t4] lK/1`.

ONE KEY, DELIBERATELY. Hash order is not guaranteed, so a case printing
the values of a two-key hash would pin an expectation perl does not
promise. With one key the list has one member and its value is
determined.

```perl
my %h = (a => 7);
my @v = values %h;
print scalar(@v), "\n";
print $v[0], "\n";
```

```behavior
parses: yes
```

```output
1
7
```

```tokens
one word whose text is "values"
```

## `each` returns a PAIR where its siblings return a flat list

`values` and `keys` take the same argument and return a flat list
of one thing per entry. `each` takes the same argument and returns TWO
-- one key and one value -- so the list-assignment on its left has two
scalars to fill. A parser that models `each` on its siblings gives the
construct the wrong arity and the pair silently collapses to the key.

Measured, the hash arrives as `padhv` and `each` is a real op, not a
flag: `padhv[%h:1,3] lRM`, `each lK/1`, then `padrange[$k:2,3; $v:2,3]
RM/LVINTRO,range=2` and `aassign[t5] vKS/COM_RC1`. The `range=2` on the
padrange is the arity, visible: two pad slots are introduced as one
range because the assignment's left side is a two-element list. A
one-element return would have had a padsv there.

NO LOOP, WHICH IS THE POINT OF THE SPELLING. T1 writes `while (my ($k,
$v) = each %h)`, and measured, that form drags in `enterloop`,
`leaveloop`, `and` and `unstack` -- ops belonging to tiers 05, 04 and
06. A bare list assignment reaches the same `each` with nothing this
tier does not claim, and the iterator's LOOPING is a control-flow claim
rather than a claim about how `each` parses.

ONE KEY, for the `values` case's reason: hash order is not guaranteed,
so a two-key hash would make the pair unpredictable and the pinned
output a guess. With one key the first iteration is determined.

```perl
my %h = (a => 7);
my ($k, $v) = each %h;
print $k, "\n";
print $v, "\n";
```

```behavior
parses: yes
```

```output
a
7
```

```tokens
one word whose text is "each"
```
