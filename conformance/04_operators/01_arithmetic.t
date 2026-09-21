#!perl
# The seven arithmetic operators, each with a RUNTIME operand so the
# optimiser cannot fold it away: add, subtract, multiply, divide, modulo,
# pow and negate.
#
# TIER 04 operators
# INTRODUCES arithmetic operators
# USES nothing from a later tier
#
# `$ARGV[0] // 12` is the runtime operand. It is the cheapest one
# available: measured, it costs `aelemfast` and `dor` and nothing else.
# `shift` costs a `shift` op no tier claims, and `$ENV{X}` costs
# `multideref` -- both tier 02's or worse, and `shift`'s op is unclaimed
# anywhere, which would fail the lint outright.
#
# With no arguments `$ARGV[0]` is undef and the defaults apply, so the
# output is deterministic and warning-free. That is the whole reason for
# the `//`: `$ARGV[0] + 5` would warn on an uninitialised value.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $a = $ARGV[0] // 12; my $b = $ARGV[1] // 5; print $a % $b, "\n"'
#   2

--- source
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

--- expect output
17 7 60 2.4 2 248832 -12

--- expect parses

--- expect tokens
one operator whose text is "%"
one operator whose text is "**"
