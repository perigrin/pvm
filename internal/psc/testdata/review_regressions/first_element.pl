# A list assignment's first scalar takes one element of what it is
# assigned, from either arm of a ternary, from a sub's list or from a
# match's captures: Math/Complex.pm, Memoize.pm, Unicode/UCD.pm, Cwd.pm.
use strict;
use warnings;

sub cexp {
    my ($z) = @_ ? @_ : $_;
    return CORE::exp($z) unless ref $z;
    return $z;
}

sub cache_opt {
    my ($fullopt) = @_;
    my ($cache_opt, @args) = ref $fullopt ? @$fullopt : $fullopt;
    return $cache_opt eq 'HASH' ? 1 : 0;
}

sub aliases {
    my @list = ("gc", "General_Category");
    return @list if wantarray;
    return;
}

sub values_of {
    my $prop = shift;
    ($prop) = aliases($prop);
    return undef if ! $prop;
    return lc $prop;
}

sub abs_path {
    my $path = shift;
    ($path) = $path =~ /(.*)/s;
    return CORE::chdir($path) ? 1 : 0;
}

print cexp(1), "\n";
print cache_opt(["HASH"]), "\n";
print values_of("gc") // "none", "\n";
print abs_path("/"), "\n";
