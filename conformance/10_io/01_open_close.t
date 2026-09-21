#!perl
# `open(my $fh, "<", ...)` autovivifies a glob into the fresh lexical: the
# handle is not a value the scalar receives, it is a glob `rv2gv` builds
# into the scalar's slot.
#
# TIER 10 io
# INTRODUCES opening and closing a filehandle
# USES nothing from a later tier
#
# The three-argument open takes a scalar ref as its file argument, and the
# result is an in-memory handle reading from that string. Nothing touches
# the filesystem, which is what makes this file reproducible on a machine
# whose `/etc/hostname` is not ours.
#
# MEASURED perl 5.42.0, the in-memory form against a real path:
#
#   open(my $fh, "<", \"x\n")        padsv rv2gv const const open
#   open(my $fh, "<", "/etc/passwd") padsv rv2gv const const open
#
# The same five ops. `\"x\n"` is constant-folded to `const[IV \"x\n"]
# s/FOLD` at compile time, so the ref is built by the optimiser rather
# than by an `srefgen` in the op stream. The in-memory handle is the
# ordinary open as far as the op table can see.
#
# `or die` is deliberately absent: `die` emits an op of its own, and no
# tier claims it. A check on the return value would widen this tier's
# declared set by an op that is not this tier's subject, so the open goes
# unchecked and the file prints a constant to show it ran.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'open(my $fh, "<", \"x\n"); print "opened\n"; close($fh)'
#   opened

--- source
open(my $fh, "<", \"x\n");
print "opened\n";
close($fh);

--- expect output
opened

--- expect parses
