#!perl
# The postfix LOOP modifier is the exception that proves postfix `if`:
# `EXPR while COND` emits `enter`/`leave` and `unstack` but NO `enterloop`.
#
# TIER 06 control
# INTRODUCES the postfix loop modifier
# USES nothing from a later tier
#
# `01_if_postfix.t` measures that postfix `if` and block `if` are the same
# construct to perl. The loop modifiers are the case where the same
# reasoning fails: same modifier syntax, different machinery. A postfix
# loop has no block, so there is nothing for `next` and `last` to target,
# so perl does not build the loop frame that names those targets -- it
# builds a plain scope and a backward jump.
#
# The consequence is observable rather than cosmetic: `next` inside a
# postfix `while` does not address that loop, because that loop has no
# `enterloop` to address.
#
# The body is `$i = $i - 1` rather than `$i--` because `postdec` is an op
# no tier here claims, and the file's subject is the loop, not the
# decrement.
#
# MEASURED perl 5.42.0:
#
#   $ perl -MO=Concise,-exec -e 'my $i = $ENV{N} // 3;
#     print $i = $i - 1 while $i > 0;'
#   ...
#   8  <0> enter v
#   9  <0> padsv[$i] s
#   a  <$> const[IV 0] s
#   b  <2> gt sK/2
#   c  <|> and(other->d) vK/1
#   d      <0> pushmark s
#   e      <0> padsv[$i] s
#   f      <$> const[IV 1] s
#   g      <2> subtract[$i] sK/TARGMY,2
#   h      <@> print vK
#   i      <0> unstack v
#              goto 9
#   j  <@> leave vK*
#
# `enter` at 8 and `leave` at j, and no `enterloop` or `leaveloop`
# anywhere -- compare `05_while.t`, where those two are the frame and
# `enter`/`leave` are absent.

--- source
my $i = $ENV{N} // 3;
print $i = $i - 1 while $i > 0;
print "\n";

--- expect output
210

--- expect parses
