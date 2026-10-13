my $x;
unless (defined($x)) { 1; } elsif (ref($x)) { 2; } else { 3; }
