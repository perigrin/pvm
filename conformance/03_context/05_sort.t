#!perl
# `sort` in list context returns the sorted elements; in scalar context
# the count idiom `() = sort @a` reports how many there were.
#
# TIER 03 context
# INTRODUCES sort
# USES nothing from a later tier
#
# The second statement is the count idiom, `my $n = () = sort @a`, and it
# is here because it is the only way to observe the list-versus-scalar
# distinction on an `aassign`. The inner assignment to an empty list runs
# in list context, and the outer assignment then asks the list assignment
# how many elements it moved -- `aassign ... s` rather than `l`.
#
# Neither half introduces an op the other does not. As with `padav`, what
# separates them is the flag B::Concise prints, not the op name.
#
# No comparator block appears here on purpose: `sort { $a <=> $b } @a`
# would drag in a block and the comparison operator, which belong to later
# tiers. Default sort is a string sort, which is why the digits come back
# in the order they do.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my @a=(3,1,2); my @b=sort @a; my $n=()=sort @a; print "@b $n\n"'
#   1 2 3 3

--- source
my @a = (3, 1, 2);
my @b = sort @a;
my $n = () = sort @a;
print "@b $n\n";

--- expect output
1 2 3 3

--- expect parses

--- expect tokens
no variable whose text is "@n"
