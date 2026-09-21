#!perl
# A named subroutine declared and then called: `sub f {...}` installs a CV
# under a package name, and `f(1)` calls it through that name.
#
# TIER 07 subroutines
# INTRODUCES named subroutine declaration and the call that reaches it
# USES nothing from a later tier
#
# The declaration itself emits NOTHING into the main program. `sub f {...}`
# is compile-time: the CV is built and installed, and the main optree that
# runs afterwards contains only the call. So this file's measurement is the
# call site and nothing else -- `pushmark`, the `gv` naming the sub, and
# `entersub`. The body, which is where this tier's subject actually lives,
# is a separate optree that `-MO=Concise,-exec` never prints. See the
# tier README's final section for what that costs the lint.
#
# MEASURED perl 5.42.0:
#
#   $ perl -MO=Concise,-exec -e 'sub f { $_[0] + 1 } print f(1), "\n"'
#   1  <0> enter v
#   2  <;> nextstate(main 2 -e:1) v:{
#   3  <0> pushmark s
#   4  <0> pushmark s
#   5  <$> const[IV 1] sM
#   6  <#> gv[IV \&main::f] s
#   7  <1> entersub lKS
#   8  <$> const[PV "\n"] s
#   9  <@> print vK
#   a  <@> leave[1 ref] vKP/REFC
#
# `gv[IV \&main::f]` rather than `gv[*f]`: the call resolved to the CV at
# compile time, because the sub was already declared when the call was
# compiled. A call compiled BEFORE the declaration gets `gv[*f]` and looks
# the name up at run time. Same op either way.

--- source
sub double { $_[0] * 2 }
print double(21), "\n";

--- expect output
42

--- expect parses
