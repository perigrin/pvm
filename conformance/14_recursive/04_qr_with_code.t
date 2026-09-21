#!perl
# A `qr//` carrying a code block freezes the embedded program into a
# value: the code travels with the compiled pattern and runs wherever the
# object is later matched.
#
# TIER 14 recursive
# INTRODUCES qr// with embedded code
# USES nothing from a later tier
#
# Tier 09 claims plain `qr//` and says it claims only the non-recursive
# form, leaving this tier "only what embedding adds". Measured, what
# embedding adds to the `qr` op itself is NOTHING: `qr/a(?{ $n = 7 })b/`
# compiles to one `qr` op whose pattern string carries the block, exactly
# as `qr/abc/` compiles to one `qr` op. The two are indistinguishable in
# the op stream by anything but their pattern text.
#
# What changes is downstream, and it is still tier 09's op. Matching
# against the compiled object emits `regcomp` then `match`, where matching
# a literal pattern emits `match` alone -- which tier 09 already documents
# as the cost of interpolation, not of embedding.
#
# So this file's whole subject is that a value can CARRY Perl code across
# statements. The assignment and the match are separated on purpose: the
# block is written on line 2 and runs on line 4, and `$n` proves it.
#
# The `qr` object is not printed. Its stringification would embed the
# block's source, which is deterministic, but printing `$n` is the claim
# this file is actually making.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $n=0; my $r=qr/a(?{ $n = 7 })b/; my $s="ab"; $s =~ $r; print "$n\n";'
#   7

--- source
my $n = 0;
my $r = qr/a(?{ $n = 7 })b/;
my $s = "ab";
$s =~ $r;
print "$n\n";

--- expect output
7

--- expect parses

--- expect tokens
one quote-like operator whose text is "qr/a(?{ $n = 7 })b/"
one operator whose text is "=~"
no numeric literal whose text is "7"
