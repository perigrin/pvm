#!perl
# `goto` is one op name over three unrelated constructs. Two of them are
# affordable at this tier: `goto("LABEL")` carrying a constant, and a unary
# `goto` over a pad value.
#
# TIER 06 control
# INTRODUCES the goto statement
# USES nothing from a later tier
#
# `goto LABEL` emits a `<">` op -- the class B::Concise uses for an op with
# an SV operand baked in -- carrying the label as a constant argument and
# taking nothing from the stack. `goto $target` emits a `<1>` unary op over
# the pad slot, and the label it lands on is not known until run time. Same
# name, different arity, different class: a parser that gives them one node
# shape with an optional operand is modelling something perl does not.
#
# The third spelling, `goto &sub`, is MEASURED OUT of this file rather than
# forgotten. It compiles to `rv2cv`/`srefgen`/`goto` inside the sub body,
# which is not visible in the main optree at all and needs `-exec,g` to
# see; `rv2cv` belongs to tier 07 and `srefgen` to tier 08, so a file
# spelling it here would fail the lint on two ops the tier cannot claim.
# The op NAME is claimed here because these two spellings emit it; the
# frame-replacing spelling waits for the tier that supplies frames.
#
# Both jumps go FORWARD to a label at the same scope depth. Perl warns
# about `goto` into or out of a construct, and the harness compares output
# byte for byte, so a warning would be a diff.
#
# MEASURED perl 5.42.0:
#
#   $ perl -MO=Concise,-exec -e 'my $c = $ENV{X} // 0;
#     goto SKIP unless $c; print "b"; SKIP: print "c";'
#   ...
#   h  <0> padsv[$c] s
#   i  <|> or(other->j) vK/1
#   j      <"> goto("SKIP") v
#   ...
#   o  <;> nextstate(SKIP: main 3 -:1) v:{
#
#   $ perl -MO=Concise,-exec -e 'my $t = $ENV{T} // "TAIL";
#     goto $t; print "d"; TAIL: print "e\n";'
#   ...
#   t  <0> padsv[$t] s
#   u  <1> goto vKS/1
#   ...
#   z  <;> nextstate(TAIL: main 3 -:1) v:{
#
# The label is not an op. It is a flag on the `nextstate` of the statement
# it precedes -- `nextstate(SKIP: ...)` -- so a label with no statement
# after it has nowhere to live, and the jump targets a statement rather
# than a position.

--- source
my $c = $ENV{X} // 0;
my $t = $ENV{T} // "TAIL";
print "a";
goto SKIP unless $c;
print "b";
SKIP:
print "c";
goto $t;
print "d";
TAIL:
print "e\n";

--- expect output
ace

--- expect parses
