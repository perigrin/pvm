#!perl
# `@{$r}` dereferences a scalar as a whole array, and the op it emits --
# `rv2av` -- is tier 02's, reached here by a route tier 02 does not have.
#
# TIER 08 references
# INTRODUCES the @{ } dereference
# USES my, print, scalar, the reference operator, interpolation
#
# This is one of the tier's two hard markers, `deref-at`, and it is the
# spelling tier 11's method bodies are written in.
#
# `@{$r}`, `@$r` and `$r->@*` all emit `rv2av` and are indistinguishable
# in the op stream. The op set cannot tell which was written, which is
# the recurring shape of this tier: the `rv2*` family collapses several
# spellings each. Only a token claim separates them -- see
# `10_deref_at_sigil.t` and `11_deref_postfix.t` for the other two.
#
# THE DEREF IS WRITTEN TWICE, and that is the point. The interpolated
# `"@{$r}\n"` is what makes the output deterministic: printing the list
# bare would run the three elements together as `102030`, which is true
# but reads as a bug. But inside a double-quoted string our lexer
# produces ONE `Quote` token and the dereference is not separately
# tokenised at all -- so a file with only the interpolated form asserts
# nothing about its own construct, and its `no operator whose text is
# "->"` fact is satisfied entirely by the `\@a` two lines above. The bare
# `my @c = @{$r};` is what gives the lexer the construct to tokenise, and
# `no variable whose text is "@{$r}"` is the claim: measured, our lexer
# emits `DerefSigil(@) Operator({) Variable($r) CloseBracket(})`, and a
# lexer that emitted the single token `Variable("@{$r}")` instead would
# satisfy every behavioural pin in this tier.
#
# The copy also keeps the deref in the optree in a second place: the
# interpolation compiles through `join`, tier 03's, while the assignment
# is a plain `rv2av` feeding an `aassign`.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my @a = (10, 20, 30); my $r = \@a; my @c = @{$r}; print scalar(@c), " ", "@{$r}\n"'
#   3 10 20 30

--- source
my @a = (10, 20, 30);
my $r = \@a;
my @c = @{$r};
print scalar(@c), " ", "@{$r}\n";

--- expect output
3 10 20 30

--- expect parses

--- expect tokens
one operator whose text is "\\"
no operator whose text is "->"
no variable whose text is "@{$r}"
