# Writing to a handle

`print` to a named destination, `say`, and the operator that changes
what "no handle" means.

**Tier 10 io.** Introduces `close`, `eof`, `open`, `readline`, `rv2gv`,
`say`, `select`, `sselect`. Depends on 03_context.

`print` IS NOT IN THAT LIST AND IS THIS TIER'S SUBJECT. Tier 01 claims
it deliberately -- a literal has to be observed somehow -- and a tier
claims an op once. That is not a filing technicality: measured, `print
"x"` and `print $fh "x"` emit the same `print vKS`, and what changes is
what precedes it. The difference between printing and printing
SOMEWHERE lives entirely in the operands, so a tier declared as a set
of op names cannot express it.

NO CASE WRITES TO STDERR, recorded because it reads as an omission. The
`output` block is compared against stdout, so a line sent to STDERR
lands in neither stream the runner reads and asserting it would assert
bytes nothing checks. The op stream does not distinguish them --
`gv[*STDERR] rv2gv print` against `gv[*STDOUT] rv2gv print` -- so the
bareword case measures its construct through STDOUT and stays
observable.

## `print $fh "x"` differs from `print "x"` only in operands

Measured under 5.42.0:

    print "x"        pushmark const      print vK
    print $fh "x"    pushmark padsv rv2gv const print vKS

The op name is `print` in both. This is why `print` stays tier 01's op
and is not reclaimed here.

A handle opened onto a scalar with `\my $buf` makes the write
observable without a file: `print $out` fills `$buf`, and printing
`$buf` to stdout is what the case is checked against. `\my $buf` emits
`srefgen` -- tier 08's op, and a legal use here -- because unlike the
open case's `\"x\n"` there is no constant to fold.

```perl
open(my $out, ">", \my $buf);
print $out "captured\n";
close($out);
print $buf;
```

```behavior
parses: yes
```

```output
captured
```

## A bareword handle reaches `print` through `gv`

Measured under 5.42.0:

    print STDOUT "x"   gv[*STDOUT] rv2gv sKR/1 const print vKS
    print $fh    "x"   padsv[$fh]  rv2gv sKR/1 const print vKS

The bareword and the lexical converge one op later. A file using only
lexical handles therefore never emits `gv`, and this tier's declared
set is the union across its cases rather than a property of any one of
them. `gv` itself is tier 02's op, claimed there with the package
scalar; reaching it through a bareword handle is a use, not an
introduction.

```perl
print STDOUT "to stdout\n";
```

```behavior
parses: yes
```

```output
to stdout
```

## `say` is feature-gated, and the featureless spelling compiles

The same bytes parse two different ways depending on a pragma written
earlier in the file, and WITHOUT THE FEATURE THE SPELLING IS NOT A
SYNTAX ERROR. Measured:

    my $x=1; say $x                        pushmark padsv method_named[PV "say"] entersub
    use feature "say"; my $x=1; say $x     pushmark padsv say vK

Without the pragma perl reads `say $x` as indirect-object method
dispatch and it compiles clean -- `perl -c` says `syntax OK` -- failing
only at RUNTIME with `Can't locate object method "say" via package "1"`
and exit code 255. That is `11_oo/10_isa_infix.t`'s finding met from
the worse side: there the featureless form is a syntax error and fails
loudly, here it is a clean compile of the WRONG PROGRAM.

THE CORPUS CANNOT PIN THAT PARSE FROM THIS TIER. The featureless form
emits `method_named`, which is tier 11's op, and 11 is LATER than 10; a
case may only emit ops its tier or an earlier one claims. (`entersub`,
the other op it emits, is tier 07's and would have been a legal use;
`method_named` alone is what blocks it.) The measurement is recorded
and asserted nowhere. A case that pins it belongs in tier 11 beside the
`isa` file whose finding it mirrors.

WHY `say` IS HERE AND NOT BESIDE `print` IN TIER 01: it takes a
filehandle through the SAME `rv2gv` -- `say $out $line` is `padsv[$out]
rv2gv sKR/1 padsv[$line] say vKS` against `print $out "x"`'s
`padsv[$out] rv2gv sKR/1 const print vKS`. The handle machinery is
identical; what `say` adds is a record separator, which is a property
of writing to a HANDLE.

