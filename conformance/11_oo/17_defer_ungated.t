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
# THE TOKEN LAYER CANNOT SEPARATE THE TWO READINGS, and saying so
# is the honest version of a claim this file got wrong twice.
#
# `one word whose text is "defer"` is true under BOTH readings --
# the keyword lexes as one word whether it is a keyword or a
# method name -- so on its own it says nothing about which reading
# applies. That much an earlier draft had right.
#
# What it concluded was wrong. It added
# `no operator whose text is "->"`, reasoning that the ungated
# reading is an indirect method call, that a method call is
# written with an arrow, and that a lexer producing one here
# would have manufactured it. A LEXER CANNOT MANUFACTURE BYTES
# THAT ARE NOT THERE: `scanOperator` matches its table against
# `l.src` at each position, so a token whose text is `->`
# requires those two bytes in the source. The fact could never
# fail, and a fact that cannot fail asserts nothing.
#
# The premise is right here and the evidence still is not.
# Measured, perl's deparse of the ungated reading keeps the
# block form and shows no arrow at all:
#
#   $ perl -MO=Deparse -e 'defer { print "D\n" } print "body\n";'
#   defer {
#       print "D\n"
#   } print("body\n");
#
# So there was no arrow to read the claim off. The fact was
# copied from `14_state_ungated.t`, whose deparse DOES show
# `$n->state`, without checking that it applied here.
#
# So the positive fact stays and does the work it can: the keyword
# lexes as ONE word, not split and not swallowed. THE OUTPUT IS
# THE DISCRIMINATOR, and it has to be -- identical bytes lex
# identically, and only running them tells the readings apart.

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
