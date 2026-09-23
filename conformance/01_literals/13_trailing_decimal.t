#!perl
# A trailing decimal point with no digits after it is still part of the
# numeric literal: `1.` is one token, not `1` followed by the
# concatenation operator.
#
# TIER 01 literals
# INTRODUCES numeric literal
# USES nothing from a later tier
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $x = 1.; print "$x\n"'
#   1
#
# The mirror image of `06_leading_decimal.t`, and the pair is why both are
# here: `.5` puts the dot where a lexer expects an operator and `1.` puts
# it where a lexer expects more digits. A scanner that requires a digit on
# BOTH sides of the point gets each of them wrong in a different way, and
# a corpus holding only one of the two would diagnose that as a single
# fault.
#
# The output cannot distinguish the spellings -- `1.` prints `1`, exactly
# as `1` does -- so the token assertion is what this file turns on.

--- source
my $x = 1.;
print "$x\n";

--- expect parses

--- expect output
1

--- expect tokens
one numeric literal whose text is "1."
no operator whose text is "."
