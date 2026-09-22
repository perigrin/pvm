# 11_oo

`bless` and method dispatch, and the `class`/`field`/`method`/`ADJUST`
object system, including indirect `new Foo`.

## Why this tier sits here

This tier needs references, and nothing else that is not already
universal. Every blessed-hash object is a tier 08 reference plus a string
-- `bless {}, "Foo"` takes exactly those two things -- and `method` needs
tier 07 for a body to mean anything. Both are satisfied well before this
point.

Check 2 asks whether this tier could move earlier. It could move as early
as immediately after `08_references`, and the numbering does not claim
otherwise: the ordering is one topological sort of a partial order, and
`09_regex` and `10_io` are not prerequisites of anything here. What this
tier cannot do is move before 08.

It also could not move LATER in the way an earlier draft had it. That
draft ordered `11_packages` then `12_oo`, on the theory that an object
system consumes `package`. Perl refutes it: `class Foo { method m { 1 } }`
is syntax OK with no `package` anywhere, and `class Foo {}` IS a package
declaration -- a different spelling of tier 12's own construct rather
than a consumer of it. The old order was circular. `use` and `require`
are the heavier construct, because they locate and load a file and then
call `import`, and they move after this tier for that reason.

The tier after this one needs nothing from it. `12_packages` does not
depend on `11_oo`; the swap was made to remove a false dependency, not to
create the reverse one.

## DEPENDS ON

    08_references

`bless` is the whole of the dependency. The adjacency file pairs with the
declared prerequisite rather than with tier 10, which this tier does not
use -- pairing with 10 would assert nothing.

## INTRODUCES

    anonhash bless emptyavhv isa method method_named method_super methstart shift stub
    tie tied

## Why those ops, and not the ones the source implies

This list is what `perl -MO=Concise,-exec` EMITS for the files in this
tier, measured, not what reading them suggests. This tier has more
distance between source and op stream than any other, and most of it is on
the `class` side.

- **`class`, `field`, `ADJUST` and `:isa` emit no op of their own.** A
  class body at file scope compiles to `enterloop` / `leaveloop` -- a bare
  block, tier 05's ops -- with `nextstate` for each declaration inside it
  and nothing else. An empty body adds `stub`. `field $z :param = 7;`
  contributes one `nextstate` and no store: the initialiser is compiled
  into a field-init tree that is not reachable from the main optree.
  `ADJUST` becomes an anonymous sub in the class stash with no CODE slot,
  so B::Concise cannot name it. `:isa(Base)` emits nothing at all; it is
  resolved entirely at compile time.

- **`methstart` and `shift` are the two spellings' prefixes, and this
  tier owns both.** Every `method` body opens with `methstart`, which
  binds the invocant and the field pad; a `sub` body doing the same job
  opens with `shift`. Two spellings of the same object system produce two
  disjoint op prefixes, so a parser that handles one learns nothing about
  the other, which is the argument for this tier covering both systems
  rather than picking one.

  Both were unclaimable until `opsOf` learned to look inside a CV.
  `perl -MO=Concise,-exec file.pl` with no sub named dumps THE MAIN
  PROGRAM ALONE, and a method body is a CV, so nothing a file could be
  written to do put `methstart` where the lint looked. `opsOf` now
  enumerates the file's own subs and dumps each beside the main program,
  and both ops are measured where they are emitted -- `methstart` from
  `00_adjacency.t` and `08_class_field_method.t`, `shift` from
  `06_indirect_new.t`.

  **`shift` is claimed HERE rather than in tier 02, and the reason is a
  gap rather than a judgement.** The op is not really this tier's
  subject: `shift @a` outside any sub emits it too, and tier 02 is where
  arrays live. But the corpus's rule is that the tier which EMITS an op
  first owns it, and no tier-02 or tier-04 file emits `shift` at all --
  claiming it there would fail `TestCorpusLints`, which reports a README
  claiming an op no file emits. A tier-02 file exercising `shift @a`
  would move the claim to where it belongs, and until one exists the
  claim sits with the first tier that actually compiles the op.

  `leavesub`, which every method body here also ends in, belongs to tier
  07: it is emitted first by the subroutine tier, which is where the
  corpus introduces sub bodies at all.

