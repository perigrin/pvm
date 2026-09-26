# Every operator, each beside another

One body holding every construct this tier introduces, each adjacent to
another -- and adjacent to tier 03's list operators, which is what this
tier depends on.

**Tier 04 operators.** Introduces nothing of its own; it is the mixture
that is the subject.

WHY A MIXTURE NEEDS ITS OWN CASE. The tier's other cases are one kind of
operator each, which is what makes them diagnosable: when the
associativity case prints `512 64`, the construct that broke is the only
one present. That same property is why a corpus of such cases cannot
reach an adjacency bug -- a parser that handles each operator class
alone and mishandles a mixture goes green over every isolated case, and
a mixture is exactly what real Perl is.

THE CLAIM IS THE SOURCE. Every construct the tier introduces appears
below, next to another one, and a reader can see that by reading it.
This used to be asserted by a Go map from file name to a spelling to
grep for, because a one-construct-per-file corpus had no way to say
"these are adjacent" except by naming files. A case whose body IS the
mixture needs no such map.

The adjacency here is dense on purpose. `$n + 1 - 2 * 3` mixes three
precedence levels in one expression; `($n <=> 3) . ($word cmp "ab")`
feeds two comparisons into a concatenation; `substr($edit, 0, 1) =
chr(ord($word) + 1)` puts an lvalue substr, chr and ord in one
statement.

WHAT THIS DOES NOT ESTABLISH is that the constructs are adjacent in any
stronger sense than "in the same body". Tier N beside tier N-3 is
unreached by construction, and the `t/` sweep is what finds it -- the
plan defers that deliberately and names where it goes.

## The whole tier in one body

ZERO Unknown nodes, and the whole tier in one body is why this case is
kept: every construct here also appears in a sibling case that parses on
its own, and for three successive gaps the MIXTURE was the only thing
that refused. A one-case-per-construct corpus would have gone green over
all of it and seen none of them.

### The last of those gaps, and what closed it

It was `undef @cleared` -- the unary spelling, which the parser read as a
complete term with `@cleared` stranded after it, because `undef` was
absent from `parse.namedUnary`. `01a0dd43` added the entry, classified by
that table's own deparse method: `undef $x, $y` gives `(undef($x), $y)`,
the comma outside, so it is a named unary.

### What this passage used to say, and why it was wrong three times

It said TWO nodes, the second being `$t x= 2`: the lexer emitted
`Word(x) Operator(=)` rather than forming the `x=` token, so the
statement had a Word where an operator belonged. `01a0ce57` fixed that in
the lexer -- `x=` is the only compound assignment perl spells with a
letter, which is why the punctuation scanner never reached it -- and
`compound-assignment.md` carries the case.

### What it said before that, and why it was wrong twice

It said FIVE nodes at three sites, and attributed three of them to `not`
arriving as a Word with nothing in `parseTerm` able to begin a term with
it.

The count is now two, because `01a0dbe8` taught the paren to hold a full
expression -- `and`, `or` and `xor` bind BELOW the comma, so a paren
parsing its elements at the comma's power never reached them.

The attribution was wrong before that landed. The site was the
parenthesised `xor`, not `not`: measured, `(not $a)` parses clean on its
own and always did, because `not` is PREFIX and `prefix` has it. Reading
a count off a failing case and then naming a cause for it is how that
happened; the cause was never measured separately.

The file test is the one construct NOT here. `31_file_test`'s claim is
`no operator whose text is "-"`, and this body writes `-$n ** 2` and
`~$a`, so the negative is false here and no spelling of it is not.

```perl
my @nums = (3, 1, 2);
my $n = $ARGV[0] // @nums;
my $word = $ARGV[1] // "ab";
my $sum = $n + 1 - 2 * 3;
my $rel = ($n <=> 3) . ($word cmp "ab");
my $pick = ($n > 2 and $word ne "zz") || ($n % 2);
my $power = -$n ** 2 / 3;
my $loose = defined $n + 1;
my @cleared = @nums;
undef @cleared;
my @filled = undef;
my $edit = $word;
substr($edit, 0, 1) = chr(ord($word) + 1);
my $a = $ARGV[2] // 6;
my $b = $ARGV[3] // 3;
my $c = $ARGV[4] // 2;
my $bits = $a & $b;
my $prec = ($a | $b) & $c;
my $shifted = $a << 1;
my $comp = ~$a & 255;
my $acc = $ARGV[5] // 5;
$acc += 2;
my $app = $ARGV[6] // "a";
$app .= "b";
my $t = $ARGV[7] // "ab";
$t x= 2;
my $def = $ARGV[8] // 0;
$def //= 99;
my $mask = $ARGV[9] // 12;
$mask |= 3;
my $count = $ARGV[10] // 5;
$count++;
$count++;
--$count;
my $tag = $ARGV[11] // "Az";
$tag++;
my $flag = !$count;
my $group = +($n + 1) * 2;
my $chain = 9 < 1 < 5;
print "sum [$sum] rel [$rel] pick [$pick] rep [", $word x 2, "] pow [$power]\n";
print join(",", reverse sort @nums), " ", ($n >= 3 xor not $n <= 3), " $loose ", scalar(@cleared), scalar(@filled), "\n";
print sprintf("%0*d", $n, $n), " [$edit] ", index($edit, "b"), " [", substr($word, 1, 1), "]\n";
print "bits [$bits] prec [$prec] shift [$shifted] comp [$comp]\n";
print "acc [$acc] app [$app] rep [$t] def [$def] mask [$mask]\n";
print "count [$count] tag [$tag] flag [$flag] group [$group] chain [$chain]\n";
```

```behavior
parses: yes
```

```output
sum [-2] rel [00] pick [1] rep [abab] pow [-3]
3,2,1 1 1 01
003 [bb] 0 [b]
bits [2] prec [2] shift [12] comp [249]
acc [7] app [ab] rep [abab] def [0] mask [15]
count [6] tag [Ba] flag [] group [8] chain []
```
