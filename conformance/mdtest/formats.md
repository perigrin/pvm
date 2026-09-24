# Formats

A `format NAME =` declaration's body runs from the `=` to a line holding
a lone `.`, and it is the clearest case in the tier of a region the lexer
must delimit WITHOUT LEXING: `@<<<<<` inside a picture line is a
left-justified column, not an array sigil followed by two left shifts.

**Tier 13 opaque.** Introduces `backtick`, `enterwrite`, `glob`, `pack`,
`unpack`. Depends on 10_io.

`enterwrite` is claimed by this tier but it belongs to `write`. The
DECLARATION compiles to nothing: measured, a file holding only a format
and a `write` emits `enter`, `nextstate`, `enterwrite`, `leave`, and the
declaration accounts for none of them. The picture lines that are the
actual lexing problem are compiled into a format no op mentions.

Both cases refuse with `unimplemented_statement`, which is not the code
the heredocs carry. Those parse their statement and then meet a body
token they have no form for; this one never starts, because `format` is a
statement keyword the parser does not implement. Measured, the Unknown
spans `format STDOUT =` through the `write;` that follows it -- the
declaration and the statement after it swallowed together, which is what
an unimplemented keyword does to whatever it cannot find an end for. The
`format body` fact passes in both, so the picture lines are already
opaque to the lexer and what is missing is above it.

## A format declaration and the `write` that uses it

The picture here is a fixed line rather than a `@<<<` field, because a
field pads its column with spaces to a width the picture sets and the
runner compares bytes. A fixed line keeps this case about the LEXING and
leaves the field to the case below.

`write` prints the format's output and `print` prints after it, so the
two share one ordered stream on STDOUT.

```perl
format STDOUT =
a fixed report line
.
write;
print "after write\n";
```

```behavior
parses: yes
refuses: unfiled
refusal: unimplemented_statement
```

```output
a fixed report line
after write
```

```tokens
one format body whose text is "a fixed report line\n.\n"
one word whose text is "format"
```

## A picture line that is not Perl

WHY A PLAIN PICTURE LINE CANNOT MAKE THIS CLAIM, which is the whole
reason this case exists beside the one above. `a fixed report line`
inside a format body is delimited identically by a lexer that treats the
body as opaque and by one that lexes it as Perl -- three Words either
way, and the region's extent is unchanged. Such a case measures
DELIMITING and says nothing about NOT LEXING, which is the other half of
this tier's thesis. `@<<<<<<< @>>>` separates them: opaque it is one
token, lexed as Perl it is `@` variables and shift operators.

Measured against our lexer, the body arrives as a single format body
token running from the first picture line through the lone `.`, with no
token inside it.

The argument line `$name,   $qty` IS ordinary Perl to perl, evaluated
when `write` runs. It is inside the opaque region all the same: the lexer
does not distinguish picture lines from argument lines, and it does not
need to.

The variables are `our` rather than `my` because a format body is
compiled in its own scope and cannot see a lexical declared beside it;
measured, `my $name` leaves the field empty.

The output has NO TRAILING WHITESPACE on either line, which the field
widths were chosen for: `widget` fills 6 of the 8 columns `@<<<<<<<` sets
and `7` is right-justified into `@>>>`, so the padding falls between the
two fields rather than after the last one.

```perl
our $name = "widget";
our $qty  = 7;
format STDOUT =
@<<<<<<< @>>>
$name,   $qty
.
write;
print "after write\n";
```

```behavior
parses: yes
refuses: unfiled
refusal: unimplemented_statement
```

```output
widget      7
after write
```

```tokens
one format body whose text is "@<<<<<<< @>>>\n$name,   $qty\n.\n"
```
