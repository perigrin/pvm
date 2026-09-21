#!perl
# A `0b` prefix makes the digits after it binary: `0b1010` is one token
# denoting ten.
#
# TIER 01 literals
# INTRODUCES numeric literal
# USES nothing from a later tier
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $x = 0b1010; print "$x\n"'
#   10
#
# The radix is invisible after compilation -- `const[IV 10]`, the same op
# `my $x = 10` produces -- so the optree cannot say which spelling was
# written. `perldata` lists this form under "Scalar value constructors",
# which is why ../GLOSSARY.md, "numeric literal", carries it.

--- source
my $x = 0b1010;
print "$x\n";

--- expect parses

--- expect output
10

--- expect tokens
one numeric literal whose text is "0b1010"
no operator whose text is "b"
