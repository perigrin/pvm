#!perl
# Every kind of operator this tier introduces, in one body, each adjacent
# to another -- and adjacent to tier 03's list operators, which is what
# this tier declares it depends on.
#
# TIER 04 operators
# INTRODUCES nothing of its own
# USES nothing from a later tier
# STATUS refuses as of this file.
#
# Three Unknown nodes, measured. That is the adjacency file earning its
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
# `my $n = $ARGV[0] // @nums` gives the runtime operand this tier cannot
# do without AND a tier-03 scalar-context array in one expression: with
# no arguments `$ARGV[0]` is undef, so `$n` is the array's count, 3.
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/04_operators/00_adjacency.t
#   sum [-2] rel [00] pick [1] rep [abab] pow [-3]
#   3,2,1 1

--- source
my @nums = (3, 1, 2);
my $n = $ARGV[0] // @nums;
my $word = $ARGV[1] // "ab";
my $sum = $n + 1 - 2 * 3;
my $rel = ($n <=> 3) . ($word cmp "ab");
my $pick = ($n > 2 and $word ne "zz") || ($n % 2);
my $power = -$n ** 2 / 3;
print "sum [$sum] rel [$rel] pick [$pick] rep [", $word x 2, "] pow [$power]\n";
print join(",", reverse sort @nums), " ", ($n >= 3 xor not $n <= 3), "\n";

--- expect output
sum [-2] rel [00] pick [1] rep [abab] pow [-3]
3,2,1 1

--- expect parses
