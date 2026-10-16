# Control: a list assignment's first scalar from a call that may return
# undef may be undef.
use strict;
use warnings;

my $n;
($n) = sysread(NOFH, my $buf, 1);
print $n + 1, "\n";
