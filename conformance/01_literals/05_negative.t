#!perl
# A leading minus is NOT part of a numeric literal: `-1` is two tokens, a
# negation operator and the literal `1`.
#
# TIER 01 literals
# INTRODUCES numeric literal
# USES nothing from a later tier
#
# This is the asymmetry `08_signed_exponent.t` is the other half of. The
# sign of an EXPONENT is part of the token; a sign in front of the number
# never is.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $x = -1; print "$x\n"'
#   -1
#
#   $ perl -MO=Concise -e 'my $x = -1;' 2>&1 | grep const
#   const[IV -1] s/FOLD
#
# The second measurement is why this file exists. The optree holds ONE
# FOLDED CONSTANT: the optimiser has already applied the negation and
# erased the operator, so nothing downstream of compilation can tell
# `-1` from a hypothetical single negative literal. Behaviour cannot see
# it either -- both would print `-1`. The token stream is the only place
# the two tokens are still two, which is this tier's argument for the
# token layer stated on the case where it is cheapest to check.
#
# See ../GLOSSARY.md, "numeric literal", which decides this against
# `perlop`'s unary minus rather than against our lexer.
#
# Our lexer gets this right today: Operator(-) Number(1). The file is
# here to keep it right, not to report that it is wrong.

--- source
my $x = -1;
print "$x\n";

--- expect parses

--- expect output
-1

--- expect tokens
one operator whose text is "-"
no numeric literal whose text is "-1"
