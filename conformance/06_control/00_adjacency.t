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
#   - `do BLOCK while` immediately after the block `while` and `until`,
#     because the post-test loop is the one that emits `unstack` with no
#     `enterloop`, and a parser that shares the `while` production across
#     both loses the ordering difference only when both are present.
#   - `do BLOCK` as an expression immediately after `do BLOCK while`,
#     which is the sharpest pair in this file. The two spell the same two
#     tokens and open the same brace, and what decides the production is
#     what FOLLOWS the closing brace. A parser that commits at the `do`
#     handles whichever one it guessed and mis-handles the other, and
#     only their adjacency catches it.
#   - `eval BLOCK` immediately after `do BLOCK`, because both put a block
#     where an expression goes with no comma after it, and both are the
#     position where perl's own `{` heuristic can answer "anonymous
#     hash". Adjacent, a parser cannot pass by special-casing one.
#   - `goto LABEL` last, with a dead statement between it and its label,
#     so the label is adjacent to a statement it is not attached to.
#
# `redo` is guarded by `$ENV{R}`, false at run time: the op compiles and
# sits in the loop body next to `print`, which is the adjacency, without
# making the loop non-terminating.
#
# `eval` traps a division by zero rather than a `die`, for the budget
# reason `14_eval_block.t` records: `die` emits an op no tier at or before
# 06 claims, and `divide` is tier 04's.
#
# This tier DEPENDS ON 05, so the pairing with the earlier tier is the
# blocks themselves -- every loop body and both `if` arms are tier 05's
# block, and the loop variables are tier 05's scoped declarations.
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/06_control/00_adjacency.t
#   if-w0-w1-u3-d2-v3-et-Dx
#   -X-t1-c0-fa-fb-p
#
# The output runs to two lines because `die`'s message ends in a newline,
# which is not decoration: a `$@` that does not end in one carries the
# file and line number perl appends, and the file here is the harness's
# temp path, which differs every run. `-Dx` is the trapped message and
# the newline is its own.
#
# The `exit` is a DEAD BRANCH -- `exit 3 if $r` with `$r` unset. It is
# here so the tier's construct appears adjacent to the others, and it
# cannot be taken, because an `exit` that fired would end the program
# before `goto DONE` and the file would measure nothing after it. An
# `eval` does NOT trap it: measured, `eval { exit 0 }` exits.
#
# The `-u3` is the one piece of output worth explaining, because it looks
# like an off-by-one and is not. `$i` is 2 when the `while` exits; the
# `until` runs it to 4 and prints only on the pass where `$i` is 3, since
# the `next` skips the rest. One line of output from that loop is the
# point -- a loop that printed nothing would still emit its ops and the
# file would measure the same thing while reading as a bug.
#
# The three fields the block-valued group adds each discriminate:
#
#   -d2  the post-test loop ran its body to $n. A `while` in its place
#        with the same initial $k would reach the same 2, so this field
#        is the WEAK one -- `09_do_while.t` carries the strong pin, where
#        a false condition still produces one pass. Here the value is
#        adjacency, not discrimination.
#   -v3  the block-as-expression yielded its LAST statement, $n + 1, not
#        the $n its first statement bound and not a hash reference. A
#        parser that read `{ my $t = $n; $t + 1 }` as an anonymous hash
#        would print `HASH(0x...)` here.
#   -et  the eval frame trapped the division by zero. Without the frame
#        the program aborts before any of the output after it is
#        printed, so every field to its right is also evidence the frame
#        held.

--- source
my $n = $ENV{N} // 2;
my $r = $ENV{R} // 0;
my @l = ($ENV{A} // "a", "b");
if ($n) { print "if" } else { print "else" }
print "-unless" unless $n;
my $i = 0;
while ($i < $n) { last if $i > 1; print "-w$i"; $i = $i + 1 }
until ($i >= $n + 2) { $i = $i + 1; next if $i > $n + 1; print "-u$i" }
my $k = 0;
do { $k = $k + 1 } while $k < $n;
print "-d$k";
my $v = do { my $t = $n; $t + 1 };
print "-v$v";
my $e = eval { 10 / ($ENV{X} // 0) } // "t";
print "-e$e";
eval { die "x\n" };
print "-D$@";
exit 3 if $r;
print "-X";
my $t = time;
print "-t", $t > 1000000000 ? 1 : 0;
for (my $j = 0; $j < 1; $j = $j + 1) { print "-c$j" }
foreach my $x (@l) { redo if $r; print "-f$x" }
print "-p" if $n;
goto DONE;
print "-skipped";
DONE:
print "\n";

--- expect output
if-w0-w1-u3-d2-v3-et-Dx
-X-t1-c0-fa-fb-p

--- expect parses
