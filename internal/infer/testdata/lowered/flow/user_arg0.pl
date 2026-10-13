sub is_def { defined $_[0] }
my $x;
if (is_def($x)) { 1; } else { 2; }
