# Token category glossary

A corpus file may assert a LEXICAL FACT: a claim about the token stream,
written as a declared fact rather than a dump of any parser's token kinds.

    --- expect tokens
    one numeric literal whose text is ".5"

For that to mean anything to somebody else's lexer, the category names have
to be defined here rather than in our enum. This page is the definition.

**It is derived from perl, not from us.** Where a boundary is contestable,
the entry cites `perldata`/`perlop` and, where those are silent, a measured
`perl` invocation. Where perl itself is ambiguous the entry says so instead
of picking.

**It grows with the corpus.** A category is added when a file needs to
assert about it, not in advance. Entries below cover tier 01 only.

---

## numeric literal

A single token denoting a number.

**Includes**, per `perldata` "Scalar value constructors":

    12345           decimal integer
    12345.67        decimal with fractional part
    .23             leading decimal point, no integer part
    6.02e23         exponent
    6.02e-23        SIGNED exponent -- the sign is PART of the token
    4_294_967_296   underscore separators
    0xff            hexadecimal
    0b1010          binary
    0377            octal by leading zero
    0o377           octal by explicit prefix (5.34+)

**Boundary cases, each decided against perl:**

**`-1` is TWO tokens**, not one: a negation operator and a numeric literal.
Perl parses the minus as an operator even where the result is constant-folded
to a single constant. Measured:

    $ perl -MO=Concise -e 'my $x = -1;' 2>&1 | grep const
    const[IV -1]

The optree shows one folded constant, which is why **the optree cannot
adjudicate this and the token stream must**. A leading `-` is never part of
a numeric literal. (The exponent sign is, which is the asymmetry that makes
this worth stating.)

**`5e-1` is ONE token.** The sign belongs to the exponent, not to a
subtraction. `5e-1` is 0.5, not `5e` minus `1` -- and `5e` alone is not a
number at all.

**`.5e3` is ONE token**: leading decimal point and exponent compose.

**`1.` is ONE token.** A trailing decimal point with no digits after it is
still one numeric literal, and equals 1.

**`v5.42` is NOT a numeric literal.** It is a v-string -- a string of
characters built from ordinals. Two dots make it a v-string unambiguously
(`5.42.0`), and a leading `v` makes it one with a single dot. `perldata`
documents these under "Version Strings" separately from numbers.

**`0.5` and `.5` and `5e-1` are the SAME VALUE.** Measured:

    $ perl -e 'printf "%.17g %.17g %.17g\n", .5, 0.5, 5e-1'
    0.5 0.5 0.5

This is the reason lexical facts exist as a category. No behavioural probe
can distinguish these spellings, to any precision, so a corpus that only
ran programs could not assert which one was written.

---

## string literal

A single token denoting a string, including its delimiters and any quoting
operator.

`'a'`, `"a"`, `q{a}`, `qq{a}`, and the heredoc INTRODUCER `<<'EOF'` are each
one token. An interpolating string is still ONE token at this layer: what
its interpolations mean is a parsing question, not a lexical one.

`qw(a b c)` is ONE token, not three. It denotes a list, but it is a single
quote-like operator; the split is semantic.

---

## word

A bare identifier: `foo`, `Foo::Bar`, `my`, `print`.

Keywords are not distinguished from other identifiers at this layer.
Whether `print` is a builtin, a user sub, or a filehandle is a parsing
question that needs context this layer does not have.

---

## variable

A sigil and the name it applies to, as one token: `$x`, `@a`, `%h`, `$#x`.

`${name}` is one variable: the braces are punctuation around a name. But
`${ $ref }` is NOT one token, because its contents are an expression
requiring a parser.

`$$` (the PID) is one variable; `$$ref` is a dereference, which is two.

---

## operator

Punctuation denoting an operation: `+`, `.`, `=~`, `->`, `?`, `:`.

The `.` in `.5` is not an operator, and `scanNumber` running before operator
scanning is what makes that true. That precedence is the subject of
`01_literals/03_leading_decimal.t`.
