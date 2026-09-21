#!perl
# Three dot-separated parts make a v-string without any `v`: `65.66.67` is
# one token denoting the three-character string `ABC`, NOT a numeric
# literal.
#
# TIER 01 literals
# INTRODUCES numeric literal
# USES nothing from a later tier
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $v = 65.66.67; print "$v\n"'
#   ABC
#
#   $ perl -e 'my $n = 65.66; print "$n\n"'
#   65.66
#
# ONE DOT IS A NUMBER AND TWO DOTS ARE A STRING. That is the whole
# boundary, and it is decided by counting dots after the token has already
# started -- a lexer cannot know which it is scanning until it reaches the
# second point or the end. `perldata` documents this under "Version
# Strings", away from numbers, and ../GLOSSARY.md, "numeric literal",
# records the decision: a v-string is not in that category.
#
# STATUS refuses as of 7711154e. Our lexer produces Number("65.66.67") --
# one token, which is right, in the wrong CATEGORY, which is not. The
# parser sees a number where perl sees a string and returns zero Unknowns,
# and the value never reaches output because we do not run the program.
# The token assertion is the only place the disagreement is visible.
#
# No issue: this file is the record. The construct was found by writing
# it, so there is nowhere earlier for it to have been filed.

--- source
my $v = 65.66.67;
print "$v\n";

--- expect parses

--- expect output
ABC

--- expect tokens
no numeric literal whose text is "65.66.67"
