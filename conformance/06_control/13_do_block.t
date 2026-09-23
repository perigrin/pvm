#!perl
# `do BLOCK` in expression position is a BLOCK where a value goes, and the
# value is its last statement. It is not an anonymous hash, and nothing in
# the token stream says which it is.
#
# TIER 06 control
# INTRODUCES the block-valued expression
# USES nothing from a later tier
#
# `09_do_while.t` carries the other `do BLOCK`, where the same two tokens
# open a post-test loop. The fork is decided by what FOLLOWS the closing
# brace -- a `while` makes it a loop, anything else makes it this -- so a
# parser cannot choose the production when it reads the `do`.
#
# THE ERASURE COMES FIRST, because it is why this file is shaped the way
# it is. MEASURED perl 5.42.0:
#
#   $ perl -MO=Concise,-exec -e 'my $r = do { 1 }; print $r;'
#   1  <0> enter v
#   2  <;> nextstate(main 1 -e:1) v:{
#   3  <$> const[IV 1] s
#   4  <1> padsv_store[$r:3,4] vKS/LVINTRO
#
# NO OP AT ALL. The construct is gone: `do { 1 }` is a bare `const`, and a
# corpus file spelled that way would pass its output test while measuring
# nothing whatsoever. That is tier 01's lesson about `my $x = 1+2`
# arriving in this tier, and this tier's README records the same trap for
# `if (1)` and `while (0)`.
#
# So the block gets a RUNTIME operand and more than one statement:
#
#   $ perl -MO=Concise,-exec -e 'my $r = do { my $t = $ENV{X} // 7;
#         $t + 1 }; print $r;'
#   3  <0> enter s
#   4  <;> nextstate(main 2 -e:1) v
#   5  <+> multideref($ENV{"X"}) sK
#   6  <|> dor(other->7) sK/1
#   7      <$> const[IV 7] s
#   8  <1> padsv_store[$t:2,3] vKS/LVINTRO
#   9  <;> nextstate(main 3 -e:1) v
#   a  <0> padsv[$t:2,3] s
#   b  <$> const[IV 1] s
#   c  <2> add[t4] sK/2
#   d  <@> leave sKP
#   e  <1> padsv_store[$r:4,5] vKS/LVINTRO
#
# and even then the construct adds no op of its own. What survives is an
# `enter`/`leave` pair -- tier 01's ops -- with an `s` flag rather than
# the `v` a statement-position block gets, and the LAST statement's value
# falling out of the `leave` into the assignment. The block-valued
# expression is a FLAG on an op perl already had, which is the same shape
# tier 03 measures for scalar versus list context.
#
# The INTRODUCES set does not grow, and that is the honest result.
#
# WHAT THE OUTPUT FALSIFIES. Printing $r is not decoration here. perl
# itself disambiguates `{` at the start of a term by heuristic, and the
# wrong answer is an anonymous hash -- which is precisely the misparse the
# task this file closes names. Read as a hash constructor, $r would hold a
# REFERENCE and this file would print `HASH(0x...)`. It prints:
#
#   $ perl conformance/06_control/13_do_block.t
#   8
#
# A scalar, not a reference. One byte of output separates the two trees.
#
# The `+ 1` is load-bearing twice over. It keeps the last statement
# unfoldable, and it makes the printed value differ from the `7` the
# block's first statement binds -- so a parser that took the FIRST
# statement as the block's value, or that treated the two statements as a
# comma list and kept the left one, prints `7` and fails here too.
#
# `ref($r)` would say the same thing more directly and is deliberately not
# used: `ref` is tier 08's op and this tier may not reach forward for it.
# The lint refused the clearer spelling and was right to.
#
# The token facts count. Exactly one `do`, and NO `while` -- the absence
# is what distinguishes this file's production from `09_do_while.t`'s,
# which holds the same `do` and the same block and is a loop. A lexer that
# emitted a `while` here would be reporting a construct this source does
# not contain.

--- source
my $r = do { my $t = $ENV{X} // 7; $t + 1 };
print "$r\n";

--- expect output
8

--- expect parses

--- expect tokens
one word whose text is "do"
