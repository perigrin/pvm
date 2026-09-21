#!perl
# Every construct this tier introduces, in one body, each adjacent to
# another.
#
# TIER 10 io
# INTRODUCES nothing of its own
# USES nothing from a later tier
#
# The tier's other files are one construct each, which is what makes them
# diagnosable. That same property is why such a corpus cannot reach an
# ADJACENCY bug: a parser handling every construct alone and mis-handling
# a pair goes green over the pair.
#
# Here the pairs that matter are the ones a one-construct file cannot
# make. A scalar readline followed by a LIST readline on the SAME handle
# is the whole of this tier's context dependency in two adjacent
# statements -- the second reads what the first left, so the printed `2`
# is only correct if both contexts were resolved and resolved
# differently. `02_readline_scalar.t` and `03_readline_list.t` each open
# a fresh handle and so can never disagree about position.
#
# The other pairing is `eof` inside a print LIST that also holds an
# array, which puts the tier's own op in an argument position rather than
# alone in a statement.
#
# This tier depends on `03_context`, and the pairing with it is the
# scalar-versus-list readline itself: context is not a separate construct
# to place beside this one, it is the thing selecting which readline
# happens.
#
# The ops the adjacency reaches beyond this tier's own, all claimed
# earlier and none new: `gv` and `padav` and `aassign` (02), `cond_expr`
# and `goto` from the ternary (06), `srefgen` from `\my $buf` (08).
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/10_io/00_adjacency.t
#   first
#   2 left, eof yes
#   round trip
#
# `expect output` is written before `expect parses` rather than last: the
# blank line after it carries the output's trailing newline, and a blank
# line at END of file is what `end-of-file-fixer` strips.

--- source
open(my $in, "<", \"first\nsecond\nthird\n");
my $head = <$in>;
my @rest = <$in>;
print STDOUT $head;
print STDOUT scalar(@rest), " left, eof ", (eof($in) ? "yes" : "no"), "\n";
close($in);
open(my $out, ">", \my $buf);
print $out "round trip\n";
close($out);
print $buf;

--- expect output
first
2 left, eof yes
round trip

--- expect parses
