# 01_literals

Numbers, strings, quoting, `qw`.

## Why this tier sits here

Nothing precedes it. A literal is the smallest thing a program can
contain, and every later tier needs one: tier 02 must put a value in a
variable before it can say anything about variables, tier 04 must have
operands before it can have operators.

Check 2 asks whether this tier could move earlier. It could not — there is
no earlier.

## What the files cover

One construct per file, numbered from 01 with no gap, and `00` reserved
for the adjacency file — which is not one of the tier's constructs but
the check that they compose.

The numbers are DERIVED. `TestTierNumberingRegenerates` rebuilds them from
the file names alone and requires the result to be what is on disk, which
is the spec's promise that "a reorder is a regeneration rather than a
hand-edit" made checkable. A file's identity is its name; the number is
its current position.

The constructs are the boundary cases `../GLOSSARY.md` names for the two
literal categories this tier owns, and `TestTierLiteralsCoversGlossary`
holds the tier to that list rather than to whichever cases occurred to
somebody. Three of them are why the token layer exists at all, because in
each the optree and the output are both blind:

- **`-1` is two tokens** (`05_negative.t`) — perl folds it to
  `const[IV -1]`, so the operator is gone before the optree exists.
- **`5e-1` is one token** (`08_signed_exponent.t`) — split into
  `5e - 1` it still evaluates to 0.5, and `5e` alone is not a number perl
  would ever accept.
- **A v-string is not a number** (`11_vstring_bare.t`, `12_vstring_v.t`)
  — `65.66.67` and `v65.66.67` both print `ABC`, and one dot versus two
  is the whole boundary.

Four of the tier's construct files refuse, and the adjacency file refuses
with them: a body holding every construct the tier introduces holds the
refusing ones too. Composing only what already works would make that file
green and make it stop covering the tier.

## DEPENDS ON

    nothing

This is the only tier that can say that.

`my` and `print` appear in these files as FIXTURES rather than as subjects:
a literal has to be bound to something and observed somehow. `my` is tier
05's subject and `print` is tier 10's. That is the partial order working as
documented — using a construct is not the same as introducing it.

## INTRODUCES

    const enter leave multiconcat nextstate padrange padsv padsv_store print pushmark

## Why those ops, and not the ones the source implies

This list is what `perl -MO=Concise,-exec` EMITS for the files in this
tier, measured, not what reading them suggests. Two places where those
differ:

- **`multiconcat`, not `concat`.** `print "$x\n"` compiles to one
  multiconcat op; the peephole optimiser fuses the interpolation.
- **No `add`, though arithmetic appears nowhere here anyway.** Worth
  stating because `my $x = 1+2` would also emit no `add`: it arrives as
  `const[IV 3] s/FOLD`. The optimiser can erase the very construct a tier
  is about, which is why the ops LINT a declared tier and cannot derive
  one.

- **`padrange`, which no single-statement file emits.** It is the
  optimiser fusing consecutive `my` declarations into one op, so it appears
  only once a file declares several in a row -- which the adjacency file is
  the first here to do. Statement machinery, same class as `enter` and
  `leave`, and claimed for the same reason.

`enter`, `leave`, `nextstate` and `pushmark` are statement machinery that
every file in every tier emits. They are claimed here because this is the
first tier, and a tier claims an op once.
