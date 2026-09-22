#!perl
# `do BLOCK while COND` is a POST-test loop: the body runs before the
# condition is ever evaluated, and it is not `while` with the parts moved.
#
# TIER 06 control
# INTRODUCES the post-test loop
# USES nothing from a later tier
#
# This is the first of two `do BLOCK` forms the tier carries, and the one
# that is a loop. `13_do_block.t` carries the other, where the same block
# is an expression. Same two tokens, two unrelated productions, and the
# fork is decided by what FOLLOWS the closing brace -- which is the
# property this slice of the corpus exists to name.
#
# MEASURED perl 5.42.0, with N unset so the condition is false:
#
#   $ perl -MO=Concise,-exec -e 'my $i = $ENV{N} // 0;
#     do { print "once" } while $i > 0;'
#   ...
#   8  <0> enter v
#   9  <0> pushmark s
#   a  <$> const[PV "once"] s
#   b  <@> print vK
#   c  <0> unstack v
#   d  <0> padsv[$i:1,4] s
#   e  <$> const[IV 0] s
#   f  <2> gt sK/2
#   g  <|> and(other->9) vK/1
#   h  <@> leave vK*
#
# Two facts in that stream, and both are this tier's.
#
# `unstack` at c is the loop op, and it is the ONLY one: there is no
# `enterloop` and no `leaveloop` anywhere, exactly as `11_postfix_while.t`
# measures for the postfix modifier. So `next` and `last` have no frame to
# address here either -- perl builds a plain scope and a backward jump.
#
# And the ORDER is the construct. The body at 9-b precedes the test at
# d-f, where `05_while.t` puts the test first and jumps over the body. A
# parser that desugars this to `while (COND) { BLOCK }` emits the same op
# NAMES in a different order, which is why this file pins behaviour and
# not an op list.
#
# The behavioural pin is the discriminating half. N is unset, so the
# condition is false on its first evaluation, and:
#
#   $ perl conformance/06_control/09_do_while.t
#   once
#
# A `while` loop with that same false condition prints NOTHING -- measured,
# `while ($i > 0) { print "once" }` with $i == 0 compiles to an empty
# program, the erasure this tier's README records. One line of output
# where the desugared spelling produces none.
#
# The token facts count rather than forbid, as `04_elsif.t`'s do. This
# source holds exactly one `do` and exactly one `while`, so a lexer that
# split `do` off as a separate statement, or that read the trailing
# `while` as opening a second loop, would produce a count this file
# refuses. `no word whose text is "until"` is the third: the post-test
# loop has an `until` spelling too, and its absence here is what makes the
# `while` fact discriminating rather than incidental.

--- source
my $i = $ENV{N} // 0;
do { print "once" } while $i > 0;
print "\n";

--- expect output
once

--- expect parses

--- expect tokens
one word whose text is "do"
one word whose text is "while"
no word whose text is "until"
