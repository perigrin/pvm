use v5.36;
sub has_ref ($v) { ref $v }
my $x;
if (has_ref($x)) { 1; }
