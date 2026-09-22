#!perl
# `values` takes an AGGREGATE where a list operator takes an expression.
#
# TIER 02 variables
# INTRODUCES the aggregate-argument operator over a hash
# USES nothing from a later tier
#
# `values %h` is not `values(%h)` with the hash flattened into a list of
# key-value pairs. The hash is the argument, whole, and the operator
# reads its values out of the container. A parser that flattens `%h`
# first hands `values` a flat list of four scalars where perl hands it
# one hash -- and for a one-key hash the printed answer would still look
# right, which is why this file pins the op and a token fact rather than
# only the output.
#
# MEASURED perl 5.42.0. The hash reaches the operator as `padhv`, the
# container op, not as an aassign'd list:
#
#   $ perl -MO=Concise,-exec -e 'my %h = (a => 7); my @v = values %h;'
#   ... <0> padhv[%h:1,3] lRM
#       <1> values[t4] lK/1
#
# `values` IS AN OP, unlike `keys`, and that asymmetry is worth recording
# because this tier's README documents the `keys` half and a reader would
# generalise it wrongly. Measured, `scalar(keys %h)` compiles `keys` away
# to the flag `sM/KEYS` on the padhv -- the README's claim, and true in
# SCALAR context. In LIST context `keys %h` emits a `keys` op of its own,
# exactly as `values` does here. So the flag is a property of the
# context, not of the keyword, and this file writes `values` in list
# context and no `keys` at all so the op budget covers what it emits.
#
# ONE KEY, DELIBERATELY. Hash order is not guaranteed, so a file printing
# the values of a two-key hash would pin an expectation perl does not
# promise -- the reason `07_hash.t` prints `scalar(keys %h)` rather than
# the contents. With one key the list has one member and its value is
# determined:
#
#   $ perl conformance/02_variables/16_values.t
#   1
#   7
#
# The token facts count. Exactly ONE word spelled `values`, and NO word
# spelled `keys`: the second is the falsifying half, because a lexer or a
# keyword table that mapped one aggregate-argument operator onto the
# other -- they share a shape, a signature and an argument rule -- would
# produce a `keys` here and be caught before the optree was consulted.

--- source
my %h = (a => 7);
my @v = values %h;
print scalar(@v), "\n";
print $v[0], "\n";

--- expect output
1
7

--- expect parses

--- expect tokens
one word whose text is "values"
no word whose text is "keys"
