#!perl
# `exit` leaves the PROGRAM, and `eval BLOCK` -- the frame that catches
# every `die` -- does not catch it.
#
# TIER 06 control
# INTRODUCES the exit statement
# USES nothing from a later tier
#
# `exit` was claimed by no tier before this file, the same gap `15_die.t`
# closes for its own op. It is this tier's on the same argument: a jump
# whose target the construct names and which nothing in the source spells,
# which is how this tier's README frames `entertry` and is why `next`,
# `last`, `redo` and `goto` all sit here.
#
# WHAT SEPARATES IT FROM `die` IS THE ONLY THING WORTH MEASURING, because
# the two are otherwise the same shape: a named unary in statement
# position that the following statement never reaches. MEASURED perl
# 5.42.0, with X unset so the status is 0:
#
#   $ perl -MO=Concise,-exec -e 'my $c = $ENV{X} // 0;
#     print "ran\n"; eval { exit $c }; print "trapped\n";'
#   ...
#   a  <@> print vK
#   b  <;> nextstate(main 2 -e:3) v:{
#   c  <|> entertry(other->d) v
#   j  <;> nextstate(main 3 -e:3) v
#   k  <0> padsv[$c:1,4] s
#   l  <1> exit vK/1
#   d  <@> leavetry vK
#   e  <;> nextstate(main 4 -e:4) v:{
#   f  <0> pushmark s
#   g  <$> const[PV "trapped\n"] s
#   h  <@> print vK
#
# The `entertry` frame is BUILT -- `c` and `d` bracket the block exactly
# as they do in `14_eval_block.t` -- and the `exit` at `l` goes straight
# through it. Swap `exit $c` for `die $c` and the same frame catches, the
# program continues, and `trapped` prints. Same syntax, same ops around
# it, opposite outcome, and only running the file shows which.
#
# THE DISCRIMINATING PIN IS AN ABSENCE, and the optree is what makes it a
# claim rather than an accident. `print "trapped\n"` COMPILES -- ops `f`
# through `h` are in the binary, at addresses after the `exit` -- and
# never runs. So this file asserts two things at once that a single
# spelling cannot fake: the statement is real perl the compiler accepted,
# and control never arrives at it.
#
#   $ perl conformance/06_control/16_exit.t
#   ran
#
# One line. A parser that treated `exit` as an ordinary bareword call
# would compile the same ops and print `trapped` as well, and a parser
# that dropped the trailing statement as unreachable would print the same
# one line while failing `expect parses` on a statement it never built.
#
# THE RUNNER SEES THIS, which is not obvious and was checked rather than
# assumed. `askPerl` in `internal/conformance/run.go` execs the file and
# reads `.Output()`, which returns stdout whether or not the process exits
# non-zero -- and `exit 0` is not an error at all, so nothing about the
# early termination is hidden from the comparison. A file exiting
# mid-stream is pinned on exactly the bytes it printed before it left.
#
# The status comes from `$ENV{X}`, unset when the runner executes, and not
# only for this tier's constant-erasure reason: an `exit` whose status is a
# literal would still terminate, but the value would be one the optimiser
# folded rather than one the program computed, and `exit` takes an
# EXPRESSION.
#
# The token facts are the pair `15_die.t` carries in mirror. Exactly one
# `exit`, which falsifies a lexer joining it to what follows; and NO
# `die`, checked against the whole source, which keeps the two files'
# claims from satisfying each other.

--- source
my $c = $ENV{X} // 0;
print "ran\n";
eval { exit $c };
print "trapped\n";

--- expect output
ran

--- expect parses

--- expect tokens
one word whose text is "exit"
no word whose text is "die"
