#!perl
# C-style `for` is a `while`: it emits `enterloop`, not `enteriter`. "for"
# names two unrelated optrees and this is the one that is not `foreach`.
#
# TIER 06 control
# INTRODUCES the three-clause for loop
# USES nothing from a later tier
#
# The init clause runs once before `enterloop` and is an ordinary
# `padsv_store`; the increment clause is compiled into the tail of the body,
# ahead of the `unstack`, so from the op stream's point of view the three
# clauses are not three things. The only structural mark the C-style form
# leaves that `04_while.t` does not is the extra `unstack v*` between the
# init and `enterloop`.
#
# The counter is the loop's subject, so it is bound to `$ENV{N}` rather
# than a literal -- `for (my $i = 0; $i < 3; ...)` still emits the loop
# here, but the tier's own rule is that the condition is a runtime value,
# and a file that relies on the optimiser declining to fold is a file that
# measures the optimiser.
#
# MEASURED perl 5.42.0:
#
#   $ perl -MO=Concise,-exec -e 'my $n = $ENV{N} // 3;
#     for (my $i = 0; $i < $n; $i = $i + 1) { print $i }'
#   ...
#   9  <1> padsv_store[$i] vKS/LVINTRO
#   a  <0> unstack v*
#   b  <{> enterloop(next->f last->n redo->c) v
#   j  <0> padsv[$i] s
#   k  <0> padsv[$n] s
#   l  <2> lt sK/2
#   m  <|> and(other->c) vK/1
#   c      <0> pushmark s
#   d      <0> padsv[$i] s
#   e      <@> print vK
#   f      <0> padsv[$i] s
#   g      <$> const[IV 1] s
#   h      <2> add[$i] vK/TARGMY,2
#   i      <0> unstack v
#              goto j
#   n  <2> leaveloop vK/2
#
# `next` targets f -- the increment clause -- and not the `unstack`, which
# is why `next` in a C-style `for` still advances the counter and `next` in
# a `while` written the same way does not.

--- source
my $n = $ENV{N} // 3;
for (my $i = 0; $i < $n; $i = $i + 1) { print $i }
print "\n";

--- expect output
012

--- expect parses
