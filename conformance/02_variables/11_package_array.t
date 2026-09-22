#!perl
# The same constant subscript on a PACKAGE array is a different op from the
# one on a lexical array.
#
# TIER 02 variables
# INTRODUCES package array
# USES nothing from a later tier
#
# `$a[0]` on a lexical is `aelemfast_lex` (`02_array_element.t`). `$::a[0]`
# is `aelemfast` behind an `rv2av` over a `gv`. Constant subscript,
# lexical or package, read or write -- four ops for what reads as one
# construct, which is why the tier's adjacency file carries all four rather
# than a representative one.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e '@::a = (1, 2, 3); print $::a[0], "\n"; print scalar(@::a), "\n"'
#   1
#   3

--- source
@::a = (1, 2, 3);
print $::a[0], "\n";
print scalar(@::a), "\n";

--- expect output
1
3

--- expect parses
