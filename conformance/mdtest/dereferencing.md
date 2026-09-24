# Dereferencing

Five spellings and one hash form, all reaching through a scalar that
holds a reference -- and the reason the op stream cannot tell most of
them apart.

**Tier 08 references.** Introduces `anonlist`, `prototype`, `ref`,
`refgen`, `rv2cv`, `rv2sv`, `srefgen`. Depends on 07_subroutines.

THE OP STREAM COLLAPSES THESE SPELLINGS. `@{$r}`, `@$r` and `$r->@*`
all emit `rv2av` and are indistinguishable in it; `$r->[2]` and
`${$r}[2]` both use `multideref` and differ only in op COUNT and a
private flag, the brace form's extra op being tier 01's `padsv`. A
tier lint that compares op-name sets alone sees these cases as
identical. The `tokens` claims are the only check that reads the
spelling rather than the result, which is why every case below carries
one.

Both of the tier's hard markers live here: `deref-brace` (`${`) and
`deref-at` (`@{`).

## The arrow dereference

`$r->[2]` is ONE op: a `multideref` that has absorbed the pad lookup of
`$r` along with the subscript.

MEASURED perl 5.42.0, the whole print statement:

    a  <0> pushmark s
    b  <+> multideref($r->[2]) sK
    c  <$> const[PV "\n"] s
    d  <@> print vK

The subscript chain is fused. `$r->[2]` does not emit `padsv` then
`rv2av` then `aelem`; it emits one op naming the whole path, base
included. This is the case the brace form below is compared against,
and the pair is this tier's version of tier 01's `.5` problem: two
spellings that mean the same thing and must still be told apart. This
one has an arrow and no `${`; the next asserts the reverse.

```perl
my $r = [10, 20, 30];
print $r->[2], "\n";
```

```behavior
parses: yes
```

```output
30
```

```tokens
one operator whose text is "->"
```

## The brace dereference of an array reference

`${$r}[2]` means what `$r->[2]` means and does NOT compile to the same
ops: the brace form leaves the pad lookup as a separate `padsv`.

MEASURED perl 5.42.0, the whole print statement:

    a  <0> pushmark s
    b  <0> padsv[$r:1,4] sM/DREFAV
    c  <+> multideref(->[2]) sK
    d  <$> const[PV "\n"] s
    e  <@> print vK

Against the arrow form's single `multideref($r->[2])`: same result,
same op NAMES plus one that tier 01 already owns. The difference is
visible only in the op COUNT and the private flag, so an op-level lint
cannot distinguish the two. This case has no arrow at all, which no
op-level check can see.

```perl
my $r = [10, 20, 30];
print ${$r}[2], "\n";
```

```behavior
parses: yes
```

```output
30
```

```tokens
no variable whose text is "${$r}"
```

## The `@{ }` whole-array dereference

`@{$r}` dereferences a scalar as a whole array, and the op it emits --
`rv2av` -- is tier 02's, reached here by a route tier 02 does not have.
This is the hard marker `deref-at`, and the spelling tier 11's method
bodies are written in.

THE DEREF IS WRITTEN TWICE, and that is the point. The interpolated
`"@{$r}\n"` is what makes the output deterministic: printing the list
bare would run the three elements together as `102030`, which is true
but reads as a bug. But inside a double-quoted string our lexer
produces ONE `Quote` token and the dereference is not separately
tokenised at all -- so a case with only the interpolated form asserts
nothing about its own construct, and its negative fact would be
satisfied entirely by the `\@a` two lines above. The bare `my @c =
@{$r};` is what gives the lexer the construct to tokenise. Measured,
our lexer emits `DerefSigil(@) Operator({) Variable($r) CloseBracket(})`,
and a lexer that emitted the single token `Variable("@{$r}")` instead
would satisfy every behavioural pin in this tier.

The copy also keeps the deref in the optree in a second place: the
interpolation compiles through `join`, tier 03's, while the assignment
is a plain `rv2av` feeding an `aassign`.

```perl
my @a = (10, 20, 30);
my $r = \@a;
my @c = @{$r};
print scalar(@c), " ", "@{$r}\n";
```

```behavior
parses: yes
```

```output
3 10 20 30
```

```tokens
one operator whose text is "\\"
no variable whose text is "@{$r}"
```

## The sigil-only `@$r` dereference

