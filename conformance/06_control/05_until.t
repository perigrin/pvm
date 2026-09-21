#!perl
# `while` and `until` differ by the same one op as `if` and `unless`: the
# test is `and` for `while` and `or` for `until`, and everything else is
# identical.
#
# TIER 06 control
# INTRODUCES the until loop
# USES nothing from a later tier
#
# So the loop family and the conditional family share their branch ops
# entirely. What makes a loop a loop is `enterloop`/`unstack`, not the
# test -- an `until` is a `while` with the other short-circuit, exactly as
# an `unless` is an `if` with the other short-circuit. There is no
# negation op in either pair.
#
# The condition is spelled `$i >= $n` rather than `!($i < $n)` for the same
# reason `02_unless.t` does not spell its condition `!$c`: a `not` op would
# be tier 04's, would appear in the stream, and would make this file
# measure a different construct from the one it names.
#
# MEASURED perl 5.42.0:
#
#   $ perl -MO=Concise,-exec -e 'my $n = $ENV{N} // 3; my $i = 0;
#     until ($i >= $n) { print $i; $i = $i + 1 }'
#   ...
#   b  <{> enterloop(next->k last->p redo->c) v
#   l  <0> padsv[$i] s
#   m  <0> padsv[$n] s
#   n  <2> ge sK/2
#   o  <|> or(other->c) vK/1
#   ...
#   k      <0> unstack v
#              goto l
#   p  <2> leaveloop vKP/2
#
# Position o is the only difference from `04_while.t`'s stream, and `ge`
# against `lt` at n is the source's own inversion, not perl's.

--- source
my $n = $ENV{N} // 3;
my $i = 0;
until ($i >= $n) { print $i; $i = $i + 1 }
print "\n";

--- expect output
012

--- expect parses
