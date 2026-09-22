# 07_subroutines

Declaration, call forms, `@_`, return, signatures.

## Why this tier sits here

A subroutine needs a body, and a body is statements: the tier cannot
precede 05_scoping, because `my $x = shift` inside a sub is the argument
idiom and it is a `my` declaration first, nor 06_control, because a sub
whose body has no branch never needs an explicit `return` at all -- the
last expression is the value, and measured under 5.42.0 `sub f { return
(1,2) }` emits no `return` op. The `return` op only survives where control
leaves early, which means the construct this tier is named for is
observable only once the control tier exists.

Check 2 asks whether this tier could move earlier. It could not move
before 06_control without the `return` op disappearing from the corpus
entirely, and it could not move before 02_variables, because `@_` is an
array and `$_[0]` is element access -- the tier's own argument protocol is
spelled in tier 02's syntax and nothing else.

What the next tiers need from it is concrete in two places. 08_references
takes `\&f` and `$code->()`, both of which are a subroutine wrapped in a
reference; the sub has to exist before there is anything to take a
reference to. And 11_oo needs this tier before `method` means anything:
`method m { ... }` is a subroutine declaration with an implicit first
argument, and `$obj->m` is `entersub` with a receiver -- so without tier
07 the OO tier would be introducing the call machinery and the dispatch
rule in the same file, and a failure in either would look like a failure
in the other. That is the forward dependency the spec records when it
places `signature` among the twelve hard markers here: the marker belongs
to this tier so that tier 11 can use signatures rather than introduce
them.

## DEPENDS ON

    06_control

The declared prerequisite is the tier without which this tier's own
subject stops being observable: with no branch in the body, `return`
compiles to nothing at all. Tiers 01 through 05 are also used -- `@_` is
tier 02's array, the body's declarations are tier 05's -- and the lint's
`ops(file) ⊆ ∪ ops(tiers ≤ N)` covers those, since every one of them sorts
earlier. `DEPENDS ON` names the single prerequisite the adjacency file
pairs against, not the whole reachable set.

## INTRODUCES

    anoncode argcheck argdefelem argelem entersub leavesub lock postinc
    return warn

Nine ops. `lock` joined the original eight and is argued for at the end
of this section; the sentence below is about those eight.

Eight ops, six of which this tier could not claim until the lint learned
to look inside a CV. `argcheck`, `argdefelem`, `argelem`, `leavesub` and
`return` are the compiled form of a signature, a body's terminator and a
branching return -- all real, all emitted by the files in this directory,
and all structurally unreachable while `opsOf` ran `perl
-MO=Concise,-exec` with no sub named. `opsOf` now enumerates the file's
own subs and dumps each beside the main program. The last section of this
file records what that changed.

`postinc` is the odd one, and it is claimed here for a gap rather than on
merit; see the note on it below.

## Why those ops, and not the ones the source implies

This list is what perl EMITS for the files in this tier -- the main
program and every named sub they declare -- measured under 5.42.0, not
what reading them suggests. Six places where those differ:

