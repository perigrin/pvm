#!perl
# `${x}` is ONE variable: the braces are punctuation around a name, not a
# block and not a dereference.
#
# TIER 02 variables
# INTRODUCES braced variable name
# USES nothing from a later tier
#
# GLOSSARY.md says this outright under `variable` -- "`${name}` is one
# variable: the braces are punctuation around a name. But `${ $ref }` is
# NOT one token, because its contents are an expression requiring a
# parser." The two spellings differ by what is inside the braces and by
# nothing else, which makes this the tier's sharpest lexing question: a
# lexer that sees `${` and commits to a dereference is wrong here, and a
# lexer that sees `${` and commits to a name is wrong at tier 08.
#
# This file takes the half that belongs to this tier. `${ $ref }` is tier
# 08's, and writing it here would be the file reaching forward.
#
# Both sigils are braced because the brace rule is about the NAME rather
# than about scalars: `@{a}` is the same array as `@a`, and a lexer that
# special-cased `${` would pass a file that only wrote the scalar form.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $x = 42; my @a = (7, 8); print ${x}, "\n"; print scalar(@{a}), "\n"'
#   42
#   2
#
# The ops are `padsv` and `padav`, identical to the unbraced spellings:
# perl resolves the braces away entirely, so the optree cannot tell this
# file from one without them. The claim is a LEXICAL one and only the
# source records it -- the same situation tier 01 recorded for `-1`, whose
# two tokens fold to one constant.

--- source
my $x = 42;
my @a = (7, 8);
print ${x}, "\n";
print scalar(@{a}), "\n";

--- expect output
42
2

--- expect parses
