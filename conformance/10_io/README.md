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

    close eof open readline rv2gv say select sselect

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
  `readline`, identical to `<ARGV>` -- measured, op for op. There is no
  distinct operator to claim; the diamond is a spelling the lexer must
  recognise, which is a token fact and not an op fact. `<*>` is glob and
  belongs to tier 13.

  The diamond itself is not a file here, and the reason is the same one
  that makes every handle in this tier in-memory: `<>` with an empty
  `@ARGV` reads STDIN, so what it prints depends on what the runner
  hands the child process rather than on the program. That is a property
  of the harness, not of perl, and a corpus file resting on it would pin
  bytes the source does not determine. The token fact the diamond would
  have carried is carried instead by the spelling this tier CAN run
  reproducibly, `<$fh>`, which produces the same token category.

**`say` is here rather than beside `print` in tier 01, and `select` is
two ops.** Both were added after the first four files and neither is a
print variant.

`say` takes a filehandle through the SAME `rv2gv` as `print` --
measured, `say $out $line` is `padsv[$out] rv2gv sKR/1 padsv[$line] say
vKS` against `print $out "x"`'s `padsv[$out] rv2gv sKR/1 const print
vKS` -- so the machinery it needs is this tier's and not tier 01's. What
it adds is a record separator, which is a property of writing to a
handle.

Its sharper property is that it is FEATURE-GATED and the featureless
spelling is NOT a syntax error. Measured, `say $x` without the pragma
compiles clean as INDIRECT-OBJECT METHOD DISPATCH -- `pushmark padsv
method_named[PV "say"] entersub` -- and fails only at runtime, exit code
255. Identical bytes, two valid parses, separated by a pragma earlier in
the file. That is `11_oo/10_isa_infix.t`'s finding met from the worse
side: there the featureless form is a syntax error and fails loudly,
here it is a clean compile of the wrong program.

**The corpus cannot pin that parse from this tier**, and the reason is
the lint working as designed. The featureless form emits `method_named`,
which is `11_oo`'s op, and 11 is LATER than 10. A file may only emit ops
its tier or an earlier tier claims, so the measurement is RECORDED in
`07_say.t`'s header and asserted nowhere. A file that pins it belongs in
tier 11 beside the `isa` file it mirrors. (`entersub`, the other op the
featureless form emits, is tier 07's and would have been a legal use;
`method_named` alone is what blocks it.)

`select` is the only construct in this tier where ONE SPELLING EMITS TWO
DIFFERENT OPS, which is why INTRODUCES names both. Measured,
`select($out)` emits `select sK/1` and `select(undef,undef,undef,0)`
emits `sselect sK/4` -- a different op name, not a different flag. The
arity decides, and it decides at COMPILE time: arity 2 and arity 3 are
both compilation errors naming the syscall. This is the exact inverse of
`rv2gv`, which does three unrelated jobs in this tier under one op name.

`08_select.t` measures the one-argument form entirely through a `print`
that names NO handle: the statement is unchanged, tier 01's `print vK`
either way, and its bytes land in a buffer because the line above
changed what "no handle" means.

The ops also cannot tell a successful open from a failed one, which is
the general caveat restated for this tier: they LINT a declared tier and
cannot derive one.

## What writing the files measured

The declared set came out of the corpus unchanged -- every op in
INTRODUCES is emitted, nothing had to be removed, and nothing outside the
earlier tiers' claims appeared. Three facts the files forced, none of
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

- **The behavioural files cannot see the angle split, and token facts
  can.** `<$fh>` and `<*.c>` share a spelling across two tiers, and the
  files here assert the first entirely through behaviour -- what gets
  read, what gets printed. Measured, that is not enough: our parser reads
  `<`, `$fh`, `>` as an ordinary comparison chain and returns ZERO
  Unknowns for it, so a lexer that split the angles would pass every
  `expect parses` and every `expect output` in this tier without a word.
  `02_readline_scalar.t` and `03_readline_list.t` therefore carry two
  token facts each -- one readline operator spelled `<$fh>`, and NO
  operator spelled `<` -- and the second is the one that falsifies a
  split lexer, which emits that operator where an unsplit one never
  does. Tier 13's `05_glob_angle.t` makes the other half of the split for
  `<*.nonexistent-xyz>`; before these facts the corpus claimed the
  category for the glob spelling alone.

One consequence of the format, recorded because it reads as an omission:
no file writes to STDERR. `--- expect output` is compared against stdout,
so a line sent to STDERR lands in neither stream the runner reads, and
asserting it would assert bytes nothing checks. The op stream does not
distinguish them -- `gv[*STDERR] rv2gv print` against `gv[*STDOUT] rv2gv
print` -- so `05_print_bareword_handle.t` measures the bareword construct
through STDOUT and stays observable.

## FILE ORDER

    accidental

The numbers are the order the files happened to be written in. This README
cites its files by name and never by position, and no claim in it would
become false if the files were renumbered.

`accidental` is a record of debt, not a convention. Renumbering this tier
into `derived` order costs nothing on disk but regenerates the ratchet,
which is a separate change; the declaration exists so the state is written
down rather than rediscovered by the next agent whose numbering test fails.
