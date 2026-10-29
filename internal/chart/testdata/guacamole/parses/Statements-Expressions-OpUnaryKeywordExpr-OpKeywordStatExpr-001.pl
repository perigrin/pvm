use feature ':5.36';
my $perm = (stat $fh)[2] & 07777;