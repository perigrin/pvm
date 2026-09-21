#!perl
# A block `if` and a postfix `if` are the same construct: same ops, same
# order, down to the `nextstate`. Neither emits an `if` op -- both emit
# tier 04's `and`.
#
# TIER 06 control
# INTRODUCES the conditional statement, block and postfix
# USES nothing from a later tier
#
# This file asserts an EQUALITY rather than a behaviour. The two statements
# print the same thing, so an output test alone would pass against a parser
# that treats them as unrelated grammar productions; what makes the claim
# checkable is that the two halves of the measured op stream below are
# byte-identical apart from the `nextstate` line number.
#
# The condition is `$ENV{X} // 0` because a CONSTANT condition erases the
# construct: `if (1) { print "y" }` emits `pushmark const print` and no
# branch op at all. Measured, not assumed.
#
# MEASURED perl 5.42.0:
#
#   $ perl -MO=Concise,-exec -e 'my $c = $ENV{X} // 0;
#     if ($c) { print "y\n" } print "y\n" if $c;'
#   ...
#   8  <0> padsv[$c] s
#   9  <|> and(other->a) vK/1
#   a      <0> pushmark s
#   b      <$> const[PV "y\n"] s
#   c      <@> print vK
#   d  <;> nextstate(main 6 -:1) v:{
#   e  <0> padsv[$c] s
#   f  <|> and(other->g) vK/1
#   g      <0> pushmark s
#   h      <$> const[PV "y\n"] s
#   i      <@> print vK
#
# The block form's braces add no `enter`/`leave` pair at this nesting --
# the `and` branches straight into `pushmark`, exactly as the postfix does.

--- source
my $c = $ENV{X} // 1;
if ($c) { print "block\n" }
print "postfix\n" if $c;

--- expect output
block
postfix

--- expect parses
