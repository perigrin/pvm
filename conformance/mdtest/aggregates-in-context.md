# An aggregate in two contexts

The tier's baseline: what an array becomes when something asks it for a
scalar, and what the comma operator does when the receiver is not a
list.

**Tier 03 context.** Introduces `join` and `list`. Depends on
02_variables.

The dependency is on the AGGREGATES, not on element access. `$a[0]` is
already a scalar and tells you nothing; `@a` in two contexts is the
whole subject. And context needs a thing to be in it -- `my $x = 1` is
the same program in either context, which is why this tier cannot come
before the one that supplies `@a`.

Two of the three cases below introduce no op the corpus can lint. That
is the tier's central fact rather than a gap in it: what separates the
contexts is the `s` versus `l` FLAG B::Concise prints on an op, and a
flag is not an op.

## `my $n = @a` and `my $n = scalar(@a)` are the same program

The keyword that names this tier's subject compiles to no op of its own.
Both statements compile to `padav[@a] s` followed by `padsv_store` --
byte-identical optrees. What separates scalar context from list context
is the `s` versus `l` flag on `padav`.

So this case introduces no new op, and it is the case that proves the
tier cannot be linted by op NAME alone. A tier whose first case emitted
something new would imply the tier is a set of ops; it is a set of
contexts, of which one leaves no trace at all.

```perl
my @a = (10, 20, 30);
my $n = @a;
my $m = scalar(@a);
print "$n $m\n";
```

```behavior
parses: yes
```

```output
3 3
```

```tokens
one word whose text is "scalar"
```

## `"@a"` is a `join`, decided at compile time

The source contains no function call and compiles to one. `"@a"` emits
`gvsv[*"]` -- fetching the list separator `$"` -- and then
`join[t5] sK/2`. No reader of the source would predict `join`, which is
the whole point: interpolation is a third context after scalar and list
in everything but name, and it is visible only in the optree.

The `gvsv` is tier 02's, claimed there with the package scalar; only
`join` is new here. That is what lets the case sit in 03 at all -- it
leans on an earlier tier's op rather than smuggling one in.

The separator is a single space by default, so three elements print as
`x y z`, unlike a bare `print @a` list which has nothing between its
arguments. That difference is the observable half of what the optree
shows.

```perl
my @a = ('x', 'y', 'z');
my $s = "@a";
print "$s\n";
```

```behavior
parses: yes
```

```output
x y z
```

```tokens
one variable whose text is "@a"
```

## The comma operator in scalar context discards its left operands

`my $last = (4,5,6)` is 6. The scalar half emits `pushmark v`,
`const[IV 6]`, `list sKP` -- the constants 4 and 5 are simply gone.
Two-thirds of the construct this case is about does not survive to the
optree, because discarding them is precisely what the comma operator
does here.

That makes this the tier's example of the lesson tier 01 records for
`my $x = 1+2` folding to `const[IV 3]`: ops LINT a declared tier and
cannot derive one, and a case demonstrating a construct may emit fewer
ops than the construct has parts.

The list half is the discriminator. The same parenthesised commas
assigned to an array keep all three values, so the pair shows that the
source text does not decide -- the assignment target does.

```perl
my $last = (4, 5, 6);
my @all = (4, 5, 6);
print "$last @all\n";
```

```behavior
parses: yes
```

```output
6 4 5 6
```

```tokens
one variable whose text is "$last"
one variable whose text is "@all"
```
