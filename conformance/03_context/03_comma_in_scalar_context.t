#!perl
# `my $last = (4,5,6)` is 6. The comma operator in scalar context
# evaluates and discards its left operands, and the optimiser then erases
# them from the optree entirely.
#
# TIER 03 context
# INTRODUCES the comma operator in scalar context
# USES nothing from a later tier
#
# The scalar half emits `pushmark v`, `const[IV 6]`, `list sKP` -- the
# constants 4 and 5 are simply gone. Two-thirds of the construct this file
# is about does not survive to the optree, because discarding them is
# precisely what the comma operator does here.
#
# That makes this the tier's example of the lesson tier 01 records for
# `my $x = 1+2` folding to `const[IV 3]`: ops LINT a declared tier and
# cannot derive one, and a file demonstrating a construct may emit fewer
# ops than the construct has parts.
#
# The list half is the discriminator. The same parenthesised commas
# assigned to an array keep all three values, so the pair shows that the
# source text does not decide -- the assignment target does.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $last=(4,5,6); my @all=(4,5,6); print "$last @all\n"'
#   6 4 5 6

--- source
my $last = (4, 5, 6);
my @all = (4, 5, 6);
print "$last @all\n";

--- expect output
6 4 5 6

--- expect parses

--- expect tokens
one variable whose text is "$last"
one variable whose text is "@all"
