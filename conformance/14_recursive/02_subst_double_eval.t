#!perl
# A second `e` evaluates the FIRST evaluation's result as Perl: the string
# `3*4` is compiled at runtime by code the compiler never saw.
#
# TIER 14 recursive
# INTRODUCES s///ee -- the replacement's value re-read as Perl
# USES nothing from a later tier
#
# This is the deepest re-entry the corpus contains and the only place an
# `entereval` op appears. `s/b/$c/ee` compiles to the same
# `subst(replstart->)` / `substcont` frame as `/e`, with `entereval`
# between them: the first `e` evaluates `$c` to the string `3*4`, the
# second compiles and runs that string.
#
# The distinction matters lexically because NEITHER `e` is visible to the
# lexer as code. The replacement region holds `$c`; the Perl that
# eventually runs is `3*4`, which appears in the source only as the
# contents of a `q{}` three lines earlier. A lexer cannot reach it at all,
# and neither can the optree -- `entereval` is where the source ends.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $s="abc"; my $c=q{3*4}; $s =~ s/b/$c/ee; print "$s\n";'
#   a12c

--- source
my $s = "abc";
my $c = q{3*4};
$s =~ s/b/$c/ee;
print "$s\n";

--- expect output
a12c

--- expect parses

--- expect tokens
one quote-like operator whose text is "s/b/$c/ee"
one quote-like operator whose text is "q{3*4}"
no operator whose text is "*"
