use feature 'state';
my sub bar { 1 }
bar();
$y = bar 1, 2;
state sub baz;
my sub qux;
$y = &qux;
