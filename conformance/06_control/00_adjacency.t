#!perl
# Every construct this tier introduces, in one body, each adjacent to
# another.
#
# TIER 06 control
# INTRODUCES nothing of its own
# USES nothing from a later tier
#
# The tier's other files are one construct each, which is what makes them
# diagnosable. That same property is why a corpus of such files cannot
# reach an ADJACENCY bug: a parser that handles each construct alone and
# mis-handles a pair goes green over the pair.
#
# The pairs this tier has to worry about are specific, and they are why
# the order below is the order it is:
#
#   - `if`/`else` immediately before a postfix `unless`, because `else`
#     and a statement modifier both end a statement without a semicolon
#     being where the reader expects one.
#   - `while` with a postfix `last` inside it, so a jump statement sits
#     directly inside a loop body rather than in a file of its own.
#   - `until` immediately after `while`, because the two differ by one op
#     and a parser that shares their production can lose the difference
#     only when both are present.
#   - C-style `for` immediately before `foreach`, because these are the
#     same keyword over two unrelated optrees and the disambiguation
#     happens at the open paren.
#   - `goto LABEL` last, with a dead statement between it and its label,
#     so the label is adjacent to a statement it is not attached to.
#
# `redo` is guarded by `$ENV{R}`, false at run time: the op compiles and
# sits in the loop body next to `print`, which is the adjacency, without
# making the loop non-terminating.
#
# This tier DEPENDS ON 05, so the pairing with the earlier tier is the
# blocks themselves -- every loop body and both `if` arms are tier 05's
# block, and the loop variables are tier 05's scoped declarations.
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/06_control/00_adjacency.t
#   if-w0-w1-u3-c0-fa-fb-p
#
# The `-u3` is the one piece of output worth explaining, because it looks
# like an off-by-one and is not. `$i` is 2 when the `while` exits; the
# `until` runs it to 4 and prints only on the pass where `$i` is 3, since
# the `next` skips the rest. One line of output from that loop is the
# point -- a loop that printed nothing would still emit its ops and the
# file would measure the same thing while reading as a bug.

--- source
my $n = $ENV{N} // 2;
my $r = $ENV{R} // 0;
my @l = ($ENV{A} // "a", "b");
if ($n) { print "if" } else { print "else" }
print "-unless" unless $n;
my $i = 0;
while ($i < $n) { last if $i > 1; print "-w$i"; $i = $i + 1 }
until ($i >= $n + 2) { $i = $i + 1; next if $i > $n + 1; print "-u$i" }
for (my $j = 0; $j < 1; $j = $j + 1) { print "-c$j" }
foreach my $x (@l) { redo if $r; print "-f$x" }
print "-p" if $n;
goto DONE;
print "-skipped";
DONE:
print "\n";

--- expect output
if-w0-w1-u3-c0-fa-fb-p

--- expect parses
