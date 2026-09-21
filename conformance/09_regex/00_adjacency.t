#!perl
# Every construct this tier introduces, in one body, each adjacent to
# another.
#
# TIER 09 regex
# INTRODUCES nothing of its own
# USES not, from tier 04
#
# The tier's other files are one construct each, which is what makes them
# diagnosable. That same property is why a corpus of such files cannot
# reach an ADJACENCY bug -- and in a tier whose operand is not Perl, the
# adjacency bug is the one to expect. A lexer that delimits a pattern by
# scanning for the next `/` is correct on every file here taken alone and
# wrong the moment a `s{a}{z}` sits between two matches.
#
# So the constructs sit consecutively: a constant match, a negated match
# through a bracketing delimiter, an interpolated match, a `qr//`, and a
# substitution with two bracketing pairs -- then one print that reads all
# five results.
#
# This tier DEPENDS ON 01_literals, and the pairing is present rather than
# decorative: every pattern here is matched against a string literal bound
# to a pad slot, and the final print interpolates five of them into one
# double-quoted string. Pairing with 08_references instead would assert
# nothing, which is the failure mode the spec warns about.
#
# The substitution runs LAST on purpose. It mutates `$s`, which the three
# matches above it read; running it earlier would make their results
# depend on statement order in a way that hides a mis-parse behind a
# plausible-looking output.
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/09_regex/00_adjacency.t
#   1 1 1 (?^:abc) zbc

--- source
my $s = "abc";
my $p = "b";
my $hit = $s =~ /abc/;
my $miss = $s !~ m{zzz};
my $interp = $s =~ /$p/;
my $re = qr/abc/;
$s =~ s{a}{z};
print "$hit $miss $interp $re $s\n";

--- expect output
1 1 1 (?^:abc) zbc

--- expect parses

--- expect tokens
one quote-like operator whose text is "m{zzz}"
one quote-like operator whose text is "s{a}{z}"
one quote-like operator whose text is "qr/abc/"
