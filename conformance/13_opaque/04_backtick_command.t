#!perl
# `qx{...}` runs a shell command and returns its output, emitting `backtick`
# over the command string.
#
# TIER 13 opaque
# INTRODUCES the command quote
# USES nothing from a later tier
#
# This file RUNS A SHELL COMMAND, and the runner compares bytes, so the
# command has to be one whose output does not depend on the machine, the
# clock or the directory. `echo hi` is that command: measured, it writes
# the three bytes `hi\n` including the trailing newline, on every system
# with a POSIX shell. A corpus file calling `date` or `hostname` would be
# unreproducible and must not be written.
#
# `qx` is the one construct in this tier whose own op is visible, and even
# that op is shared: `` `echo hi` `` compiles to the same `backtick`, which
# is why GLOSSARY.md puts the two spellings in one category.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $out = qx{echo hi}; print $out;' | od -c
#   0000000   h   i  \n
#   0000003

--- source
my $out = qx{echo hi};
print $out;

--- expect parses

--- expect output
hi

--- expect tokens
one quote-like operator whose text is "qx{echo hi}"
