our %h = (a => 1);
my $r = \%h;
print $h{a}, $r->{b}{c}, %$r, @h{'a'};
