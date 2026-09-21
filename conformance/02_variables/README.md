# 02_variables

Sigils, scalars, arrays, hashes, element access, `delete`/`exists`, `$::`.

## Why this tier sits here

A variable needs something to hold. Tier 01 establishes the values; this
tier is the first place a program can name one and get it back. Nothing
before it can do that -- tier 01's `my $x` is a fixture there, a place to
put a literal so the literal can be observed, and the file says so.

It cannot move earlier because there is only one earlier tier and every
file here begins by assigning a literal into something. It cannot move
later because tier 03 has nothing to discriminate without it: `scalar vs
list` is a question about `@a` and `%h`, and a corpus with no aggregate in
it cannot ask the question, let alone answer it. `my $n = @a` -- the
canonical context pair -- is two tier-02 constructs and one tier-03 claim
about what happens between them.

Check 2 asks whether this tier could move earlier. It could swap with
nothing: the only candidate is 01_literals, and an array with no elements
to put in it is not a smaller program, it is an untestable one.

## DEPENDS ON

    01_literals

Every file here assigns literals into a variable and prints the result,
because a variable is only observable through a value. `my`, `our` and
`print` appear as FIXTURES, as they did in tier 01: `my` is tier 05's
subject and `print` is tier 10's. Using a construct is not introducing it.

## INTRODUCES

    aassign aelem aelemfast aelemfast_lex aelemfastlex_store aslice av2arylen delete gv gvsv helem hslice multideref padav padhv rv2av rv2hv sassign

## Why those ops, and not the ones the source implies

This list is what `perl -MO=Concise,-exec` EMITS for the constructs in this
tier, measured under 5.42.0, not what reading the source suggests. Five
places where those differ, and the first is the largest gap in the corpus
so far.

- **No `exists` op, and usually no `delete` op.** `delete $h{a}` compiles
  to `multideref($h{"a"}) vK/DELETE` and `exists $h{a}` to
  `multideref($h{"a"}) sK/EXISTS`. The construct survives only as a FLAG on
  an op named after something else. `delete` appears as its own op only
  when the optimiser declines to build a multideref -- a hash SLICE
  (`delete @h{"a","b"}`) or a subscript that is itself an element lookup
  (`delete $h{$k[0]}`). A tier declared from the op names alone would
  therefore contain no notion of `exists` at all, and would contain
  `delete` for the wrong reason. This is the same failure tier 01 recorded
  for `my $x = 1+2` emitting no `add`, in its most extreme form here: the
  ops LINT a declared tier, they cannot derive one.

- **`multideref` swallows most element access, and the ops that remain are
  the exceptions.** `$h{a}` and `$a[$i]` are both multideref. `helem` and
  `aelem` survive only where the subscript is an expression the optimiser
  will not fold into the deref chain -- `$h{$k[0]}` and `$a[$#a]`. So
  `helem`/`aelem`, the ops a reader would expect to be the tier's centre,
  are in fact its edge cases, and `multideref` is the common path.

- **Array element access has four ops depending on how it is spelled.**
  `$a[0]` on a lexical reads as `aelemfast_lex` and writes as
  `aelemfastlex_store`; the same subscript on a package array is
  `aelemfast` behind `rv2av`. Constant subscript, lexical or package, read
  or write -- four spellings of one construct, which is why the tier's
  adjacency file has to contain all four rather than one representative.

- **`sassign` arrives with the package scalar, not the lexical one.** Tier
  01's `my $x = 0.5` emits `padsv_store` and no `sassign` at all; `our $x =
  1` emits `sassign` over a `gv`. The plain scalar assignment op is
  introduced by `$::`, four constructs away from where a reader would look
  for it.

- **`${name}` leaves no trace at all.** `${x}` is `padsv` and `@{a}` is
  `padav`, identical to the unbraced spellings: perl resolves the braces
  away before the optree exists. GLOSSARY.md nonetheless calls it out --
  "`${name}` is one variable: the braces are punctuation around a name" --
  because the lexing question is real and is the same `${` that opens the
  dereference `${ $ref }` at tier 08. So the boundary is covered by
  `05_braced_name.t` and asserted on the SOURCE, the way tier 01 asserts
  `-1`'s two tokens against a folded constant.

Two further measurements shape the tier's files rather than its op list.
`print "@a"` emits `join` and `gvsv` -- the `gvsv` is `$"`, which the
interpolation reads -- and `print "$a[0]"` emits `stringify`. Neither is
claimed here: `join` and `stringify` belong to what the interpolation does
with the values, not to naming them, and a file that prints `"@a"` would
drag two unclaimed ops into the tier. The tier's files print elements and
counts instead, the same way tier 01's adjacency file prints `qw` as a
`print` LIST to keep `aassign` and `padav` out of it.
