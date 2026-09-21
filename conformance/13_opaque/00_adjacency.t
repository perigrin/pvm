#!perl
# Every construct this tier introduces, in one body, each adjacent to
# another.
#
# TIER 13 opaque
# INTRODUCES nothing of its own
# USES nothing from a later tier
# STATUS refuses as of 6c231691, and this was NOT a known gap. Cites this file.
#
# Three Unknowns for three distinct constructs -- the format declaration,
# the heredoc, and the data section -- which is what 06, 01 and 08 refuse
# on their own and nothing more. `qx`, the glob and the pod block pass here
# as they pass alone (04, 05, 07). So no Unknown appears at a JOIN between
# two constructs, which is the negative worth having: the lexer leaves each
# opaque region cleanly and the failures do not compound.
#
# Every token fact below passes. The lexer handles all six constructs
# adjacent to one another; the refusals are the parser's, and the token
# section is what locates them there.
#
# Adjacency matters more here than anywhere earlier, because every
# construct in this tier CHANGES THE LEXER'S MODE and then has to hand it
# back. A format body, a pod block and a heredoc body each suspend ordinary
# lexing, consume lines by a rule of their own, and resume. A file holding
# one of them tests that the mode was entered and left once. This file
# tests that leaving one mode leaves the lexer able to enter the next.
#
# The order is not arbitrary. The format body sits FIRST, before any
# statement, because its picture lines are the region least like Perl -- if
# resuming after them is wrong, everything downstream is wrong for a reason
# the single-construct files cannot name. The heredoc body and the pod
# block follow, so the heredoc's line-oriented terminator is immediately
# followed by another line-oriented terminator of a different kind. The
# data section is last because it has to be: it ends the program text, so
# nothing can follow it.
#
# This tier depends on 10_io, and the dependency is live here: the
# `__END__` section is read through the DATA filehandle, which is tier 10's
# readline. The spelling is `__END__` rather than `__DATA__` to cover the
# other half of 08's construct -- measured, both end the program text and
# both open DATA in package main.
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/13_opaque/00_adjacency.t
#   hi
#   heredoc line
#   glob 0
#   a report line
#   from the data section
#
# `write` prints between the two `print`s rather than at the end, so format
# output and print output share one ordered stream on STDOUT.
#
# The output ends with a BLANK LINE, and it is the corpus format's own: a
# section body runs to the next `--- ` marker including the blank separator
# before it, and inside the data section that blank line is data. See
# `08_data_section.t`, which records the same measurement with the reason.
#
# A HAZARD THIS FILE HAD TO AVOID, and it is a real limitation of the
# corpus format rather than a quirk of this file: `--- ` at the start of a
# line is recognised as a SECTION MARKER anywhere after the comment block,
# with no awareness of Perl's own nesting. A heredoc body, a pod block or a
# data section whose content held such a line would be silently cut in two
# and the halves read as separate sections. The bodies here are written
# clear of it. Tier 13 is the first tier able to hit this, because it is
# the first with multi-line opaque regions, and no other tier can express
# the case that would break.
#
# `expect output` is written before `expect parses` rather than last, so
# the blank line carrying the output's trailing newline sits in the middle
# of the file where `end-of-file-fixer` does not strip it.

--- source
format STDOUT =
a report line
.

my $shell = qx{echo hi};
my $here = <<"EOT";
heredoc line
EOT

=pod

Prose the lexer eats and the optree never sees.

=cut

my @none = <*.nonexistent-xyz>;
my @data = <DATA>;
print $shell, $here, "glob ", scalar(@none), "\n";
write;
print @data;
__END__
from the data section

--- expect output
hi
heredoc line
glob 0
a report line
from the data section


--- expect parses

--- expect tokens
one format body whose text is "a report line\n.\n"
one heredoc opener whose text is "<<\"EOT\""
one heredoc body whose text is "heredoc line\nEOT\n"
one pod block whose text is "=pod\n\nProse the lexer eats and the optree never sees.\n\n=cut\n"
one quote-like operator whose text is "qx{echo hi}"
one readline operator whose text is "<*.nonexistent-xyz>"
one readline operator whose text is "<DATA>"
one data section whose text is "__END__\nfrom the data section\n\n"
