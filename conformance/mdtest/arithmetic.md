# Arithmetic, precedence and associativity

The seven arithmetic operators, and the two properties of the grammar
that leave NO OP BEHIND -- precedence and associativity.

**Tier 04 operators.** Introduces `add`, `subtract`, `multiply`,
`divide`, `modulo`, `pow`, `negate`, `concat`, `repeat`. Depends on
03_context.

Every case gives its operators a runtime operand. With constants the
optimiser folds the expression and there is no operator left in the op
stream to claim.

THE PRECEDENCE CASES ARE WHY THIS TIER PINS OUTPUT. `$a + $b * $c` and
`($a + $b) * $c` emit the same four ops in the same order; so do the two
associativity groupings. Nothing in the op stream separates them, and
nothing in a token stream does either. The VALUE is the only witness,
which is what makes a wrong pin worse here than anywhere else in the
corpus -- elsewhere a bad pin leaves the op claims standing, here it
leaves the tier measuring nothing.

## The seven arithmetic operators

Each with a runtime operand: add, subtract, multiply, divide, modulo,
pow and negate, in one statement so a parser that mis-groups any of them
prints a different number.

```perl
my $a = $ARGV[0] // 12;
my $b = $ARGV[1] // 5;
my $sum = $a + $b;
my $diff = $a - $b;
my $prod = $a * $b;
my $quot = $a / $b;
my $rem = $a % $b;
my $powr = $a ** $b;
my $neg = -$a;
print "$sum $diff $prod $quot $rem $powr $neg\n";
```

```behavior
parses: yes
```

```output
17 7 60 2.4 2 248832 -12
```

```tokens
one operator whose text is "%"
one operator whose text is "**"
```

## Concatenation and repetition

`.` compiles to `concat` only when its result is a LIST element -- here,
a direct `print` argument. `x` compiles to `repeat` either way.

```perl
my $a = $ARGV[0] // "ab";
my $b = $ARGV[1] // "cd";
print "joined: ", $a . $b, "\n";
print "repeat: ", $a x 3, "\n";
```

```behavior
parses: yes
```

```output
joined: abcd
repeat: ababab
```

```tokens
one word-shaped operator whose text is "x"
```

## `*` binds tighter than `+`

The only evidence is behavioural. Both groupings emit the same four ops
in the same order, so the two printed numbers are the whole claim.

```perl
my $a = $ARGV[0] // 2;
my $b = $ARGV[1] // 3;
my $c = $ARGV[2] // 4;
my $default = $a + $b * $c;
my $grouped = ($a + $b) * $c;
print "$default $grouped\n";
```

```behavior
parses: yes
```

```output
14 20
```

```tokens
one operator whose text is "("
```

## `**` is right associative, `-` is left

As with precedence the op stream shows neither: two `pow` ops and two
`subtract` ops in source order, either way. `2 ** 3 ** 2` is 512 right
associatively and 64 left. `2 - 3 - 2` is -3 left associatively and 1
right, so the third number separates those too.

```perl
my $a = $ARGV[0] // 2;
my $b = $ARGV[1] // 3;
my $c = $ARGV[2] // 2;
my $right = $a ** $b ** $c;
my $left = ($a ** $b) ** $c;
my $minus = $a - $b - $c;
print "$right $left $minus\n";
```

```behavior
parses: yes
```

```output
512 64 -3
```

```tokens
no operator whose text is "*"
```
