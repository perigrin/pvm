#!perl
# Without the `bitwise` feature, `&` is POLYMORPHIC: string operands get
# a string bitwise operation, and `"12" & "10"` is the string `"10"`.
#
# TIER 04 operators
# INTRODUCES nothing of its own
# USES the bitwise operators from 16_bitwise.t
#
# The file is named `bitwise_polymorphic` but introduces no construct of
# its own: it emits `bit_and`, which `16_bitwise.t` already claims. What
# it adds is the STRING operand, and the answer that follows from it.
#
# This file is one half of a pair. `21_bitwise_numeric.t` holds the other
# and differs from it by ONE LINE -- a `use v5.28` at the top -- and by
# the answer it prints. Neither file means anything alone.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $a = "12"; my $b = "10"; print $a & $b'
#   10
#   $ perl -e 'use v5.28; my $a = "12"; my $b = "10"; print $a & $b'
#   8
#
# `"12" & "10"` ANDs the characters: `'1' & '1'` is `'1'` and `'2' & '0'`
# is `'0'`, giving the string `"10"`. Numerically `12 & 10` is `8`. Two
# valid programs from identical bytes, and NO WARNING OR ERROR either
# way -- which is what makes this the sharpest version of a hazard the
# corpus has now met three times.
#
# `say` without its feature is a method call and dies at runtime. `isa`
# without its feature is a filehandle print and dies at runtime.
# `state` without its feature is a method call on undef and dies at
# runtime. All three FAIL, loudly enough to be noticed.
#
# This one SUCCEEDS with a different answer. A parser can be wrong here
# and every test that does not pin the pragma will agree with it.
#
# THE OPTREE SEES IT TOO, which was not obvious and is worth recording:
#
#   ungated   bit_and
#   gated     nbit_and
#
# Two ops, not one op with a flag. So the dependency lint distinguishes
# the pair as well as the output does -- but only because both files
# exist. A corpus with just this one would claim `bit_and` and say
# nothing about the feature that changes it.
#
# Operands are runtime values for the reason `16_bitwise.t` records: with
# constants the optimiser folds the expression and there is no operator
# left to be polymorphic. They are STRINGS here rather than numbers
# because the polymorphism is exactly about what a string operand does.
#
# The `no "&."` fact is the other gating kind in the same band. Measured,
# the DOTTED spellings are hard-gated where the undotted ones are silent:
#
#   $ perl -e 'my $a = "AB"; my $b = "ab"; print $a &. $b'
#   syntax error at -e line 1, near "&."
#
# So level 14 hides both kinds at once -- `&.` refuses without the
# feature, `&` changes meaning with it. This file's source contains no
# `&.` anywhere, which is what makes the negative a claim: a lexer that
# split `&` from a following `.` would manufacture one.

--- source
my $s1 = $ENV{X} // "12";
my $s2 = $ENV{Y} // "10";
print $s1 & $s2, "\n";

--- expect output
10

--- expect parses

--- expect tokens
one operator whose text is "&"
no operator whose text is "&."
