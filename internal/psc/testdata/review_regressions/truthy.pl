# A true value is defined: CPAN/Distribution.pm, IPC/Open3.pm,
# Test.pm, Test2/API/InterceptResult/Squasher.pm, Test2/Util/Stash.pm.
use strict;
use warnings;

sub reap {
    my $pid = fork;
    if (!defined $pid) {
        die "fork: $!";
    } elsif ($pid) {
        waitpid($pid, 0);
    } else {
        exit 0;
    }
}

sub flags {
    my $flags = sysread(STDIN, my $buf, 1);
    die "nothing read" unless $flags;
    return $flags | 1;
}

sub same {
    my ($into, $sig) = @_;
    my $fsig = $into ? lc($into) : undef;
    return $fsig && $fsig eq $sig;
}

sub pkg {
    my ($pkg, $package) = @_;
    $pkg = $pkg ? substr($pkg, 0, -2) : undef;
    return $pkg && $package && $pkg ne $package;
}

my $kind = undef;
my $flush = sub { $kind = undef };
my $collect = sub {
    my $k = shift;
    $flush->() if $kind && $kind ne $k;
    $kind = $k;
};

reap();
print flags(), "\n";
print same("A", "a") ? "same\n" : "differ\n";
print pkg("Foo::", "Foo") ? "mismatch\n" : "match\n";
$collect->("a");
$collect->("b");
