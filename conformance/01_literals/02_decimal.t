#!perl
# A decimal point between digits is part of the numeric literal: `0.5` is
# one token, not `0` `.` `5`.
#
# TIER 01 literals
# INTRODUCES numeric literal
# USES nothing from a later tier
#
# This is the case that already works, and it is here for two reasons. It is
# the baseline the two files after it deviate from -- `.5` and `5e-1` are
# the same construct with the integer part removed and an exponent sign
# added -- so a regression here would explain both without being separately
# diagnosed. And a corpus whose every file refuses cannot demonstrate that
# passing is reachable.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $x = 0.5; print "$x\n"'
#   0.5

--- source
my $x = 0.5;
print "$x\n";

--- expect parses

--- expect output
0.5

--- expect tokens
one numeric literal whose text is "0.5"
no operator whose text is "."
