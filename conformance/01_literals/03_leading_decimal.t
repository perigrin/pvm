#!perl
# A leading decimal point is part of the numeric literal: `.5` is one token,
# not the concatenation operator `.` followed by `5`.
#
# TIER 01 literals
# INTRODUCES numeric literal
# USES nothing from a later tier
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $x = .5; print "$x\n"'
#   0.5
#
#   $ perl -e 'printf "%.17g %.17g %.17g\n", .5, 0.5, 5e-1'
#   0.5 0.5 0.5
#
# The three spellings are the SAME VALUE to 17 significant digits, so no
# behavioural probe can see which was written. That is why this file carries
# a token assertion: the spelling is a lexical fact and output cannot reach
# it. See ../GLOSSARY.md, "numeric literal".
#
# STATUS refuses as of 38c95d23. Our lexer produces Operator(.) Number(5),
# because scanNumber does not accept a leading `.`. Issue 01a0c13f-97f5-7f98-b32d-07245ec6ddfe.

--- source
my $x = .5;
print "$x\n";

--- expect parses

--- expect output
0.5

--- expect tokens
one numeric literal whose text is ".5"
no operator whose text is "."
