#!perl
# An explicit `return` leaves the sub with a value, and an early `return`
# inside a branch leaves it before the last statement runs.
#
# TIER 07 subroutines
# INTRODUCES explicit return, including early return from a branch
# USES nothing from a later tier
#
# The source construct `return` and the op `return` are different things,
# and this file is where the corpus records the gap. A `return` that IS
# the last thing the body does compiles to nothing: the value is already
# on the stack and `leavesub` takes it, so perl deletes the op. Only a
# return that jumps -- one guarded by a branch, with statements after it
# -- survives as an op.
#
# That is why this tier declares 06_control as its prerequisite. Without a
# branch there is no early exit, without an early exit there is no `return`
# op, and the construct the tier is named for would be unobservable in any
# optree the corpus could produce.
#
# MEASURED perl 5.42.0. The trailing return, erased:
#
#   $ perl -MO=Concise,-exec,f -e 'sub f { return $_[0] + 1 }'
#   main::f:
#   1  <;> nextstate(main 1 -e:1) v:{
#   2  <$> aelemfast_lex_or_rv2av[*_] s
#   3  <$> const[IV 1] s
#   4  <2> add[t4] sK/2
#   5  <1> leavesub[1 ref] K/REFC
#
# The guarded return, surviving:
#
#   $ perl -MO=Concise,-exec,g -e 'sub g { return 1 if $_[0]; 0 }'
#   main::g:
#   1  <;> nextstate(main 1 -e:1) v:{
#   ...
#   5  <|> and(other->6) vK/1
#   6      <0> pushmark s
#   7      <$> const[IV 1] s
#   8      <@> return K
#   ...
#
# Both of those live in the sub's own optree, which the main-program
# measurement does not reach -- see the tier README's last section.
#
# The behaviour is what this file asserts, and it distinguishes the two:
# `classify(0)` prints "zero" only if the early return fired, and
# "positive" only if it did not.

--- source
sub classify {
    return "zero" if $_[0] == 0;
    return "negative" if $_[0] < 0;
    return "positive";
}
print classify(0), " ", classify(-5), " ", classify(7), "\n";

--- expect output
zero negative positive

--- expect parses
