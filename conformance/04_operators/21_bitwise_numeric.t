#!perl
# With `use v5.28`, the same two strings AND numerically and the answer
# changes from 10 to 8.
#
# TIER 04 operators
# INTRODUCES the numeric bitwise operator
# USES the bitwise operators from 16_bitwise.t
#
# The other half of `20_bitwise_polymorphic.t`. The two files differ by
# ONE LINE and by one digit of output, and that is the entire claim:
#
#   20_bitwise_polymorphic.t   my $a = "12"; my $b = "10"; $a & $b   ->  10
#   21_bitwise_numeric.t       use v5.28; ... same three lines        ->  8
#
# `use v5.28` enables `feature 'bitwise'`, under which `&` always treats
# its operands as NUMBERS. Ungated it is polymorphic and two strings get
# a character-wise AND.
#
# MEASURED perl 5.42.0, and the optree distinguishes them too:
#
#   ungated   bit_and
#   gated     nbit_and
#
# That is why this file INTRODUCES an op its partner does not. A reader
# might reasonably expect one op with a flag; measured, they are two, and
# the dependency lint can therefore tell the pair apart without reading
# the output.
#
# NO `no warnings` IS NEEDED, which was worth checking. Under the feature
# perl does not warn about string operands to a numeric operator -- it
# simply numifies them. An earlier draft carried `no warnings` on the
# assumption that it would, and the assumption was wrong.
#
# `use v5.28` rather than `use feature "bitwise"` because the version
# bundle is how this reaches real code: nobody writes the bare feature
# name, and the bundle is what a file at the top of a modern distribution
# carries. The bundle also turns on `strict`, which changes nothing here
# -- every variable is already declared.

--- source
use v5.28;
my $s1 = $ENV{X} // "12";
my $s2 = $ENV{Y} // "10";
print $s1 & $s2, "\n";

--- expect output
8

--- expect parses

--- expect tokens
one word whose text is "use"
one operator whose text is "&"
