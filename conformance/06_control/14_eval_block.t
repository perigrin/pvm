#!perl
# `eval BLOCK` is a CONTROL construct, not a re-entry into the parser, and
# the ops say so: it emits `entertry`/`leavetry` and never `entereval`.
#
# TIER 06 control
# INTRODUCES entertry leavetry
# USES nothing from a later tier
#
# `eval` is two constructs wearing one keyword, and measured they share no
# op at all. MEASURED perl 5.42.0:
#
#   $ perl -MO=Concise,-exec -e 'my $r = eval { 1 };'
#   3  <|> entertry(other->4) s
#   ...
#   4  <@> leavetry sK
#
#   $ perl -MO=Concise,-exec -e 'my $r = eval "1";'
#   5  <1> entereval[t256] sK/1
#
# Different ops entirely, and the split decides the placement. The STRING
# form compiles Perl the compiler never saw -- it hands a region back to
# the lexer with no delimiter at all -- and it lives in `14_recursive`,
# whose README frames the tier as "here the operand is PERL" and which
# already claims `entereval` for `s///ee`. `06_recursive_eval_string.t`
# over there is that file.
#
# This form does none of that. The block is ordinary Perl, compiled once,
# at the ordinary time. What `entertry` adds is a FRAME: a marked point
# the runtime can unwind back to when something inside the block fails.
# That is a control transfer, the same family as this tier's `next`,
# `last` and `goto` -- a jump whose target is named by the construct
# rather than by a label, and whose distinction is that nothing in the
# source SPELLS the jump. So it belongs here and its two ops are claimed
# here.
#
# `entertry` and `leavetry` were unclaimed by every tier in the corpus
# before this file. Adding them to 06's INTRODUCES is the additive change,
# not a move: `claimonce_test.go` finds no other tier holding them.
#
# THE OPERAND IS DIVISION BY ZERO, NOT `die`, and that is a budget fact
# rather than a stylistic one. `die` emits a `die` op which no tier at or
# before 06 claims, so `eval { die "x" }` reaches forward and the lint
# refuses it. Division by zero is a runtime failure this tier can already
# spell -- `divide` is tier 04's -- and it exercises the same frame.
#
# MEASURED, with X unset so the divisor is zero:
#
#   $ perl -MO=Concise,-exec -e 'my $d = $ENV{X} // 0;
#     my $r = eval { 10 / $d } // "trapped"; print "$r\n";'
#   8  <|> entertry(other->9) s
#   j  <;> nextstate(main 3 -e:1) v
#   k  <$> const[IV 10] s
#   l  <0> padsv[$d:1,5] s
#   m  <2> divide[t4] sK/2
#   9  <@> leavetry sK
#   a  <|> dor(other->b) sK/1
#   b      <$> const[PV "trapped"] s
#
# Note the ADDRESSES: `entertry` at 8 jumps to j, the block body runs at
# j-m, and `leavetry` is numbered 9 -- the frame ops bracket the block in
# source order while sitting outside it in execution order. That is the
# shape a parser gets wrong when it treats `eval` as a named unary over a
# hash constructor.
#
# THE BEHAVIOURAL PIN IS THE DISCRIMINATING HALF, and it is exit status as
# much as output:
#
#   $ perl conformance/06_control/14_eval_block.t
#   trapped
#
# The same statement without the block does not print `trapped` and does
# not print anything: measured, `my $r = 10 / $d // "trapped"` exits 255
# with `Illegal division by zero at -e line 1.` So a parser that dropped
# the block, or that evaluated its contents outside the frame, produces a
# program that aborts where this one completes. No op list is needed to
# see that; one line of output and a zero exit is the whole claim.
#
# The token facts count. Exactly one `eval`, and NO `do` -- the two
# keywords take a block in the same position and `13_do_block.t` is the
# file that does, so the absence here is what keeps the two claims
# separate rather than each satisfying the other.

--- source
my $d = $ENV{X} // 0;
my $r = eval { 10 / $d } // "trapped";
print "$r\n";

--- expect output
trapped

--- expect parses

--- expect tokens
one word whose text is "eval"
no word whose text is "do"
