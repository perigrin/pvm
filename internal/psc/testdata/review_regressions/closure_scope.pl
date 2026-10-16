# What an anonymous sub's body narrows or assigns stays in its body: the
# code after it runs before the sub is called, and the sub reads what
# any write has left. base.pm, Test.pm.
use strict;
use warnings;

sub hooks {
    my $lvl;
    my $count = sub { return if defined $lvl; 1 while defined caller ++$lvl; () };
    my $check = sub { return if defined caller $lvl; () };
    $count->();
    $check->();
}

my $kind = 'a';
my $flush = sub { $kind = undef };
print lc($kind), "\n";
$flush->();

hooks();
print "ok\n";
