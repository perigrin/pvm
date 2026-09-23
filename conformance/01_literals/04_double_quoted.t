#!perl
# A double-quoted string processes escapes: `"a\nb"` is THREE characters,
# and the same bytes in single quotes are four.
#
# TIER 01 literals
# INTRODUCES string literal
# USES nothing from a later tier
#
# This is `12_single_quoted.t`'s pair, and the pair is the point. The two
# constructs are one `const` op apiece and indistinguishable in the
# optree, so neither file can claim anything from ops; what separates them
# is the ESCAPE PROCESSING, which output can see.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'print length("a\nb")'       3
#   $ perl -e 'print length(q{a\nb})'      4
#
# So this file prints two lines and its pair prints one. A lexer that
# treated the two delimiters alike fails exactly one of the two, whichever
# way it was wrong.
#
# INTERPOLATION is deliberately NOT exercised here. `"v=$x"` needs a
# runtime operand to be worth anything, and the corpus idiom for one --
# `$ENV{X} // <default>` -- uses `dor`, which is `04_operators`' op and
# three tiers away. Interpolation on an ARRAY is already claimed by
# `03_context`, which is where the construct earns its keep. What this
# tier owns is the DELIMITER, and the delimiter is what the escape
# distinguishes.

--- source
my $s = "a\nb";
print $s, "\n";

--- expect output
a
b

--- expect parses

--- expect tokens
one string literal whose text is "\"a\\nb\""
