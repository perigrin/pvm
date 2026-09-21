#!perl
# `\&twice` on a NAMED sub is the only construct in this tier that emits
# `rv2cv`, and `$c->(21)` calls the result through tier 07's `entersub`.
#
# TIER 08 references
# INTRODUCES the code reference and the call through it
# USES sub, return, print
#
# This is what pins the tier below 07 rather than at 03. `\&twice` needs
# a named sub to exist; a tier introducing `rv2cv` before subroutines
# existed would be claiming an op for a construct it could not write.
#
# `rv2cv` is narrower than it looks. It appears ONLY for `\&foo` on a
# named sub: `&{$r}()` and `&$r()` both compile to `entersub` with no
# `rv2cv` at all. So this file is the tier's only source for that op, and
# removing it would leave the README claiming an op nothing emits.
#
# The call itself is `entersub`, tier 07's, reached through a code
# reference rather than a name -- the same "different route to an earlier
# tier's op" this tier does with `rv2av` and `multideref`.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'sub twice { return $_[0] * 2 } my $c = \&twice; print $c->(21), "\n"'
#   42

--- source
sub twice { return $_[0] * 2 }
my $c = \&twice;
print $c->(21), "\n";

--- expect output
42

--- expect parses

--- expect tokens
one operator whose text is "\\"
one operator whose text is "->"
