# Heredocs

A heredoc's body begins on the NEXT line while the statement it sits in
continues on the SAME one, so the body token arrives after the semicolon
that ended the statement it belongs to.

**Tier 13 opaque.** Introduces `backtick`, `enterwrite`, `glob`, `pack`,
`unpack`. Depends on 10_io.

None of these three cases claims an op, because a heredoc emits none of
its own. Measured, `<<"EOT"` with an interpolation gives tier 01's
`multiconcat` and `<<'EOT'` and `<<~EOT` give a plain `const[PV]` --
indistinguishable from the same strings written inline. The token facts
are what say a heredoc was written at all.

All three parse. They did not: all three refused with `trailing_tokens`,
and it was one gap rather than three -- the LEXER produced the opener and
the body as two tokens in the right order with the right text, and the
parser had no grammar rule for a body token arriving where a new statement
should start. `01a0c13f-aaf8` closed it by making the body a child of the
statement its opener sits in, consumed AFTER the terminator because that
is where the bytes are. One rule closed all three, which is what "one
gap" predicted.

This tier also settles a question GLOSSARY.md deferred: a heredoc body
token INCLUDES its terminator line. Perl cannot adjudicate -- there is no
token stream to inspect -- so the decision is made on what a consumer
needs, which is to know where the body ends.

## The interpolating heredoc

`<<"EOT"` interpolates. The statement arrives as a declaration whose two
terms are `$h` and the opener, and a THIRD child holding the body -- one
statement spanning from `my` past its own `;` to the end of the terminator
line, because that is where the body's bytes are.

It refused with `trailing_tokens` before `01a0c13f-aaf8`: the declaration
itself parsed, and the Unknown spanned what came NEXT -- the body token
and the `print` after it. The output proves the CONTENT and proves nothing
about the spelling; a case asserting only `parses` would have reported
that failure and left the boundary unlocated.

```perl
my $name = "world";
my $h = <<"EOT";
hello $name
EOT
print $h;
```

```behavior
parses: yes
```

```output
hello world
```

```tokens
one heredoc opener whose text is "<<\"EOT\""
one heredoc body whose text is "hello $name\nEOT\n"
```

## The non-interpolating heredoc

`<<'EOT'` does not interpolate: `$name` in the body is five characters of
text, not a variable. Measured, this compiles to a plain
`const[PV "hello $name\n"]` -- no multiconcat, because there is nothing
to interpolate.

This is the pair to the interpolating case, and the pair is the point.
The two spellings differ by one quote character and produce DIFFERENT
strings, which makes this the one heredoc property whose behaviour is
observable: the case above prints `hello world` and this prints
`hello $name` with `$name` in scope and holding `world`. Everything else
about a heredoc -- where it starts, where it ends, that it is a heredoc
at all -- is invisible below the token stream.

Parses like its pair, and that is the claim rather than a copied line: the
parser sees ONE opener token and ONE body token in either spelling, so a
rule that reads one reads both. Before `01a0c13f-aaf8` the two refused with
the same code, for the same reason -- the parser never saw the quoting
then either.

```perl
my $name = "world";
my $h = <<'EOT';
hello $name
EOT
print $h;
print "$name\n";
```

```behavior
parses: yes
```

```output
hello $name
world
```

```tokens
one heredoc opener whose text is "<<'EOT'"
one heredoc body whose text is "hello $name\nEOT\n"
```

## The indentation-stripping heredoc

`<<~EOT` strips the terminator's indentation from every body line, so the
heredoc can be indented with the code around it.

The `~` belongs to the OPENER token, not to a separate operator: the
opener's text is `<<~EOT`. The stripping happens in the lexer, so by the
time an op exists the body is already `trimmed\n` -- identical to what an
unindented `<<EOT` would have produced. The BODY TOKEN still carries the
leading spaces, which is what distinguishes this case from one that
simply did not indent.

```perl
my $h = <<~EOT;
    trimmed
    EOT
print $h;
```

```behavior
parses: yes
```

```output
trimmed
```

```tokens
one heredoc opener whose text is "<<~EOT"
one heredoc body whose text is "    trimmed\n    EOT\n"
```
