my $v;
setit() unless defined $v;
my $ok = ($v ge 'a');
sub setit { $v = 'x' }
