#!perl
# The `/e` flag makes the replacement a Perl EXPRESSION rather than a
# string: the lexer must hand the region between the second and third
# delimiter back to the parser.
#
# TIER 14 recursive
# INTRODUCES s///e -- an expression in the replacement half
# USES nothing from a later tier
#
# The replacement is `$n+1` rather than a constant on purpose. Measured,
# `s/a/uc("z")/e` folds to `const[PV "Z"] s/FOLD` followed by a plain
# `subst` -- the optimiser erases the `/e` completely, and the file would
# measure the same thing as tier 09's `04_substitution.t`. An unfoldable
# replacement is what makes `substcont` appear.
#
# The token claims record where the re-entry has NOT yet happened. At the
# lexical layer `s/a/$n+1/e` is ONE quote-like operator, the same as tier
# 09's `s/a/z/`: the `+` between its delimiters is inside its text, not a
# token beside it. A lexer that emitted that `+` as an operator would
# have lexed the replacement instead of delimiting it, which is the error
# this tier exists to catch -- the region is handed to the PARSER, which
# lexes it in a second pass.
#
# `$n` is declared on its own line and so does appear as a variable token
# there; only the `+` is unique to the replacement, which is why it is
# the one the `no` claim names.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $s="abc"; my $n=1; $s =~ s/a/$n+1/e; print "$s\n";'
#   2bc

--- source
my $s = "abc";
my $n = 1;
$s =~ s/a/$n+1/e;
print "$s\n";

--- expect output
2bc

--- expect parses

--- expect tokens
one quote-like operator whose text is "s/a/$n+1/e"
no operator whose text is "+"
