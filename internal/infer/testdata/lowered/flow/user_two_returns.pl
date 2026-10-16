sub check { my $v = shift; return 0 if 1; return defined $v; }
my $x;
if (check($x)) { 1; }
