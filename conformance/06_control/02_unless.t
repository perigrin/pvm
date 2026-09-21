#!perl
# `unless` is `or`, not an inverted `if`. The op streams differ by exactly
# one op, and there is no negation op anywhere.
#
# TIER 06 control
# INTRODUCES the negated conditional
# USES nothing from a later tier
#
# The discriminating pair is exact. Against `01_if_postfix.t`'s stream,
# position 9 is `and` there and `or` here; every other op is the same op in
# the same place. Perl does not invert the condition and test it, it picks
# the other short-circuit. A parser that desugars `unless COND` to
# `if (!COND)` is building a tree perl never builds -- and nothing in the
# output distinguishes the two, which is why this file exists.
#
# Both `and` and `or` are tier 04's ops. What this tier contributes is the
# measurement that the two keywords differ by exactly one of them.
#
# MEASURED perl 5.42.0:
#
#   $ perl -MO=Concise,-exec -e 'my $c = $ENV{X} // 0;
#     unless ($c) { print "n\n" } print "n\n" unless $c;'
#   ...
#   8  <0> padsv[$c] s
#   9  <|> or(other->a) vK/1
#   a      <0> pushmark s
#   b      <$> const[PV "n\n"] s
#   c      <@> print vK
#   d  <;> nextstate(main 6 -:1) v:{
#   e  <0> padsv[$c] s
#   f  <|> or(other->g) vK/1
#   ...
#
# `$ENV{X}` is unset when the harness runs, so `// 0` is a stable false and
# both branches are taken.

--- source
my $c = $ENV{X} // 0;
unless ($c) { print "block\n" }
print "postfix\n" unless $c;

--- expect output
block
postfix

--- expect parses
