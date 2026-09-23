#!perl
# Without the `defer` feature, `defer BLOCK` runs its block FIRST rather
# than last, and the inversion is silent until the method dispatch fails.
#
# TIER 11 oo
# INTRODUCES nothing of its own
# USES eval from 06_control and method dispatch from this tier
# STATUS refuses as of this file. Issue 01a0cf3e-e82c-7191-ac2b-e6d108b51317. Refusal unimplemented_statement.
#
# THE WORST OF THE SIX, and the reason is temporal rather than
# structural. The corpus's other version-gate findings change what a
# construct IS: `say` becomes a method call, `isa` becomes a filehandle
# print, `state` becomes a method call on undef. Each fails, loudly or at
# runtime.
#
# This one changes WHEN the code runs.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'sub f { defer { print "D\n" } print "body\n" } f();'
#   D
#   body
#   Can't locate object method "defer" via package "1"
#
#   $ perl -e 'use feature "defer"; sub f { defer { print "D\n" }
#              print "body\n" } f();'
#   body
#   D
#
# Ungated, `defer { ... }` is an indirect method call whose brace group
# is an anonymous hash constructor -- so the BLOCK IS EVALUATED EAGERLY
# to build the hash, printing `D`, and only then does perl look for a
# `defer` method and fail. Gated, the block runs at scope exit, after
# `body`.
#
# So the two readings produce the same two lines in the OPPOSITE ORDER,
# and the failure arrives after the damage. A parser that always treats
# `defer BLOCK` as a compound statement is wrong on pre-5.36 code; one
# that never does is wrong on modern code. Nothing at compile time
# distinguishes them.
#
# The `eval` traps the dispatch failure so STDOUT is pinnable. What the
# file claims is the ORDER: `D` before `body`, which is the ungated
# reading, where the gated one gives `body` before `D`.
#
# THE NEGATIVE FACT IS THE HALF THAT DISCRIMINATES, and a first draft
# omitted it. `one word whose text is "defer"` is true under BOTH readings
# -- the keyword lexes as one word whether it is a keyword or a method
# name -- so on its own it says nothing about which reading applies.
#
# What separates them is that perl reads this as a METHOD CALL. A method
# call is spelled with an arrow when written out, and this source
# contains no arrow anywhere -- so a lexer that produced one would have
# manufactured it. `10_isa_infix.t` and `10_io/07_say.t` both guard
# their equivalent claim exactly this way; this file cited them as models
# and left out the half that does the work.

--- source
sub f { defer { print "D\n" } print "body\n" }
eval { f() };
print "end\n";

--- expect output
D
body
end

--- expect parses

--- expect tokens
one word whose text is "defer"
no operator whose text is "->"
