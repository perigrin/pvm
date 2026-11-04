use feature 'signatures';
no warnings 'experimental::signature_named_parameters';
sub f ($x, :$n, :$m = 2) { 1 }
