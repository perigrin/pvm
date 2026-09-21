#!perl
# The six numeric relations and the three-way `<=>`: eq, ne, lt, gt, le,
# ge and ncmp. Numeric comparison imposes numeric context on both
# operands, which is the whole reason this tier sits after 03.
#
# TIER 04 operators
# INTRODUCES numeric comparison operators
# USES nothing from a later tier
#
# The op names are the trap here. Perl's `==` is the op `eq`, and Perl's
# `eq` is the op `seq` -- the op named after the SOURCE spelling is the
# string one, not the numeric one. Reading the op stream and matching op
# names to source text gets both backwards. `04_string_comparison.t` is
# the other half of the pair.
#
# Each result is printed between brackets rather than bare. A false
# comparison is `!!0`, which stringifies to the EMPTY string, so a bare
# `print "eq ", ($a == $b), "\n"` emits a line ending in a space -- and
# the repo's `trailing-whitespace` pre-commit hook would strip that space
# out of the expected output, leaving a CORPUS BUG nobody wrote. The
# brackets put a non-space byte at the end of every line.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $a = 3; my $b = 5; print "[", ($a == $b), "][", ($a < $b), "]\n"'
#   [][1]

--- source
my $a = $ARGV[0] // 3;
my $b = $ARGV[1] // 5;
print "eq [", ($a == $b), "]\n";
print "ne [", ($a != $b), "]\n";
print "lt [", ($a < $b), "]\n";
print "gt [", ($a > $b), "]\n";
print "le [", ($a <= $b), "]\n";
print "ge [", ($a >= $b), "]\n";
print "cmp [", ($a <=> $b), "]\n";

--- expect output
eq []
ne [1]
lt [1]
gt []
le [1]
ge []
cmp [-1]

--- expect parses

--- expect tokens
one operator whose text is "<=>"
