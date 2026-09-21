#!perl
# `m{abc}` and `s{a}{z}` emit op streams BYTE-IDENTICAL to `/abc/` and
# `s/a/z/`. Delimiters are invisible to the optree, so this file's whole
# measurement is its token claims.
#
# TIER 09 regex
# INTRODUCES bracketing and punctuation delimiters for m// and s///
# USES nothing from a later tier
#
# This is the same argument `conformance/README.md` makes for `5e-1`: when
# two spellings compile to the same tree and print the same bytes, only an
# assertion about the token stream distinguishes them. A delimiter file
# asserting only ops would measure nothing `01_bare_match.t` and
# `04_substitution.t` do not already measure, and would go green against a
# lexer that silently normalised every delimiter to `/`.
#
# `s{a}{z}` is the case where the second pair's opening delimiter is free
# to differ from the first's, which GLOSSARY.md records as true only when
# the first pair is bracketing. `m!abc!` is the non-bracketing case, where
# one character delimits both ends and nothing nests.
#
# MEASURED perl 5.42.0 -- the op stream for
#   perl -MO=Concise,-exec -e 'my $m = "abc" =~ m{abc}; print "$m\n"'
# is identical to the one for `/abc/`, down to the pattern text perl
# prints inside the op: match(/"abc"/) sKS.
#
#   $ perl -e 'my $a = "abc" =~ m{abc}; my $b = "abc" =~ m!abc!; my $s = "abc"; $s =~ s{a}{z}; print "$a$b$s\n"'
#   11zbc

--- source
my $a = "abc" =~ m{abc};
my $b = "abc" =~ m!abc!;
my $s = "abc";
$s =~ s{a}{z};
print "$a$b$s\n";

--- expect output
11zbc

--- expect parses

--- expect tokens
one quote-like operator whose text is "m{abc}"
one quote-like operator whose text is "m!abc!"
one quote-like operator whose text is "s{a}{z}"
no operator whose text is "{"
