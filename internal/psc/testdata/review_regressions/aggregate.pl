# An array or a hash variable stays an array or a hash whatever fills it:
# Pod/Perldoc.pm, Class/Struct.pm, App/Prove/State.pm.
use strict;
use warnings;

sub dump_self {
    my $self = shift;
    my @x = %$self;
    while (@x) {
        splice @x, 0, 2;
    }
}

sub decls {
    my $base_type = shift;
    my @decls;
    if ($base_type eq 'HASH') {
        @decls = %{ shift() };
    } elsif ($base_type eq 'ARRAY') {
        @decls = @{ shift() };
    } else {
        @decls = @_;
    }
    die "odd" if @decls % 2 == 1;
    my $idx = 0;
    while ($idx < @decls) {
        $idx += 2;
    }
    return $idx;
}

sub tests {
    my @argv = @_;
    if ($^O eq 'MSWin32') {
        @argv = map { glob "$_" } @argv;
    }
    chomp(@argv);
    return @argv;
}

dump_self({ a => 1 });
print decls('HASH', { a => 1 }), "\n";
print join(",", tests("a\n", "b\n")), "\n";
