#!perl
# `caller` returns DIFFERENT AMOUNTS in the two contexts, and at file
# scope the difference is the sharpest in the tier: one value against
# none at all.
#
# TIER 03 context
# INTRODUCES caller
# USES my, print, scalar
#
# 74 of T1's 986 files call it and the corpus named it nowhere. It is
# placed HERE and not in tier 07, where a reader expects a call-stack
# operator to live, because what makes its parse distinctive is not the
# stack: it is that the SAME call site yields a scalar or a list
# depending on what receives it, which is this tier's entire subject and
# no other tier's.
#
# MEASURED perl 5.42.0, this file's first two statements:
#
#   $ perl -MO=Concise,-exec -e 'my $s = caller; my @l = caller;'
#   3  <0> caller[t2] s
#   7  <0> caller[t4] l
#
# One op name, two context flags -- the `s`/`l` pair this tier is built
# on, exactly as `reverse` and `localtime` carry it. An op-NAME lint sees
# one `caller` and is satisfied by a file that only ever put it in one
# context, which is why `discriminatingPairs` names this file.
#
# The behavioural half is stronger here than anywhere else in the tier.
# At file scope there IS no caller, and the two contexts do not merely
# format the same answer differently -- they return different AMOUNTS of
# it. Measured:
#
#   $ perl -e 'my $s = caller; my @l = caller; my @one = $s;
#       print scalar @one, " ", scalar @l, "\n";'
#   1 0
#
# `@l` is EMPTY: list-context `caller` with no frame to report returns
# the empty list. `$s` is undef, but a scalar holding undef is still one
# element, so the count reports 1. A parser that imposed LIST context on
# both would print `0 0`; one that imposed SCALAR on both would print
# `1 1`. Both wrong readings are visible in one byte.
#
# NO SUBROUTINE, deliberately. A sub is where `caller` is USEFUL, and it
# is also where `entersub` and `leavesub` are -- tier 07's ops, four
# tiers later, which the lint would report as this file reaching forward.
# `07_wantarray.t` records the same finding for the same reason: the
# operator is legal at file scope and measures there.
#
# COUNTING rather than `defined $s ? ... : ...`, again for
# `07_wantarray.t`'s reason: a ternary emits `cond_expr`, which is tier
# 06's op. And `scalar` is not an op at all -- measured, `scalar @one`
# compiles to a bare `padav[@one] s`, which is this tier's own scalar
# half of the `padav` pair. The count is free.
#
# THE SOURCE HAS NO PARENTHESIS ANYWHERE, and that is what the token
# fact rests on. Both `caller` calls are parenless, both `scalar` calls
# are parenless, and `my @one = $s` needs none. So a lexer or a parser
# that quietly rewrote any of those into a parenthesised call -- the
# obvious way to "handle" a parenless operator, and the way tier 07's
# extent files say must not happen -- introduces a `(` this file says is
# absent. Measured, ours emits none: the whole token stream is words,
# variables, `=`, `,` and quotes.
#
# The fact is written as a FORBIDDING count because the counting form
# cannot express what is wanted here. There are TWO words spelled
# `caller`, and the grammar admits only `one` and `no`, so `one word
# whose text is "caller"` would be FALSE of a correct lex. The paren
# claim is the one this source can make that a broken lexer fails and a
# correct one passes.

# THE TOKEN FACT HERE IS WEAK AND THE FILE SAYS SO. `one word whose
# text is "print"` would pass any lexer that tokenises identifiers at
# all; it asserts nothing about `caller`.
#
# Measured, nothing better is available. `caller`, `scalar`, `$s` and
# `@l` each appear twice, `@one` twice, and the fact grammar has only
# `one` and `no` -- so no count fits, and every negative that would be
# reachable here names a spelling the source writes, which makes it
# false rather than vacuous.
#
# What this file actually claims is its OUTPUT: `1 0`, which separates
# scalar `caller` (returning a package name) from list `caller` at the
# top level (returning the empty list). The token layer has nothing to
# add, and a fact naming text this source cannot produce would be worse
# than a weak one -- that is the whole finding of `01a0cfb2`.

--- source
my $s = caller;
my @l = caller;
my @one = $s;
print scalar @one, " ", scalar @l, "\n";

--- expect output
1 0

--- expect parses

--- expect tokens
one word whose text is "print"
