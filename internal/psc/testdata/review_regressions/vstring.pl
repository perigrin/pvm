# A v-string is a version literal, which perl compares as a version
# against $^V, a version object: ExtUtils/Constant/ProxySubs.pm.
use strict;
use warnings;

my $croak_on_error = 1;
print "old perl\n" if $croak_on_error && $^V < v5.13.5;
print "ok\n";
