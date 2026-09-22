#!perl
# `undef` is TWO OPERATORS wearing one word: a UNARY one that clears what
# it is given, and a NILADIC one that is simply the undefined value.
# `undef @a` empties the array; `@a = undef` fills it with one element.
#
# TIER 04 operators
# INTRODUCES the niladic/unary split in `undef`
# USES my, print, scalar, %ENV, //
# MEASURED perl 5.42.0
# STATUS refuses as of da919402. Issue 01a0c730-b241-7765-aaa2-5260d050dce9. Refusal trailing_tokens.
#
# 191 of T1's 986 files -- 19% -- use it, and the corpus named it
# nowhere. It is placed HERE and not in tier 02, where a reader expects
# an operation on a variable to live, because the question is not what
# happens to the variable. It is how many operands the WORD takes, which
# is an operator's arity and this tier's subject.
#
# THE MEASUREMENT THAT DECIDES IT, perl 5.42.0:
#
#   $ perl -e 'my @a = (1,2,3); undef @a; print scalar @a, "\n"'
#   0
#   $ perl -e 'my @a = (1,2,3); @a = undef; print scalar @a, "\n"'
#   1
#
# Opposite results from the same three characters. `undef @a` is the
# unary operator applied to the array, which discards every element.
# `@a = undef` is a list assignment whose right-hand side is the niladic
# operator's single value, so the array ends up holding exactly one
# element and that element is undef. An array that is EMPTY and an array
# holding ONE UNDEF are not the same array, and no amount of printing
# their contents would show it -- both interpolate to nothing. The COUNT
# is what tells them apart, which is why this file prints counts.
#
# THE OP NAMES ARE IDENTICAL AND THE ARITY IS NOT. Measured, this file:
#
#   $ perl -MO=Concise,-exec conformance/04_operators/10_undef_arity.t
#   o  <1> undef vK/1      <- `undef @a`, ONE operand
#   r  <0> undef s         <- `@b = undef`, NONE
#   u  <2> aassign[t7] vKS
#
# `<1>` against `<0>` is the whole construct, and `opsOf` collects op
# NAMES: it reports `undef` twice and cannot say that one of them took an
# argument. An op-name lint is satisfied by a file that only ever spelled
# one of the two. That is the same blindness `07_precedence.t` records
# for grouping, met here through arity instead, and it is why this file's
# assertion is `expect output`.
#
# `$ENV{X} // 1` rather than a bare `1` is this tier's standing trap,
# recorded in its README: with every operand constant the optimiser folds
# the construct away and the file measures nothing. `X` is unset when the
# runner executes the file, so both arrays start as `(1, 2, 3)` with the
# first element arriving at runtime.
#
# NO BRANCH ANYWHERE. `defined $a ? ... : ...` would emit `cond_expr`,
# which is tier 06's op and two tiers forward; `scalar @a` emits nothing
# at all -- measured, it compiles to a bare `padav s`. The counts are
# free.
#
# WHY THIS REFUSES, AND WHICH HALF. Bisected against our parser, the
# two spellings do NOT fail together, and that asymmetry is the finding:
#
#   @b = undef;     PARSES
#   undef @a;       REFUSES, trailing_tokens
#
# The niladic form is fine because `undef` in value position IS a
# complete term and nothing follows it. The unary form is the one that
# breaks: our parser reads `undef` as that same complete term, then
# finds `@a` with no operator between them, which is `trailing_tokens`.
# It has no notion that this word may take an operand.
#
# That is the SAME refusal, at the same site and for the same reason, as
# tier 07's `08_parenless_extent.t` -- a bareword read as a term where
# perl reads an operator with a parenless argument. Two tiers apart, one
# missing rule. The difference is that tier 07's case is a user sub and
# this one is a builtin, so a fix that special-cases named builtins would
# clear this file and leave tier 07's refusing.
#
# THE TOKEN FACTS ARE WHAT A FIX HAS TO KEEP TRUE, and neither is the
# count a reader would reach for first. There are THREE `=` operators
# here -- two `my` initialisers and the `@b = undef` assignment, all
# three `Operator("=")` to our lexer -- and TWO words spelled `undef`,
# so the counts that would name the construct directly are not
# expressible: the grammar admits only `one` and `no`.
#
# So the facts assert around it. `one word whose text is "print"` pins
# that there is exactly one print statement, which a lexer that split or
# duplicated the statement stream changes.
#
# `no operator whose text is ")"` is the asymmetric one, and the reason
# is worth stating because it reads as false. This source contains two
# `)` characters, both closing a list literal -- and measured, our lexer
# gives them the kind `CloseBracket`, never `Operator`, so the claim is
# TRUE of a correct lex of a source that plainly holds two. A lexer that
# unified the brackets under one kind satisfies nothing here and fails
# this, which is the drift the fact exists to catch. `09_named_unary.t`
# asserts the same text for the same reason and next to its `(` twin.
#

--- source
my @a = ($ENV{X} // 1, 2, 3);
my @b = ($ENV{X} // 1, 2, 3);
undef @a;
@b = undef;
print scalar @a, scalar @b, "\n";

--- expect output
01

--- expect parses

--- expect tokens
one word whose text is "print"
no operator whose text is ")"
