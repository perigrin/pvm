#!perl
# An interpolated pattern cannot be compiled at compile time, so the
# stream gains `regcomp` between the `padsv` that fetches the variable and
# the `match` that uses the result. This is the ONLY form in the tier that
# emits `regcomp`.
#
# TIER 09 regex
# INTRODUCES a runtime-compiled pattern
# USES nothing from a later tier
#
# Against `01_bare_match.t`, one character of source differs and a whole
# op appears. Both forms belong in the tier and the tier's op set is their
# union -- which is the corpus's standing lesson that the op stream is not
# derivable from the shape of the source.
#
# Note also what the match op loses: `match()[$s:2,4]` carries no pattern
# in its own dump, because it has none until `regcomp` runs.
#
# The token claim is the sharpest in the tier. `$p` appears TWICE in the
# source and must lex as ONE variable: the declaration's, with the
# pattern's occurrence staying INSIDE the pattern token. A lexer that
# recurses into a pattern -- reasonable-looking, since the pattern really
# does interpolate -- produces two, and this is the assertion that catches
# it. The optree cannot: perl interpolates at runtime either way and
# `regcomp` is emitted regardless.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $p = "b"; my $s = "abc"; my $m = $s =~ /$p/; print "$m\n"'
#   1

--- source
my $p = "b";
my $s = "abc";
my $m = $s =~ /$p/;
print "$m\n";

--- expect output
1

--- expect parses

--- expect tokens
one variable whose text is "$p"
no operator whose text is "/"
