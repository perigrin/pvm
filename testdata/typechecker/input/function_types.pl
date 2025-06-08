sub add(Int $a, Int $b) -> Int {
    return $a + $b;
}

sub greet(Str $name) -> Str {
    return "Hello, $name!";
}

my $result = add(1, 2);  # Should be Int
my $message = greet("Alice");  # Should be Str
