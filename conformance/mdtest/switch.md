# The switch statement

`given`, `when` and `default`: a topicalizer, a guarded block that
leaves it, and the guard that always holds.

**Tier 06 control.** Introduces `cond_expr`, `die`, `entergiven`,
`enteriter`, `entertry`, `enterwhen`, `exit`, `goto`, `iter`, `last`,
`leavegiven`, `leavetry`, `leavewhen`, `next`, `redo`, `time`, `unstack`.
Depends on 05_scoping. The four this topic is for are `entergiven`,
`leavegiven`, `enterwhen` and `leavewhen`.

THE FEATURE IS OFF FROM BUNDLE 5.36 ONWARD, and every case enables it by
name. Measured under 5.42.0: `use v5.34` followed by `given` compiles and
runs; `use v5.36` followed by the same `given` is a syntax error near `) {`;
and with no `use` at all, `given` is not a keyword and the program is a
syntax error too. `use v5.36; use feature "switch";` compiles again, so the
bundle disables the feature rather than removing it. Bundles 5.10 through
5.34 enable it.

AND IT DOES NOT WARN. Every case below compiles silently under `perl -w`
and under `use warnings "deprecated"` -- measured, 5.42.0 prints no
deprecation diagnostic for `given`, `when` or `~~`. A reader who expects
one will not find it, and an implementation that emits one is producing a
diagnostic perl does not.

`when (EXPR)` smartmatches `$_` against EXPR, which is why the first two
cases emit `smartmatch`. That op is tier 04's, introduced with `~~` in
`comparison.md`, because `~~` reaches it without any switch at all. The
third case is the shape that emits NO `smartmatch`: a comparison argument
is used as a boolean.

Every case binds its topic from `$ENV{X}`, unset when the harness runs, so
the optimiser cannot see which arm is taken.

## `given` with two `when` arms and a `default`

The SECOND `when` is the one taken. A parser that read the arms as one
statement, or stopped at the first, prints `one` or nothing; `default`
following a taken `when` is skipped because `when` leaves the `given`.

```perl
use feature "switch";
my $x = $ENV{X} // 2;
given ($x) {
    when (1) { print "one\n" }
    when (2) { print "two\n" }
    default { print "other\n" }
}
```

```behavior
parses: yes
```

```output
two
```

```tokens
one word whose text is "given"
one word whose text is "default"
```

## The parenless `default BLOCK`, taken, and the statement after it

`default` takes no parenthesised argument: its block follows the keyword
directly. This is the form a parser that treats every keyword followed by
a block as `KEYWORD (EXPR) BLOCK` mis-reads. The `print` after the `given`
is there to pin that the construct ENDS at its closing brace -- no
semicolon follows it, and a parser that kept consuming would swallow the
next statement.

```perl
use feature "switch";
my $x = $ENV{X} // 9;
given ($x) {
    when (1) { print "one\n" }
    default { print "other\n" }
}
print "after\n";
```

```behavior
parses: yes
```

```output
other
after
```

## `when` over a comparison, which does not smartmatch

`when ($_ > 5)` is not `$_ ~~ ($_ > 5)`. Perl recognises a comparison
argument and tests it as a boolean: measured, this case emits `gt` and no
`smartmatch` at all, where the two cases above each emit one. A parser that
gives `when` a single expression grammar sees no difference; the optree
does.

An ARRAY or anonymous-array argument -- `when (@l)`, `when ([2, 3])` --
smartmatches through an implicit reference and emits `srefgen`, tier 08's
op, so it cannot be spelled here. `taking-references.md` carries `when
(@l)`.

```perl
use feature "switch";
my $x = $ENV{X} // 7;
given ($x) {
    when ($_ > 5) { print "big\n" }
    default { print "small\n" }
}
```

```behavior
parses: yes
```

```output
big
```
