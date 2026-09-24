# Asking what the context is

The other cases in this tier show context from the outside -- a
construct placed in one and then the other. These two ask perl directly.

**Tier 03 context.** Introduces `wantarray` and `caller`. Depends on
02_variables.

Both are placed at FILE SCOPE deliberately. A sub is where either
operator is USEFUL, not where it is legal, and a sub brings `entersub`
and `leavesub` -- tier 07's ops, four tiers forward, which the lint
would report as reaching ahead. The tier README previously claimed
`wantarray` was out of reach for exactly that reason; measured, that is
false, and the claim is withdrawn.

Observing the result is the part that needs care in both cases.
`defined $w ? ... : ...` emits `cond_expr` and `if (defined $w)` emits
`and`, and both are tier 06's. So both cases COUNT instead: a count is
an assertion about shape that needs no branch, and `scalar` is not an op
at all.

## `wantarray` at file scope is undef, and costs one op

`wantarray` is the only op in perl whose entire meaning is this tier: it
returns true in list context, false in scalar context, and undef in void
context.

Measured, at file scope it compiles to a bare `wantarray s` with no call
machinery at all, and returns undef because a file body is void context.

`my @seen = ($w)` holds exactly one element whether `$w` is undef or
not, so the printed `1` says the op ran and produced a value. That the
value is undef is what the measurement records:
`perl -e 'my $w = wantarray; print defined($w) ? "defined\n" : "undef\n"'`
prints `undef`.

```perl
my $w = wantarray;
my @seen = ($w);
print scalar(@seen), "\n";
```

```behavior
parses: yes
```

```output
1
```

```tokens
one word whose text is "wantarray"
```

## `caller` returns different AMOUNTS in the two contexts

Measured, this case's first two statements:

    $ perl -MO=Concise,-exec -e 'my $s = caller; my @l = caller;'
    3  <0> caller[t2] s
    7  <0> caller[t4] l

One op name, two context flags -- the `s`/`l` pair this tier is built
on. An op-NAME lint sees one `caller` and is satisfied by a case that
only ever put it in one context.

The behavioural half is stronger here than anywhere else in the tier. At
file scope there IS no caller, and the two contexts do not merely format
the same answer differently: `@l` is EMPTY, because list-context
`caller` with no frame returns the empty list, while `$s` is undef and a
scalar holding undef is still one element. A parser that imposed LIST
context on both prints `0 0`; one that imposed SCALAR on both prints
`1 1`. Both wrong readings are visible in one byte.

`caller` is placed in this tier rather than in tier 07, where a reader
expects a call-stack operator to live, because what makes its parse
distinctive is not the stack: it is that the SAME call site yields a
scalar or a list depending on what receives it. 74 of T1's 986 files
call it and the corpus named it nowhere.

THE TOKEN FACT HERE IS WEAK AND THE CASE SAYS SO. `one word whose text
is "print"` would pass any lexer that tokenises identifiers at all.
Measured, nothing better is available: `caller`, `scalar`, `$s`, `@l`
and `@one` each appear twice, and the fact grammar has only `one` and
`no`, so no count fits -- and every negative reachable here names a
spelling the source writes, which makes it false rather than merely
vacuous. What this case actually claims is its OUTPUT.

```perl
my $s = caller;
my @l = caller;
my @one = $s;
print scalar @one, " ", scalar @l, "\n";
```

```behavior
parses: yes
```

```output
1 0
```

```tokens
one word whose text is "print"
```
