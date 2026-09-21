#!perl
# Four ways to spell a call to the same sub -- `f(...)`, bare `f`, `&f`
# and `&f(...)` -- all compile to the same op. The differences are flags.
#
# TIER 07 subroutines
# INTRODUCES the call forms: parenthesised, bare, and ampersand
# USES nothing from a later tier
#
# This file exists because the op stream cannot tell the four apart, which
# is exactly why they need separate source rather than separate op claims.
# `entersub` appears four times below and the only thing distinguishing
# the ampersand calls is `/AMPER` in the flags column -- a lint that reads
# op NAMES, as this corpus's does, sees one construct where the language
# has four.
#
# The semantic difference the flags stand for is real. `&f` with no
# parens passes the CALLER's `@_` through untouched rather than an empty
# list, so it is not merely a noisier `f()`. Here `answer` ignores its
# arguments, which keeps the printed output the same across all four and
# isolates the call syntax as the only thing varying.
#
# MEASURED perl 5.42.0:
#
#   $ perl -MO=Concise,-exec -e 'sub f { 7 } print f, "\n"; print &f, "\n"'
#   1  <0> enter v
#   2  <;> nextstate(main 3 -e:1) v:{
#   3  <0> pushmark s
#   4  <0> pushmark s
#   5  <#> gv[*f] s
#   6  <1> entersub lKS
#   7  <$> const[PV "\n"] s
#   8  <@> print vK
#   9  <;> nextstate(main 3 -e:1) v:{
#   a  <0> pushmark s
#   b  <0> pushmark s
#   c  <#> gv[*f] s
#   d  <1> entersub[t3] lK/AMPER,TARG
#   e  <$> const[PV "\n"] s
#   f  <@> print vK
#   g  <@> leave[1 ref] vKP/REFC
#
# Note `gv[*f]` here where 01_named_sub.t measured `gv[IV \&main::f]`: a
# bare `f` with no parens is compiled as a glob lookup rather than
# resolved to the CV, because the parser had to decide whether `f` was a
# call at all. Still one `entersub`.

--- source
sub answer { 42 }
print answer(), "\n";
print answer, "\n";
print &answer, "\n";
print &answer(), "\n";

--- expect output
42
42
42
42

--- expect parses
