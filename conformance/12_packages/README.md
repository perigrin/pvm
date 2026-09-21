# 12_packages

`package` in both spellings, `use`, `require`, and `import`.

## Why this tier sits here

`use Foo LIST` locates a file, loads it, and then calls `Foo->import(LIST)`.
The last of those is a method call on a package name, which needs tier 07
for a sub to call and tier 11's dispatch to find it. `require` is the same
construct with the import step removed. So the tier cannot come before the
call machinery exists, and it is that -- not `package` -- that fixes its
position.

Check 2 asks whether this tier could move earlier. It could, and quite far:
the ops it introduces are one, `require`, and the only thing `require`
needs is a string. What it could not do is move before `07_subroutines`,
because a tier that covers `use` and stops short of the `import` call is
covering the easy half of the construct.

This tier also could not come EARLIER than `11_oo` in the way an earlier
draft had it. That draft ordered `11_packages` then `12_oo`, on the theory
that an object system is built out of packages. Perl refutes it:
`class Foo {}` IS a package declaration -- `__PACKAGE__` inside such a body
is `Foo` -- so ordering packages first would have made tier 11 depend on a
construct it also supplies. The dependency was circular, and the swap
removed it. `use` and `require` are the heavier construct anyway, since
they reach outside the file.

## DEPENDS ON

    07_subroutines

`import` is a subroutine call, and it is the whole reason `use` is not
just `require`. The adjacency file pairs with that -- a package defining
an `import`, called across a package boundary -- rather than with tier 11,
which this tier needs nothing from. The spec says so plainly, and pairing
with N-1 here would assert a dependency that does not exist.

`package` itself depends on nothing and could sit in tier 01. It is here
because splitting a tier so that `package` precedes `use` by ten rungs
would gain nothing: neither half is a prerequisite of the other.

## INTRODUCES

    require

## Why those ops, and not the ones the source implies

One op. That is the tier's finding, and it is not a shortfall in the
files -- it is what perl emits.

- **`package` emits nothing.** `package Foo;` moves the compiler's notion
  of the current package and contributes no runtime op at all. The only
  trace is inside the `nextstate` that follows, which names `Foo` instead
  of `main` -- a field of an op tier 01 already claims, not an op of its
  own. The block form `package Foo { ... }` emits `enterloop`, `stub` and
  `leaveloop`: a bare block, which is tier 05's construct, plus tier 11's
  placeholder. Nothing in that stream says `package`.

- **`use` emits nothing.** `use strict;` is a `BEGIN` block that has
  already run by the time there is an optree to dump. `use POSIX ();`
  likewise -- the file is located, read, compiled and installed in `%INC`,
  and the whole of that work is finished before the first runtime op. The
  op stream for `use strict; use warnings; print "$x\n";` is identical to
  the stream for `print "$x\n";` alone, but for the feature bits on
  `nextstate`. A tier that tried to DERIVE itself from ops would conclude
  that `use` is not in the language.

- **`require` is the one that survives**, because it is the one that
  happens at run time. `require strict;` emits `const[PV "strict.pm"]
  s/BARE` and then `require`. The bareword-to-filename rewrite is done by
  the compiler, so the op sees a string either way: `require $m` emits
  `padsv` and the same `require`, and `require 5.010` emits `const[NV
  5.01]` and the same `require` again. One op, three sources.

- **`import` is an ordinary method call.** `Foo->import("tag")` emits
  `pushmark`, two `const`, `method_named` and `entersub` -- tiers 01, 07
  and 11 between them, with nothing left over. There is no import op. The
  reason `use` is heavier than `require` is a call the op stream cannot
  distinguish from any other call.

## What a file can assert when its construct emits no op

A tier measured through ops has an op check. This one does not, and
"perl printed what we pinned" is not a substitute for it: a file can pin
output that is unchanged by deleting the very statement the file is
about. `03_use_pragma.t` is that file. It pins `ok`, and a copy with
both `use` lines deleted prints `ok` too -- measured. It establishes
that the lines lex, and nothing further.

The replacement is a DELETION TEST applied while authoring: a file
earns its pin only if removing its construct changes the output. What
satisfies it differs by construct.

- **`package` already satisfies it, by the fully qualified call.**
  `01_package_statement.t` defines `hello` under `package Greet;` and
  calls `Greet::hello()`. Delete the `package` line and the sub lands in
  `main::`, the call is to an undefined subroutine, and perl dies.
  `02_package_block.t` is the same measurement with a scope on it. So
  the construct that emits no op is pinned by a name that only resolves
  because it took effect.

- **`use` needs an EXPORTER, which is why `04_use_import.t` exists.** A
  pragma cannot be caught this way: its effect is lexical and
  compile-time, and `$^H` and `${^WARNING_BITS}` read the caller's scope
  rather than the file's when consulted at run time -- measured, both
  report the pragma OFF inside the file that turned it on. A module with
  an exporter can, through the symbol table: `use POSIX;` puts `floor`
  into `main::` and the empty-list form does not.

- **The pair is the assertion, not either file alone.**
  `04_use_import.t` and `05_use_empty_list.t` make the same two
  observations one statement apart:

        use POSIX;      loaded yes    imported yes
        use POSIX ();   loaded yes    imported no

  The first column is `require`'s half of `use`, the second is
  `import`'s. Their op streams are byte-identical to each other and to
  the two `print` statements alone.

`03_use_pragma.t` is kept rather than replaced. What it measures is
narrow but real -- `use` with a pragma is the overwhelmingly common
spelling in the surveyed corpus, and a lexer that mis-delimited it would
fail here -- and the file now says plainly which half of the construct
it does not reach.

The ops LINT this declared tier; they cannot derive it. This tier is the
sharpest case of that in the corpus after `class`: four keywords, and
between them one op, which belongs to the one keyword a reader would call
the least interesting of the four. `package` and `use` are compile-time
constructs, and a measurement of runtime ops is measuring the wrong clock
for them. Saying so is the finding; padding the list would hide it.

## FILE ORDER

    accidental

The numbers are the order the files happened to be written in. This README
cites its files by name and never by position, and no claim in it would
become false if the files were renumbered.

`accidental` is a record of debt, not a convention. Renumbering this tier
into `derived` order costs nothing on disk but regenerates the ratchet,
which is a separate change; the declaration exists so the state is written
down rather than rediscovered by the next agent whose numbering test fails.
