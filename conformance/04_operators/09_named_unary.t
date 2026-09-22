#!perl
# A NAMED UNARY OPERATOR binds LOOSER than arithmetic. `defined $x + 1`
# is `defined($x + 1)`, not `(defined $x) + 1`, and the two readings
# print different numbers.
#
# TIER 04 operators
# INTRODUCES the named unary operator's precedence level
# USES my, print, $ARGV, string interpolation
#
# 286 of T1's 986 files -- 29% -- use `defined`, and the corpus named it
# nowhere. It is placed HERE and not in tier 02, where a reader might
# expect an operator about a variable's state to live, because what is
# distinctive about its PARSE is not what it asks of a variable. It is
# where it stops. `defined` occupies a precedence level of its own,
# BELOW the arithmetic operators and ABOVE the comparisons, and that
# level is this tier's subject and no other tier's.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $x = 2; print defined $x + 1, " ", (defined $x) + 1, "\n"'
#   1 2
#
# `1` is the true the whole `defined($x + 1)` returns; `2` is that same
# true, numified, plus one. One byte apart, and every parser that got
# the level wrong prints the second where perl prints the first.
#
# THE OPS CANNOT SEE IT, which is `07_precedence.t`'s finding restated
# for the unary case and the reason this file asserts on output.
# Measured, both halves emit `add` and `defined` -- the same two names in
# the same file -- and differ only in ORDER:
#
#   $ perl -MO=Concise,-exec ... | grep -E 'defined|add'
#   a  <2> add[t4] sK/2        <- inside the loose reading
#   b  <1> defined sK/1
#   f  <1> defined sKP/1       <- inside the tight reading
#   h  <2> add[t6] sK/2
#
# `opsOf` collects op NAMES, so the lint sees `add` and `defined` twice
# each whichever way round they came. Grouping is again the part the op
# list cannot measure.
#
# TWO THINGS THE ISSUE ASSERTED THAT 5.42 DOES NOT DO, measured, and
# recorded here because both would have shaped a file written from the
# claim rather than from the interpreter:
#
#   - `defined %h` IS A FATAL ERROR, not a special parse. Measured,
#     `perl -e 'my %h = (a=>1); print defined %h'` dies with "Can't use
#     'defined(%hash)' (Maybe you should just omit the defined()?)".
#     There is no rule left to measure; the rule was REMOVED. `defined
#     $h{a}` on one element is ordinary and says nothing about `defined`.
#
#   - `defined &f` is real, and it is NOT THIS TIER'S. Measured, it
#     compiles to `rv2cv[t6] sK*/AMPER` feeding `defined sKPM/1`, and
#     `rv2cv` is `08_references`'s op -- four tiers later. A file here
#     spelling it would reach forward, which the lint reports. The
#     `&sub` half belongs wherever a tier claims code dereference, and
#     this file says so rather than smuggling it in.
#
# `$ARGV[0] // 2` is the runtime operand every file in this tier needs:
# with both sides constant the optimiser folds the operator out of
# existence and the file would measure nothing. That trap is recorded in
# this tier's README and this file honours it the same way its siblings
# do.
#
# THE TOKEN FACTS COUNT AND THE COUNTING FALSIFIES. This source holds
# exactly ONE `(` -- the one that spells the tight grouping -- so a lexer
# or a parser that supplied a paren of its own around the parenless
# `defined $x` would produce two, and one that dropped the real grouping
# would produce none. The parenless spelling IS the construct here: with
# both halves parenthesised there is no precedence question left to ask.
#
# The second fact is the ASYMMETRY, and it is the one a reader will
# doubt. Measured, our lexer gives `)` the kind `CloseBracket` and NOT
# `Operator`, while `(` is an `Operator` -- so `no operator whose text is
# ")"` is TRUE of a correct lex of a source that plainly contains one.
# It is asserted anyway, and deliberately: a lexer that unified the two
# brackets under one kind would satisfy the first fact and fail this one,
# which is exactly the drift the pair exists to catch. Tier 07's
# `09_prototype_extent.t` asserts the same two texts for the same reason.
#
# What this pair CANNOT say is how many `defined`s there are. There are
# TWO, and the grammar admits only `one` and `no`, so the count that
# would name the construct directly is not expressible. That limit is
# `GLOSSARY.md`'s and is recorded here rather than worked around: a fact
# written `one word whose text is "defined"` would be FALSE of a correct
# lex of this file, and the runner says so.

--- source
my $x = $ARGV[0] // 2;
my $loose = defined $x + 1;
my $tight = (defined $x) + 1;
print "$loose $tight\n";

--- expect output
1 2

--- expect parses

--- expect tokens
one operator whose text is "("
no operator whose text is ")"
