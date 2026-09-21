#!perl
# Adding an `else` changes the op: a bare `if` emits `and`, an `if`/`else`
# emits `cond_expr` -- the same op the ternary `?:` emits.
#
# TIER 06 control
# INTRODUCES the two-armed conditional
# USES nothing from a later tier
#
# So one source-level keyword maps to two different ops depending on
# whether an `else` is present, and only one of them is new to this tier.
# The corpus needs both `01_if_postfix.t` and this file; neither alone
# lints the tier, and a parser that emits one shape for both spellings
# passes either file in isolation.
#
# The `else` arm is what brings back the `enter`/`leave` pair that the bare
# `if` does not emit: `cond_expr` jumps into a second block, and a second
# block is a scope.
#
# MEASURED perl 5.42.0:
#
#   $ perl -MO=Concise,-exec -e 'my $c = $ENV{X} // 0;
#     if ($c) { print "y\n" } else { print "n\n" }'
#   ...
#   8  <0> padsv[$c] s
#   9  <|> cond_expr(other->a) vK/1
#   a      <0> pushmark s
#   b      <$> const[PV "y\n"] s
#   c      <@> print vK
#              goto d
#   e  <0> enter v
#   f  <;> nextstate(main 6 -:1) v
#   g  <0> pushmark s
#   h      <$> const[PV "n\n"] s
#   i  <@> print vK
#   j  <@> leave vKP
#   d  <@> leave[1 ref] vKP/REFC
#
# And the erasure this file exists to avoid: `if (1) { print "y" } else {
# print "n" }` emits `pushmark`, `const` and `print` and NO branch op --
# the else arm is gone from the binary. Such a file would pass its output
# test while measuring none of this tier.

--- source
my $c = $ENV{X} // 0;
if ($c) { print "taken\n" } else { print "else\n" }

--- expect output
else

--- expect parses
