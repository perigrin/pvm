#!perl
# `\@a` in scalar context is one op, `srefgen`, and the array it points at
# stays reachable through the scalar that holds it.
#
# TIER 08 references
# INTRODUCES the reference operator in scalar context
# USES my, print, array element access
#
# Take a reference to a VARIABLE, not a literal. `my $r = \1` arrives as
# `const[IV \1] s/FOLD` and emits no `srefgen` at all: the optimiser
# erases the very construct this file is about. The array fixture is what
# defeats that, and it is the same lesson tier 01 learned from `1+2`.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my @a = (10, 20); my $s = \@a; print "$$s[0] $$s[1]\n"'
#   10 20
#
# The optree is `padav[@a] lRM`, then `srefgen sK/1`, then
# `padsv_store[$s]` -- the array is loaded first and the reference taken
# of it, which is why this tier cannot precede tier 02.

--- source
my @a = (10, 20);
my $s = \@a;
print "$$s[0] $$s[1]\n";

--- expect output
10 20

--- expect parses

--- expect tokens
one operator whose text is "\\"
no operator whose text is "->"
