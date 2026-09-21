#!perl
# `=~` emits NO OP OF ITS OWN. The binding operator is absorbed into the
# match op, which carries its target in its own flags -- the stream shows
# `match(/"b"/)[$s:1,3]` and no `bind` anywhere.
#
# TIER 09 regex
# INTRODUCES the `=~` binding operator
# USES nothing from a later tier
#
# So this file and `01_bare_match.t` differ in their OPERAND, not in their
# op set. That is why the binding needs a file of its own: the optree
# cannot tell you the operator was written, only that the match found a
# target other than $_, and the token stream is what records the operator.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $s = "abc"; my $m = $s =~ /b/; print "$m\n"'
#   1

--- source
my $s = "abc";
my $m = $s =~ /b/;
print "$m\n";

--- expect output
1

--- expect parses

--- expect tokens
one operator whose text is "=~"
no operator whose text is "/"
