# Control: an array filled from a hash is still an array, which a
# scalar slot does not take.
use strict;
use warnings;

my %h = (a => 1);
my @x = %h;
print length(@x), "\n";