THE NEWLINE BETWEEN `said` AND `printed` IS THE WHOLE ASSERTION. It is
in no string in this source, so it can only have come from `say`. A
parser that compiled `say` as `print` emits `[saidprinted]`; one that
appended a newline to both emits a third line. The brackets make the
trailing byte visible -- without them `[said\nprinted]` and
`[said\nprinted\n]` would differ only in bytes the format cannot show.

`$ENV{X} // "said"` keeps the argument a RUNTIME value so nothing
folds; `X` is unset when the runner executes the case. `use feature
"say"` rather than `use v5.36`, which would also enable it: measured,
the version bundle sets `strict` as well -- ops carry `/STRICT`, the
nextstate flags read `fea=6` against `fea=15` -- so it turns on more
than this case is about.

THE TOKEN FACT CATCHES THE REWRITE, because behaviour cannot. A parser
that read the featureless spelling -- rewriting `say $out $line` into
`$out->say($line)`, which is what perl itself does without the pragma
-- produces an ARROW, and this source has none. The positive count says
the construct survived the lex as ONE word: it counts one and not two
even though `say` appears twice, because the other occurrence is inside
`use feature "say"` where it is a STRING and not a bare word.

```perl
use feature "say";
my $line = $ENV{X} // "said";
open(my $out, ">", \my $buf);
say $out $line;
print $out "printed";
close($out);
print "[$buf]\n";
```

```behavior
parses: yes
```

```output
[said
printed]
```

```tokens
one word whose text is "say"
```

## `select` is two operators sharing a name, told apart by arity

One argument selects a default output handle; four arguments are the
`select(2)` syscall. Measured, they are not one operator with an
optional tail:

    select($out)                    rv2gv sK/DREFSV,1  select[t7] sK/1
    select(undef, undef, undef, 0)  undef undef undef const  sselect[t11] sK/4

A different op NAME, not a different flag. That is the inverse of
`rv2gv`, which does three unrelated jobs in this tier under one name.

So this is a PARSING fact. A parser must count the arguments before it
knows which operator it has read, and the arity is not recoverable from
the name, from the first argument, or from anything a lexer can see.

THERE IS NO TWO-ARGUMENT FORM. Measured, arity 2 and arity 3 are both
COMPILE-TIME errors carrying the same message -- `Not enough arguments
for select system call` -- which is perl deciding that anything past
one argument must be the four-argument form and then finding it short.
The wrong count never reaches runtime. Arity 0 is the one-argument
operator with nothing selected: measured, `select()` returns
`main::STDOUT`, the same value `select(STDOUT)` returns, which makes
the split 0-or-1 against 4. It is not written into the source because
it would add a third `select` to a case whose subject is that there are
two.

THE ONE-ARGUMENT FORM RETURNS THE PACKAGE-QUALIFIED PREVIOUS HANDLE --
measured, `main::STDOUT` and not `STDOUT` -- which is what makes the
round trip work: the value `select` hands back is a thing `select` will
take, so restoring needs no name of its own.

THE OUTPUT IS THE FALSIFYING HALF, and the `print` with NO HANDLE is
what makes it one. Line 3 is a bare `print "captured\n"` with no
filehandle anywhere in the statement, and its bytes land in `$buf`
because line 2 changed where "no handle" points. A parser that compiled
`select` as a no-op prints `captured` to stdout and then `[][0]`: the
right bytes in the wrong stream, and an empty buffer where the capture
should be. Nothing about `print` itself changed -- it emits the same
`print vK` either way -- so this case measures `select` entirely
through what `print` DID NOT DO.

The `0` is the four-argument form's return, the number of handles
ready, zero because every handle argument is `undef`. A TIMEOUT OF ZERO
is what makes it usable: `select(undef,undef,undef,0.25)` is the
idiomatic sub-second sleep and would be the obvious thing to write, but
a corpus case cannot rest on a duration. With the timeout at 0 the call
returns immediately and deterministically; measured twice in
succession, both runs printed `[captured\n][0]`.

NO TOKEN FACTS. `select` is a bare word in both spellings and the arity
split is a parsing question, not a lexical one -- the token stream is
`word(select) operator(() ...` for both, identically, and the
glossary's vocabulary admits `one` and `no` and nothing that would say
"four". This is the mirror of the scalar readline case: there the
behaviour could not see the split and the tokens could.

```perl
open(my $out, ">", \my $buf);
my $prev = select($out);
print "captured\n";
select($prev);
close($out);
my $ready = select(undef, undef, undef, 0);
print "[$buf][$ready]\n";
```

```behavior
parses: yes
```

```output
[captured
][0]
```
