#!perl
# A constant subscript on a lexical array reads as one op and writes as
# another: `$a[0]` on the right of `=` is not the same op as `$a[0]` on
# the left.
#
# TIER 02 variables
# INTRODUCES array element with a constant subscript
# USES nothing from a later tier
#
# This is the asymmetry the tier README calls out. Read and write are the
# same three characters in the source, and a parser that treats them as one
# node is not wrong -- but perl does not, and a corpus that never writes an
# element cannot see the second op at all.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my @a = (1, 2, 3); $a[0] = 9; print $a[0], "\n"'
#   9
#
# The write is `aelemfastlex_store` and the read is `aelemfast_lex`. Both
# are the `_lex` spellings because `@a` is a lexical; the package spellings
# are in `11_package_array.t`.
#
# `print $a[0]` rather than `print "$a[0]"`: the interpolated form emits
# `stringify`, which belongs to what interpolation does with a value rather
# than to subscripting one.

--- source
my @a = (1, 2, 3);
$a[0] = 9;
print $a[0], "\n";

--- expect output
9

--- expect parses
