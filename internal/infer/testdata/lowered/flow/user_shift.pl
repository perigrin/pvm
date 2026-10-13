sub is_def { my $v = shift; return defined $v; }
my $x;
if (is_def($x)) { 1; }
