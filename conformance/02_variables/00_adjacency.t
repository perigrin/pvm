#!perl
# Every construct this tier introduces, in one body, each adjacent to
# another, and adjacent to tier 01's literals.
#
# TIER 02 variables
# INTRODUCES nothing of its own
# USES nothing from a later tier
#
# The tier's other ten files are one construct each, which is what makes
# them diagnosable: when `06_hash_element_expr.t` refuses, the construct
# that refused is the only one present. That same property is why a corpus
# of such files cannot reach an ADJACENCY bug -- a parser that handles
# every construct alone and mis-handles a pair goes green over the pair.
#
# This tier's adjacency has a specific shape the one-construct files cannot
# have. Four of the tier's ops are the SAME construct spelled four ways --
# a constant subscript, lexical or package, read or written -- and the
# files that isolate them each see one. Here all four sit in one body, so a
# parser that collapses `$a[0]` read onto `$a[0]` written, or a lexical
# array onto a package one, is visible.
#
# DEPENDS ON 01_literals, so the pairing is real and not internal: the
# literals are what the variables hold. `qw(x y)` binds into an array,
# `"$w[0]-$w[1]"` interpolates two elements back out, and the numbers and
# strings of tier 01 are the values every subscript here returns.
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/02_variables/00_adjacency.t
#   7plain23x-y
#   70.53232
#   782110
#
# The runs are unseparated because `$,` is unset and each `print` gets its
# arguments as a list, the same reason tier 01's adjacency file prints
# `qw(a b c)` as `abc`. Line by line: `$a[0]` `$a[$#a]` `$#a` `scalar(@a)`
# `$tag`; then `@a[0,1]` `$h{c}` `$h{$k[0]}` `@h{"c"}` `scalar(keys %h)`;
# then the three package variables.
#
# `print $a[0]` rather than `print "$a[0]"` throughout: the interpolated
# subscript emits `stringify` and an interpolated `@a` emits `join` plus a
# `gvsv` for `$"`, none of which this tier claims. `$tag` is built by a
# separate statement so the interpolation is tier 01's `multiconcat` over
# two already-fetched elements rather than a fetch inside the print.
#
# `expect output` is written before `expect parses` rather than last: the
# blank line after it carries the output's trailing newline, and a blank
# line at END of file is what `end-of-file-fixer` strips.

--- source
my @a = (42, 0.5, 'plain');
my @w = qw(x y);
my %h = (a => 1, b => 2, c => 3);
my @k = ("b");
$a[0] = 7;
delete @h{"a"};
my $tag = "$w[0]-$w[1]";
$::s = $a[0];
@::p = (8, 9);
%::g = (z => 10);
print $a[0], $a[$#a], $#a, scalar(@a), $tag, "\n";
print( (@a[0, 1]), $h{c}, $h{$k[0]}, (@h{"c"}), scalar(keys %h), "\n");
print $::s, $::p[0], scalar(@::p), scalar(keys %::g), $::g{z}, "\n";

--- expect output
7plain23x-y
70.53232
782110

--- expect parses
