sub is_defref { my $v = shift; defined($v) && ref($v) }
my $x;
if (is_defref($x)) { 1; }
