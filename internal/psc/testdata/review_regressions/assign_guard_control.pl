# Control: an assignment in a condition that is not a guard leaves the
# variable as the assignment typed it.
use strict;
use warnings;

my $n;
if (($n = sysread(NOFH, my $buf, 1)) || 1) {
    print $n + 1, "\n";
}
