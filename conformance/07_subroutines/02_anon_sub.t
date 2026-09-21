#!perl
# An anonymous subroutine is a value: `sub { ... }` builds a closure in the
# enclosing scope and yields it, and `$c->(4)` calls it through the
# variable that holds it.
#
# TIER 07 subroutines
# INTRODUCES anonymous subroutine, and the call through a code value
# USES nothing from a later tier
#
# This is the one place where a subroutine construct DOES emit an op into
# the main program, and it is the reason `anoncode` is in this tier's
# INTRODUCES list at all. A named declaration vanishes at compile time; an
# anonymous one cannot, because the closure has to be built where it is
# written -- `anoncode` is that build, and it runs every time control
# reaches the expression.
#
# The call `$c->(4)` is the same `entersub` a named call uses, with the CV
# coming off the pad instead of out of a `gv`. It carries `/STRICT`
# because the invocant is a reference under `use strict`-era rules rather
# than a symbolic name.
#
# MEASURED perl 5.42.0:
#
#   $ perl -MO=Concise,-exec -e 'my $c = sub { $_[0] * 2 }; print $c->(4), "\n"'
#   1  <0> enter v
#   2  <;> nextstate(main 1 -e:1) v:{
#   3  <$> anoncode[CV CODE] sR
#   4  <1> padsv_store[$c:1,2] vKS/LVINTRO
#   5  <;> nextstate(main 2 -e:1) v:{
#   6  <0> pushmark s
#   7  <0> pushmark s
#   8  <$> const[IV 4] sM
#   9  <0> padsv[$c:1,2] s
#   a  <1> entersub[t3] lKS/TARG
#   b  <$> const[PV "\n"] s
#   c  <@> print vK
#   d  <@> leave[1 ref] vKP/REFC
#
# `anoncode[CV CODE]` is opaque on purpose: the body it wraps is a CV that
# B::Concise prints only when asked for it by name, and an anonymous sub
# has no name to ask with. The tier README records what that means for the
# lint.

--- source
my $triple = sub { $_[0] * 3 };
print $triple->(14), "\n";

--- expect output
42

--- expect parses
