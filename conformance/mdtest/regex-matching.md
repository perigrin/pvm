# Matching and substitution

The tier's baseline: a pattern bound to a string, negated, interpolated,
and made to mutate its target.

**Tier 09 regex.** Introduces `match`, `pos`, `qr`, `regcomp`, `split`,
`subst`, `trans`. Depends on 01_literals -- a match needs something to
match against, and the smallest such thing is a string literal.

Two measurements run through every case here. `=~` EMITS NO OP OF ITS
OWN, so the optree cannot tell you the operator was written; and a
constant pattern is compiled at COMPILE time while an interpolated one
is not, so one character of source is worth a whole op. Both are
recorded by the pairs below rather than by any single case.

## A constant pattern compiles once

A pattern with no interpolation is compiled at COMPILE time: the match
emits `match` alone, with no `regcomp` beside it. This is the tier's
baseline, and measuring the constant form first is what makes the
interpolated case's extra op attributable.

The token claims are negative, and deliberately so. GLOSSARY.md's
`quote-like operator` is defined as a quote spelled with an OPERATOR
NAME -- q, qq, m, s, qr -- or with backticks; a bare `/abc/` is neither,
so there is no positive category to assert it under and this case does
not invent one. What it can say is what a mis-lex would produce: the
hazard for a bare-slash pattern is the slash read as DIVISION and the
pattern body read as Perl, which yields an `Operator("/")` and a
`Word("abc")`. Asserting their absence is the claim that survives.

```perl
my $m = "abc" =~ /abc/;
print "$m\n";
```

```behavior
parses: yes
```

```output
1
```

```tokens
no operator whose text is "/"
no word whose text is "abc"
```

## The binding operator emits nothing

`=~` emits NO OP OF ITS OWN. The binding operator is absorbed into the
match op, which carries its target in its own flags -- the stream shows
`match(/"b"/)[$s:1,3]` and no `bind` anywhere.

So this case and the bare match above differ in their OPERAND, not in
their op set. That is why the binding needs a case of its own: the
optree cannot tell you the operator was written, only that the match
found a target other than `$_`, and the token stream is what records the
operator.

```perl
my $s = "abc";
my $m = $s =~ /b/;
print "$m\n";
```

```behavior
parses: yes
```

```output
1
```

```tokens
one operator whose text is "=~"
no operator whose text is "/"
```

## The negated binding is two ops

`!~` is not a match op with a flag: it is `match` followed by `not`, a
second op stacked on top of the first. `not` is tier 04's op, claimed
there with the logical operators, and its appearance here is the negated
binding REUSING it rather than this tier introducing anything. The
measurement is still this tier's: reading the source would suggest a
negated match op, and there is none.

The pattern deliberately does not match, so the printed value is the
empty string rather than 0 -- Perl's false is `""` -- which is why the
brackets are there. Without them the pinned output would be a line with
nothing on it, and a mis-measured match would print the same.

```perl
my $s = "abc";
my $m = $s !~ /zzz/;
print "[$m]\n";
```

```behavior
parses: yes
```

```output
[1]
```

```tokens
one operator whose text is "!~"
no operator whose text is "!"
no operator whose text is "/"
```

## Substitution mutates in place

A substitution mutates its target in place and emits one `subst` op. No
assignment, no block, no dereference -- which is the measurement behind
this tier's claim that it could sit at 03.

The replacement `z` arrives as a `const` pushed before the `subst`, not
as a second argument to it, and there is no `sassign`: `$s` is named in
the subst op's own flags exactly as the match op names its target.

```perl
my $s = "abc";
$s =~ s/a/z/;
print "$s\n";
```

```behavior
parses: yes
```

```output
zbc
```

```tokens
one quote-like operator whose text is "s/a/z/"
```

## An interpolated pattern adds `regcomp`

An interpolated pattern cannot be compiled at compile time, so the
stream gains `regcomp` between the `padsv` that fetches the variable and
the `match` that uses the result. This is the ONLY form in the tier that
emits `regcomp`. Against the bare match above, one character of source
differs and a whole op appears; the tier's op set is their union.

Note also what the match op loses: `match()[$s:2,4]` carries no pattern
in its own dump, because it has none until `regcomp` runs.

The token claim is the sharpest in the tier. `$p` appears TWICE in the
source and must lex as ONE variable: the declaration's, with the
pattern's occurrence staying INSIDE the pattern token. A lexer that
recurses into a pattern -- reasonable-looking, since the pattern really
does interpolate -- produces two, and this is the assertion that catches
it. The optree cannot: perl interpolates at runtime either way and
`regcomp` is emitted regardless.

```perl
my $p = "b";
my $s = "abc";
my $m = $s =~ /$p/;
print "$m\n";
```

```behavior
parses: yes
```

```output
1
```

```tokens
one variable whose text is "$p"
no operator whose text is "/"
```
