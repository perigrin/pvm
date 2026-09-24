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

All three refuse, all three name `trailing_tokens`, and that is one gap
rather than three. The LEXER produces the opener and the body as two
tokens in the right order with the right text -- every token fact here
passes. The parser has no grammar rule for a body token arriving where a
new statement should start. Naming the code in all three is what makes
"one gap" falsifiable: the day one of them begins refusing elsewhere,
these cases report the change instead of skipping on in silence.

This tier also settles a question GLOSSARY.md deferred: a heredoc body
token INCLUDES its terminator line. Perl cannot adjudicate -- there is no
token stream to inspect -- so the decision is made on what a consumer
needs, which is to know where the body ends.

## The interpolating heredoc

`<<"EOT"` interpolates. The declaration itself PARSES: measured,
`my $h = <<"EOT";` arrives as a declaration whose two terms are `$h` and
the opener, with no Unknown in it. The Unknown spans what comes NEXT --
the body token and the `print` after it -- and its code is
`trailing_tokens`, an expression that parsed with bytes remaining.

That is the heredoc's reordering stated as a parser failure. The output
proves the CONTENT and proves nothing about the spelling; a case
asserting only `parses` would report the same failure and leave the
boundary unlocated.

```perl
my $name = "world";
my $h = <<"EOT";
hello $name
EOT
print $h;
```

```behavior
parses: yes
refuses: unfiled
refusal: trailing_tokens
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

Refuses with the SAME CODE, and that is the claim rather than a copied
line: a refusal that changed cause between the two would mean the parser
sees the quoting, and it does not.

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
refuses: unfiled
refusal: trailing_tokens
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
refuses: unfiled
refusal: trailing_tokens
```

```output
trimmed
```

```tokens
one heredoc opener whose text is "<<~EOT"
one heredoc body whose text is "    trimmed\n    EOT\n"
```
