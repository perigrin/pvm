sub missing { my $v = shift; !defined $v }
my $x;
if (missing($x)) { 1; } else { 2; }
