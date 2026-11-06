use experimental "isa";
$y = $x isa Foo;
no feature ":all";
$y = isa(1);
use feature qw(say isa);
$y = $x isa Foo;
use v5.10;
$y = ok(isa Foo("x"));
