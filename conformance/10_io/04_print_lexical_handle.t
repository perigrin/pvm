#!perl
# `print $fh "x"` and `print "x"` are the SAME `print` op. What differs is
# the operand pushed before it: the handle, through `rv2gv`.
#
# TIER 10 io
# INTRODUCES printing to a lexical filehandle
# USES nothing from a later tier
#
# This is why `print` is tier 01's op and is not reclaimed here. Measured
# under 5.42.0:
#
#   print "x"        pushmark const      print vK
#   print $fh "x"    pushmark padsv rv2gv const print vKS
#
# The op name is `print` in both. The difference between printing and
# printing SOMEWHERE lives entirely in the operands, so a tier declared as
# a set of op names cannot express it, and the README carries the fact
# instead.
#
# A handle opened onto a scalar with `\my $buf` makes the write
# observable without a file: `print $out` fills `$buf`, and printing
# `$buf` to stdout is what this file is checked against. Note that
# `\my $buf` emits `srefgen` -- tier 08's op, and allowed here -- because
# unlike `01_open_close.t`'s `\"x\n"` there is no constant to fold.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'open(my $o, ">", \my $b); print $o "captured\n"; close($o); print $b'
#   captured

--- source
open(my $out, ">", \my $buf);
print $out "captured\n";
close($out);
print $buf;

--- expect output
captured

--- expect parses
