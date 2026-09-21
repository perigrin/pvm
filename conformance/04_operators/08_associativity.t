#!perl
# `**` is RIGHT associative and `-` is LEFT associative, and as with
# precedence the op stream shows neither: two `pow` ops and two
# `subtract` ops, in source order, either way.
#
# TIER 04 operators
# INTRODUCES operator associativity
# USES nothing from a later tier
#
# `**` is the operator where associativity is observable with small
# integers, which is why it is the probe. `2 ** 3 ** 2` is `2 ** (3 ** 2)`
# = `2 ** 9` = 512; forced the other way, `(2 ** 3) ** 2` = `8 ** 2` = 64.
# A parser that associated `**` leftwards would print 64 for the first
# expression and this file would name it.
#
# `$a - $b - $c` is the left-associative counterpart, included so the two
# directions sit in one file: `2 - 3 - 2` is `(2 - 3) - 2` = -3, not
# `2 - (3 - 2)` = 1.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'print 2**3**2, " ", (2**3)**2, "\n"'
#   512 64

--- source
my $a = $ARGV[0] // 2;
my $b = $ARGV[1] // 3;
my $c = $ARGV[2] // 2;
my $right = $a ** $b ** $c;
my $left = ($a ** $b) ** $c;
my $minus = $a - $b - $c;
print "$right $left $minus\n";

--- expect output
512 64 -3

--- expect parses

--- expect tokens
no operator whose text is "*"
