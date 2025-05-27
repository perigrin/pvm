sub greet {
    my ($name) = @_;
    return "Hello, $name!";
}

sub add(Int $a, Int $b) -> Int {
    return $a + $b;
}
