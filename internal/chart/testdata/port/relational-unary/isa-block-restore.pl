use feature "isa";
my ($y, $x);
{
    no feature "isa";
    $y = ok(isa Foo("x"));
}
$y = $x isa Foo;
