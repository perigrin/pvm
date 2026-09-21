#!perl
# A leading `v` makes a v-string of what follows, one dot or many:
# `v65.66.67` is ONE token denoting the three-character string `ABC`.
#
# TIER 01 literals
# INTRODUCES numeric literal
# USES nothing from a later tier
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $v = v65.66.67; print "$v\n"'
#   ABC
#
#   $ perl -e 'my $v = v5.42.0; print join(".", map ord, split //, $v), "\n"'
#   5.42.0
#
# The second measurement is the form this repository writes constantly --
# `v5.42.0` is how every `use` line spells a version -- and it is the same
# construct.
#
# STATUS refuses as of 7711154e. Our lexer produces
# Word("v65") Operator(".") Number("66.67"): `v65` is a legal identifier,
# so the word scanner claims it and the rest of the v-string is read as a
# concatenation of a bareword with a number. The PARSER cannot see this.
# It receives Word Operator Number, reads a valid expression, and returns
# ZERO Unknown nodes -- the same shape `08_signed_exponent.t` records, and
# the same reason the token layer exists: the assertion below is the only
# check that reaches it.
#
# No issue: this file is the record, on the same grounds as
# `08_signed_exponent.t` -- the construct was found by writing the file.

--- source
my $v = v65.66.67;
print "$v\n";

--- expect parses

--- expect output
ABC

--- expect tokens
no operator whose text is "."
no word whose text is "v65"
