#!perl
# `elsif` is a THIRD conditional form, not `else` followed by `if`.
#
# TIER 06 control
# INTRODUCES the chained conditional
# USES nothing from a later tier
#
# The tier shipped with `if`, `unless` and `if`/`else` and no `elsif`
# anywhere, while 14 of T1's 986 files use it. A parser that desugars
# `elsif` to a nested `else { if ... }` produces a tree perl does not
# build, and nothing in the tier would have said so.
#
# MEASURED perl 5.42.0. The chain compiles to NESTED cond_expr, one per
# condition, and adds no op of its own:
#
#   $ perl -MO=Concise,-exec -e 'my $n = shift; if ($n == 1) { print "a" }
#         elsif ($n == 7) { print "b" } else { print "c" }'
#   ... eq cond_expr(other->d) ... eq cond_expr(other->l) ...
#
# So `cond_expr` -- already claimed by this tier for `if`/`else` -- covers
# it, and the INTRODUCES set does not grow. What distinguishes `elsif` is
# the SHAPE of the nesting, which ops cannot show, so the file pins the
# branch behaviourally instead: the middle branch is the one taken, which
# neither a lone `if` nor an `else` can produce.
#
#   $ perl conformance/06_control/04_elsif.t
#   zero
#
# The token facts count rather than forbid, and the counting is the
# falsifying half: this source holds exactly ONE `elsif` and ONE `else`, so
# a lexer that read `elsif` as `else` followed by `if` would produce TWO
# words spelled `else` and fail the second fact. Measured, ours emits
# `Word("elsif")` whole.
#
# The condition reads $ENV{X}, which is unset when the runner executes the
# file, because a constant condition erases the construct entirely --
# `if (1) {...}` emits no branch op at all. That trap is recorded in this
# tier's README and every file here honours it.

--- source
my $c = $ENV{X} // 0;
if ($c == 1) { print "one\n" } elsif ($c == 0) { print "zero\n" } else { print "other\n" }

--- expect output
zero

--- expect parses

--- expect tokens
one word whose text is "elsif"
one word whose text is "else"