`@$r` is the SAME op as `@{$r}` -- `rv2av` -- and the op stream cannot
tell which was written. Only the tokens can, and the interpolated case
above cannot make that claim, because there the whole string is one
token.

Measured, our lexer produces two tokens, `DerefSigil(@) Variable($r)`,
and a lexer that produced the single token `Variable("@$r")` instead
would satisfy every behavioural pin in this tier: perl prints the same
bytes and compiles the same ops.

MEASURED perl 5.42.0, the assignment statement -- identical to the
postfix case's except for the subscript constant its last statement
reads:

    f  <0> pushmark s
    g  <0> padsv[$r:2,4] s
    h  <1> rv2av[t5] lK/1
    i  <0> pushmark s
    j  <0> padav[@c:3,4] lRM*/LVINTRO
    k  <2> aassign[t6] vKS/COM_AGG

Copying into a named array rather than printing the list keeps `join`
in tier 03 where it belongs: printing `@$r` bare would run the three
elements together as `102030`, which is true and reads as a bug.

This case once also carried `no operator whose text is "->"`, and an
earlier note said that fact was "satisfied by the `\@a` on the line
above". That was wrong about the mechanism: a negative fact is
satisfied by ABSENCE, not by some other token standing in for it. The
fact was vacuous because the source contains no `->` bytes at all, so
no lexing of it could produce one. It was deleted along with eight
siblings by `01a0cfb2`.

```perl
my @a = (10, 20, 30);
my $r = \@a;
my @c = @$r;
print scalar(@c), " ", $c[0], "\n";
```

```behavior
parses: yes
```

```output
3 10
```

```tokens
one operator whose text is "\\"
no variable whose text is "@$r"
```

## The postfix `$r->@*` dereference

`$r->@*` is the THIRD spelling of the same `rv2av`, and the only one of
the three that puts an arrow in front of a sigil. This case is the
proof rather than the assertion: its optree and the sigil case's are
identical op for op, differing only in the subscript constant the last
statement reads.

MEASURED perl 5.42.0, the assignment statement:

    f  <0> pushmark s
    g  <0> padsv[$r:2,4] s
    h  <1> rv2av[t5] lK/1
    i  <0> pushmark s
    j  <0> padav[@c:3,4] lRM*/LVINTRO
    k  <2> aassign[t6] vKS/COM_AGG

So the tokens carry the whole distinction, and they carry it in a
direction the other two spellings do not go. `@*` is a SIGIL AND A STAR
after an arrow, which a lexer may reasonably read as the glob `*` --
measured, ours emits `Operator(->) Variable(@*)`, and a lexer that
emitted `Operator(->) Operator(@) Operator(*)` would parse as a
multiplication and return no Unknown at all. The `one operator whose
text is "->"` claim pairs with the arrow case's identical one to say
the arrow survived; the negative says the lexer did not swallow the
construct whole.

Postfix dereference has been stable since 5.24 and is not experimental;
`perlref` documents it under "Postfix Dereference Syntax".

```perl
my @a = (10, 20, 30);
my $r = \@a;
my @c = $r->@*;
print scalar(@c), " ", $c[2], "\n";
```

```behavior
parses: yes
```

```output
3 30
```

```tokens
one operator whose text is "->"
no variable whose text is "$r->@*"
```

## The `%{ }` and `${ }{ }` hash dereferences

`%{$r}` dereferences a scalar as a whole hash, and `${$r}{a}` reaches
one of its values -- both without any anonymous hash constructor. This
is the tier's other hard marker, `deref-brace`.

The hash is a NAMED one taken a reference to, not `{ a => 1 }`. The
anonymous hash constructor compiles to `emptyavhv`/`anonhash`, which
tier 11 claims where objects are built, and a tier-08 case emitting a
tier-11 op is exactly what the lint refuses. `\%h` reaches the same
place through `srefgen`, which this tier owns.

`scalar(keys %copy)` emits no `keys` op: measured, the optimiser fuses
it into `padhv[%copy] sM/KEYS`. Another instance of more construct,
fewer ops.

```perl
my %h = (a => 1, b => 2);
my $r = \%h;
my %copy = %{$r};
print scalar(keys %copy), " ", ${$r}{a}, "\n";
```

```behavior
parses: yes
```

```output
2 1
```

```tokens
one operator whose text is "\\"
```
