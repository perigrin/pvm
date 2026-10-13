my $x;
if (is_def($x)) { 1; }
sub is_def { my $v = shift; defined $v }
