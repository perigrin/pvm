#!perl
# A `0x` prefix makes the digits after it hexadecimal, letters included:
# `0xff` is one token denoting 255.
#
# TIER 01 literals
# INTRODUCES numeric literal
# USES nothing from a later tier
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $x = 0xff; print "$x\n"'
#   255
#
# The trap this file covers is that `ff` is also a legal identifier. A
# lexer scanning the digit `0`, stopping at the first non-digit, and
# handing `xff` to the word scanner produces `Number(0) Word(xff)` --
# which the parser reads as a number beside a bareword, not as an error.
# Measured, the three radix spellings and one decimal are the SAME value:
#
#   $ perl -e 'printf "%s %s %s\n", 0xff, 0377, 0o377'
#   255 255 255

--- source
my $x = 0xff;
print "$x\n";

--- expect parses

--- expect output
255

--- expect tokens
one numeric literal whose text is "0xff"
no word whose text is "xff"
