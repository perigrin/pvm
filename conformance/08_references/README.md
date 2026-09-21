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

    anonlist ref refgen rv2cv rv2sv srefgen

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
