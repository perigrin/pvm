my $count = 42;  # Should infer Int
my $name = "Alice";  # Should infer Str
my $list = [1, 2, 3];  # Should infer ArrayRef[Int]
my $hash = {a => 1, b => 2};  # Should infer HashRef[Int]
