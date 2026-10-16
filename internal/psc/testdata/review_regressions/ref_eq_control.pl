# Control: an undef pattern is still reported.
use strict;
use warnings;

my $re = undef;
print "x" =~ $re ? "match\n" : "no\n";
