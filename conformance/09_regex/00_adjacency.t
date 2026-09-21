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
# through a bracketing delimiter, a match through a non-bracketing one, a
# match through the comment character, a match whose pattern NESTS its own
# delimiter, an interpolated match, a `qr//`, and four substitutions
# delimited by brackets, slashes, hashes and nested brackets -- then one
# print that reads all nine results.
#
# EVERY DELIMITER FORM THE TIER TEACHES APPEARS HERE, which is the part
# that needed fixing. The file shipped holding `m{}`, `s{}{}` and `qr//`
# alone -- so the one bug this file exists to reach was unreachable for
# every form most likely to produce it. A lexer that scans to the next `/`
# mis-terminates a `s/c/y/` sitting after a `s{a}{z}`, and nothing here
# would have run that pair.
#
# Three of the added forms are the ones that cannot NEST. `m!abc!`,
# `s/c/y/` and `s#d#w#` each close with the character they opened with, so
# a lexer has no bracket depth to count and must simply stop at the next
# occurrence. The fourth is the opposite case and is here for the
# contrast: `m{a{b}c}` and `s{a{b}c}{ok}` carry their own opening brace
# inside the pattern, where a lexer MUST count depth and stopping at the
# first `}` truncates the pattern. That pair of rules is what GLOSSARY.md
# records, and neither `m{zzz}` nor `s{a}{z}` above exercises either half.
#
# The hash forms are the sharpest pair in the file. `#` is Perl's comment
# character, so a lexer that has not yet recognised the `m` or `s` has
# already discarded the rest of the line. Putting them BETWEEN other
# delimiter forms rather than alone is the adjacency claim in its strongest
# version: recovering from `m#abc#` is not enough if the `s#d#w#` three
# statements later is then read as a comment.
#
# This tier DEPENDS ON 01_literals, and the pairing is present rather than
# decorative: every pattern here is matched against a string literal bound
# to a pad slot, and the final print interpolates nine of them into one
# double-quoted string. Pairing with 08_references instead would assert
# nothing, which is the failure mode the spec warns about.
#
# The substitutions run LAST on purpose. They mutate `$s`, which the five
# matches above them read; running any earlier would make those results
# depend on statement order in a way that hides a mis-parse behind a
# plausible-looking output. Each of the three that touch `$s` mutates a
# DIFFERENT character -- `a`, `c` and `d` -- so the final `zbyw` records
# that all three ran, where two substitutions of one character would leave
# the second's failure invisible. The nested substitution needs its own
# target, `$n`, because its pattern is five characters `$s` does not hold.
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/09_regex/00_adjacency.t
#   1 1 1 1 1 1 (?^:abc) zbyw ok

--- source
my $s = "abcd";
my $p = "b";
my $hit = $s =~ /abc/;
my $miss = $s !~ m{zzz};
my $bang = $s =~ m!abc!;
my $hash = $s =~ m#abc#;
my $nest = "a{b}c" =~ m{a{b}c};
my $interp = $s =~ /$p/;
my $re = qr/abc/;
$s =~ s{a}{z};
$s =~ s/c/y/;
$s =~ s#d#w#;
my $n = "a{b}c";
$n =~ s{a{b}c}{ok};
print "$hit $miss $bang $hash $nest $interp $re $s $n\n";

--- expect output
1 1 1 1 1 1 (?^:abc) zbyw ok

--- expect parses

--- expect tokens
one quote-like operator whose text is "m{zzz}"
one quote-like operator whose text is "m!abc!"
one quote-like operator whose text is "m#abc#"
one quote-like operator whose text is "m{a{b}c}"
one quote-like operator whose text is "s{a}{z}"
one quote-like operator whose text is "s/c/y/"
one quote-like operator whose text is "s#d#w#"
one quote-like operator whose text is "s{a{b}c}{ok}"
one quote-like operator whose text is "qr/abc/"
