#!perl
# `*` binds tighter than `+`, and the op stream cannot see it. The only
# evidence is behavioural: the two groupings print different numbers.
#
# TIER 04 operators
# INTRODUCES operator precedence
# USES nothing from a later tier
#
# `$a + $b * $c` and `($a + $b) * $c` emit the SAME four ops in the SAME
# order -- `add`, `multiply` and the two operand loads. `-exec` walks
# execution order, and execution order is identical; only the tree
# differs, and the tree is what `-exec` does not print. Grouping is what
# this tier is most about and it is the part the op list cannot measure,
# which is why this file's assertion is `expect output` rather than a
# claim about ops.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my ($a,$b,$c)=(2,3,4); print $a + $b * $c, " ", ($a + $b) * $c, "\n"'
#   14 20

--- source
my $a = $ARGV[0] // 2;
my $b = $ARGV[1] // 3;
my $c = $ARGV[2] // 4;
my $default = $a + $b * $c;
my $grouped = ($a + $b) * $c;
print "$default $grouped\n";

--- expect output
14 20

--- expect parses

--- expect tokens
one word whose text is "print"
