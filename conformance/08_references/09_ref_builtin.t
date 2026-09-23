#!perl
# `ref` reports what a reference points at, and it is the one operation
# in this tier that has no home in an earlier one.
#
# TIER 08 references
# INTRODUCES ref
# USES my, print, the reference operator, brace dereference
#
# The tier description does not mention `ref`; the op measurement is
# where it came from. It is a reference operation whose operand is a
# reference and whose result is a plain string, so no earlier tier can
# claim it -- there is nothing for it to be about before this tier.
#
# It is also what every other file here leans on for determinism. A
# reference stringifies as `ARRAY(0x5606f0a12345)` and the address
# changes every run, so a corpus file can never print one. `ref` gives
# the stable category name instead, which is why it appears in
# `03_anonymous_array.t` as well.
#
# The three categories are measured together because the op is the same
# for all of them: `ref` does not discriminate by sigil, the referent
# does.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my @a = (1); my %h = (k => 1); my $n = 5; my $ar = \@a; my $hr = \%h; my $sr = \$n; print ref($ar), " ", ref($hr), " ", ref($sr), " ", ${$sr}, "\n"'
#   ARRAY HASH SCALAR 5

--- source
my @a = (1);
my %h = (k => 1);
my $n = 5;
my $ar = \@a;
my $hr = \%h;
my $sr = \$n;
print ref($ar), " ", ref($hr), " ", ref($sr), " ", ${$sr}, "\n";

--- expect output
ARRAY HASH SCALAR 5

--- expect parses

--- expect tokens
no operator whose text is "->"
one operator whose text is "{"
