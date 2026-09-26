# Comparison, logic and the nonassoc levels

The two comparison families, the logical operators, and the levels where
repetition is a SYNTAX ERROR rather than a grouping.

**Tier 04 operators.** Introduces `eq` `ne` `lt` `gt` `le` `ge` `ncmp`,
`seq` `sne` `slt` `sgt` `sle` `sge` `scmp`, `and` `or` `dor` `not`
`xor`, `cmpchain_and`, `cmpchain_dup`. Depends on 03_context.

THE TWO COMPARISON FAMILIES ARE THE SAME RELATION OVER THE SAME
OPERANDS, differing only in the context each imposes -- which is the
whole reason this tier sits after 03.

The last three cases are about a question no op stream can answer: which
levels ACCEPT repetition. perly.y declares eleven `%nonassoc` levels and
measurement says eight of them accept it anyway; `TestTierOperatorsNonassocMeasured`
runs all eleven against perl. What the corpus adds here is the two that
refuse, and the chaining answer that has no diagnostic either way.

## The numeric relations and `<=>`

Six relations and the three-way comparison, each imposing numeric
context on both operands.

```perl
my $a = $ARGV[0] // 3;
my $b = $ARGV[1] // 5;
print "eq [", ($a == $b), "]\n";
print "ne [", ($a != $b), "]\n";
print "lt [", ($a < $b), "]\n";
print "gt [", ($a > $b), "]\n";
print "le [", ($a <= $b), "]\n";
print "ge [", ($a >= $b), "]\n";
print "cmp [", ($a <=> $b), "]\n";
```

```behavior
parses: yes
```

```output
eq []
ne [1]
lt [1]
gt []
le [1]
ge []
cmp [-1]
```

```tokens
one operator whose text is "<=>"
```

## The string relations and `cmp`

The same operands and the same relations as the numeric case above,
different ops. The difference is only the context each imposes, which is
why both cases exist and neither would be enough alone.

```perl
my $a = $ARGV[0] // "aa";
my $b = $ARGV[1] // "bb";
print "seq [", ($a eq $b), "]\n";
print "sne [", ($a ne $b), "]\n";
print "slt [", ($a lt $b), "]\n";
print "sgt [", ($a gt $b), "]\n";
print "sle [", ($a le $b), "]\n";
print "sge [", ($a ge $b), "]\n";
print "scmp [", ($a cmp $b), "]\n";
```

```behavior
parses: yes
```

```output
seq []
sne [1]
slt [1]
sgt []
sle [1]
sge []
scmp [-1]
```

```tokens
one word-shaped operator whose text is "cmp"
```

## The five logical operators

`&&` is the op `and`, `||` is `or`, `//` is `dor`. Every operator is
parenthesised here, so this measures RETURN VALUES rather than binding
-- the binding question is the next case.

```perl
my $a = $ARGV[0] // 1;
my $b = $ARGV[1] // 0;
my $and = ($a && $b);
my $or = ($a || $b);
my $dor = ($b // $a);
my $not = (not $a);
my $xor = ($a xor $b);
print "and [$and] or [$or] dor [$dor] not [$not] xor [$xor]\n";
```

```behavior
parses: yes
```

```output
and [0] or [1] dor [0] not [] xor [1]
```

```tokens
one word-shaped operator whose text is "xor"
one operator whose text is "||"
```

## `&&` and `and` are one op at two precedences

The same op, differing only in precedence, and the op names cannot show
it. `my $tight = $a && 9` binds tighter than `=`; `my $loose = $a and 9`
does not, so the assignment happens first and the `and` is discarded.
Different trees, different output, identical op names.

```perl
my $a = $ARGV[0] // 1;
my $tight = $a && 9;
my $loose = $a and 9;
print "$tight $loose\n";
```

```behavior
parses: yes
```

```output
9 1
```

```tokens
one operator whose text is "&&"
one word-shaped operator whose text is "and"
```

## Punctuation `!`, and the double negative

`!0` is 1 and `!!0` is the EMPTY STRING, not 0 -- perl's false is a
dual-valued empty-string-and-zero and `print` shows the string half. A
parser folding `!!` to a no-op prints `[1][0]`.

Measured, `!!$a` lexes as TWO `Operator("!")` tokens. There is no `!!`
entry in the lexer's operator table and there should not be, but the
source writes the bytes adjacently, so they are there for a greedy scan
to fuse.

```perl
my $a = $ENV{X} // 0;
print "[", !$a, "][", !!$a, "]\n";
```

```behavior
parses: yes
```

```output
[1][]
```

```tokens
no operator whose text is "!!"
```

## Comparisons chain, and the answer flips

`9 < 1 < 5` is FALSE. Chained it is `9 < 1 && 1 < 5`; read
left-associatively it is `("") < 5`, which is `0 < 5` and TRUE. One
source, opposite answers, no diagnostic either way.

The second line is the control and has to ASCEND: `1 < 5 < 9` is true
both ways, so a parser that prints the empty string twice fails it.
`1 < 9 < 5` would not serve -- false under both readings.

Constants are correct here, which is the one place in this tier that is
true: chaining is a PARSE-time n-ary grouping, not a foldable binary
expression, and `cmpchain_dup` survives constant operands.

```perl
print "[", (9 < 1 < 5), "]\n";
print "[", (1 < 5 < 9), "]\n";
```

```behavior
parses: yes
```

```output
[]
[1]
```

## Level 13 is `chain/na` and the halves differ

`1 == 1 == 1` compiles. `1 <=> 2 <=> 3` does not, nor does
`"a" cmp "b" cmp "c"`. A parser modelling perlop's level 13 as ONE
associativity class is wrong whichever it picks.

This is a `parses: no` case, which is what the claim IS: perl rejects
the source, so there is no output to pin and no tree to describe.

```perl
print 1 <=> 2 <=> 3;
```

```behavior
parses: no
```

```tokens
no operator whose text is "<="
no operator whose text is ">"
```

## Level 11 is the one true `%nonassoc`

`1 .. 2 .. 3` is a syntax error, and level 11 is the ONE `%nonassoc`
level in perly.y that means what the keyword suggests -- eight of the
eleven accept the repetition, because a conflict needs a left operand to
fight over and those are prefix forms whose second occurrence is the
first one's argument.

Without this case the corpus claims only the ACCEPTING direction, and a
parser refusing nothing would pass.

A `parses: no` case emits NO OPS, which is what makes this legal here:
`..` is tier 03's `range`/`flip`/`flop` and a parsing case spelling it
would fail the dependency lint.

```perl
my @x = (1 .. 2 .. 3);
```

```behavior
parses: no
```

```tokens
no operator whose text is "."
```
