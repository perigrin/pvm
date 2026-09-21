#!perl
# `foreach` emits `enteriter`/`iter` and NO `enterloop` at all. It is a
# different optree from the C-style `for` that shares its keyword.
#
# TIER 06 control
# INTRODUCES the foreach loop
# USES nothing from a later tier
#
# `enteriter` does what `enterloop` does -- it names the same three jump
# targets in its own dump -- and additionally allocates the loop variable's
# pad slot with `LVINTRO`, which is why this tier cannot sit before 05. The
# per-pass `iter` is the op that has no analogue in the `while` family: it
# advances the list cursor and pushes the truth of "there was another one"
# for the `and` to test, so the loop's condition is not written anywhere in
# the source.
#
# Both families close with `leaveloop`.
#
# MEASURED perl 5.42.0:
#
#   $ perl -MO=Concise,-exec -e 'my @l = ($ENV{A} // "a", "b");
#     foreach my $x (@l) { print $x }'
#   ...
#   c  <0> pushmark sM
#   d  <0> padav[@l] sRM
#   e  <{> enteriter(next->j last->m redo->f)[$x:3,6] vKS/LVINTRO
#   k  <0> iter s
#   l  <|> and(other->f) vK/1
#   f      <;> nextstate(main 5 -:1) v
#   g      <0> pushmark s
#   h      <0> padsv[$x] s
#   i      <@> print vK
#   j      <0> unstack v
#              goto k
#   m  <2> leaveloop vK/2
#
# The list is built from `$ENV{A}` for the same reason every other file
# here binds a runtime value, though the erasure is milder: `foreach my $x
# (1, 2)` does keep its loop ops. The rule is uniform so that no file in
# the tier has to be argued about individually.

--- source
my @l = ($ENV{A} // "a", "b", "c");
foreach my $x (@l) { print $x }
print "\n";

--- expect output
abc

--- expect parses
