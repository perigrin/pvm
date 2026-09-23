#!perl
# `AUTOLOAD` is a sub name that is special to the LANGUAGE rather than to
# the program: perl calls it when a named sub does not exist.
#
# TIER 10 io
# INTRODUCES nothing of its own
# USES entersub from 07_subroutines and the symbol table from 09_glob_assign.t
#
# The third of this tier's symbol-table slice, and the one where the
# table is read implicitly. `09_glob_assign.t` installs a name and
# `10_glob_ref.t` references one; here perl fails to find a name and
# falls back, which is the same table seen from the other side.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'our $AUTOLOAD;
#              sub AUTOLOAD { my $n = $AUTOLOAD; $n =~ s/.*:://;
#                             return "auto:$n" }
#              print missing(), "\n"'
#   auto:missing
#
# `missing` is never defined. The call finds nothing in the symbol table,
# perl dispatches to `AUTOLOAD` instead, and `$AUTOLOAD` holds the fully
# qualified name that was sought -- `main::missing`, which the
# substitution trims to `missing`.
#
# WHAT A PARSER GETS WRONG HERE is the call itself. `missing()` has no
# declaration anywhere in the file, so a parser that resolves calls at
# parse time has nothing to resolve; one that treats an unknown bareword
# followed by parens as a call gets it right and defers the question to
# runtime, which is what perl does. `07_subroutines/10_undeclared_callee.t`
# makes the same claim for an ordinary sub; the difference here is that
# the sub genuinely does not exist and the program still works.
#
# THE TOKEN FACT COUNTS ONE `AUTOLOAD` IN A SOURCE THAT SPELLS IT THREE
# TIMES, and that is the claim rather than an oversight. Two of the three
# are `$AUTOLOAD` -- a Variable -- and only the sub name is a bare Word.
# So the fact separates the special sub name from the special variable
# that carries its argument, which are different things wearing one
# spelling. A lexer that let the sigil fall off, or that read the bare
# name as a variable, fails it.
#
# The `s///` is tier 09's op and reached rather than introduced -- the
# file needs it because `$AUTOLOAD` arrives package-qualified and the
# unqualified name is what makes the output readable as a claim.

--- source
our $AUTOLOAD;
sub AUTOLOAD { my $n = $AUTOLOAD; $n =~ s/.*:://; return "auto:$n" }
print missing(), "\n";

--- expect output
auto:missing

--- expect parses

--- expect tokens
one word whose text is "AUTOLOAD"
