# ABOUTME: Fixture whose calls to Carp's declared subs fit their declarations, for TestPSCCheckDeclarationsNoFalsePositive.
# ABOUTME: perl -w runs it with no warning of its own; carp and cluck print their messages.
use strict;
use warnings;
use Carp qw(carp cluck longmess);

my $count = 3;
my $ok = carp("seen ", $count, " items");
print "carped\n" if $ok;
cluck("trace");
my $mess = longmess("x");
print length($mess) > 0 ? "long\n" : "short\n";
