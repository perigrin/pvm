# Control: a variable no code writes keeps its undef in a sub that reads
# it.
use strict;
use warnings;

my $lvl = undef;
my $check = sub { return 1 if defined caller $lvl; 0 };
print $check->(), "\n";
