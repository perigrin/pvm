#!perl
# `~` is a unary operator at perlop's level 5, and it is the only one of
# that level's six the corpus reaches other than `\` and unary minus.
#
# TIER 04 operators
# INTRODUCES the bitwise complement
# USES the bitwise operators from 16_bitwise.t
#
# MEASURED perl 5.42.0. `~` complements every bit of a 64-bit integer, so
# its bare result is platform-sized and enormous:
#
#   $ perl -e 'my $a = 12; print ~$a'
#   18446744073709551603
#
# That number is not a good thing to pin -- it says as much about the
# integer width as about the operator. Masking with `& 255` keeps the
# claim to the low byte and makes the arithmetic checkable by hand:
# `~12` has its low byte `11110011`, which is 243.
#
#   $ perl -e 'my $a = $ENV{X} // 12; print ~$a & 255'
#   243
#
# The mask is why this file depends on `16_bitwise.t` rather than
# standing alone: `&` has to exist before a complement can be observed
# without pinning a word size.
#
# PRECEDENCE comes free and is worth stating. `~` is level 5 and `&` is
# level 14, so `~$a & 255` is `(~$a) & 255` -- the complement binds first.
# A parser that read it as `~($a & 255)` would print 18446744073709551355
# instead of 243, which is the same discriminating shape as
# 17_bitwise_precedence.t one level up.

--- source
my $a = $ENV{X} // 12;
print ~$a & 255, "\n";

--- expect output
243

--- expect parses

--- expect tokens
one operator whose text is "~"
