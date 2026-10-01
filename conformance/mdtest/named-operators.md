# Named operators and argument extent

Operators spelled as words, and the question they all raise: HOW FAR
DOES THE ARGUMENT RUN.

**Tier 04 operators.** Introduces `defined`, `undef`, `chr`, `ord`,
`index`, `sprintf`, `substr`, `uc`. Depends on 03_context.

Four of these fold away when their operands are constant -- `chr(74)`
arrives as `const[PV "J"]` -- so each case spells its construct with a
runtime operand, the rule the rest of this tier lives under.

WHAT MAKES THEM A TOPIC rather than eight unrelated builtins is that the
op name answers almost nothing about them. `substr` is one op for three
arities and for the lvalue form; `sprintf` is one op whose argument
count is decided by its format's CONTENTS; `undef` is one word for two
operators. In every case the extent or the arity is the claim, and the
op stream cannot carry it.

## A named unary binds looser than arithmetic

`defined $x + 1` is `defined($x + 1)`, not `(defined $x) + 1`, and the
two readings print different numbers. This is the argument-extent
question in its smallest form.

```perl
my $x = $ARGV[0] // 2;
my $loose = defined $x + 1;
my $tight = (defined $x) + 1;
print "$loose $tight\n";
```

```behavior
parses: yes
```

```output
1 2
```

```tokens
one operator whose text is "("
no operator whose text is ")"
```

## `undef` is two operators wearing one word

A UNARY one that clears what it is given, and a NILADIC one that is
simply the undefined value. `undef @a` empties the array; `@a = undef`
fills it with one element. One word, opposite effects on the same
variable.

```perl
my @a = ($ENV{X} // 1, 2, 3);
my @b = ($ENV{X} // 1, 2, 3);
undef @a;
@b = undef;
print scalar @a, scalar @b, "\n";
```

```behavior
parses: yes
```

```output
01
```

```tokens
one word whose text is "print"
no operator whose text is ")"
```

## `chr` and `ord`, folded and unfolded

Inverses over one character, and BOTH FOLD AWAY when their argument is a
literal. Each is written twice -- once on a runtime operand, once on a
constant -- so the optree holds one op where the source holds two. That
is the fold made visible rather than worked around.

```perl
my $n = $ENV{X} // 74;
my $live = chr($n);
my $folded = chr(74);
my $back = ord($live);
print "[$live][$folded][$back]\n";
```

```behavior
parses: yes
```

```output
[J][J][74]
```

```tokens
one word whose text is "ord"
```

## `index` reports failure as -1, not undef

Which makes its result a NUMBER that is always defined -- so `//` cannot
test it, and a truth test is wrong at position zero. The three pinned
values are a hit, a hit at a later offset, and the sentinel.

```perl
my $s = $ENV{X} // "hello world";
my $first = index($s, "o");
my $next = index($s, "o", $first + 1);
my $none = index($s, "z");
print "[$first][$next][$none]\n";
```

```behavior
parses: yes
```

```output
[4][7][-1]
```

```tokens
one operator whose text is "+"
```

## `%*d` makes the format consume an argument

`sprintf`'s first argument is a FORMAT, and `%*d` makes it take an extra
argument to supply its own width -- so how many arguments a call takes
is decided by the format's contents and not by its syntax. No parser can
know the arity without reading the string.

```perl
my $n = $ENV{X} // 5;
my $zero = sprintf("%03d", $n);
my $star = sprintf("%*d", $n, $n);
my $left = sprintf("%-*d", $n, $n);
print "[$zero][$star][$left][", $n % 3, "]\n";
```

```behavior
parses: yes
```

```output
[005][    5][5    ][2]
```

```tokens
one operator whose text is "%"
no word whose text is "printf"
```

## `substr` takes two, three or four arguments

A negative offset counts from the end, and a missing length means "to
the end" -- three meanings for one word, none visible in the op.

```perl
my $s = $ENV{X} // "hello world";
my $tail = substr($s, 6);
my $mid = substr($s, 3, 2);
my $neg = substr($s, -5, 3);
print "[$tail][$mid][$neg]\n";
```

```behavior
parses: yes
```

```output
[world][lo][wor]
```

```tokens
one operator whose text is "-"
```

## `substr` is an lvalue

`substr($s,0,1) = "J"` assigns INTO the string, and the four-argument
form does the same replacement while RETURNING the text it displaced.
Same op name, same result in the string, and only the return value tells
them apart.

```perl
my $a = $ENV{X} // "hello";
my $b = $ENV{X} // "hello";
substr($a, 0, 1) = "J";
my $old = substr($b, 0, 1, "J");
print "[$a][$b][$old]\n";
```

```behavior
parses: yes
```

```output
[Jello][Jello][h]
```

```tokens
one word whose text is "print"
```

## `uc` is a named unary, so a comma ends its argument

`uc $s, "-", "def"` upper-cases `$s` alone: the comma is below a named
unary, so the rest are `print`'s further arguments. With a constant
operand it folds -- `uc "abc"` arrives as `const[PV "ABC"]` -- so the
case uses a runtime one. Measured 5.42.0, the stream holds `uc[t2] sK/1`
and then `print`.

```perl
my $s = "abc";
print uc $s, "-", "def", "\n";
```

```behavior
parses: yes
```

```output
ABC-def
```
