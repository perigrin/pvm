#!perl
# `.=` is ONE operator token, and it emits `multiconcat` rather than the
# `concat` a reader would predict from `+=` giving `add`.
#
# TIER 04 operators
# INTRODUCES nothing of its own
# USES the string operators from 02_string.t
#
# Separated from `22_compound_arithmetic.t` because the op is not the
# binary one reused. Measured perl 5.42.0:
#
#   $s .= "b"    ->  multiconcat
#   $x += 2      ->  add
#
# The peephole optimiser fuses the append, the same fusion tier 01's
# README records for `print "$x\n"`. A reader expecting `concat` by
# analogy with the arithmetic group finds something else, which is why
# this file exists rather than a fifth line in that one.
#
# The claim is a TOKEN claim for the reason `22_compound_arithmetic.t`
# records: `$s .= "b"` and `$s = $s . "b"` produce the same output and a
# different token count. The output below is a sanity check on the
# operand, not the discriminator.
#
#   $ perl -e 'my $s = "a"; $s .= "b"; print $s'
#   ab
#
# `x=` IS NOT HERE. It is the other string compound assignment and our
# lexer does not form the token at all -- `24_compound_repeat.t` records
# that refusal and bisects it.

--- source
my $s = $ENV{X} // "a";
$s .= "b";
print "$s\n";

--- expect output
ab

--- expect parses

--- expect tokens
one operator whose text is ".="
