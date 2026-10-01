# Hashes

Naming a hash, subscripting it with braces, and the two words -- `exists`
and `delete` -- that mostly leave no op behind.

**Tier 02 variables.** Introduces `aassign`, `aelem`, `aelemfast`,
`aelemfast_lex`, `aelemfastlex_store`, `aslice`, `av2arylen`, `delete`,
`each`, `gv`, `gvsv`, `helem`, `hslice`, `multideref`, `padav`, `padhv`,
`push`, `rv2av`, `rv2hv`, `sassign`, `shift`, `unshift`, `values`. Depends on
01_literals.

Hash order is not guaranteed, so nothing below prints a hash's contents.
`scalar(keys %h)` asks how many rather than which, and that is the count
every case here pins.

## A hash, and a bareword key

A hash is named with `%` and subscripted with braces; `$h{a}` is
one element of it, and the bareword key needs no quotes.

`$h{a}` emits no `helem`. It compiles to `multideref($h{"a"})`, the op
that swallows most element access in this tier; `helem` survives only
where the subscript is an expression, which is the next case. `keys`
likewise emits no op of its own here -- it becomes a FLAG on the
`padhv`, `sM/KEYS`. That flag is a property of SCALAR context, not of
the keyword; in list context `keys` emits an op, which is why the
aggregate-operators topic writes `values` and no `keys` at all.

```perl
my %h = (a => 1, b => 2);
print scalar(keys %h), "\n";
print $h{a}, "\n";
```

```behavior
parses: yes
```

```output
2
1
```

## A bareword key may be spelled like a quote-like operator

`$h{m}` is the key `"m"`, not a match. Every name in perl's quote-like
set -- `q qq qw qx m qr s tr y` -- autoquotes as a lone bareword
subscript, measured on 5.42.0:

    $ perl -MO=Deparse -e 'my %h; my $a=$h{m}; my $b=$h{s}; my $c=$h{tr};'
      ->  $h{'m'}  $h{'s'}  $h{'tr'}

THE OPTREE CANNOT SEE THIS CASE, which is why the token fact carries it.
The ops are the same `multideref` the case above emits, so a corpus that
asserted only ops would hold `$h{m}` and `$h{a}` to be the same claim.
They are not: the LEXER has to decide whether `m` opens an operator, and
getting it wrong swallows source. Before the fix, `$h{m}` lexed as a
match whose delimiter was `}`, so the body ran to the NEXT `}` and the
emission gained a spurious `};`.

A `}` never delimits one of these operators in a program perl will
COMPILE, in any context and not only a subscript: `perl -e 'sub f { m }'`
is "Search pattern not terminated". So the subscript is the only spelling
that compiles, and a lexer needs no bracket-stack knowledge to tell the
two apart -- the byte after the name settles it. A brace-DELIMITED
operator in the same position stays an operator, because there that byte
is `{`: `$h{ m{a} }` deparses to `$h{/a/}`.

The hash is built with QUOTED keys so each bareword below appears
exactly once, which is what a `one ...` fact requires.

```perl
my %h = ("m" => 1, "s" => 2, "tr" => 3, "qw" => 4);
print $h{m}, $h{s}, $h{tr}, $h{qw}, "\n";
```

```behavior
parses: yes
```

```output
1234
```

```tokens
one word whose text is "m"
one word whose text is "s"
one word whose text is "tr"
one word whose text is "qw"
```

## A computed subscript is the only route to `helem`

`$h{a}` is `multideref`. `$h{$k[0]}` is `helem` over an
`aelemfast_lex` -- same construct in the source, different op, decided
entirely by what is inside the braces. This case exists because it is
the only way to reach `helem` at all, and a tier that claimed `helem`
without it would be claiming an op no file emits.

```perl
my %h = (a => 1, b => 2);
my @k = ("b");
print $h{$k[0]}, "\n";
```

```behavior
parses: yes
```

```output
2
```

## `exists` and `delete` on an element leave no op of their own

Measured, `exists $h{a}` compiles to `multideref($h{"a"})
sK/EXISTS` and `delete $h{a}` to `multideref($h{"a"}) sK/DELETE`: both
survive only as a FLAG on an op named after something else. A tier
derived from op names alone would contain no notion of `exists` at all,
and the slice case below would be the only evidence `delete` exists --
for the wrong reason, since the slice is the spelling where the
optimiser DECLINES to fold. This case is the common path; that one is
the exception.

BOTH TRUTH VALUES OF `exists` ARE HERE because the false one is a
different claim. Perl's false is the empty string, so `print exists
$h{z}` prints NOTHING and the pinned output carries a blank line -- a
case that tested only the true value could not tell `exists` from a
construct that always yields 1.

`print exists $h{a}` rather than `print exists $h{a} ? 1 : 0`: measured,
the conditional adds `cond_expr` and `0+(...)` adds `add`, and both are
ops later tiers introduce. The plain print is the only spelling of this
construct that stays inside the tier.

`delete` in scalar context returns the value removed, which is what
makes it observable without printing the hash.

```perl
my %h = (a => 1, b => 2);
print exists $h{a}, "\n";
print exists $h{z}, "\n";
my $gone = delete $h{a};
print $gone, "\n";
print scalar(keys %h), "\n";
```

```behavior
parses: yes
```

```output
1

1
1
```

## A hash slice, where `delete` becomes a real op

`@h{...}` is a hash slice. `delete $h{a}` emits NO delete op, but
deleting a SLICE is different: the optimiser declines to build a
multideref for it, so `delete vK/SLICE` appears as a real op. That is
why the tier claims `delete` for the slice and not for the element, and
why `exists` is not claimed at all -- it has no op in any spelling.

```perl
my %h = (a => 1, b => 2, c => 3);
delete @h{"a", "b"};
print scalar(keys %h), "\n";
print( (@h{"c"}), "\n");
```

```behavior
parses: yes
```

```output
1
3
```
