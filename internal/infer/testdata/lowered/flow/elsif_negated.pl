use feature 'isa';
my $x;
if ($x isa Foo) { 1; } elsif (!defined($x)) { 2; } else { 3; }
