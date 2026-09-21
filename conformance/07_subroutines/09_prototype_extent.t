#!perl
# The same parenless call site, under a callee with a `($)` prototype:
# the extent is ONE argument, and the second falls through to the
# enclosing `print`.
#
# TIER 07 subroutines
# INTRODUCES the prototype's effect on parenless argument extent
# USES print, @_, return, string interpolation
# MEASURED perl 5.42.0
# STATUS refuses as of dc1bea2c. Issue 01a0c432-fbd5. Refusal trailing_tokens.
#
# THIS FILE IS HALF OF A MEASUREMENT. `08_parenless_extent.t` is the
# other half, and the point is the difference between them. Set the two
# call sites side by side:
#
#   sub f     { return "f[@_]" }   print f 1, 2;   # f[1 2]
#   sub g ($) { return "g[@_]" }   print g 1, 2;   # g[1]2
#
# The call sites differ only in the callee's name. Everything that
# decides how far `1, 2` runs is in the DECLARATION -- its prototype,
# and indeed whether there is a declaration at all, since an undeclared
# callee makes the whole statement a syntax error. This is why a parser
# cannot answer the extent question from the call site's tokens, and why
# the corpus measures it rather than assuming it.
#
# MEASURED perl 5.42.0, this file:
#
#   $ perl -MO=Concise,-exec conformance/07_subroutines/09_prototype_extent.t
#   3  <0> pushmark s
#   4  <0> pushmark s
#   5  <$> const[IV 1] sM
#   6  <#> gv[IV \"$"] s
#   7  <1> entersub lKS
#   8  <$> const[IV 2] s
#   9  <@> print vK
#
# `const[IV 2]` is AFTER the `entersub`, so it is print's argument and
# not g's. In `08_parenless_extent.t` both constants precede the
# `entersub`. The optree says the same thing the output says.
#
# Note `gv[IV \"$"]` rather than `gv[IV \&main::g]`: with a prototype in
# force perl has resolved the call differently again. Still one
# `entersub`, which is why the op lint cannot tell these two files apart
# and the outputs have to.
#
# WHY THIS REFUSES. The same `trailing_tokens` as its pair, at the same
# place and for the same reason: our parser reads `g` as a complete term
# and finds `1` next with no operator between them. The prototype is not
# what it stumbles on -- it has not got as far as caring.
#
# The token facts assert that `($)` is NOT three punctuation tokens. A
# lexer that read it as `(`, `$`, `)` would produce a stream in which
# this file looks like a parenthesised call, which is the other call
# form entirely.

--- source
sub g ($) { return "g[@_]" }
print g 1, 2;
print "\n";

--- expect output
g[1]2

--- expect tokens
no operator whose text is "("
no operator whose text is ")"
one operator whose text is ","
one numeric literal whose text is "2"

--- expect parses
