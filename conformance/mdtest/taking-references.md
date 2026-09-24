# Taking a reference

The constructs that PRODUCE a reference -- the backslash in both
contexts, the anonymous array constructor -- and the builtin that reads
one back.

**Tier 08 references.** Introduces `anonlist`, `prototype`, `ref`,
`refgen`, `rv2cv`, `rv2sv`, `srefgen`. Depends on 07_subroutines.

Every case takes a reference to a VARIABLE, never to a literal. `my $r
= \1` arrives as `const[IV \1] s/FOLD` and emits no `srefgen` at all:
the optimiser erases the very construct the tier is about. The array
and hash fixtures are what defeat that, and it is the same lesson tier
01 learned from `1+2`.

Nor does any case print a reference. A reference stringifies as
`ARRAY(0x5606f0a12345)` and the address changes every run, so `ref`
supplies the stable category name instead.

## The backslash in scalar context

`\@a` in scalar context is one op, `srefgen`, and the array it points
at stays reachable through the scalar that holds it.

The optree is `padav[@a] lRM`, then `srefgen sK/1`, then
`padsv_store[$s]` -- the array is loaded first and the reference taken
of it, which is why this tier cannot precede tier 02.

```perl
my @a = (10, 20);
my $s = \@a;
print "$$s[0] $$s[1]\n";
```

```behavior
parses: yes
```

```output
10 20
```

```tokens
one operator whose text is "\\"
```

## The same backslash in list context

The SAME backslash in list context is a DIFFERENT op: `\(@a)` emits
`refgen`, not `srefgen`, and distributes over the array's elements.

This case and the scalar one above differ in source only by the
parentheses and the assignment target, and perl compiles them to two
different ops. That is why the tier claims both: one spelling in the
source is two ops in the optree, and a corpus that wrote only the
scalar form would claim an op it never exercised.

The distribution is the second surprise. `\(@a)` is not a reference to
the array -- it is a LIST of references, one per element, so `$r[0]` is
a SCALAR reference and `${$r[0]}` is the element behind it. Writing
`$r[0]->[0]` instead dies with "Not an ARRAY reference", which is how
this case arrived at its current form.

```perl
my @a = (10, 20);
my @r = \(@a);
print ${$r[0]}, " ", ${$r[1]}, "\n";
```

```behavior
parses: yes
```

```output
10 20
```

```tokens
one operator whose text is "\\"
```

## The anonymous array constructor

`[10, 20, 30]` builds an array and yields a reference to it in one op,
`anonlist`, with no named array anywhere in the program.

The anonymous HASH constructor is deliberately NOT here. `{}` and
`{ %a }` compile to `emptyavhv` and `anonhash`, which tier 11 claims
where objects are built; writing one in this tier would use an op a
LATER tier owns and the lint would refuse it. `[...]` alone is tier
08's.

`ref` is what makes the assertion deterministic -- it prints the stable
category name rather than an address that changes every run.

```perl
my $r = [10, 20, 30];
print scalar(@$r), " ", ref($r), "\n";
```

```behavior
parses: yes
```

```output
3 ARRAY
```

```tokens
one operator whose text is "["
```

## `ref` over all three referent kinds

`ref` reports what a reference points at, and it is the one operation
in this tier that has no home in an earlier one. The tier description
does not mention it; the op measurement is where it came from. Its
operand is a reference and its result is a plain string, so no earlier
tier can claim it -- there is nothing for it to be about before this
tier.

It is also what every other case here leans on for determinism.

The three categories are measured together because the op is the same
for all of them: `ref` does not discriminate by sigil, the referent
does.

```perl
my @a = (1);
my %h = (k => 1);
my $n = 5;
my $ar = \@a;
my $hr = \%h;
my $sr = \$n;
print ref($ar), " ", ref($hr), " ", ref($sr), " ", ${$sr}, "\n";
```

```behavior
parses: yes
```

```output
ARRAY HASH SCALAR 5
```

```tokens
one operator whose text is "{"
```
