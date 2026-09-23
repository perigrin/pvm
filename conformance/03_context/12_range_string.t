#!perl
# A STRING range counts by perl's magic increment, not by arithmetic:
# `"az" .. "bb"` is three elements, and none of them is a number.
#
# TIER 03 context
# INTRODUCES nothing of its own
# USES the range operator from 11_range_list.t
#
# The second of `..`'s three readings. It looks like the first and is a
# separate claim, because the SUCCESSOR function is different: a numeric
# range adds one, a string range applies the magic increment that also
# drives `$s++` on a string.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my @r = ("az" .. "bb"); print "@r"'
#   az ba bb
#
# `az` to `ba` is the carry: the last character wraps from `z` to `a` and
# the one before it advances. A parser that read this as an arithmetic
# range over numified endpoints would get `0 .. 0` and print a single
# zero, since `"az"` numifies to 0.
#
# THE FOLD APPLIES HERE TOO, and for the same placement reason
# `11_range_list.t` records. Measured, constant string endpoints fold
# exactly as numeric ones do:
#
#   my @r = ("a" .. "c")            no range op
#   my @r = ($a[0] .. "c")          range flip flop
#
# The fold is an optimiser effect on the op stream and does not touch
# lexing or parsing -- `("a" .. "c")` parses identically either way. The
# array element is here so the tier's `range` declaration has something
# to check, not because the constant form would parse differently.
#
# The COUNT is the second half of the claim. `scalar(@r)` is 3, and it is
# this tier's own scalar-context idiom applied to the result -- a range
# in list context producing a list whose length is then taken in scalar
# context, which is the tier's subject twice over.

--- source
my @seed = ("az");
my @r = ($seed[0] .. "bb");
print "@r ", scalar(@r), "\n";

--- expect output
az ba bb 3

--- expect parses

--- expect tokens
one operator whose text is ".."
