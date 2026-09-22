# 08_references

Taking references with `\`, dereferencing, the arrow, anonymous constructors,
`${ }` and `@{ }`.

## Why this tier sits here

A reference is a value that names another value, so this tier cannot come
before there is something to point at. Tier 02 supplies the arrays and
hashes that `\@a` and `\%h` take a reference to, and tier 07 supplies the
subroutines that `\&foo` and `$c->()` need -- a code reference called
through the arrow is tier 07's `entersub` reached by a route tier 07 does
not have.

Could this tier move earlier? Only as far as tier 03. The op measurements
say no further: `\@a` emits `padav` before `srefgen`, and `@{$r}` emits
`rv2av` whose operand must already be a scalar holding a reference, so
both spellings need tier 02's aggregates to exist first. What pins it below
07 rather than at 03 is the code reference: `\&foo` compiles through
`rv2cv` on a named sub, and a tier that introduced that before subroutines
existed would be claiming an op for a construct it could not write.

What the next tiers need from it is not symmetric. Tier 09 needs nothing
here -- the numbering is one topological sort of a partial order, and
regex could sit beside references rather than after them. Tier 11 is the
real consumer, and it is load-bearing: `bless` takes a reference and a
string, and the spec's statement that "every blessed-hash object is tier 08
plus a string" is literally an op-level claim. Without this tier, `bless
{}, "Foo"` has no anonymous hash to bless, `$self->{name}` has no
`multideref`, and `$self->method` has nothing for the arrow to mean.
`bless` without tier 08 is a function of two arguments neither of which the
corpus can construct. The two hard markers `deref-brace` and `deref-at`
land here for the same reason -- they are the spellings tier 11's method
bodies are written in.

## DEPENDS ON

    07_subroutines

Tier 02 is the deeper dependency for the data side, and tier 07 for the
code side; only one can be named, and 07 is the later of the two, so naming
it implies both under the partial order.

## INTRODUCES

    anonlist prototype ref refgen rv2cv rv2sv srefgen

## Why those ops, and not the ones the source implies

Measured with `perl -MO=Concise,-exec` under 5.42.0, taking the union
across the tier's files and subtracting what earlier tiers claim. Five
things the source does not predict:

- **`$r->[0]` and `${$r}[0]` are NOT the same ops.** This was worth
  measuring because if they had been identical the op set could not tell
  them apart. They differ: the arrow form emits a single
  `multideref($r->[0])` that has absorbed the pad lookup, while the brace
  form emits `padsv[$r] sM/DREFAV` followed by `multideref(->[0])` with an
  empty base. So the op stream does discriminate the pair -- but only by
  op COUNT and by the private flag, not by op NAME, since both spellings
  use `multideref` and the brace form's extra op is tier 01's `padsv`.
  A tier lint that compares name sets alone sees the two files as
  identical. The files must therefore distinguish them in `expect tokens`,
  which is the one check that reads the spelling rather than the result.

- **`multideref`, not `helem`/`aelem`/`rv2av`.** The peephole optimiser
  fuses an entire subscript chain into one op: `$r->[0][1]` and
  `$r->{a}{b}` each emit exactly one `multideref`, not two element
  lookups. A file exercising a three-level chain emits FEWER ops than one
  exercising a single subscript plus a separate deref, which is the same
  effect `padrange` has in tier 01 -- more construct, fewer ops. The
  declared set is the union across the tier's files and is not a property
  of any one of them. `multideref` itself is tier 02's op, claimed there
  with ordinary element access; this tier reaches it by a different route
  rather than introducing it.

- **`srefgen` and `refgen` are two ops for one `\`.** Scalar context gives
  `srefgen`; list context gives `refgen`. `my $s = \@a` emits `srefgen`
  and `my @r = \(@a)` emits `refgen`, from source that differs only in
  parentheses and the assignment target. Both are claimed because the tier
  writes both.

- **The `rv2*` family collapses several spellings each.** `@{$r}` and
  `@$r` and `$r->@*` all emit `rv2av` and are indistinguishable in the op
  stream; likewise `%{$r}`/`%$r`/`$r->%*` give `rv2hv` and
  `${$s}`/`$$s`/`$s->$*` give `rv2sv`. `rv2av` and `rv2hv` are tier 02's,
  emitted there by package aggregates; only `rv2sv` and `rv2cv` are new
  here. `rv2cv` appears only for `\&foo` on a named sub -- `&{$r}()` and
  `&$r()` both compile to `entersub` instead. `$$$rr` is two stacked
  `rv2sv`, so depth is visible as repetition but spelling is not visible at
  all. Everything this family cannot see has to be asserted in
  `expect tokens`.

  Measured: the optrees of `10_deref_at_sigil.t` and `11_deref_postfix.t`
  are identical op for op, differing only in the subscript constant their
  last statement reads. Saying that and asserting nothing lexical would be
  the same corpus shape tiers 09 and 10 were found in -- behaviour pinned,
  spelling unmeasured -- so `TestTierReferencesSpellingsAreDistinguished`
  requires a token fact per spelling, and requires the file making it to
  contain the spelling OUTSIDE a string. That second half is not
  decoration: `06_deref_at.t` originally wrote `@{$r}` only inside
  `"@{$r}\n"`, where our lexer emits one `Quote` token and the deref is
  never tokenised, so its facts were satisfied entirely by the `\@a` two
  lines above and the file asserted nothing about its own construct.

- **`prototype`, which takes this tier's own construct as its argument.**
  `prototype \&f` emits `rv2cv` and `srefgen` -- both above -- and then
  `prototype` on top of them, so the op measurement places it here
  without a judgement call: the construct is not writable in any earlier
  tier, because `\&f` is not. Tier 07 owns what a prototype DOES to a
  call site; this op is the prototype read back out as a string, and the
  string is the text the parser was given. `12_prototype_builtin.t`
  carries it and says explicitly what it does not repeat from
  `07_subroutines/09_prototype_extent.t`.

- **`ref`, which the tier description does not mention, and `anonlist`,
  which it does.** `ref $r` emits `ref`, a reference operation with no home
  in an earlier tier. `[1,2]` emits `anonlist`. The anonymous HASH
  constructor does not appear here: `{}` and `{ %a }` are tier 11's
  `emptyavhv` and `anonhash`, first emitted where objects are built.
  `$#{$r}` emits `av2arylen` on top of `rv2av`, both tier 02's.

