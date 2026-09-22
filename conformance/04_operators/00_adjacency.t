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
# Three Unknown nodes, measured, at TWO sites. The cited code is the
# FIRST one reached and the one the other two follow from:
#
#   not_a_term       span `)`   -- from `not $n <= 3` on the last line
#   trailing_tokens  span `print join(",", reverse sort @nums), " ",
#                                ($n >= 3 xor not $n <= 3)`
#   trailing_tokens  span `, "\n";`
#
# `not` arrives as a Word and nothing in parseTerm can begin a term with
# it, so the Unknown runs to the closing paren; the statement around it
# then has bytes left over, which is what the two `trailing_tokens`
# report. One declined term, two statements that could not finish -- the
# same refusal seen from three spans, which is why one code is cited
# rather than three.
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
# MEASURED perl 5.42.0:
#
#   $ perl conformance/04_operators/00_adjacency.t
#   sum [-2] rel [00] pick [1] rep [abab] pow [-3]
#   3,2,1 1 1 01

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
print "sum [$sum] rel [$rel] pick [$pick] rep [", $word x 2, "] pow [$power]\n";
print join(",", reverse sort @nums), " ", ($n >= 3 xor not $n <= 3), " $loose ", scalar(@cleared), scalar(@filled), "\n";

--- expect output
sum [-2] rel [00] pick [1] rep [abab] pow [-3]
3,2,1 1 1 01

--- expect parses
