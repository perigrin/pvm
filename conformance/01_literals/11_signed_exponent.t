#!perl
# The sign of an exponent is part of the numeric literal: `5e-1` is one
# token, not `5e` minus `1`.
#
# TIER 01 literals
# INTRODUCES numeric literal
# USES nothing from a later tier
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'print 5e-1, "\n"'
#   0.5
#
#   $ perl -e 'print 5e, "\n"'
#   Bareword found where operator expected (Missing operator before "e"?)
#   syntax error at -e line 1, near "5e"
#
# The second measurement is what makes the split WRONG rather than merely
# different: `5e` is not a number, so a lexer emitting Number("5e") has
# produced a token perl would reject. See ../GLOSSARY.md, "numeric literal",
# and note the asymmetry recorded there -- the exponent sign is part of the
# literal, a leading `-` never is.
#
# STATUS refuses as of 38c95d23. Our lexer produces Number(5e) Operator(-)
# Number(1), and the same split affects `5e+1`, `5E-1` and `1.5e-3`.
# Unsigned `5e1` lexes correctly, which is why it went unnoticed.
#
# No issue: this file is the record. The construct was found by writing it,
# so there is nowhere earlier for it to have been filed, and duplicating the
# token stream into a tracker would give it a second place to go stale.

--- source
my $x = 5e-1;
print "$x\n";

--- expect parses

--- expect output
0.5

--- expect tokens
one numeric literal whose text is "5e-1"
no operator whose text is "-"
