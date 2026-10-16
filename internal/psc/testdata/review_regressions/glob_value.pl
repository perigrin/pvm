# A scalar may hold a glob, which ref and fileno take: Dumpvalue.pm.
use strict;
use warnings;

sub unwrap {
    my ($v) = shift;
    my $address;
    if (ref \$v eq 'GLOB') {
        $address = "$v";
    }
    if (ref $v eq 'Regexp') {
        print "regexp\n";
    }
    my $fileno;
    if (defined($fileno = fileno($v))) {
        print "fileno $fileno\n";
    }
    return $address;
}

print unwrap(*STDOUT) // "none", "\n";
