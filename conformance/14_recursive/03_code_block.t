#!perl
# `(?{ CODE })` puts a Perl STATEMENT inside a pattern, and it runs when
# the engine reaches that point in the match.
#
# TIER 14 recursive
# INTRODUCES (?{ }) -- a code block embedded in a regex
# USES nothing from a later tier
#
# This file emits NO op of its own, and that is the finding. Measured, the
# whole statement compiles to a single `match` op whose pattern is the
# string `"2(?{ $k = 5 })"`; the assignment is inside the op's data, not
# in the op stream. The full tree does show the block's optree hanging off
# the match -- `sassign`, `const`, `padsv` -- but every op in it is marked
# `-`, optimised out of the execution path. It lives in a separate CV that
# the regex engine calls, and `-exec` walks the outer path only.
#
# So the construct that most plainly re-enters Perl is invisible to a
# measurement of ops. The tokens are the only place this file can make a
# claim, which is why they carry it: the block is INSIDE the match's one
# token, not a brace and statements beside it.
#
# The `no` claims name `5` and `$k` deliberately. `$k` appears as a real
# variable token on its own declaration line, so a bare "no variable" claim
# would be false -- but the `5` inside the block appears NOWHERE else in
# the file, so a numeric literal with that text can only come from a lexer
# that read into the pattern. That is the claim worth making.
#
# No `use re 'eval'` and no warnings suppression. Measured, a LITERAL code
# block compiles and runs clean under `use strict; use warnings`; the
# pragma is required only when the pattern itself is interpolated from a
# variable, which this file does not do.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $s="abc"; my $k=0; $s =~ /b(?{ $k = 5 })/; print "$k\n";'
#   5

--- source
my $s = "abc";
my $k = 0;
$s =~ /b(?{ $k = 5 })/;
print "$k\n";

--- expect output
5

--- expect parses

--- expect tokens
no numeric literal whose text is "5"
no operator whose text is "/"
