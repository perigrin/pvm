# Control: the false arm of a truth test may hold undef.
use strict;
use warnings;

my $v = sysread(NOFH, my $buf, 1);
print $v ? 1 : $v + 1, "\n";
