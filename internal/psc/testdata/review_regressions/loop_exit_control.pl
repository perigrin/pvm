# Control: a variable a last leaves undefined stays so after the loop.
use strict;
use warnings;

my $len;
while (1) {
    $len = sysread(NOFH, my $buf, 1);
    last unless defined $len;
    print "read $len\n";
}
print $len + 1, "\n";
