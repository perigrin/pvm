use feature 'isa';
my $x;
if (defined($x)) { 1; } elsif (ref($x)) { 2; } elsif ($x isa Foo) { 3; }
