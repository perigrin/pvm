#!perl
# `&&` and `and` are the SAME op and differ only in precedence, and the
# difference is invisible in the op names -- it shows up as a changed
# tree shape and as different output.
#
# TIER 04 operators
# INTRODUCES the precedence gap between `&&` and `and`
# USES nothing from a later tier
# STATUS refuses as of this file.
#
# The refusal is the token claim, and it is the sharpest instance of the
# word-operator gap in this tier: our lexer gives `&&` the kind
# `Operator` and `and` the kind `Word`, so the two halves of a construct
# whose whole point is that they are THE SAME OPERATOR arrive as
# different kinds of thing. Anything downstream that decides on kind
# alone will treat one as an operator and the other as a bareword.
#
# Measured: `my $c = ($a && $b)` and `my $c = ($a and $b)` produce
# byte-identical op streams down to the order. With the parentheses
# removed the two diverge, because `=` binds tighter than `and` but
# looser than `&&`:
#
#   my $tight = $a && 9;    (($tight = ($a && 9)))        -> 9
#   my $loose = $a and 9;   ((($loose = $a)) and 9)       -> 1
#
# The second compiles the ASSIGNMENT INSIDE the left operand of `and`,
# which is what puts a `sassign` in the op stream -- `padsv $a`, `padsv
# $loose`, `sassign`, then `and`. `sassign` is tier 02's op, claimed with
# the package scalar; only its REASON for appearing is new here. An op
# that shows up exactly when precedence goes the surprising way is what
# makes the cliff observable at all.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $a = 1; my $t = $a && 9; my $l = $a and 9; print "$t $l\n"'
#   9 1

--- source
my $a = $ARGV[0] // 1;
my $tight = $a && 9;
my $loose = $a and 9;
print "$tight $loose\n";

--- expect output
9 1

--- expect parses

--- expect tokens
one operator whose text is "&&"
one operator whose text is "and"
