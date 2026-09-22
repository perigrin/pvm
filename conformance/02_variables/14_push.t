#!perl
# `push`'s first argument is an ARRAY, not an expression.
#
# TIER 02 variables
# INTRODUCES the aggregate-argument operator
# USES nothing from a later tier
#
# Every other list operator in the corpus takes expressions and flattens
# them. `push` does not: its first argument slot holds the array ITSELF,
# and only the arguments after the comma are flattened into it. A parser
# that treats `push @a, @tail` as `push(@a, @tail)` with both arrays
# flattened builds a tree perl does not build, while producing output
# plausible enough that nothing behavioural would notice.
#
# MEASURED perl 5.42.0. The two arrays sit adjacent, separated by one
# comma, and compile to the SAME OP with DIFFERENT FLAGS:
#
#   $ perl -MO=Concise,-exec -e 'my @a = (10); my @tail = (20, 30);
#         push @a, @tail;'
#   ... <0> padav[@a:1,3] lRM
#       <0> padav[@tail:2,3] l
#       <@> push[t5] vK/2
#
# `lRM` is the aggregate slot -- lvalue, ref-modify, the array as a
# container. `l` is the flattened slot. The flags are the whole
# distinction, and the file pins THREE consequences of them rather than
# one, because each alone is reachable by a wrong tree:
#
#   $ perl conformance/02_variables/14_push.t
#   2
#   3
#   30
#
# `@tail` still holds TWO elements. The flattened slot is read, not
# consumed, so a parser that moved the elements rather than copying them
# fails here and nowhere else. `@a` holds THREE, which is the count a
# flattening parser would also reach -- so the count alone proves
# nothing, and the file prints it only to make the third line legible.
# `$a[2]` is 30, the LAST element of `@tail`, which pins the order the
# flattening put them in.
#
# `push` emits `push`, an op no earlier tier claims and which this tier's
# README now does. There is no constant-folding trap here, and the tier's
# `$ENV{X}` idiom would be wrong to copy: measured, `push @a, 3` with
# everything constant still emits `push`, because an array is a runtime
# container and the optimiser has nothing to fold it into. (`//` is
# tier 04's `dor`, so the idiom is out of budget here regardless.)
#
# The token facts COUNT, and the counting is the falsifying half.
# Exactly ONE word spelled `push` and NO word spelled `unshift`: the two
# operators share a signature, an argument rule and an op shape, and
# differ only in which end they write, so a keyword table that mapped one
# onto the other would spell it the other way here and be caught before
# the optree was consulted. `15_unshift.t` carries the mirror of this
# pair.

--- source
my @a = (10);
my @tail = (20, 30);
push @a, @tail;
print scalar(@tail), "\n";
print scalar(@a), "\n";
print $a[2], "\n";

--- expect output
2
3
30

--- expect parses

--- expect tokens
one word whose text is "push"
no word whose text is "unshift"
