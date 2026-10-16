# A statement modifier narrows its statement by its condition, and what
# the statement assigns joins what the variable held: Math/BigInt.pm,
# IO/Socket.pm.
use strict;
use warnings;

sub round {
    my ($x, $acc) = @_;
    $acc = undef if defined $acc && $acc == 0;
    $acc = int($acc) if defined $acc;
    return defined $acc ? sprintf("%.*f", $acc, $x) : $x;
}

sub opt {
    my $r = getsockopt(STDIN, 1, 1);
    $r = unpack("i", $r) if defined $r && length($r) == 4;
    return $r;
}

print round(3.14159, 2), "\n";
my $o = opt();
print defined $o ? "opt\n" : "no opt\n";
