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
# THIS FILE DECLARES NO TOKEN FACT, which is a decision rather than an
# omission. It carried one and the fact was VACUOUS.
#
# `internal/lexer/scan.go` holds a CLOSED operator table and
# `scanOperator` emits only entries from it. The discarded fact was
# `no operator whose text is "+("`, and `+(` is not a table entry, so no
# input can produce it -- the negative could never match and asserted
# nothing. Its header argued the opposite, that the spelling "appears
# nowhere in this source, so the negative is a claim rather than an
# accident". It was the accident.
#
# No replacement is honest. Every numeric literal in this source appears
# exactly twice, so no `one` fact is available; the source writes a bare
# `+`, so `no operator whose text is "+"` is false; and it contains no
# `=` at all, so `+=` is vacuous for the same reason `+(` was.
#
# THE OUTPUT IS THE WHOLE CLAIM and it is enough: 3 against 9, from two
# statements that differ by one character. A parser that reads the
# disambiguator wrong prints the other number.
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
