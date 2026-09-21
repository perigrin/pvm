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

    anoncode entersub

Two ops, for a tier whose subject is subroutines. That is not a mistake
and it is not modesty: it is the whole measurable content of this tier
under the lint that reads it. The five ops the first draft of this list
claimed -- `argcheck`, `argdefelem`, `argelem`, `leavesub`, `return` --
are all real, all emitted by the files in this directory, and all
invisible to the lint. Why, and what that costs, is the last section of
this file.

## Why those ops, and not the ones the source implies

This list is what `perl -MO=Concise,-exec` EMITS for the files in this
tier, measured under 5.42.0, not what reading them suggests. Five places
where those differ:

- **The sub body is not in the main program's op stream.** `perl
  -MO=Concise,-exec -e 'sub f { ... } f()'` prints `enter`, `pushmark`,
  `gv[IV \&main::f]`, `entersub`, `leave` and stops -- the body is absent.
  Naming the sub gets it: `-MO=Concise,-exec,f`. This is the fact that
  determines the INTRODUCES list above, because the lint runs the first
  invocation and never the second: everything below that lives in a body
  is described here for the reader and claimed by no one.

- **`return` is mostly erased.** `sub f { return $_[0] + 1 }` emits
  `add` then `leavesub` and no `return` op; so does `sub f { return
  (1,2) }`, which additionally loses its `pushmark`. The op appears only
  where the return is not the last thing the body does -- `return 1 if
  $_[0]` compiles `and` guarding a `pushmark`/`return` pair. So the
  source construct `return` and the op `return` are different things.
  `05_return.t` branches and so does emit the op -- into its own CV,
  where the lint does not look, which is why the tier does not claim it.

- **A signature is NOT erased.** This is worth stating because it is the
  opposite of what the rest of this file warns about. `sub f ($a, $b)`
  compiles to `argcheck(2,0)` followed by one `argelem` per parameter --
  distinct ops carrying the arity, not ordinary pad assignments that
  happen to read `@_`. A default adds `argdefelem` guarding the default
  expression, and a slurpy `@rest` shows in the `argcheck` flags as
  `argcheck(1,0,@)`. Compare `sub f { my ($a,$b) = @_ }`, which is
  `padrange` and `aassign` and nothing else: the two spellings are
  genuinely different optrees, so here the ops DO describe the construct
  -- and are still unclaimable, because they are inside the sub. This is
  the most painful case: the one construct in the tier whose ops would
  have made a precise claim is the one the lint is least able to reach.
  `use v5.36` itself adds no ops -- measured, it changes the `nextstate`
  hint flags from `v:{` to `v:us,*,&,{,$,fea=6` and nothing more, so the
  pragma `06_signature.t` needs costs the op stream nothing.

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

Not claimed here, though they appear in the measurements:

- **`shift`** is tier 02's. Bare `shift` inside a sub does default to
  `@_`, and the optimiser folds that default into the op's flags -- `shift
  s*` with no `gv[*_]` and no `rv2av` in front of it -- but `shift @a`
  outside any sub emits the same op, so the op is not this tier's to
  introduce.
- **`aelemfast`, `multideref`, `rv2av`, `gv`, `aassign`, `padav`** are all
  how `@_` and `$_[0]` are read, and all of them are tier 02's array and
  package-variable surface.
- **`wantarray`** belongs to 03_context, which is what it asks about.
- **`rv2cv` and `srefgen`**, which `\&f` emits, belong to 08_references.
  This tier stops at declaring and calling; taking a reference to the
  result is the next tier's subject.

## What the lint cannot see here, and why that is worth recording

The lint measures a file by running `perl -MO=Concise,-exec` on it and
reading the op names out of the output. That invocation prints the MAIN
program's optree and nothing else. A subroutine body is a separate CV
with its own optree, printed only when the sub is named on the command
line -- `-MO=Concise,-exec,f` -- and the lint never names one, because it
has no way to know what subs a file declares without parsing it, which is
the thing the corpus exists to avoid depending on.

The consequence is specific and it is this tier's central finding: **the
ops most characteristic of subroutines are structurally unreachable by
this lint.** `leavesub` ends every sub body and appears in no main
program. `argcheck`, `argelem` and `argdefelem` are the entire compiled
form of a signature and appear in no main program. `return`, where it
survives at all, appears in no main program. All five are emitted by the
files in this directory, under 5.42.0, and none can be claimed, because a
claimed op that never appears is exactly as bad as an unclaimed op that
does -- `TestCorpusLints` would pass a tier whose INTRODUCES list was
pure fiction, and the next author would inherit a list they could not
reproduce.

`00_adjacency.t` is the sharpest instance. Its source carries a
signature, a parameter default, a `for` loop, a `next`, three `return`s
and an ampersand call; its main optree contains `anoncode` and
`entersub` and not one other op from tier 06 or 07. Six constructs, two
visible ops.

So the boundary is worth stating plainly, because it bounds what check 1
proves for every tier from here on:

- **Ops prove things about the main program only.** For tiers 01 through
  06 that happened to be the whole program, so the limit never showed.
  This is the first tier where the construct under test lives somewhere
  the measurement does not go, and 11_oo, whose methods are subs, will
  inherit the same blind spot in full.
- **A tier whose subject is a body gets a weak lint, not a wrong one.**
  `anoncode` and `entersub` are correctly this tier's, and the check that
  no file here uses a LATER tier's op still runs over the main program
  and still has force -- it is why `08_references`' `srefgen` stays out
  of these files. The lint did not become unsound; it became less
  sensitive, and only for the ops it cannot see.
- **The behaviour assertions carry the tier instead.** `--- expect
  output` runs the file through real perl, bodies and all, and compares
  bytes. `05_return.t` prints `zero negative positive` only if the early
  return actually fired; `06_signature.t` prints `4 11` only if the
  default was actually applied. That is what measures the ops the lint
  cannot, and it is why this tier's files lean on output rather than on
  their op claims.

Extending the lint to descend into sub bodies is possible -- read the
declared names out of `-MO=Concise` output and re-run per name, or use
`-MO=Concise,-main,-exec` plus each CV -- and it would make this tier's
five erased ops claimable. It is not done here, and the reason is that
it changes the lint for all fourteen tiers, which wants measuring before
it is assumed to be an improvement. Recording the limit is the smaller
and honest first step; this section is that record.
