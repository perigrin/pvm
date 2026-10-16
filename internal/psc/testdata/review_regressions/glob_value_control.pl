# Control: a glob is no number.
use strict;
use warnings;

my $v = *STDOUT;
if (ref \$v eq 'GLOB') {
    print $v + 1, "\n";
}
