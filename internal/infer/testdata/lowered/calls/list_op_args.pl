my @todo;
my $o = 1;
push @todo, [1, 2] unless $o;
push @todo, 'x' if $o;
my @a;
my $s = join ",", @a;
