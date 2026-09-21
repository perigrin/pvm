#!perl
# `localtime` is the sharpest discriminator in the tier: its two contexts
# return unrelated types -- a nine-element list, or a single string.
#
# TIER 03 context
# INTRODUCES localtime
# USES nothing from a later tier
#
# `reverse` and `sort` return rearrangements of the same data in both
# contexts. `localtime` does not: in list context it is
# (sec,min,hour,mday,mon,year,wday,yday,isdst) and in scalar context it is
# a formatted string like `Sun Sep 21 14:03:11 2026`. Nothing about the
# call site says which; the assignment target decides.
#
# What this file pins is deliberately NOT the time. `localtime` depends on
# the clock and the timezone, so pinning either result verbatim would make
# the file fail tomorrow and in another zone. Both assertions here are
# facts about SHAPE: the list has nine elements always, and the scalar
# string is one element always. That is exactly the contextual difference
# the file exists to show, and it is the part that does not move.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $n = () = localtime; my $s = localtime; my @c = ($s); print "$n ", scalar(@c), "\n"'
#   9 1

--- source
my $n = () = localtime;
my $s = localtime;
my @c = ($s);
print "$n ", scalar(@c), "\n";

--- expect output
9 1

--- expect parses

--- expect tokens
no variable whose text is "@n"
