# Control: a variable a list assignment does not write keeps its undef.
use strict;
use warnings;

my ($a1, $b1);
my $c1 = undef;
($a1, $b1) = (1, 2);
print lc($c1), "\n";
