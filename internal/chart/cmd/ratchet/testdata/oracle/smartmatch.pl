use feature 'switch'; no warnings;
my (@a, %h, $x, $r);
my $m = ($x ~~ @a);
$m = (%h ~~ $x);
$m = (@a ~~ %h);
$m = ($x ~~ $r->@*);
$m = ($x ~~ [@a]);
given ($x) { when (@a) { 1 } }
for ($x) { 1 when %h }
