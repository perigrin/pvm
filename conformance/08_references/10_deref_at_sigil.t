#!perl
# `@$r` is the SAME op as `@{$r}` -- `rv2av` -- and the op stream cannot
# tell which was written. Only the tokens can.
#
# TIER 08 references
# INTRODUCES the sigil-only @ dereference
# USES my, print, scalar, the reference operator, array element access
#
# This file exists because `06_deref_at.t` cannot make the claim. That
# file interpolates `@{$r}` inside a double-quoted string, where the
# whole string is one token and the dereference is not separately
# tokenised at all -- its `no operator whose text is "->"` fact is
# satisfied by the `\@a` on the line above and says nothing about the
# construct the file is named for.
#
# The claim here is `no variable whose text is "@$r"`. Measured, our
# lexer produces two tokens, `DerefSigil(@) Variable($r)`, and a lexer
# that produced the single token `Variable("@$r")` instead would satisfy
# every behavioural pin in this tier: perl prints the same bytes and
# compiles the same ops. The README's own measurement is that `@{$r}`,
# `@$r` and `$r->@*` "all emit `rv2av` and are indistinguishable in the
# op stream", so this is the one check that can separate them.
#
# MEASURED perl 5.42.0, the assignment statement -- identical to
# `11_deref_postfix.t`'s except for the subscript constant:
#
#   f  <0> pushmark s
#   g  <0> padsv[$r:2,4] s
#   h  <1> rv2av[t5] lK/1
#   i  <0> pushmark s
#   j  <0> padav[@c:3,4] lRM*/LVINTRO
#   k  <2> aassign[t6] vKS/COM_AGG
#
# Copying into a named array rather than printing the list keeps `join`
# in tier 03 where it belongs: printing `@$r` bare would run the three
# elements together as `102030`, which is true and reads as a bug.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my @a = (10, 20, 30); my $r = \@a; my @c = @$r; print scalar(@c), " ", $c[0], "\n"'
#   3 10

--- source
my @a = (10, 20, 30);
my $r = \@a;
my @c = @$r;
print scalar(@c), " ", $c[0], "\n";

--- expect output
3 10

--- expect parses

--- expect tokens
one operator whose text is "\\"
no operator whose text is "->"
no variable whose text is "@$r"