Ops the tier's files emit but do not claim: `entersub` belongs to tier 07,
`aslice`/`hslice`/`padav`/`padhv`/`aassign`/`av2arylen`/`multideref`/`rv2av`/`rv2hv`
to tier 02, and `gv`/`shift` are fixtures used to defeat constant folding
-- `my $r = \1` folds to `const[IV \1] s/FOLD` and emits no `srefgen` at
all, which is the tier-01 lesson recurring here. The optimiser can erase
the construct a tier is about, so these ops lint a declared tier and cannot
derive one.

## HARD MARKERS

    deref-brace	${
    deref-at	@{

Two of the twelve `hardMarkers` place here, and they are the two the list
itself records as having MOVED: re-measured when 01a0ad52 landed,
`deref-brace` went 27.1% -> 35.7% clean and `deref-at` 14.5% -> 23.6%. The
list came from `internal/parse/easy_test.go`, which used it to carve an
"easy" subset out of T1; in a graded corpus each entry is a tier placement
rather than a filter, so that file was deleted and the placements moved to
the tiers.

Both place here for the same reason and it is not a close call: `${...}`
and `@{...}` are dereference syntax, and this is the dereference tier.
`05_brace_deref.t` and `07_deref_brace_hash.t` back the first,
`06_deref_at.t` and `10_deref_at_sigil.t` the second, and the adjacency
file carries both.

The probes need no trailing space -- a brace cannot begin an identifier, so
`${` and `@{` are already unambiguous. `TestEveryHardMarkerPlaced` in
`internal/conformance` requires a file here that each probe finds.

## FILE ORDER

    accidental

The numbers are the order the files happened to be written in. This README
cites its files by name and never by position, and no claim in it would
become false if the files were renumbered.

`accidental` is a record of debt, not a convention. Renumbering this tier
into `derived` order costs nothing on disk but regenerates the ratchet,
which is a separate change; the declaration exists so the state is written
down rather than rediscovered by the next agent whose numbering test fails.
