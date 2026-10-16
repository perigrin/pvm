use feature 'isa';
sub is_foo { my $v = shift; $v isa Foo }
my $x;
if (is_foo($x)) { 1; }
