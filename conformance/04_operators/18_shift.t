#!perl
# `<<` is a SHIFT OPERATOR here and a heredoc opener three tiers away,
# and until this file the corpus only ever wrote the heredoc.
#
# TIER 04 operators
# INTRODUCES the shift operators
# USES nothing from a later tier
#
# This is a trap the corpus set for itself. Measured before this file,
# every `<<` in the tree was a heredoc introducer in `13_opaque`, and the
# single `>>` was a `format` picture line. perlop's level 9 had no
# coverage at all.
#
# So a lexer that ALWAYS read `<<` as a heredoc opener passed every file
# in the corpus. The two readings are separated by POSITION -- a shift
# where an operator is expected, a heredoc where a term is -- which is
# the same expect-state mechanism the spec describes for a leading `%`,
# `<`, `&` and `/`. The corpus asked about none of them here.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $a = $ENV{X} // 1; print $a << 3'
#   8
#   $ perl -e 'my $a = $ENV{X} // 16; print $a >> 2'
#   4
#
# Shifting is multiplication and division by powers of two, which is why
# the output is checkable without a table: `1 << 3` is 8 and `16 >> 2`
# is 4.
#
# The token facts are the load-bearing half. `one operator whose text is
# "<<"` fails against a lexer that produced a heredoc opener, because a
# heredoc opener is a different category in GLOSSARY.md -- not an
# operator at all. That is the claim this file makes and the tier could
# not make before.

--- source
my $a = $ENV{X} // 1;
my $b = $ENV{Y} // 16;
print $a << 3, " ", $b >> 2, "\n";

--- expect output
8 4

--- expect parses

--- expect tokens
one operator whose text is "<<"
one operator whose text is ">>"
no heredoc opener whose text is "<< 3"
