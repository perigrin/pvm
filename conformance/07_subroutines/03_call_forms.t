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
# The semantic difference the flags stand for is real, and THIS FILE
# ONCE NAMED IT WITHOUT ASSERTING IT. `&f` with no parens passes the
# CALLER's `@_` through untouched rather than an empty list, so it is
# not merely a noisier `f()`. The earlier source proved none of that: it
# used `sub answer { 42 }`, which ignores its arguments, so all four
# spellings printed 42 and a parser that never implemented the
# forwarding passed. The header said the distinction was real and the
# pin said nothing.
#
# Two changes make it observable, and both are necessary. The callee
# REPORTS its arguments, and the calls happen INSIDE a sub whose own
# `@_` is non-empty -- at file scope there is no caller `@_` to forward
# and the four spellings would still agree.
#
# MEASURED perl 5.42.0, `outer(1, 2)` calling a callee that prints
# `"[@_]"`:
#
#   answer()    []
#   answer      []
#   &answer     [1 2]
#   &answer()   []
#
# THE THIRD LINE IS THE CLAIM. Only the parenless ampersand forwards.
#
# The FOURTH is the third reading and is why `&f()` is here rather than
# treated as a noisier spelling of `&f`: an EXPLICIT empty list passes
# nothing, so the ampersand alone does not cause forwarding -- the
# ABSENCE of an argument list does. A parser that read `&f` and `&f()`
# as the same call would print `[1 2]` twice.
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
sub answer { "[@_]" }
sub outer {
    print answer(), "\n";
    print answer, "\n";
    print &answer, "\n";
    print &answer(), "\n";
}
outer(1, 2);

--- expect output
[]
[]
[1 2]
[]

--- expect parses
