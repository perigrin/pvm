#!perl
# A leading zero makes the digits after it octal: `0377` is one token
# denoting 255, not three hundred and seventy-seven.
#
# TIER 01 literals
# INTRODUCES numeric literal
# USES nothing from a later tier
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $x = 0377; print "$x\n"'
#   255
#
# The radix is carried by a character a decimal literal may also begin
# with, so this is the one radix form a lexer can get wrong by doing
# NOTHING: scan the digits, read them as decimal, and `0377` becomes 377
# with no token boundary out of place and nothing to report. Only the
# VALUE differs, which is why this file pins output as well as tokens.

--- source
my $x = 0377;
print "$x\n";

--- expect parses

--- expect output
255

--- expect tokens
one numeric literal whose text is "0377"
no numeric literal whose text is "377"
