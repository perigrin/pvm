#!perl
# `!` is the most common operator in Perl that this corpus never wrote,
# and in this corpus it was a DELIMITER before it was an operator.
#
# TIER 04 operators
# INTRODUCES nothing of its own
# USES the logical operators from 05_logical.t
#
# Measured before this file, every `!` in the corpus was either `!=` or a
# quote delimiter -- `q!...!`, `qq!...!`, `m!...!`, all of which tier 01
# and tier 09 use to prove a delimiter is arbitrary. So the character
# appeared often and the OPERATOR never did, which is the shape of gap a
# lexer can pass without implementing.
#
# MEASURED perl 5.42.0, and the return value is the part worth pinning:
#
#   $ perl -e 'my $a = 0; print "[", !$a, "][", !!$a, "]"'
#   [1][]
#
# `!0` is 1 and `!!0` is the EMPTY STRING, not 0. Perl's false is a
# dual-valued empty-string-and-zero, and `print` shows the string half.
# A parser that folded `!!` to a no-op prints `[1][0]`; one that treated
# `!` as numeric negation prints `[-0][0]` or similar. The two brackets
# are what separate them.
#
# `not` is the same operation at a different precedence and is already
# covered by `05_logical.t`, which measures that `not` binds TIGHTER than
# `and` -- the finding this tier's README calls its sharpest. This file
# is the punctuation spelling, whose precedence is level 5 rather than
# level 23, and the two are different operators to the parser even
# though they share an op.

--- source
my $a = $ENV{X} // 0;
print "[", !$a, "][", !!$a, "]\n";

--- expect output
[1][]

--- expect parses

--- expect tokens
no operator whose text is "!="
