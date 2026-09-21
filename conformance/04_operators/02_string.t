#!perl
# `.` compiles to `concat` only when its result is a LIST element -- here,
# a direct `print` argument. `x` compiles to `repeat` either way.
#
# TIER 04 operators
# INTRODUCES string concatenation and repetition
# USES nothing from a later tier
# STATUS refuses as of this file.
#
# The refusal is the token claim, not the parse: our lexer reads `x` as
# `Word("x")` where the glossary calls it an operator. Perl's word-spelled
# operators -- `x`, `cmp`, `eq`, `and`, `xor` -- are operators that happen
# to be spelled with letters, and a lexer that files them under words
# leaves the parser to tell `$a x 3` from a call to a sub named `x`. The
# claim below is written in the glossary's vocabulary and left failing
# rather than softened to match what we emit, because a claim rewritten to
# match the lexer stops measuring the lexer.
#
# This is the file that pins the tier's most misleading measurement. The
# same `.` written as `my $c = $a . $b` emits `multiconcat` and no
# `concat` at all -- and `multiconcat` is tier 01's op, claimed there with
# interpolation. Writing the concatenation as a `print` argument is what
# makes `concat` appear, so reading the source for `.` and expecting
# `concat` is wrong half the time. The destination decides, not the
# operator.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $a = "ab"; my $b = "cd"; print $a . $b, "\n"'
#   abcd

--- source
my $a = $ARGV[0] // "ab";
my $b = $ARGV[1] // "cd";
print "joined: ", $a . $b, "\n";
print "repeat: ", $a x 3, "\n";

--- expect output
joined: abcd
repeat: ababab

--- expect parses

--- expect tokens
one operator whose text is "x"
