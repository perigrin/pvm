print STDERR "hello\n";
open my $fh, '>', '/dev/null';
print $fh "hello\n";
printf STDOUT "%s\n", 1;