- **`Foo->new` and `new Foo` are the same op stream.** Indirect object
  notation is a hard marker placed here, but it is a LEXING problem, not a
  compilation one: both spellings emit `pushmark`, `const[PV "Foo"] sM/BARE`,
  `method_named`, `entersub`. Nothing downstream of the lexer can tell them
  apart, so the corpus has to assert on tokens.

  **Both halves of that assertion, not one.** `06_indirect_new.t` declares
  `no operator whose text is "->"`, which is where the construct is
  visible. A negative alone is satisfied VACUOUSLY by a lexer that never
  emits `->` at all -- one that folded the arrow into the word beside it
  passes the indirect file's claim perfectly while getting every direct
  call in this tier wrong. So `03_method_named.t`, the direct spelling,
  declares the matching `one operator whose text is "->"`. The pair is
  falsifiable where either half alone is not.

- **`method` AND `method_named`, which are different ops.** `$o->hi`
  resolves the name at compile time and emits `method_named`; `$c->$m`
  with the name in a variable emits `method`. `$s->SUPER::hi()` emits a
  third, `method_super`. Measuring only the first spelling would miss two
  of the three. The `entersub` and `leavesub` that surround all three are
  tier 07's; what this tier adds is how the callee is found.

  `SUPER::` resolves against the package the call is COMPILED in, not the
  invocant's class, so `05_method_super.t` makes the call at file scope
  inside `package Derived;` rather than from a method body. Written the
  usual way the op would sit in a sub's own optree, where -- see
  `methstart` above -- nothing measures it.

- **`isa` is an INFIX OPERATOR and is not dispatch at all.** `$o isa Foo`
  emits `<2> isa` -- a binary op -- with no `entersub` and no
  `method_named`, where `$o->isa("Foo")` emits both. One word, two
  unrelated parses, and only the method spelling goes through the
  machinery the rest of this tier is about. Measuring only that spelling
  would have named the half that is not distinctive.

  It is FEATURE-GATED, which is the second half of its parse: measured,
  without the feature the infix form is a SYNTAX ERROR rather than a
  weaker parse, so whether the word is an operator at all depends on a
  pragma earlier in the file. It is no longer experimental under 5.42 --
  `use v5.36` with `$o isa Foo` warns about nothing, where `class` still
  needs its `no warnings` line. `10_isa_infix.t` carries it, and asserts
  lexically for this tier's usual reason: both spellings print the same
  thing, so only the tokens separate them.

- **`can` introduces NO op, and that is why it earns a file.** The whole
  construct is how `$o->can("hi")->($o)` is parsed, and measured, that is
  TWO `entersub` under ONE `method_named`: the first arrow is a method
  call and the second dereferences the code ref `can` returned. A parser
  that read both arrows as method calls, or that read the second as part
  of the first call, produces an ordinary-looking op stream with the
  wrong counts -- and `opsOf` collects op NAMES, so it can see neither
  the counts' shape nor the flags (`sKRS` for the method call, `lKS` for
  the deref) that say the same thing. `11_can_chain.t` carries it.

  The receiver is passed explicitly in that spelling and must be:
  calling a code ref passes no invocant, so `$o->can("hi")->()` calls
  the method with an EMPTY `@_`. That is the semantic content of "only
  the first is a method call".

- **`emptyavhv` and `anonhash`, not one anon-hash op.** `bless {}, $c`
  emits `emptyavhv ... /ANONHASH`; `bless { %a }, $c` emits `pushmark`,
  `padhv`, `anonhash`. The optimiser has a dedicated op for the empty
  case, so the file that looks simpler is the one with the op nobody
  expects. `ref`, which a constructor's caller uses to check the result,
  is tier 08's.

