#!perl
# An array is named with `@` and holds a list; `scalar(@a)` gives back how
# many elements it holds.
#
# TIER 02 variables
# INTRODUCES array variable
# USES nothing from a later tier
#
# This is the tier's baseline, the way `02_decimal.t` is tier 01's: every
# other file here starts by filling an array or a hash, so a regression in
# this file explains all of them at once rather than being diagnosed nine
# times.
#
# `scalar(@a)` rather than `print "@a"`: the interpolation emits `join` and
# a `gvsv` for `$"`, neither of which is naming a variable, and both of
# which this tier would then have to claim.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my @a = (1, 2, 3); print scalar(@a), "\n"'
#   3
#
# The ops are `padav` for the array itself and `aassign` for the list
# assignment. There is no `scalar` op: `scalar(@a)` compiles to the padav
# in scalar context, so the keyword leaves no trace of its own.

--- source
my @a = (1, 2, 3);
print scalar(@a), "\n";

--- expect output
3

--- expect parses
