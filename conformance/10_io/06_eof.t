#!perl
# `eof($fh)` is its own op, and the `rv2gv` before it carries `FAKE`.
#
# TIER 10 io
# INTRODUCES eof
# USES nothing from a later tier
#
# MEASURED perl 5.42.0:
#
#   close($fh)   padsv rv2gv sK*/1      close
#   eof($fh)     padsv rv2gv sK*/FAKE,1 eof
#
# The same `rv2gv` with a different flag: `eof` does not want the glob
# made real, only inspected. The op table records `rv2gv` for both, which
# is the third job this tier's one claimed `rv2gv` does -- the other two
# being autovivifying a glob into `my $fh` at open, and resolving a
# handle operand at print.
#
# The handle holds one line and the file reads it before asking, so `eof`
# is true. Asking before the read would print nothing, and a file whose
# output is empty cannot distinguish "eof was false" from "the program
# did not run".
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'open(my $fh, "<", \"only\n"); my $l = <$fh>; print "eof\n" if eof($fh)'
#   eof

--- source
open(my $fh, "<", \"only\n");
my $l = <$fh>;
print "eof\n" if eof($fh);
close($fh);

--- expect output
eof

--- expect parses
