# Bitwise and shift

Three of perlop's precedence levels -- 9 (`<< >>`), 14 (`| ^`) and 15
(`&`) -- and the pragma that silently changes what two of them mean.

**Tier 04 operators.** Introduces `bit_and`, `bit_or`, `bit_xor`,
`left_shift`, `right_shift`, `complement`, `nbit_and`. Depends on
03_context.

Every case gives its operators a RUNTIME operand. With constants the
optimiser folds the whole expression and there is no operator left in
the op stream to claim -- measured, `print 12 & 10` compiles to a single
`const`. The `$ENV{X} //` idiom is how this tier keeps an operator
alive, and it is why a `dor` appears in every case below.

## The three binary operators

`&`, `|` and `^` over a runtime integer. One case rather than three
because they share a precedence band and a failure mode: a lexer that
reads any of them as the start of a longer operator (`&&`, `||`, `^.`)
gets a different program with no diagnostic.

```perl
my $a = $ENV{X} // 12;
print $a & 10, " ", $a | 3, " ", $a ^ 3, "\n";
```

```behavior
parses: yes
```

```output
8 15 15
```

```tokens
one operator whose text is "&"
one operator whose text is "|"
one operator whose text is "^"
```

## Bitwise precedence: `&` binds tighter than `|`

perlop puts `&` at level 15 and `|` at level 14, so `$a | $b & $c` is
`$a | ($b & $c)`. The grouped form is written beside it because
PRECEDENCE LEAVES NO OP BEHIND -- both spellings emit the same two ops
in the same order, and only the value separates them.

```perl
my $a = $ENV{X} // 6;
my $b = $ENV{Y} // 3;
my $c = $ENV{Z} // 2;
print $a | $b & $c, " ", ($a | $b) & $c, "\n";
```

```behavior
parses: yes
```

```output
6 2
```

```tokens
one operator whose text is "("
```

## The shift operators

`<<` and `>>` at level 9. The `<<` is the interesting half: it is also
the heredoc opener, and the lexer decides between them by what follows.

```perl
my $a = $ENV{X} // 1;
my $b = $ENV{Y} // 16;
print $a << 3, " ", $b >> 2, "\n";
```

```behavior
parses: yes
```

```output
8 4
```

```tokens
one operator whose text is "<<"
one operator whose text is ">>"
no heredoc opener whose text is "<< 3"
```

## Complement

`~` is unary and at level 4, far above the binary band. Masked with
`& 255` so the answer does not depend on integer width.

```perl
my $a = $ENV{X} // 12;
print ~$a & 255, "\n";
```

```behavior
parses: yes
```

```output
243
```

```tokens
one operator whose text is "~"
```

## `&` on strings, ungated: the string bitwise operation

Without the `bitwise` feature, `&` is POLYMORPHIC. String operands get a
STRING bitwise operation: `'1' & '1'` is `'1'`, `'2' & '0'` is `'0'`, so
`"12" & "10"` is the string `"10"`.

This case and the next differ by ONE LINE and by the answer they print.
Neither means anything alone, which is why the old one-construct-per-file
format could not hold them together -- a file-scoped pragma cannot be
scoped to half a file.

```perl
my $s1 = $ENV{X} // "12";
my $s2 = $ENV{Y} // "10";
print $s1 & $s2, "\n";
```

```behavior
parses: yes
```

```output
10
```

```tokens
one operator whose text is "&"
```

## `&` on strings, gated: the numeric operation

The same three lines under `use v5.28`. Now `&` is numeric: `12 & 10` is
`8`.

THE HAZARD IS THAT NEITHER READING WARNS. `say` without its feature is a
method call and dies. `isa` without its feature is a filehandle print
and dies. `state` without its feature dies on an undefined invocant. All
three fail loudly enough to notice. This one SUCCEEDS with a different
answer, and every test that does not pin the pragma agrees with whatever
the parser chose.

The optree sees it too, which was not obvious: ungated emits `bit_and`
and gated emits `nbit_and` -- two ops, not one op with a flag. So the
dependency lint distinguishes the pair as well as the output does, but
only because both cases exist.

```perl
use v5.28;
my $s1 = $ENV{X} // "12";
my $s2 = $ENV{Y} // "10";
print $s1 & $s2, "\n";
```

```behavior
parses: yes
```

```output
8
```

```tokens
one word whose text is "use"
one operator whose text is "&"
```
