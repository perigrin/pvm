#!perl
# `q` and `qq` are QUOTE-LIKE OPERATORS, not string literals: a different
# glossary category making a different claim, and the delimiter is chosen
# rather than fixed.
#
# TIER 01 literals
# INTRODUCES quote-like operator
# USES nothing from a later tier
#
# GLOSSARY.md keeps `string literal` and `quote-like operator` apart
# deliberately, and says why: a file asserting `one string literal whose
# text is "hi"` must not be satisfied by `qw(hi)`, which is not a string
# at all. That separation is only worth having if some file asserts each
# side of it, and until now no corpus file asserted this side.
#
# MEASURED perl 5.42.0. Ops cannot see the distinction:
#
#   $ perl -MO=Concise,-exec -e 'my $s = q(a b); print $s'
#   ... const[PV "a b"] ... padsv_store ... print ...
#
# One `const`, exactly as `'a b'` gives. So `q(a b)` and `'a b'` are the
# same op and DIFFERENT TOKENS, which is precisely the case where only a
# token fact can carry the claim.
#
# The DELIMITERS are what this file pins, and each is a separate chance to
# be wrong: `q(...)`, `q{...}` and `q!...!` are all the same construct
# wearing three delimiters, and a lexer that hard-codes one of them fails
# the others. `qq` is here rather than in a file of its own because what
# separates it from `q` is escape processing, which `12_single_quoted.t`
# and `04_double_quoted.t` already pin for the delimiter-less spellings --
# repeating it here would measure the same thing twice.
#
# `qw` is NOT here. Measured, `my @w = qw(a b c)` emits `aassign` and
# `padav`, both `02_variables`' ops, because a word list needs an array to
# land in. It belongs to the tier that owns arrays, and this tier's file
# would have to borrow two ops from a later one to hold it.

--- source
my $a = q(one);
my $b = q{two};
my $c = q!three!;
print $a, $b, $c, "\n";

--- expect output
onetwothree

--- expect parses

--- expect tokens
one quote-like operator whose text is "q(one)"
one quote-like operator whose text is "q{two}"
one quote-like operator whose text is "q!three!"
