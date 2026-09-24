# Strings and quoting

What the DELIMITER decides. Three cases, and the op stream can see none
of the distinctions any of them make.

**Tier 01 literals.** Introduces `const`, `enter`, `leave`,
`multiconcat`, `nextstate`, `padrange`, `padsv`, `padsv_store`, `print`,
`pushmark`. Depends on nothing -- it is the only tier that can say that.
`my` and `print` appear throughout as FIXTURES rather than as subjects: a
literal has to be bound to something and observed somehow.

MEASURED perl 5.42.0:

    $ perl -MO=Concise,-exec -e 'my $s = q{plain}; print $s'
    ... const[PV "plain"] ... padsv_store ... print ...
    $ perl -MO=Concise,-exec -e 'my $s = "plain"; print $s'
    ... const[PV "plain"] ... padsv_store ... print ...

Byte-identical. A single-quoted string, a double-quoted string with
nothing to interpolate, and `q{...}` all arrive as one `const`. That is
why GLOSSARY.md keeps `string literal` and `quote-like operator` as
separate categories, and why these claims are lexical: a case asserting
`one string literal whose text is "hi"` must not be satisfied by
`qw(hi)`, which is not a string at all.

What output CAN still see is escape processing, which is what separates
the first two cases:

    $ perl -e 'print length("a\nb")'       3
    $ perl -e 'print length(q{a\nb})'      4

So one prints two lines and the other prints one, and a lexer that
treated the two delimiters alike fails exactly one of them, whichever
way it was wrong.

## Double quotes process escapes

`"a\nb"` is THREE characters, and the same bytes in single quotes are
four.

INTERPOLATION is deliberately NOT exercised here. `"v=$x"` needs a
runtime operand to be worth anything, and the corpus idiom for one --
`$ENV{X} // <default>` -- uses `dor`, which is 04_operators' op and
three tiers away. Interpolation on an ARRAY is already claimed by
03_context, which is where the construct earns its keep. What this tier
owns is the DELIMITER, and the delimiter is what the escape
distinguishes.

```perl
my $s = "a\nb";
print $s, "\n";
```

```behavior
parses: yes
```

```output
a
b
```

```tokens
one string literal whose text is "\"a\\nb\""
```

## Single quotes do not

A single-quoted string does not interpolate, and its backslash escapes
are not the double-quoted set: `'a\nb'` is FOUR characters.

The tier was chartered for "numbers, strings, quoting and qw" and
shipped twelve numeric files and no string file at all. The three string
boundaries its glossary check names were satisfied only by the adjacency
body, which introduces nothing -- so the tier asserted the numeric half
of its vocabulary and borrowed the rest from a case that exists to
compose, not to introduce.

A lexer that applied double-quoted escape processing here would print
`a`, a newline and `b`. `length` is claimed by no tier, so the count is
not measured in the case: the OUTPUT is the four characters themselves,
and a lexer that collapsed `\n` would print three characters across two
lines instead.

The token fact spells the backslash DOUBLED, `'a\\nb'`, because the
fact's text is compared against the source bytes and the source holds a
literal backslash. The trailing `"\n"` is a second quote token, which is
why the fact counts `one` of a specific text rather than one quote: a
bare count would be two and the claim would be about the case's
punctuation rather than about its construct.

```perl
my $p = 'plain';
my $s = 'a\nb';
print $p, $s, "\n";
```

```behavior
parses: yes
```

```output
plaina\nb
```

```tokens
one string literal whose text is "'a\\nb'"
```

## `q` with three delimiters

`q` and `qq` are QUOTE-LIKE OPERATORS, not string literals: a different
glossary category making a different claim, and the delimiter is chosen
rather than fixed. That separation is only worth having if some case
asserts each side of it, and until this one no corpus file asserted this
side.

MEASURED perl 5.42.0:

    $ perl -MO=Concise,-exec -e 'my $s = q(a b); print $s'
    ... const[PV "a b"] ... padsv_store ... print ...

One `const`, exactly as `'a b'` gives. So `q(a b)` and `'a b'` are the
same op and DIFFERENT TOKENS, which is precisely the case where only a
token fact can carry the claim.

The DELIMITERS are what this case pins, and each is a separate chance to
be wrong: `q(...)`, `q{...}` and `q!...!` are all one construct wearing
three delimiters, and a lexer that hard-codes one of them fails the
others. `qq` is not spelled separately because what separates it from
`q` is escape processing, which the two quoted-string cases above
already pin for the delimiter-less spellings -- repeating it here would
measure the same thing twice.

`qw` is NOT here. Measured, `my @w = qw(a b c)` emits `aassign` and
`padav`, both 02_variables' ops, because a word list needs an array to
land in. It belongs to the tier that owns arrays, and a case here would
have to borrow two ops from a later tier to hold it.

```perl
my $a = q(one);
my $b = q{two};
my $c = q!three!;
print $a, $b, $c, "\n";
```

```behavior
parses: yes
```

```output
onetwothree
```

```tokens
one quote-like operator whose text is "q(one)"
one quote-like operator whose text is "q{two}"
one quote-like operator whose text is "q!three!"
```
