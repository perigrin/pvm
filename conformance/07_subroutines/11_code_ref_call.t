#!perl
# `$code->(...)` -- the callee is a VALUE rather than a name, and the
# argument list is parenthesised because there is no other way to spell
# it.
#
# TIER 07 subroutines
# INTRODUCES the code-reference call form
# USES my, print, @_, return, anonymous sub, string interpolation
# MEASURED perl 5.42.0
#
# WHY THIS FORM HAS NO EXTENT QUESTION, which is the reason it sits
# beside the two files that do. `08_parenless_extent.t` and
# `09_prototype_extent.t` measure an extent decided by a DECLARATION the
# parser must have already seen. Here there is no name to have seen: the
# callee is whatever `$code` holds at run time, so the parens are
# mandatory and the argument list ends where they close. The same
# construct that makes the callee unknowable until run time makes its
# argument extent knowable at parse time.
#
# MEASURED perl 5.42.0:
#
#   $ perl -MO=Concise,-exec conformance/07_subroutines/11_code_ref_call.t
#   3  <$> anoncode[CV CODE] sR
#   4  <1> padsv_store[$code:3,4] vKS/LVINTRO
#   ...
#   a  <0> padsv[$code:3,4] s
#   b  <1> entersub[t3] lKS/TARG
#
# `padsv` where the named forms carry `gv`: that is the whole difference
# in the optree, and it is the difference between a callee resolved by
# name and one read out of a pad slot. Still `entersub`, as every call
# form in this tier is -- which is exactly why the op lint cannot
# distinguish the forms and the corpus needs one file per form.
#
# The empty call is the second statement rather than a file of its own.
# `$code->()` with no arguments is where a parser that treated `->(` as
# an operator taking an operand would refuse, and it costs one line.
#
# The token facts pin the arrow-call spelling NEGATIVELY, which is the
# only form available: `->` appears twice here, and the count vocabulary
# is `one` or `no`. What they assert is that the arrow is one token and
# not two -- neither a minus nor a greater-than is anywhere in this
# source, so a lexer that split `->` would fail both at once. The
# positive half is carried by the file parsing at all.

--- source
my $code = sub { return "c[@_]" };
print $code->(1, 2), "\n";
print $code->(), "\n";

--- expect output
c[1 2]
c[]

--- expect tokens
no operator whose text is "-"
no operator whose text is ">"

--- expect parses
