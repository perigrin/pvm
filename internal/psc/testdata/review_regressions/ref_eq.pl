# ref($v) eq 'TYPE' guards $v as a reference of that type, and a Regexp
# is a pattern: Encode/Alias.pm.
use strict;
use warnings;

sub define_alias {
    my $alias = shift;
    if (ref($alias)) {
        for my $k (qw(latin1 utf8)) {
            if (ref($alias) eq 'Regexp' && $k =~ $alias) {
                print "$k\n";
            }
        }
    }
}

define_alias(qr/^lat/);
