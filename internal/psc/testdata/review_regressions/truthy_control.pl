# Control: a false value may still be undef, so the else of a truth test
# is not narrowed.
use strict;
use warnings;

my $v = sysread(NOFH, my $buf, 1);
if ($v) {
    print "read\n";
} else {
    print $v + 1, "\n";
}
