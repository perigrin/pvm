#!perl
# `//=`, `||=` and `&&=` SHORT-CIRCUIT, and they are the only compound
# assignments whose behaviour output can see.
#
# TIER 04 operators
# INTRODUCES the short-circuit assignment operators
# USES the logical operators from 05_logical.t
#
# The rest of the family reuses its binary op -- `+=` emits `add`, `.=`
# emits `multiconcat`. These three do not. Measured perl 5.42.0, each
# emits a DEDICATED op:
#
#   $u //= 5   ->  dorassign
#   $u ||= 5   ->  orassign
#   $u &&= 9   ->  andassign
#
# Three ops no tier claimed. That is the measured argument for splitting
# them off rather than writing one compound-assignment file: they are a
# different construct wearing the same shape.
#
# AND THEY ARE THE ONLY ONES OUTPUT CAN DISCRIMINATE. `$x += 2` and
# `$x = $x + 2` compute the same value, so `22_compound_arithmetic.t`
# rests entirely on token facts. Here the short circuit is behavioural:
#
#   $ perl -e 'my $u = 0; $u //= 99; print $u'    0
#   $ perl -e 'my $u = 0; $u ||= 99; print $u'    99
#
# `0` is DEFINED but FALSE, which is exactly the case that separates `//=`
# from `||=`. A parser that treated them as synonyms -- or that expanded
# either to `$u = $u // 99` without the short circuit -- prints the wrong
# one of those two numbers. This is the same distinction
# `05_logical.t` makes for `//` against `||`, moved into assignment
# position where the left operand is also the target.
#
# `&&=` is the third and the mirror: it assigns only when the left side is
# already TRUE.
#
#   $ perl -e 'my $u = 1; $u &&= 99; print $u'    99
#   $ perl -e 'my $u = 0; $u &&= 99; print $u'    0
#
# All four numbers appear below, from four variables that start at the
# two values the operators disagree about.
#
# THE TOKEN FACTS COUNT RATHER THAN FORBID. A first draft added
# `no operator whose text is "//"`, on the reasoning that a lexer
# splitting `//=` into `//` and `=` should be caught. The runner rejected
# it immediately: this source uses `$ENV{X} // 0` four times to get its
# runtime operands, so a bare `//` is present by construction and the
# negative was simply false. The counting facts carry the claim instead --
# a lexer that split `//=` would produce FIVE `//` tokens where the
# source has four, and zero `//=`.

--- source
my $p = $ENV{X} // 0;
my $q = $ENV{X} // 0;
my $r = $ENV{Y} // 1;
my $s = $ENV{X} // 0;
$p //= 99;
$q ||= 99;
$r &&= 99;
$s &&= 99;
print "$p $q $r $s\n";

--- expect output
0 99 99 0

--- expect parses

--- expect tokens
one operator whose text is "//="
one operator whose text is "||="
