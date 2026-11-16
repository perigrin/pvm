my @a;
{
  Foo::bar { a => 1 }; Foo::bar { a => 1 }; Foo::bar { a => 1 };
  Foo::bar { a => 1 }; Foo::bar { a => 1 }; Foo::bar { a => 1 };
  my $r = \@a;
}
my $s = \@a;
