#!perl
# Plain `qr/abc/` emits a single `qr` op and nothing else: no `regcomp`,
# no `match`, no block, no recursion. It is closer to a literal than to
# anything tier 14 does.
#
# TIER 09 regex
# INTRODUCES qr// in its non-recursive form
# USES nothing from a later tier
#
# Tier 14 owns `qr//` WITH EMBEDDED CODE, which is a re-entry problem.
# Claiming the plain form there would make tier 14 depend on this tier for
# the non-recursive half of one construct, so it stays here and tier 14
# claims only what embedding adds.
#
# Printing the object is safe because its stringification is deterministic
# -- and it is `(?^:abc)`, NOT `(?^u:abc)`. The `u` appears only under a
# unicode_strings-like feature bundle; this file runs without one.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $r = qr/abc/; print "$r\n"'
#   (?^:abc)

--- source
my $r = qr/abc/;
print "$r\n";

--- expect output
(?^:abc)

--- expect parses

--- expect tokens
one quote-like operator whose text is "qr/abc/"
