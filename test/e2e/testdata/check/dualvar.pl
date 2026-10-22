# ABOUTME: Fixture using Scalar::Util's dualvar as a number and as a string, for the psc check e2e test.
# ABOUTME: perl 5.42 -w runs it silently; each use reads one of the dualvar's two values and drops the other.
use strict;
use warnings;
use Scalar::Util qw(dualvar);

my $status = dualvar(404, "Not Found");
my $next = $status + 1;
my $text = lc $status;
my @parts = split dualvar(1, " "), $text;
print "$next $text @parts\n";
