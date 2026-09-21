#!perl
# A pattern with no interpolation is compiled once, at COMPILE time: the
# match emits `match` alone, with no `regcomp` beside it.
#
# TIER 09 regex
# INTRODUCES a match against a constant pattern
# USES nothing from a later tier
#
# This is the tier's baseline. `05_interpolated.t` is the same construct
# with one character changed -- `/abc/` becomes `/$p/` -- and that one
# character adds a whole op. Measuring the constant form first is what
# makes that difference attributable.
#
# The token claims are negative, and deliberately so. GLOSSARY.md's
# `quote-like operator` is defined as a quote spelled with an OPERATOR
# NAME -- q, qq, m, s, qr -- or with backticks; a bare `/abc/` is neither,
# so there is no positive category to assert it under and this file does
# not invent one. What it can say is what a mis-lex would produce: the
# hazard for a bare-slash pattern is the slash read as DIVISION and the
# pattern body read as Perl, which yields an Operator("/") and a
# Word("abc"). Asserting their absence is the claim that survives.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $m = "abc" =~ /abc/; print "$m\n"'
#   1

--- source
my $m = "abc" =~ /abc/;
print "$m\n";

--- expect output
1

--- expect parses

--- expect tokens
no operator whose text is "/"
no word whose text is "abc"
