# V-strings

Two dots make a string. A v-string is the boundary where a thing that
looks entirely like a number is not one, and GLOSSARY.md records the
decision under "numeric literal": a v-string is not in that category.

**Tier 01 literals.** Introduces `const`, `enter`, `leave`,
`multiconcat`, `nextstate`, `padrange`, `padsv`, `padsv_store`, `print`,
`pushmark`. Depends on nothing -- it is the only tier that can say that.
`my` and `print` appear throughout as FIXTURES rather than as subjects: a
literal has to be bound to something and observed somehow.

MEASURED perl 5.42.0:

    $ perl -e 'my $n = 65.66; print "$n\n"'
    65.66

ONE DOT IS A NUMBER AND TWO DOTS ARE A STRING. That is the whole
boundary, and it is decided by COUNTING DOTS after the token has already
started -- a lexer cannot know which it is scanning until it reaches the
second point or the end. `perldata` documents the form under "Version
Strings", away from numbers.

Both cases below REFUSED, and both refusals were LEXICAL: the parser
received a well-formed expression, returned ZERO Unknown nodes, and left
no refusal code to name. The token assertions were the only place the
disagreement was visible, which is what the token layer is for. Both pass
under 01a0db78: `scanNumber` counts dots, and a run with two of them is
emitted as a string rather than a number.

## Bare: three dot-separated parts

`65.66.67` is one token denoting the three-character string `ABC`, with
no `v` anywhere.

MEASURED perl 5.42.0:

    $ perl -e 'my $v = 65.66.67; print "$v\n"'
    ABC

REFUSED as of 7711154e. Our lexer produced `Number("65.66.67")` -- one
token, which was right, in the wrong CATEGORY, which was not. The parser
saw a number where perl sees a string.

Fixed under 01a0db78: `scanNumber` scans the run before choosing a kind,
and two dots make it a string. One dot is untouched -- `65.66` is still
`Number("65.66")`, which is the negative half of the same rule.

```perl
my $v = 65.66.67;
print "$v\n";
```

```behavior
parses: yes
```

```output
ABC
```

```tokens
no numeric literal whose text is "65.66.67"
```

## The `v` prefix

A leading `v` makes a v-string of what follows, one dot or many:
`v65.66.67` is ONE token denoting `ABC`.

MEASURED perl 5.42.0:

    $ perl -e 'my $v = v65.66.67; print "$v\n"'
    ABC

    $ perl -e 'my $v = v5.42.0; print join(".", map ord, split //, $v), "\n"'
    5.42.0

The second measurement is the form this repository writes constantly --
`v5.42.0` is how every `use` line spells a version -- and it is the same
construct.

REFUSED as of 7711154e. Our lexer produced `Word("v65") Operator(".")
Number("66.67")`: `v65` is a legal identifier, so the word scanner
claimed it and the rest of the v-string was read as a concatenation of a
bareword with a number. The PARSER could not see this. It received Word
Operator Number, read a valid expression, and returned ZERO Unknown
nodes -- the same shape the signed exponent records, and the same reason
the token layer exists.

Fixed under 01a0db78: `scanVString` runs BEFORE the word scanner and
claims a `v` followed by digits and TWO dots. It declines on one dot, so
`use v5.36` still lexes as three tokens and `internal/parse/use.go`
reassembles it as before. `use v5.42.0` has two dots, so it is now one
token, and `noteSignatures` reads the version out of it rather than out
of the split -- otherwise the bundle would silently stop turning on.

STILL OPEN: the one-dot `v5.36` is a v-string in perl too -- measured,
`length(v5.36)` is 2 -- and we lex it as a Word, an Operator and a
Number. That is a second gap, not this one, and correcting it means
teaching `internal/parse/use.go` to take a version that is one token.

```perl
my $v = v65.66.67;
print "$v\n";
```

```behavior
parses: yes
```

```output
ABC
```

```tokens
no operator whose text is "."
no word whose text is "v65"
```
