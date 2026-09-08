# ABOUTME: An explicit \@a, which perl reports as srefgen just like a prototype does.
# ABOUTME: Scoring this WRONG is the false-positive generator the comparison must avoid.
sub f { }
my @a;
f(\@a);
