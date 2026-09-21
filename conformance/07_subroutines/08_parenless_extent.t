#!perl
# A parenless call to a declared callee with no prototype is GREEDY: it
# consumes the whole remaining list, even when it is nested inside
# another list operator that looks like it owns those arguments.
#
# TIER 07 subroutines
# INTRODUCES the parenless call form and its argument extent
# USES print, @_, return, string interpolation
# MEASURED perl 5.42.0
# STATUS refuses as of dc1bea2c. Issue 01a0c432-fbd5. Refusal trailing_tokens.
#
# This file and `09_prototype_extent.t` are a PAIR and neither means much
# alone. Their call sites are byte-identical but for the callee's name --
# `print f 1, 2;` here and `print g 1, 2;` there -- and only the
# declaration above differs. The outputs differ, so the extent of a
# parenless argument list is decided by something that is not at the call
# site at all.
#
# `"f[@_]"` rather than a `join` because the body must not introduce a
# `(` or a second `,`: the token facts below assert on both, and a body
# that spelled its own would make those assertions about the wrong
# punctuation.
#
# MEASURED perl 5.42.0, this file:
#
#   $ perl -MO=Concise,-exec conformance/07_subroutines/08_parenless_extent.t
#   3  <0> pushmark s
#   4  <0> pushmark s
#   5  <$> const[IV 1] sM
#   6  <$> const[IV 2] sM
#   7  <#> gv[IV \&main::f] s
#   8  <1> entersub lKS
#   9  <@> print vK
#
# BOTH constants sit inside the inner pushmark, so both are f's. Compare
# the same lines in 09_prototype_extent.t, where `const[IV 2]` falls
# AFTER the `entersub` and belongs to print. That is the whole
# difference, and it is visible in the optree as well as in the output.
#
# The extent also depends on the callee being one perl has ALREADY SEEN.
# Measured 5.42.0, `f 1, 2` with no prior declaration is not an ambiguous
# extent at all but a syntax error -- "Number found where operator
# expected (Do you need to predeclare \"f\"?)" -- and defining the sub
# later in the file does not help, because the parse happens first. That
# half is measured in TestTierCallFormsParenlessExtent rather than in a
# corpus file: perl compiles no optree for a program it refuses, and the
# corpus lint asks every file for one.
#
# WHY THIS REFUSES. Our parser reads `f` as a complete term and then finds
# `1` with no operator between them, which is `trailing_tokens`: it has
# no notion that a bareword followed by a list may be a call. The token
# facts are what a fix has to keep true: there is no `(` anywhere and
# exactly one `,`, so nothing may quietly rewrite this into the
# parenthesised form while claiming to have handled the parenless one.
# They are asserted on the two argument literals rather than on `@_`,
# because `"f[@_]"` is ONE Quote token -- the array never reaches the
# token stream, and a fact about it would be satisfied vacuously.

--- source
sub f { return "f[@_]" }
print f 1, 2;
print "\n";

--- expect output
f[1 2]

--- expect tokens
no operator whose text is "("
one operator whose text is ","
one numeric literal whose text is "1"
one numeric literal whose text is "2"

--- expect parses
