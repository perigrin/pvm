my %h = (a => 1, b => 2);
my $d = delete $h{a};
my @d = delete @h{"a", "b"};
