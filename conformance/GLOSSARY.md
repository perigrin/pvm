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
assert about it, not in advance. Entries below cover tier 01 and the
lexical categories tier 13 needs.

**Every `##` heading is a category name, and nothing else uses that
level.** `TestGlossaryMatchesCategories` reads these headings and requires
each to have an entry in `internal/conformance/categories.go` and each
mapping to have a heading here, so a `## Notes` section would be read as a
category with nothing behind it. Use `###` for anything that is not a
category.

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

A single token denoting a string, spelled with DELIMITERS ALONE:

    'a'         no interpolation
    "a"         interpolating
    "$x and $y" still ONE token

An interpolating string is one token at this layer: what its interpolations
mean is a parsing question, not a lexical one.

**A quote spelled with an operator name is NOT in this category.** `q{a}`,
`qq{a}` and `qw(a b c)` are each one token, but they are
[quote-like operators](#quote-like-operator) -- the category exists
precisely so a file asserting a string literal is not satisfied by
`qw(a b)`, which denotes a list rather than a string.

**The heredoc introducer `<<'EOF'` is not in this category either.** It is
a [heredoc opener](#heredoc-opener), and what follows it is a
[heredoc body](#heredoc-body).

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
`01_literals/04_leading_decimal.t`.

**Punctuation is the whole of it.** Perl also has operators spelled as words,
and they are [word-shaped operators](#word-shaped-operator), not this. A
file asserting `operator` about `x` or `and` is asserting the wrong category,
and `TestCategoryBoundaries` holds the line in both directions.

---

## word-shaped operator

An operator perl spells with letters rather than punctuation: `x`, `cmp`,
`eq` `ne` `lt` `gt` `le` `ge`, `and` `or` `not` `xor`, `isa`, and the
compound `x=`.

`perlop` names them operators in so many words -- "Binary `x` is the
repetition operator", "Binary `and` returns the logical conjunction",
"Unary `not` returns the logical negation", "Binary `isa` evaluates to
true when the left argument is an object instance of..." -- and lists `x=`
among the assignment operators. So the category is derived from perl, as
this page's own rule requires, and not from our lexer, which gives them
all Kind Word.

**The list above is prose and the predicate is the authority.**
`parse.IsWordShapedOperator` reads the precedence tables, so a spelling
perl has and this paragraph forgot is still in the category. `isa` was
exactly that on the first pass: the predicate had it, the prose did not.

**These are ALSO [words](#word), and that is not a contradiction.** At the
token layer `word` is a claim about SPELLING: a bare identifier, keywords
not distinguished. `word-shaped operator` is the narrower claim, for a file
that means perl's operator rather than any identifier that happens to be
there. A lexer with a distinct kind for these maps that kind here and
ignores the text; ours decides by text.

**`not` is in the list and is unary.** The category is about spelling, not
arity -- `!` and `not` are the same operation at different precedences, and
`perlop` says so.

---

## quote-like operator

A quote spelled with an OPERATOR NAME and a delimiter -- `q`, `qq`, `qw`,
`qr`, `qx`, `m`, `s`, `tr`, `y` -- rather than with delimiters alone, plus
backticks, which run a command without naming an operator.

    q(a b)          single-quoted, no interpolation
    qq{hi $x}       double-quoted, interpolating
    qw(a b c)       a LIST of words, not a string
    qr/pat/         a compiled pattern
    qx/ls/          runs a command
    m{pat}          a match
    s/a/b/          substitution
    tr/a/b/         transliteration
    y/a/b/          transliteration, the other spelling

**Backticks are in this category despite having no operator name.**
`` `ls` `` and `qx/ls/` are the same operation -- measured, both run the
command and return its output -- so a corpus file must not be able to
assert one as a string literal and the other as an operator. Here the
delimiter alone carries the meaning.

**This is a separate category from `string literal` because the two make
different claims.** A corpus file asserting `one string literal whose text
is "hi"` must not be satisfied by `qw(hi)`, which is not a string at all:
measured, `my @w = qw(a b c)` gives a three-element LIST, while
`my $s = q(a b)` gives the three-character string `a b`.

Our lexer gives both the same `Quote` kind, so the category is decided by
the token's TEXT. A lexer that splits them by kind maps these two entries
onto two kinds and ignores the text. That choice is `categories.go`'s, not
the corpus's.

**Any non-whitespace delimiter is accepted**, and bracketing delimiters
nest: measured, `q{a{b}c}` is the five-character string `a{b}c`, so the
inner braces are content rather than a terminator. Non-bracketing
delimiters do not nest.

**`s`, `tr` and `y` take a second pair**, and only when the first pair is
bracketing is the second pair's opening delimiter free to differ.

---

## heredoc opener

The `<<EOT` that appears in the statement, as one token.

    my $h = <<EOT;

The opener is where the heredoc STARTS; it is not the content. A corpus
file asserting the opener is making a claim about the statement's token
stream, which is why it is a separate category from the body.

`<<~EOT` is one opener: the `~` requests indentation stripping and belongs
to the token. Measured, `<<~EOT` with an indented terminator prints the
body with the common indentation removed.

---

## heredoc body

The lines between the opener's line and the terminator, as one token.

    my $h = <<EOT;
    body line
    EOT

The body arrives AFTER the semicolon in the token stream, because the
heredoc's content begins on the next LINE while the statement continues on
the same one. That reordering is the whole reason a heredoc is hard to
lex, and asserting it is what a corpus file is for.

A body is one token however many lines it spans.

**The terminator line is PART of the body, and tier 13 settles it that
way.** Perl does not say, because perl has no token stream to consult -- the
heredoc is gone by the time anything observable exists, which is tier 13's
whole subject. So the question is decided by what a consumer needs, and a
consumer needs to know where the body ENDS: a body token stopping before
`EOT\n` leaves the terminator as unclaimed bytes that some later rule has
to skip. Including it makes the token cover every byte the heredoc
occupies. Measured against our lexer, `<<EOT` with the single body line
`hi` produces a body whose text is `hi\nEOT\n`.

---

## readline operator

The angle-bracket term `<FH>`, `<$fh>` or `<*.c>`, as one token.

    my $line = <STDIN>;
    my @all  = <DATA>;
    my @files = <*.txt>;

The category exists because the same two characters are comparison
operators in operator position, and only position decides. `$a < $b` is
three tokens; `<$b>` is one.

**A glob pattern and a filehandle read are the SAME token here**, which is
perl's own conflation rather than ours. Measured under 5.42.0, `<*.c>`
compiles to a `glob` op and `<DATA>` to a `readline` op, so the two differ
at the optree -- but the lexer cannot tell them apart without knowing
whether `DATA` names a handle, which is a parsing question. The category
is named for the syntax, not for either op.

---

## pod block

A documentation block: a line starting `=` followed by an identifier,
through the matching `=cut` line, as one token.

    =pod

    text

    =cut

**POD leaves no trace at all.** Measured under 5.42.0, a program with POD
between two statements compiles to exactly the ops of the program without
it. That is why the category exists: the token stream is the only place a
pod block is observable, so a corpus file asserting one cannot fall back on
behaviour or on the optree.

The terminator is part of the token, on the same reasoning as the heredoc
body: every byte of a pod block belongs to the pod block.

---

## data section

`__END__` or `__DATA__` and every byte after it, as one token.

    print "hi\n";
    __END__
    not perl

Both spellings are one category because both do the same thing to the
lexer: they end the program text. What follows is readable at runtime
through the `DATA` filehandle and is never lexed as code.

**Like pod, this leaves nothing in the optree** -- measured, a program with
a `__DATA__` section compiles to the same ops as one without. Reading the
section needs `readline`, which is tier 10's op and says nothing about the
marker that created the handle.

---

## format body

The picture lines of a `format NAME =` declaration, through the lone `.`
that ends them, as one token.

    format STDOUT =
    a fixed line
    .

The body is not Perl and must not be lexed as Perl: `@<<<<<` is a picture
field naming a left-justified column, not an array sigil followed by two
left-shift operators. The `format NAME =` introducer is ordinary tokens --
two words and an operator -- and only the body is opaque, which is why the
category covers the body alone.

Measured under 5.42.0, the declaration emits no ops whatsoever; only the
`write` that uses it emits `enterwrite`.
