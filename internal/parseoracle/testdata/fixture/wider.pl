# ABOUTME: A hash read keyed by a call is rv2hv, unless the sub is a constant like `sub f(){1}`, which folds the key.
# ABOUTME: The wider bucket's fixture: we hedge rather than commit, because whether the key folds is not in the tree.
sub f { "a" }
my $r;
my $x = $r->{f()};
