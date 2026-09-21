#!perl
# A signature names the parameters in the declaration: `sub f ($a, $b = 3)`
# binds them by position and supplies a default for the ones omitted.
#
# TIER 07 subroutines
# INTRODUCES subroutine signatures, with and without defaults
# USES nothing from a later tier
#
# A signature is the one construct in this tier that is NOT erased by the
# optimiser, and that is worth stating because it is the opposite of what
# the rest of the tier warns about. `sub f ($a, $b)` compiles to
# `argcheck(2,0)` followed by one `argelem` per parameter -- distinct ops
# carrying the declared arity, not ordinary pad assignments that happen to
# read `@_`. A default adds `argdefelem` guarding the default expression.
# Compare `sub f { my ($a,$b) = @_ }`, which is `padrange` and `aassign`
# and nothing else: the two spellings really are different optrees.
#
# All of which the main-program lint cannot see, because every one of
# those ops is inside the sub. That is the finding in the tier README.
#
# MEASURED perl 5.42.0, asking for the body by name:
#
#   $ perl -MO=Concise,-exec,f -e 'use v5.36; sub f ($a, $b = 3) { $a + $b }'
#   main::f:
#   1  <;> nextstate(main 2 -e:1) v:{
#   2  <0> argcheck(1,1,-) v
#   3  <0> argelem[$a:2,3] vK/SV
#   4  <|> argdefelem(other->5)[1] sKM/1
#   5      <$> const[IV 3] s
#   6  <1> argelem[$b:2,3] vKS/SV
#   ...
#
# `use v5.36` itself adds no ops. It changes the `nextstate` hint flags --
# `v:us,*,&,{,$,fea=6` instead of `v:{` -- and nothing more, so the
# feature pragma this file needs costs the op stream nothing and the tier
# claims nothing for it.
#
# MEASURED perl 5.42.0, the behaviour:
#
#   $ perl -e 'use v5.36; sub f ($a, $b = 3) { $a + $b } print f(1), " ", f(1,10), "\n"'
#   4 11

--- source
use v5.36;
sub add_up ($a, $b = 3) { $a + $b }
print add_up(1), " ", add_up(1, 10), "\n";

--- expect output
4 11

--- expect parses
