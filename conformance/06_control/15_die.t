#!perl
# `die` is a LIST OPERATOR, and where its argument list ends is a parse
# question the optree cannot answer and `$@` can.
#
# TIER 06 control
# INTRODUCES the die statement
# USES nothing from a later tier
#
# `die` was claimed by no tier at all until this file. `10_io`'s README
# records the gap in its own terms -- the idiomatic `open(...) or die` was
# written UNCHECKED there because `die` "is not tier 10's to claim, since
# failing is not this tier's subject" -- and `14_eval_block.t` in this
# tier records the other half: it traps a division by zero rather than a
# `die` because at the time the tier held no claim for the op.
#
# It is this tier's because `die` is a CONTROL TRANSFER whose target is
# the nearest enclosing `entertry` frame, which is the same family as
# `next`, `last` and `goto`, and because the frame is now spelled here:
# `entertry`/`leavetry` arrived with `14_eval_block.t`, so `eval { die }`
# became writable in this tier and was not before.
#
# MEASURED perl 5.42.0. The only op new to the tier is `die` itself:
#
#   $ perl -MO=Concise,-exec -e 'my $g = $ENV{X} // "a\n";
#     eval { die $g, "b\n" }; my $wide = $@;
#     eval { die($g), "b\n" }; print "wide=", $wide, "narrow=", $@;'
#   3  <|> entertry(other->4) v
#   l  <;> nextstate(main 2 -e:1) v:{
#   m  <0> pushmark s
#   n  <0> padsv[$g:1,7] s
#   o  <$> const[PV "b\n"] s
#   p  <@> die[t3] vK/1
#   4  <@> leavetry vKP
#   ...
#   9  <|> entertry(other->a) v
#   h  <;> nextstate(main 5 -e:3) v:{
#   i  <0> pushmark s
#   j  <0> padsv[$g:1,7] s
#   k  <@> die[t6] vK/1
#   a  <@> leavetry vKP
#
# THE EXTENT QUESTION IS THE FILE'S SUBJECT, and the two dumps above are
# why it needs a behavioural pin. `die $g, "b\n"` takes BOTH arguments --
# `pushmark`, two operands, `die`. `die($g), "b\n"` takes one, and the
# `"b\n"` does not appear in the optree AT ALL: it is a constant in void
# context and the optimiser deletes it. So the two parses differ by an op
# present in one and absent in the other, which reads like a difference an
# op check could see -- but a parser binding too LITTLE produces exactly
# the narrow stream, and nothing in it says which parse it came from.
#
# `$@` says. MEASURED, with X unset:
#
#   $ perl conformance/06_control/15_die.t
#   wide=a
#   b
#   narrow=a
#
# The wide parse concatenates both arguments into the message; the narrow
# one carries `$g` alone. One line of output separates them.
#
# BOTH MESSAGES END IN A NEWLINE ON PURPOSE. `die` appends
# ` at FILE line N.` to a message that does not, and FILE is the path the
# harness wrote its temp copy to -- a different string on every run and on
# every machine. A trailing newline suppresses the suffix, which is what
# makes this file's output reproducible rather than merely correct once.
#
# The operand comes from `$ENV{X}`, unset when the runner executes, for
# the reason this tier's README gives for every file here: a constant perl
# can see through is a construct perl can delete.
#
# The token facts are both negatives and both are checked against the
# WHOLE source. `exit` appears nowhere -- it is `16_exit.t`'s keyword, and
# the two files sit in the same group for opposite reasons, so neither
# claim may be satisfied by the other's file. And no BARE word `wide`
# exists: the name appears only as the variable `$wide` and inside the
# string `"wide="`, so the fact falsifies a lexer that split the sigil
# from its name and one that lexed inside a string literal.

--- source
my $g = $ENV{X} // "a\n";
eval { die $g, "b\n" };
my $wide = $@;
eval { die($g), "b\n" };
print "wide=", $wide, "narrow=", $@;

--- expect output
wide=a
b
narrow=a

--- expect parses

--- expect tokens
no word whose text is "exit"
no word whose text is "wide"
