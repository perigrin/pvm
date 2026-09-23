#!perl
# `$s++` on a string is the MAGIC INCREMENT -- a successor function no
# other operator in the language has.
#
# TIER 04 operators
# INTRODUCES nothing of its own
# USES postinc from 07_subroutines
#
# `27_incdec.t` pins the numeric readings; this is the one that is not
# arithmetic at all. perlop documents it as a special case of `++` and it
# is the successor `12_range_string.t` relies on for `"az" .. "bb"` --
# so the two files are the same rule seen from two operators.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $s = "Az"; $s++; print $s'      Ba
#   $ perl -e 'my $s = "zz"; $s++; print $s'      aaa
#   $ perl -e 'my $s = "a9"; $s++; print $s'      b0
#
# THREE DIFFERENT CARRIES. `Az` to `Ba` carries within the alphabet;
# `zz` to `aaa` carries and GROWS, adding a character rather than
# wrapping to `aa`; `a9` to `b0` carries across a digit run into a
# letter. A parser that numified the operand gets 1 for all three, and
# one that incremented only the last character gets `A{`, `z{` and `a:`.
#
# `zz` is the discriminating one and is why the file uses three operands
# rather than one: growth is the behaviour a naive implementation is
# least likely to have, and the other two would pass against an
# implementation that only handled a single carry.
#
# The operand is `$ENV{X} // <default>` for this tier's usual reason --
# with a constant the optimiser folds the whole expression and there is
# no increment left to measure.

--- source
my $a = $ENV{X} // "Az";
my $b = $ENV{Y} // "zz";
my $c = $ENV{Z} // "a9";
$a++;
$b++;
$c++;
print "$a $b $c\n";

--- expect output
Ba aaa b0

--- expect parses

--- expect tokens
no operator whose text is "+"
