#!perl
# `..` in SCALAR context is a stateful FLIP-FLOP, not a list: it
# remembers whether it is on, and selects elements whose own operands are
# both false.
#
# TIER 03 context
# INTRODUCES nothing of its own
# USES the range operator from 11_range_list.t
#
# The third of `..`'s readings and the one that makes the operator worth
# three files. `11_range_list.t` and `12_range_string.t` build lists;
# this one builds nothing and carries state between evaluations.
#
# WHICH READING APPLIES IS DECIDED BY CONTEXT, which is this tier's whole
# subject. The same two characters, the same operand shape, and a
# different operator -- chosen by whether the surrounding expression
# wants a list or a scalar.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my @on=(0,1,0,0,0,0); my @off=(0,0,0,0,1,0);
#              my @i=(0,1,2,3,4,5);
#              my @o = grep { $on[$_] .. $off[$_] } @i; print "@o"'
#   1 2 3 4
#
# INDICES 2 AND 3 ARE THE CLAIM. At both of them `$on[$_]` is 0 and
# `$off[$_]` is 0 -- both operands false -- and the flip-flop selects
# them anyway, because it turned on at index 1 and does not turn off
# until index 4. No truthiness test can produce that: a parser that read
# `..` here as a list range, or as a boolean `or`, or as anything
# stateless, selects only indices 1 and 4.
#
# That is why the operands are ARRAY ELEMENTS rather than the `$_ .. $_`
# a first draft used. Measured, `grep { $_ .. $_ }` over `(0,1,0,0,1,0)`
# prints `1 1` -- and so does `grep { $_ }` over the same list. The
# stateful reading and plain truthiness agree, so the file would have
# passed against a parser that had never heard of a flip-flop. Two
# separate operand arrays are what make the state observable.
#
# The `grep` form is deliberate: the loop spelling of the same claim
# needs `enteriter` and `leaveloop`, which are tiers 05 and 06, while
# `grepstart` and `grepwhile` are this tier's own.

--- source
my @on = (0, 1, 0, 0, 0, 0);
my @off = (0, 0, 0, 0, 1, 0);
my @i = (0, 1, 2, 3, 4, 5);
my @o = grep { $on[$_] .. $off[$_] } @i;
print "@o\n";

--- expect output
1 2 3 4

--- expect parses

--- expect tokens
one operator whose text is ".."
