#!perl
# The bitwise compound assignments reuse the ops `16_bitwise.t` and
# `18_shift.t` introduce, which is why they could not be written until
# those existed.
#
# TIER 04 operators
# INTRODUCES nothing of its own
# USES the bitwise and shift operators from 16_bitwise.t and 18_shift.t
#
# This file is the reason the compound assignment issue declared a
# dependency on the bitwise band: `|=` cannot be claimed before `|` is,
# and the dependency lint enforces that rather than leaving it to
# reading order.
#
# MEASURED perl 5.42.0. Every op is the binary one, reused:
#
#   $x |= 3    ->  bit_or
#   $x &= 10   ->  bit_and
#   $x ^= 3    ->  bit_xor
#   $x <<= 3   ->  left_shift
#   $x >>= 2   ->  right_shift
#
#   $ perl -e 'my $a=12; my $b=12; my $c=12; my $d=1; my $e=16;
#              $a|=3; $b&=10; $c^=3; $d<<=3; $e>>=2;
#              print "$a $b $c $d $e"'
#   15 8 15 8 4
#
# `15 8 15` is the same three numbers `16_bitwise.t` pins for the binary
# forms on the same operands, which is the point: the compound spelling
# must compute what the binary one does. A parser that mis-read `^=` as
# `|=` would print `15 8 15` unchanged, so the OUTPUT does not separate
# those two -- the token facts do.
#
# That is the general shape of this whole family, recorded once in
# `22_compound_arithmetic.t`: the value a compound assignment computes is
# reachable by a longhand spelling, so output can confirm the operands
# and only the token stream can confirm the operator.

--- source
my $a = $ENV{X} // 12;
my $b = $ENV{X} // 12;
my $c = $ENV{X} // 12;
my $d = $ENV{Y} // 1;
my $e = $ENV{Z} // 16;
$a |= 3;
$b &= 10;
$c ^= 3;
$d <<= 3;
$e >>= 2;
print "$a $b $c $d $e\n";

--- expect output
15 8 15 8 4

--- expect parses

--- expect tokens
one operator whose text is "|="
one operator whose text is "&="
one operator whose text is "^="
one operator whose text is "<<="
one operator whose text is ">>="
