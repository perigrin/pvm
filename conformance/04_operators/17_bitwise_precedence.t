#!perl
# `&` binds TIGHTER than `|`, and nothing in the corpus said so.
#
# TIER 04 operators
# INTRODUCES nothing of its own
# USES the bitwise operators from 16_bitwise.t
#
# perlop puts `&` at level 14 and `|` at level 15, one apart, so
# `$a | $b & $c` is `$a | ($b & $c)`. A parser that gave them equal
# precedence and read left-to-right would compute `($a | $b) & $c`
# instead.
#
# This file exists because the whole band arrived at once. A tier that
# introduces two operators at adjacent levels and never writes them
# together has claimed each one's existence and nothing about their
# relationship -- and the relationship is what a precedence table IS.
#
# MEASURED perl 5.42.0, and the two readings disagree:
#
#   $ perl -e 'my $a=6; my $b=3; my $c=2; print $a | $b & $c'
#   6
#   $ perl -e 'my $a=6; my $b=3; my $c=2; print(($a | $b) & $c)'
#   2
#
# 6 against 2. The operands are chosen so the two groupings differ:
# `3 & 2` is 2 and `6 | 2` is 6, while `6 | 3` is 7 and `7 & 2` is 2.
# Many triples give the same answer either way, which is why these
# particular ones are pinned.
#
# Every operand is a runtime value for the reason 16_bitwise.t records:
# with constants the optimiser folds the expression and there is no
# precedence left to observe.
#
# The file prints BOTH groupings so the claim is a comparison rather than
# a number to look up. A parser with the levels reversed prints `2 2`.
#
# The token facts are NEGATIVE here rather than counting, because this
# source holds two `&` and two `|` and a count would be a claim about the
# file's punctuation rather than about its construct. What they rule out
# is the doubling: a lexer that read `$b & $c` as `&&` or `$a | $b` as
# `||` would produce a program with SHORT-CIRCUIT semantics and a
# plausible answer. Both spellings appear nowhere else here, so the
# negatives are claims rather than accidents.

--- source
my $a = $ENV{X} // 6;
my $b = $ENV{Y} // 3;
my $c = $ENV{Z} // 2;
print $a | $b & $c, " ", ($a | $b) & $c, "\n";

--- expect output
6 2

--- expect parses

--- expect tokens
one operator whose text is "("
