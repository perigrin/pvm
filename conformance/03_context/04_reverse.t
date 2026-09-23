#!perl
# `reverse` answers two different questions depending on what asked. In
# list context it reverses the elements; in scalar context it concatenates
# them and reverses the characters.
#
# TIER 03 context
# INTRODUCES reverse
# USES nothing from a later tier
#
# This is a discriminating pair, and the pair is the point. The same op
# and the same operand appear twice: `my @b = reverse @a` emits
# `reverse[t6] lK/1` and `my $s = reverse @a` emits `reverse[t4] sK/1`.
# Same name, different `l`/`s` flag, unrelated answers.
#
# Note the scalar result is `321` and not `3 2 1`: reversing in scalar
# context first joins the list with nothing between the elements, then
# reverses the resulting string. A parser that treats the two as one
# operation on "a value" has no way to say which of these it means.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my @a=(1,2,3); my @b=reverse @a; my $s=reverse @a; print "@b $s\n"'
#   3 2 1 321

--- source
my @a = (1, 2, 3);
my @b = reverse @a;
my $s = reverse @a;
print "@b $s\n";

--- expect output
3 2 1 321

--- expect parses

--- expect tokens
one variable whose text is "@b"
one variable whose text is "$s"
