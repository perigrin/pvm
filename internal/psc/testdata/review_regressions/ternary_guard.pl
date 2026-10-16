# A ternary's arms are narrowed by its condition: IPC/Semaphore.pm.
use strict;
use warnings;

sub count {
    my $v = sysread(STDIN, my $buf, 1);
    return $v ? 0 + $v : undef;
}

my $n = count();
print defined $n ? $n : "none", "\n";
