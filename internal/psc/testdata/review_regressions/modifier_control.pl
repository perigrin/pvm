# Control: an assignment a modifier may skip leaves what the variable
# held before.
use strict;
use warnings;

my $acc = undef;
$acc = 5 if @ARGV;
print $acc + 1, "\n";
