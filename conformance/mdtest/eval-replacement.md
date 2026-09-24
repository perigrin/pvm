# The replacement that is a program

`s///e`, `s///ee` and `eval STRING` -- the three places where a region of
source is handed back to the compiler rather than read as text.

**Tier 14 recursive.** Introduces `entereval`, `substcont`. Depends on
09_regex.

`substcont` is what `/e` actually adds. Plain `s/a/z/` is one `subst` op
taking a constant; `s/a/$n+1/e` is `subst(replstart->)` with the
replacement's own ops -- `padsv`, `const`, `add` -- executed between it and
a `substcont` that loops back for the next match. `substcont` is the
re-entry made visible: the op that returns control to the substitution
after the embedded program has run.

`entereval` is what the SECOND `e` adds, and what a bare `eval STRING`
is. It is the only op in this corpus that compiles Perl the compiler
never saw.

Every replacement here is deliberately unfoldable. Measured,
`s/a/uc("z")/e` compiles to `const[PV "Z"] s/FOLD` and a plain `subst` --
the optimiser erases the `/e` completely, and the case would measure the
same thing as tier 09's plain substitution.

## The `/e` flag: an expression in the replacement half

`/e` makes the replacement a Perl EXPRESSION rather than a string: the
lexer must hand the region between the second and third delimiter back to
the parser.

The token claims record where the re-entry has NOT yet happened. At the
lexical layer `s/a/$n+1/e` is ONE quote-like operator, the same as tier
09's `s/a/z/`: the `+` between its delimiters is inside its text, not a
token beside it. A lexer that emitted that `+` as an operator would have
lexed the replacement instead of delimiting it, which is the error this
tier exists to catch -- the region is handed to the PARSER, which lexes it
in a second pass.

`$n` is declared on its own line and so does appear as a variable token
there; only the `+` is unique to the replacement, which is why it is the
one the `no` claim names.

```perl
my $s = "abc";
my $n = 1;
$s =~ s/a/$n+1/e;
print "$s\n";
```

```behavior
parses: yes
```

```output
2bc
```

```tokens
one quote-like operator whose text is "s/a/$n+1/e"
no operator whose text is "+"
```

## A second `e`: the replacement re-read as Perl

A second `e` evaluates the FIRST evaluation's result as Perl: the string
`3*4` is compiled at runtime by code the compiler never saw.

This is the deepest re-entry the corpus contains. `s/b/$c/ee` compiles to
the same `subst(replstart->)` / `substcont` frame as `/e`, with
`entereval` between them: the first `e` evaluates `$c` to the string
`3*4`, the second compiles and runs that string.

The distinction matters lexically because NEITHER `e` is visible to the
lexer as code. The replacement region holds `$c`; the Perl that eventually
runs is `3*4`, which appears in the source only as the contents of a `q{}`
three lines earlier. A lexer cannot reach it at all, and neither can the
optree -- `entereval` is where the source ends.

```perl
my $s = "abc";
my $c = q{3*4};
$s =~ s/b/$c/ee;
print "$s\n";
```

```behavior
parses: yes
```

```output
a12c
```

```tokens
one quote-like operator whose text is "s/b/$c/ee"
one quote-like operator whose text is "q{3*4}"
no operator whose text is "*"
```

## `eval STRING`: the re-entry with no delimiter

Every other case in this tier hands back a region the lexer at least had
to FIND: `s///e` has three delimiters, `(?{ })` has a brace pair inside a
pattern, `qr//` freezes one. This has none. The operand is an ordinary
string expression, and what it holds is not known until the expression has
been evaluated.

WHY HERE AND NOT WITH THE BLOCK FORM, since `eval` is one keyword: the two
forms share no op. Measured under 5.42.0, `eval { 1 }` compiles to
`entertry` / `leavetry` -- a control transfer that marks a frame to unwind
to and compiles nothing new, claimed in `06_control`. `eval "1"` compiles
to `entereval`, which compiles Perl the compiler never saw.

Measured, with X unset so the interpolated operand is a runtime value:

    $ perl -MO=Concise,-exec -e 'my $n = $ENV{X} // 2;
      my $r = eval "$n + 1"; print "$r\n";'
    8  <0> padsv[$n:1,3] s
    9  <+> multiconcat(" + 1",-1,4)[t5] sK/STRINGIFY
    a  <1> entereval[t256] sK/1

THERE IS NO `add` IN THAT STREAM, and that absence is the whole case. The
program adds two numbers and emits no addition op, because the `+` is a
character in a string at compile time and becomes an operator only inside
the second compilation `entereval` triggers. `multiconcat` BUILDS the
program text; `entereval` compiles and runs it. One `e` puts the region's
`add` in the outer stream; two `e`s and a bare `eval` do not.

The string is interpolated rather than constant on purpose. `eval "1 + 1"`
still emits `entereval`, but a parser could constant-fold the operand and
the case would no longer measure that the operand is a VALUE.

The output pins `3`, not `2 + 1`. A parser that treated the string as an
ordinary string expression -- never re-entering -- would assign the text
and print it verbatim: a one-token difference in the tree and a five-byte
difference in the output.

```perl
my $n = $ENV{X} // 2;
my $r = eval "$n + 1";
print "$r\n";
```

```behavior
parses: yes
```

```output
3
```

```tokens
one string literal whose text is "\"$n + 1\""
no operator whose text is "+"
```
