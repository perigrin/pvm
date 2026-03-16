use strict;
use warnings;

my @numbers = (1, 2, 3);
my $count = push(@numbers, 4);

my %lookup = (foo => 1, bar => 2);
my @keys = keys(%lookup);
my $exists = exists($lookup{foo});
