#!perl
# `time` is NILADIC: it parses with no argument at all, and it is the one
# construct in the corpus whose value cannot be asserted.
#
# TIER 06 control
# INTRODUCES the time builtin
# USES nothing from a later tier
#
# WHY THIS FILE IS IN A CONTROL TIER, stated plainly rather than argued
# around: `time` is not control flow. It is here because 06 is the
# EARLIEST tier that can hold it, and no earlier tier can.
#
# The reason is the op budget and nothing else. A non-deterministic
# builtin is observable only through a COMPARISON -- see below -- and the
# comparison needs `gt`, which `04_operators` introduces, and the ternary
# that renders the result needs `cond_expr`, which is this tier's. The
# lint is `ops(file) ⊆ ∪ ops(tiers ≤ N)`, so the first tier that can hold
# both is 06. Measured, the file emits:
#
#   cond_expr const enter gt leave nextstate padsv padsv_store print
#   pushmark time
#
# Tier 03 was the first placement proposed, on the ground that it already
# claims `localtime` and the two share a clock. MEASURED, that does not
# survive. Tier 03's subject is the DISCRIMINATING PAIR -- an op that
# answers differently in scalar and list context -- and `time` has none:
#
#   $ perl -e 'my $n = () = time; my $s = time; my @c = ($s);
#       print "$n ", scalar(@c), "\n"'
#   1 1
#
# `localtime` prints `9 1` there, which is why it is tier 03's. `time`
# returns ONE value in both contexts, so a file placing it beside
# `localtime` would measure nothing about context. Same clock, different
# subject.
#
# WHAT IS DISTINCTIVE IS THE ARITY. `time` takes no argument and no
# parentheses, and unlike every other named operator in the corpus there
# is no form of it that takes one: `time 1` is a syntax error, not a call
# with an ignored argument. MEASURED perl 5.42.0:
#
#   $ perl -e 'my $t = time 1;'
#   Number found where operator expected (Do you need to predeclare "time"?) at -e line 1, near "time 1"
#   syntax error at -e line 1, near "time 1"
#
# So a parser that treats named operators uniformly -- callee, then a
# greedy argument extent, which is what `07_subroutines`'s extent files
# are about -- gets this one wrong in the direction of accepting too
# much. The source below puts `time` immediately before a `;` with
# nothing between, which is the whole of its grammar.
#
# THE COMPARISON IS WHAT MAKES IT OBSERVABLE, and it is not a workaround.
# `time` returns seconds since the epoch, so its value differs on every
# run and pinning it would make the file fail one second later. Pinning a
# PROPERTY of the value is the only assertion available:
# 1000000000 is 2001-09-09, so any clock later than that gives `past`.
# The file therefore asserts that `time` returned a number in the right
# half of the line, which is everything that is true of it on every run
# and nothing that is not.
#
#   $ perl -e 'my $t = time; print $t > 1000000000 ? "past" : "impossible", "\n"'
#   past
#
# `impossible` is the other branch's text on purpose: it names what a
# failure would mean rather than describing the branch. A file printing
# `no` there would leave a reader unsure whether the comparison or the
# clock was wrong.
#
# NO CONSTANT ANYWHERE NEAR THE CALL. This tier's README records that a
# constant condition erases the construct -- `if (1)` emits no branch op
# at all -- and the same risk applies to the comparison here. It does not
# fire: `$t` is a runtime value the optimiser cannot see through, so the
# `gt` and the `cond_expr` are both in the binary. That is why the value
# is bound to `$t` first and compared second, rather than written as
# `time > 1000000000` in one expression.
#
# The token fact is the niladic claim in the only form the grammar has.
# Exactly one word spelled `time`, which falsifies a lexer that joined it
# to the `;` or split it from itself. `localtime` does NOT appear in this
# source, checked against the whole file including these comments, so the
# count is not satisfied by a substring of some other word -- and it
# could not be in any case, which is worth recording because it is not
# obvious: `checkTokenFact` (`internal/conformance/fact.go:50`) compares
# `tokenText == text` on the token's FULL text, never as a substring. So
# `one word whose text is "tie"` counts exactly one even in a file that
# also contains `tied`, and the same protects `time` from `localtime`.

--- source
my $t = time;
print $t > 1000000000 ? "past" : "impossible", "\n";

--- expect output
past

--- expect parses

--- expect tokens
one word whose text is "time"
