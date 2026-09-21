#!perl
# `$r->@*` is the THIRD spelling of the same `rv2av`, and the only one of
# the three that puts an arrow in front of a sigil.
#
# TIER 08 references
# INTRODUCES the postfix dereference
# USES my, print, scalar, the reference operator, array element access
#
# The README measures `@{$r}`, `@$r` and `$r->@*` as emitting one op each
# and being indistinguishable in the op stream. This file is the proof
# rather than the assertion: its optree and `10_deref_at_sigil.t`'s are
# identical op for op, differing only in the subscript constant the last
# statement reads.
#
# MEASURED perl 5.42.0, the assignment statement:
#
#   f  <0> pushmark s
#   g  <0> padsv[$r:2,4] s
#   h  <1> rv2av[t5] lK/1
#   i  <0> pushmark s
#   j  <0> padav[@c:3,4] lRM*/LVINTRO
#   k  <2> aassign[t6] vKS/COM_AGG
#
# So the tokens carry the whole distinction, and they carry it in a
# direction the other two spellings do not go. `@*` is a SIGIL AND A
# STAR after an arrow, which a lexer may reasonably read as the glob
# `*` -- measured, ours emits `Operator(->) Variable(@*)` and a lexer
# that emitted `Operator(->) Operator(@) Operator(*)` would parse as a
# multiplication and return no Unknown at all. The `one operator whose
# text is "->"` claim pairs with `04_arrow_deref.t`'s identical one to
# say the arrow survived; the `no variable whose text is "$r->@*"` claim
# says the lexer did not swallow the construct whole.
#
# Postfix dereference has been stable since 5.24 and is not experimental;
# `perlref` documents it under "Postfix Dereference Syntax".
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my @a = (10, 20, 30); my $r = \@a; my @c = $r->@*; print scalar(@c), " ", $c[2], "\n"'
#   3 30

--- source
my @a = (10, 20, 30);
my $r = \@a;
my @c = $r->@*;
print scalar(@c), " ", $c[2], "\n";

--- expect output
3 30

--- expect parses

--- expect tokens
one operator whose text is "->"
no variable whose text is "$r->@*"
