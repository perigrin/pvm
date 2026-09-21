#!perl
# `"@a"` contains no function call, and compiles to one: interpolating an
# array into a string IS a join, decided at compile time.
#
# TIER 03 context
# INTRODUCES interpolation context on an array
# USES nothing from a later tier
#
# No reader of the source would predict `join`, which is the whole point
# of the file. `"@a"` emits `gvsv[*"]` -- fetching the list separator `$"`
# -- and then `join[t5] sK/2`. Interpolation is a third context after
# scalar and list in everything but name, and it is visible only in the
# optree.
#
# The `gvsv` is tier 02's, claimed there with the package scalar; only
# `join` is new here. That is why this file can sit in 03 at all: it leans
# on an earlier tier's op rather than smuggling one in.
#
# The separator is a single space by default, so the three elements print
# as `x y z` -- unlike a bare `print @a` list, which has nothing between
# its arguments. That difference is the observable half of what the optree
# shows.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my @a=("x","y","z"); my $s="@a"; print "$s\n"'
#   x y z

--- source
my @a = ('x', 'y', 'z');
my $s = "@a";
print "$s\n";

--- expect output
x y z

--- expect parses

--- expect tokens
one variable whose text is "@a"
