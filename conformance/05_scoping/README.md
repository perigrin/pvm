# 05_scoping

`my`, `our`, `state`, `local`, and the bare block.

## Why this tier sits here

Tier 01 already writes `my $x = 1`. That is the fact this tier has to
account for before anything else: `my` is not new here. It appears in tier
01 as a FIXTURE -- a literal has to be bound to something before it can be
observed -- and tier 01's README says so. Using a construct is not the same
as introducing it. This is the tier where `my` stops being scaffolding and
becomes the thing under test: where a file asks what `my` DOES rather than
leaning on it to hold a value still.

It cannot come earlier because the question this tier asks is comparative.
A scoping form means nothing alone; it means something next to the other
four. `our $x` is interesting because it is not `my $x`, `local $x` is
interesting because it is not `our $x`, and `state $x` is interesting
because it is not `my $x` on the second call. To ask those questions a file
needs a value to declare (tier 01) and a variable to declare it as (tier
02). Both precede this tier, and nothing else does.

Check 2 asks whether this tier could move earlier. It could not sit before
02, since there is no declaration without a variable to declare. It could
sit later -- `06_control` needs blocks, but a block is also where `my`
becomes visible as scoping rather than as binding, and putting control flow
first would mean introducing the block as loop machinery and only later
noticing it was a scope. The partial order permits the swap; this
numbering does not take it, because the bare block is the smaller
construct and the one whose only job is scope.

What the next tier needs from it is the block. `06_control` is `if`,
`while` and `for` wrapped around a block, and every one of them is a scope
before it is a branch.

## DEPENDS ON

    04_operators

A scoping form is a declaration, and a declaration needs something to
declare. The deeper dependency is tier 02 -- these five constructs do not
ask about context or apply an operator -- but only one tier can be named,
and naming the later one implies the earlier under the partial order.

## INTRODUCES

    enterloop leaveloop once

## Why those ops, and not the ones the source implies

Three ops for five constructs is not tidy, and it is what perl emits.
Measured under 5.42.0 with `perl -MO=Concise,-exec`:

- **No op for `my`.** `my $x = 1` compiles to `const` then
  `padsv_store[$x] vKS/LVINTRO`, and both of those are tier 01's, claimed
  there because tier 01's files emit them. `my`'s pad machinery is already
  spoken for. What this tier introduces is the ops that DISTINGUISH the
  other forms from `my`, not `my`'s own. The honest statement is that the
  op stream cannot see the difference between tier 01's `my` and this
  tier's; only the declared tier can, which is the point of declaring one.

- **No op of their own for `our` or `local` either.** They emit `gvsv` and
  `sassign`, both claimed by tier 02 with the package scalar. What
  distinguishes them from each other is a flag -- `gvsv[*x] s/OURINTR`
  against `gvsv[*x] s/LVINTRO` -- and the ops list does not record flags.
  Two constructs the corpus treats as distinct subjects share one op set
  with a third tier. The files still have to exist separately, because
  their OUTPUT differs and that is what catches a parser that conflates
  them.

- **`once`, which is the whole of `state`.** `state $x = 1` wraps the
  initialisation in `once(other->...)` around an otherwise ordinary
  `padsv_store[$x] vKS/LVINTRO,STATE`. Again the pad op is tier 01's;
  `once` is what `state` adds. It needs `use feature "state"` or a
  `use v5.x` line to compile at all, so the files carry one.

- **`enterloop` and `leaveloop` for a bare block, which is not what the
  name suggests.** A bare block is a loop that runs once -- perl compiles
  `{ ... }` to `enterloop`/`leaveloop` so that `last` and `next` work
  inside it. A block attached to an `if` emits `enter`/`leave` instead,
  tier 01's ops, because nothing can jump out of it. So the same braces
  produce different ops depending on what precedes them, and only the bare
  form is this tier's subject.

- **No op for a block that does nothing.** `{ 1; }` keeps its
  `enterloop`/`leaveloop` under 5.42.0 -- measured, not assumed -- but the
  general warning stands: the optimiser erases constructs. `my $x = 1+2`
  emits no `add`; it folds to `const[IV 3] s/FOLD`. The ops LINT a declared
  tier and cannot derive one, and no op here is claimed that was not seen
  in real output.

This tier is also where the union rule is most visible, because it is the
`my` tier. `padrange` ABSORBS `pushmark`: three consecutive `my`
declarations followed by `print $x, $y` emit one
`padrange[$x; $y] /range=2` where two `padsv` and a `pushmark` would
otherwise stand. A file with MORE declarations can emit FEWER ops. So the
declared set is a UNION ACROSS THE TIER'S FILES and never a property of any
one of them -- the adjacency file that packs declarations together emits a
strictly smaller stream than the construct files it is assembled from.

## FILE ORDER

    accidental

The numbers are the order the files happened to be written in. This README
cites its files by name and never by position, and no claim in it would
become false if the files were renumbered.

`accidental` is a record of debt, not a convention. Renumbering this tier
into `derived` order costs nothing on disk but regenerates the ratchet,
which is a separate change; the declaration exists so the state is written
down rather than rediscovered by the next agent whose numbering test fails.
