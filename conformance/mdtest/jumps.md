# Jumps

`next`, `last` and `redo` are the three targets `enterloop` and
`enteriter` already name in their own dump -- `enteriter(next->w last->z
redo->g)` -- so the loop ops and the jump ops are one measurement, not
two. A tier that claimed the loops and left the jumps for later would be
claiming half of a single line of output. `goto` joins them: it is a
loop-control statement in everything but name, the same code path as the
other three.

**Tier 06 control.** Introduces `cond_expr`, `die`, `enteriter`,
`entertry`, `exit`, `goto`, `iter`, `last`, `leavetry`, `next`, `redo`,
`time`, `unstack`. Depends on 05_scoping.

Every case binds a runtime value and branches on it second.
A constant condition ERASES the construct: measured, `if (1) { print "y" }`
emits `pushmark`, `const` and `print` and no branch op at all, and
`while (0) { print "y" }` emits `enter`, `nextstate` and `leave` -- an
empty program. `$ENV{X}` is unset when the harness runs, so `// N` is a
stable operand the optimiser cannot see through.

## The three loop-control statements

All three take no operands. The labelled forms emit the SAME op carrying a
string -- `next("OUTER")` -- so the corpus distinguishes the spellings by
the argument rather than by the op name, and a parser that produces a
different node kind for the labelled form is producing a distinction perl
does not make.

`redo` is exercised under a guard that is false at run time, which is the
only way to emit the op without writing a loop that does not terminate:
`$ENV{R}` is unset, so `// 0` is a stable false and the op is compiled,
reachable and not taken.

Measured, the `v*` flag is what marks a jump: the op never returns to its
successor, so the stream's textual order is not its execution order.

```perl
my $r = $ENV{R} // 0;
my @l = ($ENV{A} // "a", "b", "c", "d");
foreach my $x (@l) {
    redo if $r;
    next if $x eq "b";
    last if $x eq "d";
    print $x;
}
print "\n";
```

```behavior
parses: yes
```

```output
ac
```

## `goto` is one op name over unrelated constructs

Two spellings are affordable at this tier. `goto LABEL` emits a `<">` op
-- the class B::Concise uses for an op with an SV operand baked in --
carrying the label as a constant and taking nothing from the stack.
`goto $target` emits a `<1>` unary op over the pad slot, and the label it
lands on is not known until run time. Same name, different arity,
different class: a parser that gives them one node shape with an optional
operand is modelling something perl does not.

The third spelling, `goto &sub`, is MEASURED OUT rather than forgotten. It
compiles to `rv2cv`/`srefgen`/`goto` inside the sub body, which is not
visible in the main optree at all and needs `-exec,g` to see; `rv2cv`
belongs to tier 07 and `srefgen` to tier 08, so a case spelling it here
would fail the lint on two ops the tier cannot claim. The op NAME is
claimed here because these two spellings emit it; the frame-replacing
spelling waits for the tier that supplies frames.

Both jumps go FORWARD to a label at the same scope depth. Perl warns about
`goto` into or out of a construct and the harness compares output byte for
byte, so a warning would be a diff.

The label is not an op. It is a flag on the `nextstate` of the statement
it precedes -- `nextstate(SKIP: ...)` -- so a label with no statement after
it has nowhere to live, and the jump targets a statement rather than a
position.

```perl
my $c = $ENV{X} // 0;
my $t = $ENV{T} // "TAIL";
print "a";
goto SKIP unless $c;
print "b";
SKIP:
print "c";
goto $t;
print "d";
TAIL:
print "e\n";
```

```behavior
parses: yes
```

```output
ace
```
