#!perl
# `each` is an aggregate-argument operator that returns a LIST of two.
#
# TIER 02 variables
# INTRODUCES the two-element list return
# USES nothing from a later tier
#
# `values` and `keys` take the same argument and return a flat list of
# one thing per entry. `each` takes the same argument and returns a PAIR
# -- one key and one value -- so the list-assignment on its left has two
# scalars to fill. A parser that models `each` on its siblings gives the
# construct the wrong arity and the pair silently collapses to the key.
#
# MEASURED perl 5.42.0. The hash arrives as `padhv`, the container op,
# and `each` is a real op, not a flag:
#
#   $ perl -MO=Concise,-exec -e 'my %h = (a => 7); my ($k, $v) = each %h;'
#   ... <0> padhv[%h:1,3] lRM
#       <1> each lK/1
#       <0> padrange[$k:2,3; $v:2,3] RM/LVINTRO,range=2
#       <2> aassign[t5] vKS/COM_RC1
#
# The `range=2` on the padrange is the arity, visible: two pad slots are
# introduced as one range because the assignment's left side is a
# two-element list. A one-element return would have had a padsv there.
#
# NO LOOP, WHICH IS THE POINT OF THE SPELLING. T1 writes `while (my ($k,
# $v) = each %h)`, and measured, that form drags in `enterloop`,
# `leaveloop`, `and` and `unstack` -- ops belonging to tiers 05, 04 and
# 06. A bare list assignment reaches the same `each` with nothing this
# tier does not claim, and the iterator's LOOPING is a control-flow claim
# rather than a claim about how `each` parses. That separation is what
# keeps the file in budget honestly rather than by a trick.
#
# ONE KEY, for `16_values.t`'s reason: hash order is not guaranteed, so a
# two-key hash would make the pair unpredictable and the pinned output a
# guess. With one key the first iteration is determined:
#
#   $ perl conformance/02_variables/06_each.t
#   a
#   7
#
# The token facts count. Exactly ONE word spelled `each`, and NO word
# spelled `values`: the second falsifies the collapse this file exists to
# catch, since a keyword table that mapped the pair-returning operator
# onto the flat-list one would spell it the other way here.

--- source
my %h = (a => 7);
my ($k, $v) = each %h;
print $k, "\n";
print $v, "\n";

--- expect output
a
7

--- expect parses

--- expect tokens
one word whose text is "each"
no word whose text is "values"
