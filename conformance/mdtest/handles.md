# Opening a handle and reading from it

`open`, `close`, `readline` and `eof`: a handle held in a lexical, and
the one op whose meaning is decided by what receives it.

**Tier 10 io.** Introduces `close`, `eof`, `open`, `readline`, `rv2gv`,
`say`, `select`, `sselect`. Depends on 03_context.

THIS IS WHERE THE TIER'S PLACEMENT COMES FROM. `<$fh>` is ONE op with
two behaviours: measured under 5.42.0 it compiles to `readline sKS/1`
when a scalar receives it and `readline lK/1` when an array does -- the
same op, a different flag, one line against every remaining line. A
parser that cannot say which context an expression is in cannot say
what `<$fh>` returns, and it cannot learn that from the readline
itself.

Every handle here is in-memory, opened on a scalar ref. These files
must be reproducible and a file that opens a real path is not --
`/etc/hostname` differs per machine. Measured, that costs nothing:
`\"x\n"` is constant-folded to `const[IV \"x\n"] s/FOLD` at compile
time, and the op stream of `open(my $fh, "<", \"x\n")` is identical op
for op to `open(my $fh, "<", "/etc/passwd")`.

No case checks its open. The idiomatic `open(...) or die` emits a `die`
op and no tier claims `die`, so a check would widen this tier's
declared set by an op belonging to a tier that does not exist yet. The
cases open unchecked and print a constant to show they ran.

## A three-argument open autovivifies the glob

`open(my $fh, "<", ...)` does not hand the fresh lexical a value: it
autovivifies a glob into the scalar's slot, through `rv2gv` with
`DREFSV` set. The handle IS the glob, which is the fact the rest of the
tier rests on.

Measured, the in-memory form against a real path -- `padsv rv2gv const
const open` for both. Five ops, the same five. The scalar ref is the
ordinary open as far as the op table can see.

```perl
open(my $fh, "<", \"x\n");
print "opened\n";
close($fh);
```

```behavior
parses: yes
```

```output
opened
```

## `<$fh>` in scalar context reads ONE line

The half of `readline` a scalar receives, and the baseline the list
case deviates from by one sigil. The handle holds two lines and this
case prints one, so the output is also the evidence that the scalar
form STOPPED -- a case reading a one-line handle could not tell the two
contexts apart by behaviour at all.

THE TOKEN FACTS ARE WHERE THE ANGLE SPLIT IS MADE. `<$fh>` and `<*.c>`
share a spelling and belong to different tiers: the first is this
tier's IO construct, the second is tier 13's delimiting problem.
Measured, they part at the optree -- `padsv readline` against `pushmark
const gv glob` -- and they are the same thing to a lexer, which cannot
tell a handle name from a pattern without knowing what the name means.

The behavioural half cannot make that claim. Measured, our parser reads
`<`, `$fh`, `>` as a perfectly good comparison chain and returns ZERO
Unknowns for it, so a lexer that split the angles would satisfy every
behavioural assertion here unnoticed. `no operator whose text is "<"`
is what falsifies it: a split lexer emits that operator, an unsplit one
never does.

```perl
open(my $fh, "<", \"one\ntwo\n");
my $l = <$fh>;
print $l;
close($fh);
```

```behavior
parses: yes
```

```output
one
```

```tokens
one readline operator whose text is "<$fh>"
no operator whose text is "<"
```

## `<$fh>` in list context reads EVERY remaining line

The same `readline` op, flagged `lK/1` rather than `sKS/1`. One op, two
behaviours, selected by what receives it -- which is why this tier
declares `DEPENDS ON 03_context` and cannot precede it.

The count is printed rather than the lines, because the count is what
separates this from the scalar case. Printing the lines would show
`one` first in both and differ only in what followed.

The token facts repeat the scalar case's pair rather than sharing them,
because the spelling is what they assert and each case has its own.
`<$fh>` is ONE token in both, and the context that tells the two cases
apart is invisible to the lexer -- which is the point: everything
separating this tier's `<$fh>` from tier 13's `<*.c>` sits above the
lexer.

```perl
open(my $fh, "<", \"one\ntwo\n");
my @lines = <$fh>;
print scalar(@lines), "\n";
close($fh);
```

```behavior
parses: yes
```

```output
2
```

```tokens
one readline operator whose text is "<$fh>"
no operator whose text is "<"
```

## `eof($fh)` inspects the glob without making it real

`eof` is its own op, and the `rv2gv` before it carries `FAKE`.
Measured: `close($fh)` is `padsv rv2gv sK*/1 close` and `eof($fh)` is
`padsv rv2gv sK*/FAKE,1 eof` -- the same `rv2gv` with a different flag,
because `eof` does not want the glob made real, only inspected. That is
the third job this tier's single claimed `rv2gv` does; the other two
are autovivifying a glob into `my $fh` at open and resolving a handle
operand at print.

The handle holds one line and the case reads it before asking, so `eof`
is true and the output is non-empty. Asking first would print nothing,
and a case whose output is empty cannot distinguish "eof was false"
from "the program did not run".

```perl
open(my $fh, "<", \"only\n");
my $l = <$fh>;
print "eof\n" if eof($fh);
close($fh);
```

```behavior
parses: yes
```

```output
eof
```
