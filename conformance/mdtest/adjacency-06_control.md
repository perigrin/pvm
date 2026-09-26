# Every control construct, each beside another

One body holding every construct this tier introduces, each adjacent to
another -- and adjacent to tier 05's blocks, which is what this tier
depends on.

**Tier 06 control.** Introduces `cond_expr`, `die`, `enteriter`,
`entertry`, `exit`, `goto`, `iter`, `last`, `leavetry`, `next`, `redo`,
`time`, `unstack`. Depends on 05_scoping.

WHY A MIXTURE NEEDS ITS OWN CASE. The tier's other cases are one construct
each, which is what makes them diagnosable. That same property is why a
corpus of such cases cannot reach an ADJACENCY bug: a parser that handles
each construct alone and mis-handles a pair goes green over the pair.

The pairs this tier has to worry about are specific, and they are why the
order below is the order it is:

- `if`/`else` immediately before a postfix `unless`, because `else` and a
  statement modifier both end a statement without a semicolon being where
  the reader expects one.
- `while` with a postfix `last` inside it, so a jump statement sits
  directly inside a loop body rather than in a case of its own.
- `until` immediately after `while`, because the two differ by one op and
  a parser that shares their production can lose the difference only when
  both are present.
- `do BLOCK while` immediately after the block `while` and `until`,
  because the post-test loop is the one that emits `unstack` with no
  `enterloop`.
- `do BLOCK` as an expression immediately after `do BLOCK while`, which is
  the sharpest pair here. The two spell the same two tokens and open the
  same brace, and what decides the production is what FOLLOWS the closing
  brace. A parser that commits at the `do` handles whichever one it
  guessed and mis-handles the other, and only their adjacency catches it.
- `eval BLOCK` immediately after `do BLOCK`, because both put a block
  where an expression goes with no comma after it, and both are the
  position where perl's own `{` heuristic can answer "anonymous hash".
- C-style `for` immediately before `foreach`, because these are the same
  keyword over two unrelated optrees and the disambiguation happens at the
  open paren.
- `goto LABEL` last, with a dead statement between it and its label, so
  the label is adjacent to a statement it is not attached to.

## The whole tier in one body

`redo` is guarded by `$ENV{R}`, false at run time: the op compiles and
sits in the loop body next to `print`, which is the adjacency, without
making the loop non-terminating. The `exit` is a DEAD BRANCH for the same
reason -- `exit 3 if $r` -- because an `exit` that fired would end the
program before `goto DONE` and the body would measure nothing after it. An
`eval` does NOT trap it: measured, `eval { exit 0 }` exits.

The `eval` here traps a division by zero rather than a `die`, for the
budget reason the `eval BLOCK` case records.

The output runs to two lines because `die`'s message ends in a newline,
which is not decoration: a `$@` that does not end in one carries the file
and line number perl appends, and the file here is the harness's temp
path, which differs every run. `-Dx` is the trapped message and the
newline is its own.

The `-u3` is the one field worth explaining, because it looks like an
off-by-one and is not. `$i` is 2 when the `while` exits; the `until` runs
it to 4 and prints only on the pass where `$i` is 3, since the `next`
skips the rest. One line of output from that loop is the point -- a loop
that printed nothing would still emit its ops and the case would measure
the same thing while reading as a bug.

The three fields the block-valued group adds each discriminate:

- `-d2`: the post-test loop ran its body to `$n`. A `while` in its place
  with the same initial `$k` would reach the same 2, so this field is the
  WEAK one -- the `do BLOCK while` case carries the strong pin, where a
  false condition still produces one pass. Here the value is adjacency,
  not discrimination.
- `-v3`: the block-as-expression yielded its LAST statement, `$n + 1`, not
  the `$n` its first statement bound and not a hash reference. A parser
  that read `{ my $t = $n; $t + 1 }` as an anonymous hash would print
  `HASH(0x...)` here.
- `-et`: the eval frame trapped the division by zero. Without the frame
  the program aborts before any of the output after it is printed, so
  every field to its right is also evidence the frame held.

```perl
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
```

```behavior
parses: yes
```

```output
if-w0-w1-u3-d2-v3-et-Dx
-X-t1-c0-fa-fb-p
```

## Two `next` guards in a row, where the first arm carries the second

Two consecutive statement-modifier `next` guards inside one loop body.

WHY TWO AND NOT ONE. One `next if` guard is a single conditional jump and
every implementation handles it. TWO make the first guard's FALL-THROUGH ARM
carry the second, so an implementation that ends the first arm without
threading the rest of the body into it loses the second guard silently -- the
loop then runs an iteration it should have skipped, with no diagnostic. That
is an adjacency property: each guard alone is fine and the pair is not.

Measured on B::SoN, which is why this case exists: one guard translates, two
refuse. Reduced from perl's own t/cmd/switch.t:5, which wraps the same shape
in a sub with a `continue` block -- dropped here because `shift` would pin the
case to tier 11 and bury a control-flow finding in the OO tier.

The output is the assertion: 2 and 4 must be ABSENT, and an implementation
that loses the second guard prints 4.

```perl
my @out;
for my $i (1 .. 6) {
    next if $i == 2;
    next if $i == 4;
    push @out, $i;
}
print "@out\n";
```

```behavior
parses: yes
```

```output
1 3 5 6
```

```ir
GAP: a loop control (`next`) inside a branch arm is not yet lowered
     -- only `last` carries an exit edge
L: GAP
```
