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

Both cases below refuse, and both refusals are LEXICAL: the parser
receives a well-formed expression, returns ZERO Unknown nodes, and
leaves no refusal code to name. The token assertions are the only place
the disagreement is visible. Neither has an issue, because each was
found by writing the case.

## Bare: three dot-separated parts

`65.66.67` is one token denoting the three-character string `ABC`, with
no `v` anywhere.

MEASURED perl 5.42.0:

    $ perl -e 'my $v = 65.66.67; print "$v\n"'
    ABC

REFUSES as of 7711154e. Our lexer produces `Number("65.66.67")` -- one
token, which is right, in the wrong CATEGORY, which is not. The parser
sees a number where perl sees a string and the value never reaches
output because we do not run the program.

```perl
my $v = 65.66.67;
print "$v\n";
```

```behavior
parses: yes
refuses: unfiled
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

REFUSES as of 7711154e. Our lexer produces `Word("v65") Operator(".")
Number("66.67")`: `v65` is a legal identifier, so the word scanner
claims it and the rest of the v-string is read as a concatenation of a
bareword with a number. The PARSER cannot see this. It receives Word
Operator Number, reads a valid expression, and returns ZERO Unknown
nodes -- the same shape the signed exponent records, and the same reason
the token layer exists.

```perl
my $v = v65.66.67;
print "$v\n";
```

```behavior
parses: yes
refuses: unfiled
```

```output
ABC
```

```tokens
no operator whose text is "."
no word whose text is "v65"
```
