#!perl
# The SAME backslash in list context is a DIFFERENT op: `\(@a)` emits
# `refgen`, not `srefgen`, and distributes over the array's elements.
#
# TIER 08 references
# INTRODUCES the reference operator in list context
# USES my, print, array element access, brace dereference
#
# This file and `01_backslash_scalar.t` differ in source only by the
# parentheses and the assignment target, and perl compiles them to two
# different ops. That is why the tier claims both: one spelling in the
# source is two ops in the optree, and a corpus that wrote only the
# scalar form would claim an op it never exercised.
#
# The distribution is the second surprise. `\(@a)` is not a reference to
# the array -- it is a LIST of references, one per element, so `$r[0]` is
# a SCALAR reference and `${$r[0]}` is the element behind it. Writing
# `$r[0]->[0]` instead dies with "Not an ARRAY reference", which is how
# this file arrived at its current form.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my @a = (10, 20); my @r = \(@a); print ${$r[0]}, " ", ${$r[1]}, "\n"'
#   10 20

--- source
my @a = (10, 20);
my @r = \(@a);
print ${$r[0]}, " ", ${$r[1]}, "\n";

--- expect output
10 20

--- expect parses

--- expect tokens
one operator whose text is "\\"
