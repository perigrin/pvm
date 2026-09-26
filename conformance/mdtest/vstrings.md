# V-strings

Two dots make a string, and so does a `v` and one dot. A v-string is the
boundary where a thing that looks entirely like a number is not one, and
GLOSSARY.md records the decision under "numeric literal": a v-string is
not in that category.

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
claims a `v` followed by digits and dots. `use v5.42.0` is now one token,
and `noteSignatures` reads the version out of it rather than out of the
split -- otherwise the bundle would silently stop turning on.

01a0db78 required TWO dots, which left the one-dot `v5.36` splitting.
01a0dc26 closed that: the `v` is what decides, not the dot count.
MEASURED perl 5.42.0 --

    $ perl -e 'my $v = v5.36; print length($v)'
    2
    $ perl -e 'print 5.36'
    5.36

-- so the same single dot lands in two categories and the prefix is the
only difference between them. `v5.36` is one `Quote` and bare `5.36` is
still `Number("5.36")`, which is the negative half of the same rule and
the reason `scanNumber` still counts to two.

`internal/parse/use.go` needed NO change, which the issue had expected to
be the cost. Measured: a `Quote` never matches the Word-shaped name path
`use` takes its module from, so the version falls through to `parseExpr`
and yields the same single `Term` the three-token reassembly built --
`use v5.36;`, `require v5.36;` and `use v5.36.0;` all parse to one
version Term.

A DOT IS STILL REQUIRED. Perl reads bare `v5` as a v-string too --
measured, `length(v5)` is 1 -- and we read it as the identifier `v5`.
That is a third gap, and it is narrower than it looks: `v5` with no dot
is a legal bareword, so claiming it takes every `v`-plus-digits name with
it.

AND THE `v` IS LOWERCASE. Capital V is never a v-string, which is easy to
assume symmetrical and is not:

    $ perl -e 'my $v = V5.36; print "[$v]"'
    [V536]
    $ perl -Mstrict -e 'my $v = V5.36; print $v'
    Bareword "V5" not allowed while "strict subs" in use
    $ perl -e 'use V5.36; print "ok"'
    Can't locate V5.pm in @INC

So `V5.36` is the bareword `V5` concatenated with `.36`, and `use V5.36`
loads a module. A revision of this corpus briefly asserted the opposite,
which would have made `use V5.36` enable signatures for a module load.

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
