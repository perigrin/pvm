my Int $count = 42;
my Str $name = "test";
my ArrayRef[Int] $numbers = [1, 2, 3, 4, 5];

sub add(Int $a, Int $b) -> Int {
    return $a + $b;
}

my $result = add($count, 10);
