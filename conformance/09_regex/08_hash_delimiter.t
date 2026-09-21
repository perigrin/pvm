#!perl
# `#` is PERL'S COMMENT CHARACTER, and it is also a legal delimiter. A
# lexer reading left to right has already decided `m#abc#` is a comment
# before it can know it was wrong.
#
# TIER 09 regex
# INTRODUCES the comment character used as a delimiter
# USES nothing from a later tier
#
# This is the sharpest delimiter case in the tier, and the one its issue
# names alongside `m//` and `m{}`. Every other delimiter here is a
# character with no other job at that position; `#` has one, and the two
# readings differ by everything -- `m#abc#` is a match, while `m` followed
# by a comment is a bareword and then nothing at all to end of line.
#
# Measured under 5.42.0, the op stream is BYTE-IDENTICAL to the one for
# `/abc/` and `s/a/z/`: `match(/"abc"/) sKS` and `subst(/"a"/)`. So
# neither the optree nor the printed output can report a lexer that got
# this wrong, and the token facts below are the only place the claim can
# live. That is the tier's standing argument in its strongest form.
#
# Both spellings are here rather than the match alone, because the
# substitution is where the hazard compounds: `s#a#z#` has THREE `#`
# characters, and a lexer that recovers from the first by luck still has
# two more to mis-read.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $a = "abc" =~ m#abc#; my $s = "abc"; $s =~ s#a#z#; print "$a$s\n"'
#   1zbc

--- source
my $a = "abc" =~ m#abc#;
my $s = "abc";
$s =~ s#a#z#;
print "$a$s\n";

--- expect output
1zbc

--- expect parses

--- expect tokens
one quote-like operator whose text is "m#abc#"
one quote-like operator whose text is "s#a#z#"
