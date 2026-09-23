#!perl
# An explicit `0o` prefix is octal too: `0o377` is one token denoting 255,
# the same value `0377` denotes.
#
# TIER 01 literals
# INTRODUCES numeric literal
# USES nothing from a later tier
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $x = 0o377; print "$x\n"'
#   255
#
# Perl 5.34 added this spelling so octal need not be signalled by a
# leading zero alone. It is a separate file from `08_octal_leading_zero.t`
# because it is a separate LEXICAL form -- the two agree on the value and
# compile to the same `const[IV 255]`, so the optree cannot tell them
# apart and only the token stream can say which was written.
#
# The trap is the same as `05_hexadecimal.t`'s: `o377` is a legal
# identifier, so a lexer that stops the number at the first non-digit
# produces `Number(0) Word(o377)` and the parser sees a number beside a
# bareword rather than an error.

--- source
my $x = 0o377;
print "$x\n";

--- expect parses

--- expect output
255

--- expect tokens
one numeric literal whose text is "0o377"
no word whose text is "o377"