- **`tie` and `tied` are here because `tie` needs a BLESSED object**, and
  that is a measurement rather than a theme. Both words were unclaimed by
  every tier of this corpus. `tie` was proposed for `07_subroutines`, on
  the ground that `TIESCALAR` and `FETCH` are ordinary named subs -- true,
  and not sufficient. Measured 5.42.0, a constructor that returns an
  unblessed reference does not error; the tied variable silently reads as
  empty:

      $ perl -e 'package C; sub TIESCALAR { my $s = "x"; return \$s }
          sub FETCH { return 42 }
          package main; tie my $c,"C"; print "[", $c, "]\n";'
      []

  So the blessing is what the protocol requires, and the minimal blessed
  constructor emits `emptyavhv` and `bless` -- both this tier's, both
  counted against the file because `opsOf` reads inside a file's own CVs.
  Tiers 02, 07 and 12 were each tried and each failed the op-budget lint.
  This is the earliest tier that can hold either word.
  `07_subroutines/README.md` records the same finding from the other side.

  **They are two argument grammars, not one.** `tie` is a LIST OPERATOR
  whose first argument is a VARIABLE -- `tie my $x, "C", @args` declares
  `$x` in the call itself, measured as `padsv ... /LVINTRO` under the
  `tie`'s own mark -- and `tied` is a NAMED UNARY. Measured, `<@> tie
  vK/3` against `<1> tied sK/1`: a mark and three children against one
  child and no mark. A parser that gave the two one grammar is caught by
  the pair. `12_tie_variable.t` and `13_tied_boolean.t` carry them.

  `tied` needs neither tier 08's `ref` nor tier 04's `defined`, which an
  earlier draft assumed. In boolean position it answers on its own: undef
  for an untied variable, a blessed reference for a tied one. Both would
  have been legal -- both tiers are earlier -- and neither is needed.

- **`Foo::new` under `class` is XS code.** The generated constructor has
  no Perl optree to dump. Nothing this tier introduces can be measured
  through it.

- **The pragmas `class` needs, and `package`, are free.** Measured:
  `use feature 'class'`, `no warnings 'experimental::class'` and a bare
  `package Foo;` each emit NO runtime op -- the op stream of every file
  here begins at the first statement after them. `use` is tier 12's
  construct and a runtime cost would have put this tier's `class` files
  out of budget; a compile-time pragma does not. `use v5.42;` is not
  enough on its own: it does not enable `feature 'class'` in 5.42.0, and
  `class Foo` under it is a syntax error, so the files name the feature.

The ops LINT this declared tier; they cannot derive it. `class` is the
sharpest case in the corpus of a construct the optimiser erases: four
keywords that a reader would call the subject of the tier produce, between
them, tier 05's `enterloop` and `leaveloop` plus `stub` and `methstart`.

One more thing the op stream cannot see, and the reason this tier's
adjacency file exists:

    class Foo { ADJUST { 1 } }                   parses, 0 Unknowns
    class Foo { ADJUST { 1 } method m { 2 } }    1 Unknown, swallowing both

`ADJUST` alone parses. `ADJUST` followed by anything -- `method`, `field`,
or a second `ADJUST` -- does not. A corpus of one construct per file goes
green over this by construction, because every construct in it passes in
isolation. The adjacency file

    class Foo { field $x; ADJUST { 1 } method m { 2 } }

puts every construct this tier introduces in one body, each adjacent to
another, and is what catches it. This is the case the whole
adjacency-file design was written for.

## HARD MARKERS

    indirect-new	new·

One of the twelve `hardMarkers` places here, and it is the widest of them:
119 T1 files, 28.6% clean. The list came from
`internal/parse/easy_test.go`, which used it to carve an "easy" subset out
of T1; in a graded corpus each entry is a tier placement rather than a
filter, so that file was deleted and the placements moved to the tiers.

`indirect-new` is `new Foo(...)`, spec §4.14.2's `MethodCall.Indirect` row,
and it places here because indirect object syntax is a method call -- it is
`Foo->new(...)` wearing the other word order, and there is no earlier tier
at which a method exists to call. `06_indirect_new.t` is the file.

The probe is deliberately the crude one the marker list used: a bare
`new` followed by a space. It over-matches in prose and under-matches
`new($x)`, and neither costs anything here, because the probe's job is to
prove this tier exercises the construct rather than to classify T1.

`TestEveryHardMarkerPlaced` in `internal/conformance` requires a file here
that the probe finds, `·` standing for the significant trailing space.

## FILE ORDER

    accidental

The numbers are the order the files happened to be written in. This README
cites its files by name and never by position, and no claim in it would
become false if the files were renumbered.

`accidental` is a record of debt, not a convention. Renumbering this tier
into `derived` order costs nothing on disk but regenerates the ratchet,
which is a separate change; the declaration exists so the state is written
down rather than rediscovered by the next agent whose numbering test fails.
