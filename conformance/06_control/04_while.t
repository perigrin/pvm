#!perl
# A block `while` emits `enterloop`, tests with tier 04's `and`, closes each
# pass with `unstack` and exits through `leaveloop`.
#
# TIER 06 control
# INTRODUCES the while loop
# USES nothing from a later tier
#
# `enterloop` and `leaveloop` are tier 05's, claimed there for the bare
# block -- a bare `{ ... }` is a loop that runs once. What this tier adds is
# `unstack`, which the bare block does NOT emit, and that is the op that
# distinguishes a loop that iterates from a block that does not.
#
# `enterloop` names all three jump targets in its own dump, which is why
# `next`, `last` and `redo` are claimed in this tier rather than deferred:
# the loop ops and the jump ops are one measurement.
#
# MEASURED perl 5.42.0:
#
#   $ perl -MO=Concise,-exec -e 'my $n = $ENV{N} // 3; my $i = 0;
#     while ($i < $n) { print $i; $i = $i + 1 }'
#   ...
#   b  <{> enterloop(next->k last->p redo->c) v
#   l  <0> padsv[$i] s
#   m  <0> padsv[$n] s
#   n  <2> lt sK/2
#   o  <|> and(other->c) vK/1
#   c      <;> nextstate(main 5 -:1) v
#   ...
#   k      <0> unstack v
#              goto l
#   p  <2> leaveloop vKP/2
#
# `while (0) { print "y" }` is the erasure worth recording here: it emits
# `enter`, `nextstate` and `leave` -- an empty program, no loop ops at all.
# Hence the runtime bound.

--- source
my $n = $ENV{N} // 3;
my $i = 0;
while ($i < $n) { print $i; $i = $i + 1 }
print "\n";

--- expect output
012

--- expect parses
