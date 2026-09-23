#!perl
# Underscores inside a numeric literal are separators, part of the token
# and absent from the value: `4_294_967_296` is one token denoting
# 4294967296.
#
# TIER 01 literals
# INTRODUCES numeric literal
# USES nothing from a later tier
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $x = 4_294_967_296; print "$x\n"'
#   4294967296
#
# The underscore is a WORD character everywhere else in perl, which is
# what makes this a boundary rather than a detail: a lexer that stops a
# number at the first non-digit produces `Number(4) Word(_294_967_296)`,
# and a lexer that scans word characters greedily after a digit produces
# one token whose text is right and whose value is unparseable. Both are
# reachable, and the token assertion below distinguishes them from the
# correct answer, which the printed value cannot -- the separators are
# gone by the time anything prints.

--- source
my $x = 4_294_967_296;
print "$x\n";

--- expect parses

--- expect output
4294967296

--- expect tokens
one numeric literal whose text is "4_294_967_296"
no word whose text is "_294_967_296"
