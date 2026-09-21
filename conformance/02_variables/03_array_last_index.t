#!perl
# `$#a` is the last index of `@a`, and using it as a subscript is what
# stops the optimiser folding the element access away.
#
# TIER 02 variables
# INTRODUCES last-index variable
# USES nothing from a later tier
#
# Two claims in one body because the second only exists in terms of the
# first. `$#a` alone is `av2arylen`. `$a[$#a]` is the case where the
# subscript is an expression rather than a constant, so perl builds a plain
# `aelem` instead of the `aelemfast_lex` of `02_array_element.t` -- which
# makes `aelem`, the op a reader would expect to be this tier's centre, an
# edge case reachable only by writing the subscript this way.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my @a = (1, 2, 3); print $#a, "\n"; print $a[$#a], "\n"'
#   2
#   3

--- source
my @a = (1, 2, 3);
print $#a, "\n";
print $a[$#a], "\n";

--- expect output
2
3

--- expect parses
