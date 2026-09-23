#!perl
# `x=` is the thirteenth compound assignment and the only WORD-SHAPED one,
# and our lexer does not form the token.
#
# TIER 04 operators
# INTRODUCES nothing of its own
# USES the string operators from 02_string.t
# STATUS refuses as of this file. Issue 01a0ce57-db92-78e5-bbe4-7c28e74db6f3. Refusal trailing_tokens.
#
# THE BISECT IS SHARP, and it is why this file is separate from
# `23_compound_string.t` rather than a second line in it. Measured against
# our own lexer:
#
#   $t x= 2       Variable($t) Word(x) Operator(=) Number(2)   REFUSES
#   $t .= "b"     Variable($t) Operator(.=) Quote("b")         parses
#   $t = $t x 2   ... Operator(=) Variable($t) Word(x) ...     parses
#
# So the binary `x` is lexed correctly AND the punctuation compound `.=`
# is lexed correctly; only the combination fails. `x=` arrives as two
# tokens and the parser then meets a `Word` where an operator belongs,
# leaving the statement with bytes it cannot place -- `trailing_tokens`.
#
# The cause is structural rather than an oversight about one operator.
# `x` is the ONLY word-shaped member of the thirteen; the other twelve
# are punctuation. A lexer that scans punctuation runs to build compound
# assignment never reaches a word, which is exactly what the token dump
# shows.
#
# MEASURED perl 5.42.0, so the construct is real:
#
#   $ perl -e 'my $t = "ab"; $t x= 2; print $t'
#   abab
#
# THE SPACING IS LOAD-BEARING and any fix has to respect it. `$t x= 2`
# needs the space before `x`, because `$tx= 2` names a different variable
# entirely. The token has to be formed from a word-position `x` followed
# immediately by `=`, without absorbing an identifier -- which is the
# same hazard that makes `x` unusual as a binary operator and is recorded
# in `02_string.t`.
#
# The file pins the answer perl gives rather than the one we give, which
# is what a refusing file is for: when the lexer learns the token, this
# file starts passing and the ratchet reports it.

--- source
my $t = $ENV{X} // "ab";
$t x= 2;
print "$t\n";

--- expect output
abab

--- expect parses

--- expect tokens
one operator whose text is "x="
