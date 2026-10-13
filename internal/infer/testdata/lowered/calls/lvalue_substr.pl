my $f = "abc";
substr($f, 1, 0) = "-";
my $x = "k " . substr $f, 1;
sub g { return join ' ', @_; }
