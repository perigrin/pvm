# 10_io

Filehandles, readline, print to a handle.

## Why this tier sits here

A filehandle has to be held somewhere before it can be used, and a line
has to land somewhere after it is read -- `open(my $fh, ...)` and
`my $l = <$fh>` are tier 02's scalar with a tier 10 value in it. That
much every tier from 03 on could supply.

The dependency that decides the placement is context. `<$fh>` is ONE op
with two behaviours: measured under 5.42.0 it compiles to
`readline sKS/1` when a scalar receives it and `readline lK/1` when an
array does -- the same op, a different flag, one line versus every
remaining line. A parser that cannot say which context an expression is
in cannot say what `<$fh>` returns, and that is not a fact it can learn
from the readline itself. So this tier sits after `03_context` and
cannot precede it.

Could it move earlier? To anywhere after 03, yes, and the dependency
check would not object -- tiers 04 through 09 supply nothing it needs.
That is check 2's revised form working as documented: the requirement is
that the prerequisite sits among the earlier tiers, not that the tier is
pinned to one slot. What it cannot do is move above 03, and it is placed
here rather than at 04 because `11_oo` is the first tier that wants
something from it and the numbering is one topological sort, not the
only one.

## DEPENDS ON

    03_context

Declared rather than assumed, in the same spirit as `09_regex` not
depending on `08_references`. The neighbour above is `09_regex`, and
nothing here reads a pattern; naming 09 because it is N-1 would record a
dependency that does not exist.

## INTRODUCES

    close eof open readline rv2gv

## Why those ops, and not the ones the source implies

`print` is missing from that list, and it is this tier's subject.

Tier 01 claims `print`, and says so deliberately: a literal has to be
observed somehow, so `print` appears there as a fixture. A tier claims an
op once, so it cannot be claimed again here -- and that is not a filing
technicality, because the op table cannot express the difference at all.
Measured: `print "x"` and `print $fh "x"` emit the same `print vKS`. What
changes is what precedes it. Tier 01's form pushes a constant; this
tier's pushes the handle through `rv2gv` first, and the bareword forms
push `gv[*STDERR]` or `gv[*STDOUT]` through the same `rv2gv`. The
distinction between printing and printing SOMEWHERE lives in the
operands, and the op stream records only that both are `print`.

So what this tier introduces is the machinery around it: `rv2gv`, which
is how a name or a scalar becomes a handle, and `open`, `close`,
`readline`, `eof`. Using a construct is not the same as introducing it,
and here the lint cannot see the difference between the two uses. It is
the tier README that carries that fact; nothing checks it.

Three other things the measurements say that reading the source does
not:

- **`gv` for barewords, `rv2gv` for both.** `print STDERR "x"` compiles
  to `gv[*STDERR]` then `rv2gv sKR/1`; `print $fh "x"` compiles to
  `padsv[$fh]` then the same `rv2gv sKR/1`. The bareword and the lexical
  converge one op later, so a file using only lexical handles never
  emits `gv` and a file using only barewords never emits `padsv` for a
  handle. `gv` is tier 02's op, claimed there with the package scalar,
  and reaching it through a bareword handle does not make it new. The
  declared set is the union across the tier's files, never a property of
  one of them.
- **`open` with a lexical handle emits `rv2gv` too**, with
  `DREFSV` set -- autovivifying the glob into the fresh `my $fh`. The
  same op name does three jobs in this tier and is claimed once.
- **No op for the diamond.** `<>` compiles to `gv[*ARGV]` plus
  `readline`, identical to `<ARGV>`. There is no distinct operator to
  claim; the diamond is a spelling the lexer must recognise, which is a
  token fact and not an op fact. `<*>` is glob and belongs to tier 13.

The ops also cannot tell a successful open from a failed one, which is
the general caveat restated for this tier: they LINT a declared tier and
cannot derive one.

## What writing the files measured

The declared set came out of the corpus unchanged -- every op in
INTRODUCES is emitted, nothing had to be removed, and nothing outside the
earlier tiers' claims appeared. Two facts the files forced, neither of
them visible before there were files:

- **The in-memory handle is the ordinary open.** These files must be
  reproducible, and a file that opens a real path is not: `/etc/hostname`
  differs per machine. So they open a scalar ref -- `open(my $fh, "<",
  \"line one\n")` -- which touches no filesystem. Measured, that is not a
  weaker measurement of `open`: `\"line one\n"` is constant-folded to
  `const[IV \"line one\n"] s/FOLD` at compile time, and the op stream is
  identical to `open(my $fh, "<", "/etc/passwd")` op for op. The REF is
  built by the optimiser, so no `srefgen` appears. The write direction is
  different and worth stating: `\my $buf` has no constant to fold and so
  does emit `srefgen`, which is tier 08's op and allowed here as a use.

- **`die` is claimed by no tier, so these files do not check their
  opens.** The idiomatic `open(...) or die` emits a `die` op, and a sweep
  of every tier's INTRODUCES block finds `die` in none of them -- it is
  not tier 10's to claim, since failing is not this tier's subject. An
  `or die` here would therefore widen the declared set by an op belonging
  to a tier that does not exist yet. The files open unchecked and print a
  constant to show they ran. When a tier claims `die`, the check can come
  back.

One consequence of the format, recorded because it reads as an omission:
no file writes to STDERR. `--- expect output` is compared against stdout,
so a line sent to STDERR lands in neither stream the runner reads, and
asserting it would assert bytes nothing checks. The op stream does not
distinguish them -- `gv[*STDERR] rv2gv print` against `gv[*STDOUT] rv2gv
print` -- so `05_print_bareword_handle.t` measures the bareword construct
through STDOUT and stays observable.
