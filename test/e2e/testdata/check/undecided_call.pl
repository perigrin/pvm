# ABOUTME: Fixture whose localtime call is in a sub's return, a context the call site cannot decide, for TestPSCCheckUndecidedCallReportsNothing.
# ABOUTME: Its candidates are joined, and perl -w runs it silently in both calling contexts.
use strict;
use warnings;

sub stamp { return localtime($_[0]) }

my @parts = stamp(0);
my $when = stamp(0);
print scalar(@parts), " $when\n";
print length(stamp(0)), "
";