- **The sub body is not in the main program's op stream.** `perl
  -MO=Concise,-exec -e 'sub f { ... } f()'` prints `enter`, `pushmark`,
  `gv[IV \&main::f]`, `entersub`, `leave` and stops -- the body is absent.
  Naming the sub gets it: `-MO=Concise,-exec,f`. This fact used to
  determine the INTRODUCES list above, because the lint ran the first
  invocation and never the second. `opsOf` now runs both: it enumerates
  the CVs the file itself declared and dumps each alongside the main
  program, so a body's ops are measured where they are emitted. What
  stays out of reach is an ANONYMOUS sub's body -- it has no name to pass
  and no glob to walk to -- so `sub { ... }` still contributes
  `anoncode` at the point the closure is built and nothing from inside.

- **`return` is mostly erased.** `sub f { return $_[0] + 1 }` emits
  `add` then `leavesub` and no `return` op; so does `sub f { return
  (1,2) }`, which additionally loses its `pushmark`. The op appears only
  where the return is not the last thing the body does -- `return 1 if
  $_[0]` compiles `and` guarding a `pushmark`/`return` pair. So the
  source construct `return` and the op `return` are different things.
  `05_return.t` branches and so does emit the op, into its own CV --
  which the lint now reads, so the tier claims it.

- **A signature is NOT erased.** This is worth stating because it is the
  opposite of what the rest of this file warns about. `sub f ($a, $b)`
  compiles to `argcheck(2,0)` followed by one `argelem` per parameter --
  distinct ops carrying the arity, not ordinary pad assignments that
  happen to read `@_`. A default adds `argdefelem` guarding the default
  expression, and a slurpy `@rest` shows in the `argcheck` flags as
  `argcheck(1,0,@)`. Compare `sub f { my ($a,$b) = @_ }`, which is
  `padrange` and `aassign` and nothing else: the two spellings are
  genuinely different optrees, so here the ops DO describe the construct
  -- and the tier now claims them, because the lint reads inside the sub.
  This was the most painful case while it lasted: the one construct in
  the tier whose ops make a precise claim was the one the lint was least
  able to reach.
  `use v5.36` itself adds no ops -- measured, it changes the `nextstate`
  hint flags from `v:{` to `v:us,*,&,{,$,fea=6` and nothing more, so the
  pragma `06_signature.t` needs costs the op stream nothing.

- **`warn` is claimed here because the BUILTIN extent question is this
  tier's, and it is the only one that can be observed.** `warn "a", "b"`
  takes both arguments and `warn("a"), "b"` takes one, which is exactly
  the greedy/cut split `08_parenless_extent.t` and `09_prototype_extent.t`
  measure for user subs -- with the extent decided by a PAREN at the call
  site rather than by a declaration. Two routes to one question, and a
  parser can have either and not the other. `12_builtin_extent.t` carries
  it. `die` asks the identical question and 218 T1 files use it, and it
  is claimed by no tier: measured, `$SIG{__DIE__}` does not prevent
  termination and a `die` under a false guard is a dead branch that says
  nothing about extent, so the only observable spelling is `eval { die }`
  and the corpus has no `eval` yet. When a tier claims `eval`, `die`'s
  extent file can be written against `12_builtin_extent.t` as its pair.

  The handler is what makes `warn` observable at all. It writes to
  STDERR, which `--- expect output` does not read -- the same constraint
  `10_io/README.md` records -- and `local $SIG{__WARN__} = sub { }` makes
  perl call the handler INSTEAD of writing, leaving `warn`'s return value
  as the only thing to count.

- **`entersub` covers every call form, and the differences are flags.**
  `f(1)`, bare `f`, `&f` and `&f(2)` all emit `entersub`; the ampersand
  forms carry `/AMPER` and the strict-mode call carries `/STRICT`. One op,
  four syntaxes, which is why this tier needs its own files per form
  rather than trusting the op list to distinguish them.

- **An anonymous sub's body is indistinguishable from a named one's.**
  Installed under a name at BEGIN time so B::Concise can reach it, `sub {
  my $x = shift; $x }` emits exactly the body a named sub emits, ending in
  `leavesub` -- there is no `leaveanonsub`. The only op the anonymous form
  adds is `anoncode`, in the enclosing scope, where the closure is built.

- **`@_`'s ALIASING is not an op, and no op comparison can find it.**
  Measured, `sub bump_alias { $_[0]++ }` and `sub bump_copy { my ($n) =
  @_; $n++ }` differ only in what the increment is applied to --
  `aelemfast[*_]` then `postinc` against `padsv` then `postinc` -- and
  the second body is the longer one, since the copy costs a `padrange`
  and an `aassign` the alias does not. Nothing in either optree says that
  the first reaches the caller's variable and the second does not. That
  is established earlier, when `entersub` filled `@_` with the caller's
  SVs rather than with copies, and it is a property of the values rather
  than of the code operating on them. `07_args_alias.t` is the file that
  measures it, and it can only do so by RUNNING: its pinned `2 1` is the
  whole of the evidence, which is why that file's own comment says the
  ops are a distraction there.

Not claimed here, though they appear in the measurements:

- **`shift`** is not this tier's. Bare `shift` inside a sub does default
  to `@_`, and the optimiser folds that default into the op's flags --
  `shift s*` with no `gv[*_]` and no `rv2av` in front of it -- but `shift
  @a` outside any sub emits the same op, so the op is not this tier's to
  introduce. No file HERE emits it; `11_oo/06_indirect_new.t` is the
  first that does, so tier 11 carries the claim, with the same note about
  it being a gap rather than a judgement.
- **`aelemfast`, `multideref`, `rv2av`, `gv`, `aassign`, `padav`** are all
  how `@_` and `$_[0]` are read, and all of them are tier 02's array and
  package-variable surface.
- **`postinc`** is claimed here, and it should not have to be. `$_[0]++`
  emits it, and the increment operator is tier 04's subject, not this
  one's. But no tier-04 file emits `postinc` -- tier 04's files compare
  and arithmetic, and none of them increments -- so claiming it there
  fails `TestCorpusLints`, which reports a README claiming an op no file
  emits. The corpus's rule is that the tier which EMITS an op first owns
  it, and `07_args_alias.t` is the only file in the corpus that compiles
  one. A tier-04 file exercising `$x++` would move the claim to where it
  belongs. What IS this tier's is the aliasing that makes the increment
  visible to the caller, and that is not an op at all.
- **`wantarray`** belongs to 03_context, which is what it asks about.
- **`rv2cv` and `srefgen`**, which `\&f` emits, belong to 08_references.
  This tier stops at declaring and calling; taking a reference to the
  result is the next tier's subject.

### `lock`, and why `tie` and `tied` are NOT here

`lock` is this tier's by placement rather than by argument, and `tie` and
`tied` were measured OUT of it. Both facts are recorded because the
second was a surprise that cost a draft.

**`tie` CANNOT be written at this tier, and the reason is structural.**
The proposal was that `TIESCALAR` and `FETCH` are ordinary named subs, so
`tie` is subroutine dispatch arriving through a construct that never
spells a call -- which is true, and is not sufficient. `tie` requires the
constructor to return a BLESSED object. Measured 5.42.0, a `TIESCALAR`
returning a plain string or an unblessed reference makes the tied
variable read as empty and the program prints nothing at all, silently.
And the minimal blessed constructor is already out of budget:

    $ perl -MO=Concise,-exec,C::TIESCALAR -e 'package C;
        sub TIESCALAR { return bless {}, "C" } sub FETCH { return 42 }
        package main; tie my $c,"C"; print $c,"\n";'
    C::TIESCALAR:
    1  <;> nextstate(C 2 -e:1) v
    2  <0> emptyavhv[t1] s/ANONHASH
    3  <$> const[PV "C"] s
    4  <@> bless sK/2
    5  <1> leavesub[1 ref] K/REFC,1

`bless` and `emptyavhv` are both 11_oo's. The lint reads inside a file's
own CVs -- the capability the last section of this README records -- so
those ops count against the file even though the main optree shows only
`tie`. There is no spelling of a working `tie` that avoids them, because
blessing is what the protocol requires. So `tie` needs 11_oo, not 07, and
the argument that its methods are "just subs" is an argument about what
they are DECLARED as rather than about what the construct needs.

`tied` inherits the same floor: observing it requires a tied variable to
observe, so any file exercising it carries a `TIESCALAR` and its `bless`.
Its own op is free of that, and the boolean form does avoid 08's `ref`
which an earlier draft assumed it needed -- but neither saves the file.

**`lock` is here for proximity and nothing else**, and the file says so
in its own header. It declares no sub, calls no sub, and has no argument
protocol; its ops outside `lock` itself are all tier 01's, so nothing
about the op budget places it. With a free choice it would sit among the
named unary operators in 04_operators. It arrived in the same pass as the
`tie` proposal and stayed when that was withdrawn, which is a worse
reason than a placement usually has, and saying so is better than
inventing a theme it fits.

One claim `13_lock.t` deliberately does NOT make: that `lock` works
without threads. The pinned interpreter is a threaded build
(`useithreads=define`, `x86_64-linux-thread-multi`), so what is
established is that `lock` compiles and runs whether or not `threads.pm`
is loaded, and nothing about an unthreaded perl. Uncontended, it has no
observable effect, which is why the file pins its operand rather than
anything about locking.

## What the lint could not see here, and what changed when it could

The lint measures a file by asking perl for its optree and reading the op
names out. It used to ask with `perl -MO=Concise,-exec`, which prints the
MAIN program's optree and nothing else. A subroutine body is a separate
CV with its own optree, printed only when the sub is named on the command
line -- `-MO=Concise,-exec,f` -- and the lint never named one, because it
had no way to know what subs a file declares without parsing it, which is
the thing the corpus exists to avoid depending on.

The consequence was specific, and it was this tier's central finding
while it held: **the ops most characteristic of subroutines were
structurally unreachable by this lint.** `leavesub` ends every sub body
and appears in no main program. `argcheck`, `argelem` and `argdefelem`
are the entire compiled form of a signature and appear in no main
program. `return`, where it survives at all, appears in no main program.
All five are emitted by the files in this directory under 5.42.0, and
none could be claimed, because a claimed op that never appears is exactly
as bad as an unclaimed op that does -- `TestCorpusLints` would pass a
tier whose INTRODUCES list was pure fiction.

`00_adjacency.t` was the sharpest instance. Its source carries a
signature, a parameter default, a `for` loop, a `next`, three `return`s
and an ampersand call; its MAIN optree contains `anoncode` and `entersub`
and not one other op from tier 06 or 07. Six constructs, two visible ops.
Measured with the sub bodies included, the same file adds seventeen:
`aelemfast and argcheck argdefelem argelem enteriter eq iter join
leaveloop leavesub lt multiconcat next postinc return rv2av unstack`.
That difference is the size of the blind spot, in one file.

**The lint no longer has that blind spot.** `opsOf` enumerates the CVs
the file itself declared -- walking the symbol table and keeping only
those whose `FILE` is the file under test -- and dumps each beside the
main program. Issue 01a0c547-516b is where the change was measured, and
the measurement is the reason it was made rather than assumed: across all
fourteen tiers, twenty-one files' op streams grew, and not one of them
started emitting an op a LATER tier claims. The widening exposed ops that
had no owner; it did not disturb the ordering.

What remains out of reach, and is worth recording in its place:

- **An anonymous sub's body.** It has no name to pass to B::Concise and
  no glob to walk to, so `sub { ... }` still contributes `anoncode` where
  the closure is built and nothing from inside. `02_anon_sub.t`'s op
  stream is byte-identical before and after the widening -- measured.
  An `ADJUST` block is the same shape: a nameless CV in the class stash.
- **The enumeration stops at the file.** Every CV the process has loaded
  is reachable from the symbol table, `Exporter::import` included, and
  dumping those would add their ops to every corpus file's measured set.
  The `FILE` filter is what prevents that, and
  `TestOpsOfStopsAtTheFileBoundary` is what keeps the filter honest.
- **The behaviour assertions still carry the tier.** `--- expect output`
  runs the file through real perl, bodies and all, and compares bytes.
  `05_return.t` prints `zero negative positive` only if the early return
  actually fired; `06_signature.t` prints `4 11` only if the default was
  actually applied. Ops now describe those constructs, but only running
  them proves they work.

## HARD MARKERS

    signature	use·v5.3

One of the twelve `hardMarkers` places here. The list came from
`internal/parse/easy_test.go`, which used it to carve an "easy" subset out
of T1; in a graded corpus each entry is a tier placement rather than a
filter, so that file was deleted and the placements moved to the tiers.

`signature` was measured at 20.0% clean over five files, and its probe is
not the signature syntax but the `use v5.3x` line that enables it -- which
is the honest thing to detect, since a signature is only a signature once
the feature is on. It places here because a signature is a parameter list,
and this tier is where a sub first has parameters at all. `06_signature.t`
is the file; the note above about what the lint cannot see through a CV
applies to it, which is why the tier's coverage of signatures is a corpus
file rather than an op.

The probe column is the SOURCE probe verbatim, `·` standing for a space.
`TestEveryHardMarkerPlaced` in `internal/conformance` requires a file here
that the probe finds.

## FILE ORDER

    accidental

The numbers are the order the files happened to be written in. This README
cites its files by name and never by position, and no claim in it would
become false if the files were renumbered.

`accidental` is a record of debt, not a convention. Renumbering this tier
into `derived` order costs nothing on disk but regenerates the ratchet,
which is a separate change; the declaration exists so the state is written
down rather than rediscovered by the next agent whose numbering test fails.
