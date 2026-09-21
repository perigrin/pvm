#!perl
# `@a[0, 2]` is a slice: the `@` sigil on an array name with a LIST
# subscript, which returns several elements rather than one.
#
# TIER 02 variables
# INTRODUCES array slice
# USES nothing from a later tier
#
# The sigil is the whole point. `$a[0]` and `@a[0]` name the same array and
# differ only in what they hand back, so a lexer that binds the sigil to
# the name and a parser that binds it to the subscript form disagree here
# and nowhere else in the tier.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my @a = (1, 2, 3); print( (@a[0, 2]), "\n")'
#   13
#
# `13`, not `1 3`: the slice's two elements arrive as separate arguments to
# `print` and `$,` is unset, so nothing separates them -- the same reason
# tier 01's adjacency file prints `qw(a b c)` as `abc`.

--- source
my @a = (1, 2, 3);
print( (@a[0, 2]), "\n");

--- expect output
13

--- expect parses
