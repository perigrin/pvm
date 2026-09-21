#!perl
# `<*.pattern>` in term position is a GLOB, not a pair of comparisons, and
# it compiles to `glob` rather than to `readline`.
#
# TIER 13 opaque
# INTRODUCES the angle-bracket glob
# USES nothing from a later tier
#
# A glob reads the filesystem, so the pattern has to be one whose result is
# the same everywhere. `*.nonexistent-xyz` is that pattern: measured, it
# matches nothing in any directory the runner might sit in, so the array is
# empty and `scalar(@none)` is 0 whatever the working directory. A file
# globbing `*.t` would print a different number depending on where it ran.
#
# The token fact is what makes this file a glob test rather than a count
# test. `<*.nonexistent-xyz>` and `<DATA>` are the SAME token category --
# our lexer cannot tell a handle name from a pattern without knowing what
# the name means, which is a parsing question -- so this asserts the
# angle-bracket term and 07 asserts the other use of it. perl separates
# them at the optree, where this is `glob` and that is `readline`.
#
# Note `scalar` emits no op: perl imposes scalar context on `@none` at
# compile time and the array is simply evaluated there.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my @none = <*.nonexistent-xyz>; print "got ", scalar(@none), "\n";'
#   got 0

--- source
my @none = <*.nonexistent-xyz>;
print "got ", scalar(@none), "\n";

--- expect parses

--- expect output
got 0

--- expect tokens
one readline operator whose text is "<*.nonexistent-xyz>"
