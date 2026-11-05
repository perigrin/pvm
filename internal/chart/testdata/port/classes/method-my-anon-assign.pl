use feature 'class';
no warnings 'experimental::class';
class Foo { method m { my $x = my method { 1 }; } }
