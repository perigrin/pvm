#!perl
# A BRACKETING DELIMITER NESTS. `m{a{b}c}` is a match on the pattern
# `a{b}c`: the inner braces are CONTENT, not the terminator, and a lexer
# must count depth rather than stop at the first `}`.
#
# TIER 09 regex
# INTRODUCES a pattern containing its own delimiter
# USES nothing from a later tier
#
# This is the half of the tier's scope that a delimiter FORM cannot
# express. Every other bracketing pattern here -- `m{abc}`, `m{zzz}`,
# `s{a}{z}` -- has a body with no bracket in it, so all of them are
# satisfied by a lexer that stops at the first closing character. Only a
# pattern carrying its own opening delimiter can tell the two apart.
#
# The asymmetry is what makes the rule a rule, and GLOSSARY.md records
# both halves: bracketing delimiters nest, non-bracketing ones do not.
# Measured, the non-bracketing counterpart is not merely different but
# ILL-FORMED -- `"a!b!c" =~ m!a!b!c!` gives `Unknown regexp modifier "/b"`
# and then a syntax error -- so it cannot appear in a file expected to
# parse, and that is the reason `06_alternative_delimiters.t` keeps its
# `m!abc!` to a body with no `!` in it.
#
# THE SUBSTITUTION IS WHAT PINS THE PATTERN'S EXTENT, and the `x` and `y`
# around the target are why. A lexer that stopped at the first `}` would
# take the pattern as `a{b` -- which still matches, so a bare match would
# report 1 either way -- but the replacement would then cover three
# characters instead of five and leave `}c` behind. `xoky` says the whole
# five characters went, bounded on both sides by text the pattern must not
# have touched.
#
# Perl's false is the empty string, which is why the match prints inside
# brackets: without them a mis-measured match and a correct one both
# produce a line with little on it.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $hit = "a{b}c" =~ m{a{b}c}; my $s = "xa{b}cy"; $s =~ s{a{b}c}{ok}; print "[$hit][$s]\n"'
#   [1][xoky]

--- source
my $hit = "a{b}c" =~ m{a{b}c};
my $s = "xa{b}cy";
$s =~ s{a{b}c}{ok};
print "[$hit][$s]\n";

--- expect output
[1][xoky]

--- expect parses

--- expect tokens
one quote-like operator whose text is "m{a{b}c}"
one quote-like operator whose text is "s{a{b}c}{ok}"
