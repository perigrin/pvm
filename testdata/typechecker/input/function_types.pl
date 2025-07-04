sub Int add(Int $a, Int $b) {
    return $a + $b;
}

sub Str greet(Str $name) {
    return "Hello, $name!";
}

my $result = add(1, 2);  # Should be Int
my $message = greet("Alice");  # Should be Str
