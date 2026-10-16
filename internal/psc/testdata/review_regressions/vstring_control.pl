# Control: a string literal that is no number is still reported where a
# number is expected.
use strict;
use warnings;

print "abc" < 5, "\n";
