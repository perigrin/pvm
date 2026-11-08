sub first(&@);
first { $_ } 1, 2;
$y = first { $_ } @a;
