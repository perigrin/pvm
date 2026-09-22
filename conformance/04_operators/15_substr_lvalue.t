#!perl
# `substr` is an LVALUE. `substr($s,0,1) = "J"` assigns INTO the string,
# and the four-argument form does the same replacement while RETURNING
# the text it displaced -- same op name, same result in the string, and
# only the return value tells them apart.
#
# TIER 04 operators
# INTRODUCES nothing; `substr` is 14_substr_arity.t's
# USES my, print, %ENV, //, string interpolation
# MEASURED perl 5.42.0
#
# THIS IS THE ONE OPERATOR IN THE TIER THAT CAN APPEAR ON THE LEFT OF AN
# `=`. Every other operator here produces a value and stops. `substr`
# produces a WINDOW onto its first argument, and assigning through that
# window edits the original:
#
#   $ perl -e 'my $s = "hello"; substr($s,0,1) = "J"; print "$s\n"'
#   Jello
#
# A parser that reads `substr(...)` as a call and calls are not lvalues
# gets a syntax error where perl gets `Jello`. That is a PARSE question
# about where an lvalue may appear, which is why the construct is in an
# operator tier rather than filed with the string builtins.
#
# THE LVALUE-NESS IS A FLAG AND NOT A DIFFERENT OP, which is the finding
# and the reason this file cannot assert on ops. Measured, the two halves
# below:
#
#   substr($a, 0, 1) = "J"       <@> substr[t5] vKS/REPL1ST,3
#   substr($b, 0, 1, "J")        <@> substr[t7] sK/4
#
# Both named `substr`, and `opsOf` collects NAMES, so it reports the op
# once and says nothing about which spelling produced it -- nor that one
# of them was an assignment target. There is no `sassign` either: the
# `REPL1ST` flag means the replacement text was folded INTO the substr op
# and the assignment disappeared as a separate step, so even the op that
# would betray an assignment is absent. An op-name lint is satisfied by a
# file that only ever spelled the rvalue form.
#
# THE FOUR-ARGUMENT FORM RETURNS THE OLD TEXT, and that is the only
# observable difference between the two:
#
#   $ perl -e 'my $b = "hello"; my $old = substr($b,0,1,"J"); print "$b $old\n"'
#   Jello h
#
# Both strings end up `Jello`. The `h` is what the four-argument form
# hands back and the lvalue form discards, so a file printing only the
# strings would show two identical results and measure nothing. Printing
# the displaced character is what makes the pair discriminate -- the
# standard this tier's README sets for a fixture, and the standard half
# its first-draft precedence fixtures failed.
#
# OFFSET ZERO IS REACHABLE HERE AND NOT IN THE SIBLING FILE, which is
# worth stating because it looks like an inconsistency. Measured,
# `substr($s,0,1)` in RVALUE position compiles to `substr_left`, an op no
# tier in this corpus claims, so `14_substr_arity.t` keeps every offset
# nonzero to stay inside the dependency lint. Neither form here is an
# rvalue at offset zero: one is an assignment target and one is the
# four-argument replacement, and measured, both emit plain `substr`. So
# this file covers the position its neighbour cannot, and the corpus
# reaches offset zero only by the two spellings that are not the ordinary
# one.
#
# NO `sassign` AND NO BRANCH. The assignment folds into the op as
# recorded above, and nothing here tests a condition -- a `defined ? :`
# would emit `cond_expr`, which is tier 06's op two tiers forward. Both
# strings start from the same `$ENV{X} // "hello"`, which is this tier's
# runtime operand and tier 02's op, so the two halves begin identical and
# any difference at the end is the construct's.
#
# THE TOKEN FACTS. There are FOUR words spelled `substr` here and the
# grammar admits only `one` and `no`, so the count that names the
# construct is not expressible -- the limit `14_substr_arity.t` records
# and this file inherits. `=` is unavailable for the same reason: the
# source holds three `my` initialisers and the lvalue assignment, four
# `Operator("=")` to our lexer, and the grammar cannot say four. The
# `=` that carries the construct is indistinguishable from the three
# that do not, so the facts assert around it.
#
# `one word whose text is "print"` pins the single print statement, which
# a lexer that split or duplicated the statement stream changes;
# `10_undef_arity.t` asserts the same text for the same reason.
# `no word whose text is "index"` is checked against the whole source,
# `$ENV{X}` included, and holds -- `12_index_sentinel.t` is where that
# word belongs, and the two files are the corpus's only users of a
# string-position builtin, so a drift that borrowed one into the other
# breaks this first.

--- source
my $a = $ENV{X} // "hello";
my $b = $ENV{X} // "hello";
substr($a, 0, 1) = "J";
my $old = substr($b, 0, 1, "J");
print "[$a][$b][$old]\n";

--- expect output
[Jello][Jello][h]

--- expect parses

--- expect tokens
one word whose text is "print"
no word whose text is "index"
