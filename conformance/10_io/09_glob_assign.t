#!perl
# `*name` is a SIGIL introducing a fourth namespace, and it is the same
# character as multiplication.
#
# TIER 10 io
# INTRODUCES nothing of its own
# USES rv2gv from this tier and anoncode from 07_subroutines
#
# The typeglob did not appear in the corpus at all. Measured across 1,141
# lines of source before this file, `*` occurred three times and every
# one was a `sprintf` width specifier -- `"%0*d"`, `"%*d"`, `"%-*d"`.
# Zero typeglobs.
#
# So the corpus could not tell a lexer that treats `*` as always
# multiplication from one that gets it right. That is a one-character
# lexical fork with no coverage, and the fork is resolved by POSITION --
# a sigil where a term is expected, an operator where one just ended,
# which is the expect-state mechanism the spec describes for `%`, `<`,
# `&` and `/`.
#
# PLACEMENT IS TIER 10 RATHER THAN TIER 08, and the issue that asked for
# this file guessed wrong. Measured, `*alias = sub {...}` emits `rv2gv`,
# which `10_io` claims for its filehandles -- a glob and a filehandle are
# the same thing to perl, which is the whole reason `open(my $fh, ...)`
# and `*STDOUT` live in one namespace. Tier 08 owns references and could
# not have this.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $n = 3; *sq = sub { $_[0] * $_[0] };
#              print sq($n), " ", $n * 2, "\n"'
#   9 6
#
# BOTH READINGS OF `*` IN ONE STATEMENT. The sigil installs `sq` into the
# symbol table; the two multiplications inside and after it are ordinary
# arithmetic. A lexer that resolved `*` by looking only at the character
# produces either a syntax error or a multiplication where the
# installation belongs, and `9 6` is unreachable either way.
#
# The installed sub is called by NAME on the next line, which is what
# makes the installation observable: a parser could accept the assignment
# and do nothing, and `sq($n)` would then be a call to an undefined sub.
#
# THE TOKEN FACT CANNOT BE ABOUT THE SIGIL, and that is itself the
# measurement. Our lexer emits `Operator("*")` for all three occurrences
# -- the glob sigil and both multiplications -- so no count of `*`
# separates them and a fact naming it would be a claim about the file's
# punctuation. The GLOSSARY has no typeglob category to assert instead.
#
# What the file can claim is the shape around it: `sub` appears exactly
# once, in the anonymous constructor. The behavioural claim -- `9 6`,
# which is unreachable unless the installation happened -- is what
# carries the rest.

--- source
my $n = $ENV{X} // 3;
*sq = sub { $_[0] * $_[0] };
print sq($n), " ", $n * 2, "\n";

--- expect output
9 6

--- expect parses

--- expect tokens
one word whose text is "sub"
