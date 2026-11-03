use feature 'keyword_all';
no warnings 'experimental::keyword_all';
$r = all { $_ > 1 } @a, @b;
$r = all { $_ } @a;
