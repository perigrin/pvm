# Conditionals

There is no `if` op, and `if` is not one thing. A bare `if` compiles to
tier 04's `and`; an `if`/`else` compiles to `cond_expr`, the same op the
ternary `?:` emits; `elsif` compiles to nested `cond_expr` and adds no op
of its own. So three source constructs share two ops between them, and
`unless` is a fourth spelling that differs from `if` by exactly one op.

**Tier 06 control.** Introduces `cond_expr`, `die`, `enteriter`,
`entertry`, `exit`, `goto`, `iter`, `last`, `leavetry`, `next`, `redo`,
`time`, `unstack`. Depends on 05_scoping.

Every case binds a runtime value and branches on it second.
A constant condition ERASES the construct: measured, `if (1) { print "y" }`
emits `pushmark`, `const` and `print` and no branch op at all, and
`while (0) { print "y" }` emits `enter`, `nextstate` and `leave` -- an
empty program. `$ENV{X}` is unset when the harness runs, so `// N` is a
stable operand the optimiser cannot see through.

## Block `if` and postfix `if` are one construct

The two statements emit the same ops in the same order, down to the
`nextstate` -- measured, the block form's braces add no `enter`/`leave`
pair at this nesting, and the `and` branches straight into `pushmark`
exactly as the postfix does.

This case asserts an EQUALITY rather than a behaviour. Both statements
print, so an output test alone passes against a parser that treats them as
unrelated grammar productions; what makes the claim checkable is that the
two halves of the measured op stream are byte-identical apart from the
`nextstate` line number.

```perl
my $c = $ENV{X} // 1;
if ($c) { print "block\n" }
print "postfix\n" if $c;
```

```behavior
parses: yes
```

```output
block
postfix
```

## `unless` is `or`, not an inverted `if`

Against the previous case's stream, position 9 is `and` there and `or`
here; every other op is the same op in the same place. There is no
negation op anywhere. Perl does not invert the condition and test it --
it picks the other short-circuit.

A parser that desugars `unless COND` to `if (!COND)` builds a tree perl
never builds, and nothing in the output distinguishes the two, which is
why this case exists. Both `and` and `or` are tier 04's ops; what this
tier contributes is the measurement that the two keywords differ by
exactly one of them.

```perl
my $c = $ENV{X} // 0;
unless ($c) { print "block\n" }
print "postfix\n" unless $c;
```

```behavior
parses: yes
```

```output
block
postfix
```

## Adding an `else` changes the op

A bare `if` emits `and`; an `if`/`else` emits `cond_expr` -- the same op
the ternary `?:` emits. One source-level keyword maps to two different ops
depending on whether an `else` is present, and only one of them is new to
this tier. The corpus needs both this case and the postfix one; neither
alone lints the tier, and a parser that emits one shape for both spellings
passes either in isolation.

The `else` arm is what brings back the `enter`/`leave` pair the bare `if`
does not emit: `cond_expr` jumps into a second block, and a second block
is a scope.

The erasure this case exists to avoid is sharp here. Measured,
`if (1) { print "y" } else { print "n" }` emits `pushmark`, `const` and
`print` and NO branch op -- the else arm is gone from the binary. Such a
case would pass its output test while measuring none of this tier.

```perl
my $c = $ENV{X} // 0;
if ($c) { print "taken\n" } else { print "else\n" }
```

```behavior
parses: yes
```

```output
else
```

## `elsif` is a third form, not `else` followed by `if`

The tier shipped with `if`, `unless` and `if`/`else` and no `elsif`
anywhere, while 14 of T1's 986 files use it. Measured, the chain compiles
to NESTED `cond_expr`, one per condition, and adds no op of its own -- so
`cond_expr`, already claimed for `if`/`else`, covers it and the INTRODUCES
set does not grow.

What distinguishes `elsif` is the SHAPE of the nesting, which ops cannot
show, so the case pins the branch behaviourally instead: the MIDDLE branch
is the one taken, which neither a lone `if` nor an `else` can produce.

The token facts count rather than forbid, and the counting is the
falsifying half. This source holds exactly ONE `elsif` and ONE `else`, so
a lexer that read `elsif` as `else` followed by `if` would produce TWO
words spelled `else` and fail the second fact. Measured, ours emits
`Word("elsif")` whole.

```perl
my $c = $ENV{X} // 0;
if ($c == 1) { print "one\n" } elsif ($c == 0) { print "zero\n" } else { print "other\n" }
```

```behavior
parses: yes
```

```output
zero
```

```tokens
one word whose text is "elsif"
one word whose text is "else"
```
