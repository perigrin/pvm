#!perl
# A bareword handle reaches `print` by a different first op than a lexical
# one and by the SAME second: `gv[*STDOUT]` then `rv2gv`, where the
# lexical form has `padsv[$fh]` then `rv2gv`.
#
# TIER 10 io
# INTRODUCES printing to a bareword filehandle
# USES nothing from a later tier
#
# MEASURED perl 5.42.0:
#
#   print STDOUT "x"   gv[*STDOUT] rv2gv sKR/1 const print vKS
#   print $fh    "x"   padsv[$fh]  rv2gv sKR/1 const print vKS
#
# They converge one op later. A file using only lexical handles therefore
# never emits `gv`, and this tier's declared set is the union across its
# files rather than a property of any one of them. `gv` itself is tier
# 02's op, claimed there with the package scalar; reaching it through a
# bareword handle is a use, not an introduction.
#
# STDERR is the other bareword this tier cares about and it is NOT written
# to here, for a reason the format forces: `--- expect output` is compared
# against stdout, and a line sent to STDERR appears in neither stream the
# runner reads. It would be an assertion about bytes nothing checks. The
# op stream is identical either way -- `gv[*STDERR] rv2gv print` against
# `gv[*STDOUT] rv2gv print` -- so STDOUT measures the construct and stays
# observable.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'print STDOUT "to stdout\n"'
#   to stdout

--- source
print STDOUT "to stdout\n";

--- expect output
to stdout

--- expect parses
