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
# how many elements it moved -- `aassign[t6] sKS`.
#
# The third statement is that same idiom with an ARRAY receiving it, and
# it is what makes the second a pair rather than a single measurement:
# `my @c = (() = sort @a)` emits `aassign[t8] lKPS`. Same op, same
# operand, same spelling of the idiom -- and the answers are unrelated.
# In scalar context the empty-list assignment reports how many elements
# it moved, which is 3; in list context it yields what it assigned, which
# is the empty list, so `@c` has 0 elements. Nothing in the source says
# which; the receiver does.
#
# Without the third statement this file's `aassign` is only ever seen
# with `s`, and the claim that context is a flag would be made by a file
# that never shows the other value of the flag.
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
#   $ perl -e 'my @a=(3,1,2); my @b=sort @a; my $n=()=sort @a; my @c=(()=sort @a); print "@b $n ", scalar(@c), "\n"'
#   1 2 3 3 0

--- source
my @a = (3, 1, 2);
my @b = sort @a;
my $n = () = sort @a;
my @c = (() = sort @a);
print "@b $n ", scalar(@c), "\n";

--- expect output
1 2 3 3 0

--- expect parses

--- expect tokens
one variable whose text is "$n"
one variable whose text is "@b"
