# Opaque regions inside one term

Four constructs whose opaque region fits inside a single term: the
command quote, the angle-bracket glob, the `pack`/`unpack` template, and
`qq`'s arbitrary delimiters. Everything above this heading suspends
ordinary lexing across LINES; these are delimited within an expression,
and all five cases PASS.

**Tier 13 opaque.** Introduces `backtick`, `enterwrite`, `glob`, `pack`,
`unpack`. Depends on 10_io.

Four of the five cases emit an op and none of the ops is the construct.
`qx{...}` emits `backtick` over a `const[PV]`, the same op backticks
emit, which is why GLOSSARY.md puts the two spellings in one category.
`<*.glob>` emits `glob`, which says a match was attempted and nothing
about the pattern that was matched.
`pack` and `unpack` emit `pack` and `unpack`, and the CONSTRUCT is the
template, which emits nothing. `qq` emits nothing of its own at all:
measured, `qq{plain}` is tier 01's `const[PV "plain"]` and `qq{v=$x}` is
tier 01's `multiconcat` -- so this tier REACHES `multiconcat` rather than
introducing it, and the delimiter is observable only in the token stream.

Two of the five touch the world, so both were written for a result that
does not depend on the machine: `echo hi` writes the same three bytes on
every system with a POSIX shell, and `*.nonexistent-xyz` matches nothing
in any directory the runner might sit in. A case calling `date` or
globbing `*.t` would be unreproducible and must not be written.

## The command quote

`qx{...}` runs a shell command and returns its output, emitting
`backtick` over the command string. Measured, `echo hi` writes the three
bytes `hi\n` including the trailing newline.

`qx` is the one construct in this tier whose own op is visible, and even
that op is shared with the backtick spelling.

```perl
my $out = qx{echo hi};
print $out;
```

```behavior
parses: yes
```

```output
hi
```

```tokens
one quote-like operator whose text is "qx{echo hi}"
```

## The angle-bracket glob

`<*.pattern>` in term position is a GLOB, not a pair of comparisons, and
it compiles to `glob` rather than to `readline`.

The token fact is what makes this a glob case rather than a count case.
`<*.nonexistent-xyz>` and `<DATA>` are the SAME token category -- our
lexer cannot tell a handle name from a pattern without knowing what the
name means, which is a parsing question -- so this asserts the
angle-bracket term and the data-section cases assert the other use of it.
perl separates them at the optree, where this is `glob` and that is
`readline`.

Note `scalar` emits no op: perl imposes scalar context on `@none` at
compile time and the array is simply evaluated there.

```perl
my @none = <*.nonexistent-xyz>;
print "got ", scalar(@none), "\n";
```

```behavior
parses: yes
```

```output
got 0
```

```tokens
one readline operator whose text is "<*.nonexistent-xyz>"
```

## The `pack` template

`C3` is a count and a type code to `pack` and three ordinary characters
to the lexer, which reads it as a string literal and never looks inside.

THE FOLDING TRAP, and it is why the argument comes from the environment.
Measured, `print unpack("A3", pack("A3","abc"))` emits NO `pack` op at
all: the pack ran at compile time and left a `const[PV "abc"] s/FOLD` for
the `unpack` to read. A case written over constant arguments would claim
`pack` and emit none, so the lint would have nothing to check and a
compiler that had never heard of `pack` would pass it. `$ENV{X}` is unset
when the runner executes, so `//` yields the default as a RUNTIME value.

THE WRONG PARSE THIS RULES OUT: a lexer that reads `C3` -- or `A3`, or
`x2` -- as anything but three characters of string. A template is the one
argument shape in this tier that LOOKS like it wants lexing, and the
temptation to give it its own scanner is exactly the temptation a
format's picture lines present. The token fact says it is one string
literal, whole, with `C` and `3` never separated.

`C3` rather than `A3`, because `A` pads with SPACES: `pack("A5","abc")`
is `abc  `, and pinning that would put trailing whitespace in the pinned
output for the repository's hook to strip. `C` takes ordinals and
produces printable letters, and 65, 66, 67 are `A`, `B`, `C` in every
encoding perl builds against.

