# ABOUTME: A \@ prototype changes the parse at the call site invisibly.
# ABOUTME: The wider bucket's fixture: we hedge rather than commit.
sub f (\@) { }
my @a;
f(@a);
