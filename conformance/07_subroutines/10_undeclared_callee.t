#!perl
# A parenless call to a callee perl has NOT YET SEEN is not an ambiguous
# argument extent. It is a syntax error, and there is no call to have an
# extent at all.
#
# TIER 07 subroutines
# INTRODUCES the parenless call's precondition: the callee must be known
# USES print, @_, return, string interpolation
# MEASURED perl 5.42.0
#
# THIS IS THE PRECONDITION OF `08_parenless_extent.t` AND
# `09_prototype_extent.t`. Those two measure how far `1, 2` runs under a
# declared callee, one greedy and one cut to a single argument by its
# prototype. Both answers presuppose that perl read `f 1, 2` as a call.
# This file is the measurement that says when it does not.
#
# MEASURED perl 5.42.0, this source:
#
#   $ perl -c conformance/07_subroutines/10_undeclared_callee.t
#   Number found where operator expected (Do you need to predeclare "f"?)
#     at ... line 1, near "f 1"
#   syntax error at ... line 1, near "f 1"
#   had compilation errors.
#
# THE `sub f` ON THE SECOND LINE IS THE POINT, not an oversight. The
# obvious reading is that the definition below fixes the call above, and
# it does not: perl parses the file top to bottom and the call site is
# read before the declaration is reached. Delete that line and the error
# is identical, which is why it is here -- a file that merely omitted the
# sub would measure "undefined subroutine", a weaker and different fact,
# and one that happens at RUNTIME rather than at parse time.
#
# Measured the same with and without `use strict`, so this is the
# PARSER's answer rather than strictness. `08_parenless_extent.t` carries
# no `use strict` either, which keeps the pair comparable.
#
# WHAT THIS FILE MAY NOT CARRY. It has no `--- expect output`: perl never
# runs a program it will not compile, so there are no bytes to pin, and
# an empty pin would be the claim that the program RUNS and prints
# nothing, which is a different and false statement. It carries no
# `STATUS refuses` line either -- the runner consults our parser only for
# a file claiming `--- expect parses`, so there is no refusal of ours for
# a marker to go stale about.
#
# The token facts are what remains assertable, and they are the two the
# extent pair assert, for the same reason: nothing may quietly read this
# as the parenthesised form while claiming to have handled the parenless
# one. `"f[@_]"` is ONE Quote token, so a fact about `@_` would be
# satisfied vacuously and is not made.

--- source
f 1, 2;
sub f { return "f[@_]" }

--- expect parsent

--- expect tokens
one operator whose text is ","
one numeric literal whose text is "1"
one numeric literal whose text is "2"
