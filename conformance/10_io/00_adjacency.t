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
# THE THIRD PAIRING IS `select` WITH `say`, and it is the one that needs
# two constructs most. `say "round trip"` names NO HANDLE, and its bytes
# land in `$buf` rather than on stdout, because the `select($out)` above
# it changed where "no handle" means. Neither file alone can make that
# claim: `07_say.t` passes its handle explicitly, and `08_select.t`
# redirects a `print` rather than a `say`. Here the two constructs are
# the same assertion -- a parser that dropped either one puts `round
# trip` on stdout and leaves the buffer empty.
#
# The four-argument `select` follows, which is the other operator sharing
# that name: measured, `select($out)` emits `select` and
# `select(undef,undef,undef,0)` emits `sselect`, a different op reached
# by a different argument count. Both spellings are here because the
# tier's declared set holds both and an adjacency file must reach every
# op the tier introduces.
#
# This tier depends on `03_context`, and the pairing with it is the
# scalar-versus-list readline itself: context is not a separate construct
# to place beside this one, it is the thing selecting which readline
# happens.
#
# The ops the adjacency reaches beyond this tier's own, all claimed
# earlier and none new: `gv` and `padav` and `aassign` (02), `cond_expr`
# and `goto` from the ternary (06), `undef` from the syscall arguments
# (04), `srefgen` from `\my $buf` (08).
#
# `use feature "say"` is required and is not decoration: without it `say`
# is not this tier's op at all but a method call, which `07_say.t`
# measures. Measured, the pragma changes no op in this file beyond
# enabling `say` itself.
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/10_io/00_adjacency.t
#   first
#   2 left, eof yes
#   round trip
#   ready 0
#
# `expect output` is written before `expect parses` rather than last: the
# blank line after it carries the output's trailing newline, and a blank
# line at END of file is what `end-of-file-fixer` strips.

--- source
use feature "say";
open(my $in, "<", \"first\nsecond\nthird\n");
my $head = <$in>;
my @rest = <$in>;
print STDOUT $head;
print STDOUT scalar(@rest), " left, eof ", (eof($in) ? "yes" : "no"), "\n";
close($in);
open(my $out, ">", \my $buf);
my $prev = select($out);
say "round trip";
select($prev);
close($out);
my $ready = select(undef, undef, undef, 0);
print $buf, "ready $ready\n";

--- expect output
first
2 left, eof yes
round trip
ready 0

--- expect parses
