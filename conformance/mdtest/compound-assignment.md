# Compound assignment

perlop's level 20: thirteen operators the corpus wrote none of until
issue `01a0cc04-32c9`.

**Tier 04 operators.** Introduces `andassign`, `dorassign`, `orassign`.
The arithmetic, string and bitwise forms emit the ops their binary
counterparts do, which is why those cases could not be written before
the binary ones existed. Depends on 03_context.

EACH IS ONE OPERATOR TOKEN, not two. `$x += 2` is `Operator("+=")` and
not `+` followed by `=`, which is the claim the token facts carry and
the reason this is a topic: thirteen spellings of one lexical rule, and
one of them our lexer still cannot read.

## The arithmetic forms

`+=` `-=` `*=` `/=` and the rest, each one token.

```perl
my $a = $ENV{X} // 5;
my $b = $ENV{X} // 5;
my $c = $ENV{X} // 5;
my $d = $ENV{Y} // 6;
$a += 2;
$b -= 2;
$c *= 2;
$d /= 2;
print "$a $b $c $d\n";
```

```behavior
parses: yes
```

```output
7 3 10 3
```

```tokens
one operator whose text is "+="
one operator whose text is "-="
one operator whose text is "*="
one operator whose text is "/="
```

## `.=` emits `multiconcat`, not `concat`

Which a reader would not predict from `+=` giving `add`. The op is
chosen by what the optimiser can fuse, not by the operator's spelling.

```perl
my $s = $ENV{X} // "a";
$s .= "b";
print "$s\n";
```

```behavior
parses: yes
```

```output
ab
```

```tokens
one operator whose text is ".="
```

## `x=` is word-shaped, and we cannot lex it

The thirteenth compound assignment and the only one whose operator is a
WORD. Measured, our lexer emits `Word(x) Operator(=)` rather than
forming the `x=` token at all, so the statement has a Word where an
operator belongs.

This case bisects that refusal against `.=` and the binary `x`, both of
which parse.

```perl
my $t = $ENV{X} // "ab";
$t x= 2;
print "$t\n";
```

```behavior
parses: yes
refuses: 01a0ce57-db92-78e5-bbe4-7c28e74db6f3
refusal: trailing_tokens
```

```output
abab
```

```tokens
one operator whose text is "x="
```

## `//=`, `||=` and `&&=` short-circuit

The only compound assignments whose behaviour output can see: the
right-hand side is not evaluated when the left already decides the
answer. Four pinned values separate the four readings.

```perl
my $p = $ENV{X} // 0;
my $q = $ENV{X} // 0;
my $r = $ENV{Y} // 1;
my $s = $ENV{X} // 0;
$p //= 99;
$q ||= 99;
$r &&= 99;
$s &&= 99;
print "$p $q $r $s\n";
```

```behavior
parses: yes
```

```output
0 99 99 0
```

```tokens
one operator whose text is "//="
one operator whose text is "||="
```

## The bitwise and shift forms

`&=` `|=` `^=` `<<=` `>>=`, reusing the ops the bitwise topic
introduces.

```perl
my $a = $ENV{X} // 12;
my $b = $ENV{X} // 12;
my $c = $ENV{X} // 12;
my $d = $ENV{Y} // 1;
my $e = $ENV{Z} // 16;
$a |= 3;
$b &= 10;
$c ^= 3;
$d <<= 3;
$e >>= 2;
print "$a $b $c $d $e\n";
```

```behavior
parses: yes
```

```output
15 8 15 8 4
```

```tokens
one operator whose text is "|="
one operator whose text is "&="
one operator whose text is "^="
one operator whose text is "<<="
one operator whose text is ">>="
```
