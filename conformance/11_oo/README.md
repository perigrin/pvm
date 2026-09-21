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

    anonhash bless emptyavhv method method_named method_super stub

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

- **`methstart` exists, and this tier cannot claim it.** Every `method`
  body opens with `methstart`, which binds the invocant and the field
  pad; a `sub` body doing the same job opens with `shift` -- tier 02's
  op. Two spellings of the same object system produce two disjoint op
  prefixes, so a parser that handles one learns nothing about the other,
  which is the argument for this tier covering both systems rather than
  picking one.

  But `methstart` is not in the INTRODUCES list, because the lint cannot
  see it. `perl -MO=Concise,-exec file.pl` with no sub named dumps THE
  MAIN PROGRAM ALONE; reaching a method body takes `-exec,Foo::m`, and
  the lint runs the fixed command. A method body is a CV, so nothing a
  file can be written to do puts `methstart` in what the lint measures.
  Claiming it anyway would fail `TestCorpusLints`, which reports an op a
  README claims and no file emits -- measured, and the reason this entry
  was removed rather than a file being written to back it.

  The same wall stands in front of `leavesub` and `argcheck` in tier 07
  and every other op that lives only inside a sub. What would move it is
  an `opsOf` that names each sub in the file and dumps those too; until
  then the corpus measures main-scope ops, and the op streams quoted in
  `08_class_field_method.t` record what the method body holds even though
  nothing lints it.

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

- **`emptyavhv` and `anonhash`, not one anon-hash op.** `bless {}, $c`
  emits `emptyavhv ... /ANONHASH`; `bless { %a }, $c` emits `pushmark`,
  `padhv`, `anonhash`. The optimiser has a dedicated op for the empty
  case, so the file that looks simpler is the one with the op nobody
  expects. `ref`, which a constructor's caller uses to check the result,
  is tier 08's.

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
