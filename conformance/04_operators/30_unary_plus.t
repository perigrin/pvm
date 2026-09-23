#!perl
# Unary `+` computes nothing and exists to DISAMBIGUATE: it is the only
# way to stop `print (...)` being read as a complete call.
#
# TIER 04 operators
# INTRODUCES nothing of its own
# USES nothing from a later tier
#
# perlop documents it as a no-op whose whole purpose is this, and the
# corpus never wrote it.
#
# MEASURED perl 5.42.0, and the two readings print different numbers:
#
#   $ perl -e 'print (1+2)*3; print "\n"'
#   3
#   $ perl -e 'print +(1+2)*3; print "\n"'
#   9
#
# `print (1+2)*3` is `(print(1+2)) * 3` -- the parens after a list
# operator are its ARGUMENT LIST, so `print` takes `1+2`, prints 3, and
# the `*3` multiplies print's return value and is discarded. The unary
# plus makes the parens an ordinary grouping again and the whole
# expression becomes the argument.
#
# THE DANGEROUS HALF IS THAT NEITHER IS AN ERROR. Both compile, both
# print a number, and the number is wrong only if you meant the other
# one. Perl warns under `-w`, but a parser has to make the choice before
# any warning exists, and a corpus that never writes the disambiguator
# cannot tell which choice a parser made.
#
# THE TOKEN FACT IS A NEGATIVE, and it is the one this construct needs.
# Measured, the lexer emits `Operator("+")` and `Operator("(")`
# separately, which is correct -- unary plus is not a compound token.
# A lexer that fused them into `+(` would have invented an operator
# perl does not have, and that spelling appears nowhere in this source,
# so the negative is a claim rather than an accident.
#
# The OUTPUT is what carries the rest: 3 against 9, from two statements
# that differ by one character.
#
# This is the same class as `09_named_unary.t`'s `defined $x + 1` --
# where an operator's argument stops -- and the reason both live in this
# tier rather than with the constructs they appear to be about.

--- source
print (1+2)*3;
print "\n";
print +(1+2)*3;
print "\n";

--- expect output
3
9

--- expect parses

--- expect tokens
no operator whose text is "+("