`length` would be the natural way to observe a packed string and NO TIER
CLAIMS IT, so the output is the packed bytes themselves.

```perl
my $n = $ENV{X} // 65;
print pack("C3", $n, $n + 1, $n + 2), "\n";
```

```behavior
parses: yes
```

```output
ABC
```

```tokens
one string literal whose text is "\"C3\""
no numeric literal whose text is "3"
```

## The `unpack` template

`unpack` reads the same opaque template `pack` writes, and returns a LIST
where `pack` returns one string. Measured, the two differ in the optree
beyond their names: `pack[t5] sK/2` against `unpack lK/2` -- LIST
context, `l` where pack has `s`. A compiler that implemented `pack` and
stopped would pass the case above and fail here.

`unpack` folds over constant arguments exactly as `pack` does, so the
source uses the same `$ENV{X} // default` idiom for the same reason.

THE WRONG PARSE THIS RULES OUT: `unpack` read as a scalar-returning call.
The pinned output is three numbers, which only exists if the result was a
list of three that `@c` absorbed -- so this is one of the few cases in a
tier of token facts where the BEHAVIOUR carries a real claim.

`C3` is the template above, deliberately: the round trip is the
assertion. That case packs 65, 66, 67 into `ABC` and this unpacks `ABC`
back into 65, 66, 67. Either alone could be satisfied by a compiler that
had the template wrong in a compensating way; the pair cannot.

`"@c"` interpolates the array with `$"` between elements, which is a
space by default -- the `gvsv[*"]` and `join` in the optree, tiers 02 and
03. `print @c` would print `656667` with no separator and would not show
the list had three elements.

```perl
my $s = $ENV{X} // "ABC";
my @c = unpack("C3", $s);
print "@c\n";
```

```behavior
parses: yes
```

```output
65 66 67
```

```tokens
one string literal whose text is "\"C3\""
no word whose text is "C3"
```

## `qq`'s arbitrary delimiters

WHAT IS PARSING-DISTINCTIVE ABOUT `qq` IS ARBITRARY DELIMITERS. `qq{}`,
`qq()`, `qq[]` and `qq!!` are one operator with four terminators, and a
lexer that hardcodes `"` or scans for a fixed closer gets three of them
wrong. GLOSSARY.md already records that bracketing delimiters NEST, and
this is where the corpus measures it.

THE WRONG PARSE THIS RULES OUT, stated as the token a broken lexer would
produce: `Quote("qq{a{b}")`. A lexer that scans forward to the FIRST `}`
ends the string there, leaving `c=$x}` as loose tokens and a stray
CloseBracket. That is the single most likely qq bug, and it is why the
nesting case is here rather than a fourth plain delimiter. Measured under
5.42.0, `print length(qq{a{b}c}), "\n"` prints `5` -- `a`, `{`, `b`,
`}`, `c` -- so the inner `}` is not a terminator. The negative fact names
the wrong token directly, and `qq{a{b}` appears nowhere else here, so it
is a real claim rather than one defeated by its own source.

`$ENV{X} // 1` is the corpus idiom for a runtime value, and here it also
gives the interpolating case something to interpolate that is not a
constant. `$ENV{X}` lexes as Variable, Operator, Word, CloseBracket --
NOT as one token -- so it cannot satisfy or defeat either fact.

Measured against our lexer, the four spellings arrive as four Quote
tokens, the first of them spanning both inner braces.

```perl
my $x = $ENV{X} // 1;
print qq{a{b}c=$x}, qq(p), qq[q], qq!r!, "\n";
```

```behavior
parses: yes
```

```output
a{b}c=1pqr
```

```tokens
one quote-like operator whose text is "qq{a{b}c=$x}"
no quote-like operator whose text is "qq{a{b}"
one quote-like operator whose text is "qq(p)"
one quote-like operator whose text is "qq[q]"
one quote-like operator whose text is "qq!r!"
```
