my $x = 'foo';
print 1 if $x =~ /o/;
print 2 if $x =~ $x;
$x =~ s/o/0/;
print 3 if /f/;
my $y = 4 / 2;
