#!perl
# Every kind of operator this tier introduces, in one body, each adjacent
# to another -- and adjacent to tier 03's list operators, which is what
# this tier declares it depends on.
#
# TIER 04 operators
# INTRODUCES nothing of its own
# USES nothing from a later tier
# STATUS refuses as of this file. Refusal not_a_term.
#
# FIVE Unknown nodes, measured, at THREE INDEPENDENT sites. The cited
# code is the first one reached:
#
#   not_a_term       span `)`
#   trailing_tokens  span `undef @cleared;`
#   trailing_tokens  span `print join(",", reverse sort @nums), " ",
#                                ($n >= 3 xor not $n <= 3)`
#   trailing_tokens  span `, " $loose ", scalar(@cleared),
#                                scalar(@filled), "\n";`
#
# Three of the four are one failure. `not` arrives as a Word and nothing
# in parseTerm can begin a term with it, so the Unknown runs to the
# closing paren; the statement around it then has bytes left over, which
# is what two of the `trailing_tokens` report. One declined term, two
# statements that could not finish.
#
# The fourth is SEPARATE and is `undef @cleared` -- the unary spelling of
# `undef`, which our parser reads as a complete term with `@cleared`
# stranded after it. `10_undef_arity.t` bisects that refusal and records
# it as the same missing rule tier 07's parenless-call files hit. It is
# named here because a reader counting three spans from the `not` failure
# and finding four would otherwise look for a fourth consequence of it.
#
# The FIFTH is `$t x= 2`, added with the compound assignment family, and
# is a third independent site: our lexer emits `Word(x) Operator(=)`
# rather than forming the `x=` token at all, so the statement has a Word
# where an operator belongs. `24_compound_repeat.t` bisects it against
# `.=` and the binary `x`, both of which parse.
#
# The five string operators added later -- `chr`, `ord`, `index`,
# `sprintf`, `substr`, on the third print line -- contribute NO Unknown.
# Measured: the count was four before they were written and four after,
# and each of them parses in isolation. The adjacency claim they make is
# about placement, not about refusal.
#
# That is the adjacency file earning its
# keep: every construct here appears in a sibling file that parses, and
# only the mixture refuses. A one-construct-per-file corpus would have
# gone green over all of it.
#
# The tier's other files are one KIND of operator each, which is what
# makes them diagnosable: when `08_associativity.t` prints 64 512, the
# construct that broke is the only one present. That same property is why
# a corpus of such files cannot reach an adjacency bug -- a parser that
# handles each operator class alone and mis-handles a mixture goes green
# over the mixture, and a mixture is exactly what real Perl is.
#
# Here the adjacency is dense on purpose. `$n + 1 - 2 * 3` mixes three
# precedence levels in one expression. `($n <=> 3) . ($word cmp "ab")`
# feeds two comparisons into a concatenation, so a parser that got the
# relational precedence wrong would concatenate the wrong things.
# `-$n ** 2 / 3` is the pairing that catches the most parsers: `**` binds
# TIGHTER than unary minus, so this is `-(($n ** 2) / 3)` and not
# `((-$n) ** 2) / 3`. With `$n` 3 the first is -3 and the second is 3, so
# the two readings are distinguishable in one byte of output.
#
# The tier-03 pairing is the last line: `join`, `sort` and `reverse` are
# 03's, and `$n >= 3 xor not $n <= 3` puts this tier's logical and
# relational operators directly beside them in one `print` list.
#
# `reverse sort @nums` emits `sort` alone -- the optimiser folds the
# reverse into the sort's direction flag and no `reverse` op survives.
# Tier 03 claims `reverse` and must emit it from a file of its own; this
# file cannot do it for them.
#
# `$loose` is the NAMED UNARY level: `defined $n + 1` is
# `defined($n + 1)` and prints 1, where `(defined $n) + 1` would print
# 2. It sits directly beside `$power`'s unary minus, which is the other
# precedence trap in this body, so a parser that got the unary levels
# right in isolation and wrong in a mixture is caught here and nowhere
# else. `09_named_unary.t` is the one-construct half.
#
# `@cleared` and `@filled` are `undef`'s two ARITIES, adjacent: `undef
# @cleared` is the unary form and empties the array, `my @filled = undef`
# is the niladic form and fills one with a single undef. Measured, `0`
# and `1`, which is the `01` this file's second line ends on.
# `10_undef_arity.t` is the one-construct half and records that our
# parser refuses the unary spelling alone.
#
# `my $n = $ARGV[0] // @nums` gives the runtime operand this tier cannot
# do without AND a tier-03 scalar-context array in one expression: with
# no arguments `$ARGV[0]` is undef, so `$n` is the array's count, 3.
#
# THE FIVE STRING OPERATORS ARE ON THE THIRD LINE, and one of them had
# to be kept off the first. `substr` is this tier's only LVALUE, so
# `substr($word, 0, 1) = ...` would have edited `$word` in place and
# changed `rep [abab]` to `rep [bbbb]` six lines later -- measured, not
# reasoned. That is the adjacency risk the construct carries and it is
# real enough that this file dodges it: `$edit` is a copy, so the
# mutation is observable without reaching back into a pin two statements
# above it. `15_substr_lvalue.t` is the one-construct half and measures
# the mutation directly.
#
# `chr(ord($word) + 1)` nests the two inverses around an `add`, so the
# named-unary argument extent and the arithmetic level are adjacent in
# one expression -- the same pairing `$loose` makes for `defined`.
# `sprintf("%0*d", $n, $n)` puts a format whose arity is decided inside a
# string beside `index`, whose failure answer is an ordinary `-1`;
# `substr($word, 1, 1)` closes the line with the rvalue form at a
# NONZERO offset, which is what keeps `substr_left` -- an op no tier
# claims -- out of this body. `14_substr_arity.t` records that
# constraint.
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/04_operators/00_adjacency.t
#   sum [-2] rel [00] pick [1] rep [abab] pow [-3]
#   3,2,1 1 1 01
#   003 [bb] 0 [b]
#   bits [2] prec [2] shift [12] comp [249]
#   acc [7] app [ab] rep [abab] def [0] mask [15]
#
# THE BITWISE BAND is the fourth line, and it is the one place in this
# body where the same characters mean something else elsewhere in the
# tier: `&` is the code sigil in `\&foo`, `|` is half of `//` two lines
# above, and `<<` is a heredoc opener three tiers away. Putting them
# beside those spellings is the adjacency claim -- a lexer that resolved
# `&` by looking only at the character would produce a code reference
# here and still parse.
#
# `($a | $b) & $c` is the PRECEDENCE pair's parenthesised half; the
# unparenthesised `$a | $b & $c` lives in `17_bitwise_precedence.t`,
# which prints both and pins that they differ. Only the grouped form is
# here, because this file's job is placement rather than discrimination.
#
# THE PRAGMA PAIR IS DELIBERATELY ABSENT, and the reason is a tier
# boundary rather than a preference. `20_bitwise_polymorphic.t` and
# `21_bitwise_numeric.t` differ only by a `use v5.28`, and the two
# readings can be put in ONE program -- measured, `use v5.28` is
# lexically scoped:
#
#   my $poly = $s1 & $s2;                   10
#   { use v5.28; $num = $s1 & $s2; }         8
#
# One pair of string operands, two answers, four lines apart. It would be
# the strongest claim this file could make, and it cannot be made here: a
# bare block emits `enterloop` and `leaveloop`, which are `05_scoping`'s
# ops, and a tier-04 file may not use them. Measured by writing it and
# letting the dependency lint refuse it.
#
# `no feature "bitwise"` does not substitute: measured, it does NOT
# restore the polymorphic reading, so both halves print 8 and the pair
# says nothing. The block is the only spelling, and the block is out of
# budget. The claim therefore lives in the two construct files, each
# holding one half.
#
# THE COMPOUND ASSIGNMENTS are the fifth line, one from each of the
# family's groups: `+=` reuses its binary op, `.=` fuses to
# `multiconcat`, `//=` emits a dedicated `dorassign`, `|=` reuses the
# bitwise op this file introduced four lines above.
#
# `$t x= 2` is here too and it is the one that REFUSES -- our lexer emits
# `Word(x) Operator(=)` rather than forming the token, so this body now
# carries a fifth Unknown that `24_compound_repeat.t` bisects. That is
# the adjacency file honouring its own stated principle: a body holding
# every construct the tier introduces holds the refusing ones too, and
# composing only what already works would make it green and make it stop
# covering the tier.
#
# `def [0]` is the short circuit visible in output: `$def` starts at 0,
# which is DEFINED but false, so `//=` leaves it alone where `||=` would
# have replaced it. Every other compound assignment on this line is
# confirmable only by its token, which is why `25_compound_shortcircuit.t`
# is the only one of the five whose claim output can reach.

--- source
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
print "sum [$sum] rel [$rel] pick [$pick] rep [", $word x 2, "] pow [$power]\n";
print join(",", reverse sort @nums), " ", ($n >= 3 xor not $n <= 3), " $loose ", scalar(@cleared), scalar(@filled), "\n";
print sprintf("%0*d", $n, $n), " [$edit] ", index($edit, "b"), " [", substr($word, 1, 1), "]\n";
print "bits [$bits] prec [$prec] shift [$shifted] comp [$comp]\n";
print "acc [$acc] app [$app] rep [$t] def [$def] mask [$mask]\n";

--- expect output
sum [-2] rel [00] pick [1] rep [abab] pow [-3]
3,2,1 1 1 01
003 [bb] 0 [b]
bits [2] prec [2] shift [12] comp [249]
acc [7] app [ab] rep [abab] def [0] mask [15]

--- expect parses
