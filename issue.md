---
title: Decide what a runtime-required test.pl means for this goal's ceiling
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T16:29:00.320261836Z
updated: 2026-09-26T16:36:29.160218782Z
---

**MEASURED AND DECIDED 2026-09-26 at `2dc301fd`. Decision: read the required
file. The question's framing was wrong in a way that changes the answer.**

Perl's own suite gets `ok`, `is`, `like` and `plan` from `t/test.pl`, which it
`require`s at RUNTIME. `internal/parse/use.go:118` leaves `require` unresolved
on purpose.

## The size of the question: it is the goal, not a footnote

    files requiring a .pl helper:   dirty=394   clean=108
    files NOT requiring one:        dirty= 41   clean= 77
    Unknown nodes in require-ers:   7466
    Unknown nodes in the rest:       279

    CEILING WITHOUT RESOLVING require:  226 of 620

394 of 435 dirty files require a helper, and they hold **7,466 of 7,745 Unknown
nodes -- 96.4%**. The other 41 dirty files hold 279 between them.

So "parse all of perl.git t/ cleanly" is unreachable without an answer here. Not
"harder"; unreachable. 226 of 620 is the arithmetic ceiling.

Spellings, counted:

    372  require './test.pl'
     78  require "./test.pl"
     22  require './charset_tools.pl'
     17  require './t/test.pl'
     15  require './loc_tools.pl'
     14  require "unicore/UCD.pl"
     11  require './thread_it.pl'
      5  require 'test.pl'

All literal relative paths. None computed.

## The framing was wrong: perl DOES build a tree, and it is the wrong one

The issue said perl knows `ok` is declared because the `require` has already run.
Measured, that is false -- and the truth is more useful. Perl parses the
undeclared form as INDIRECT OBJECT NOTATION:

    perl -MO=Deparse -e 'like $@, qr/x/, "msg";'
      ->  $@->like, qr/x/, '???';

    perl -MO=Deparse -e 'is $a, 123;'
      ->  $a->is, '???';

    perl -MO=Deparse -e 'sub like {1} like $@, qr/x/, "msg";'
      ->  like $@, qr/x/, 'msg';          the DECLARED form, a real call

`perl -c base/lex.t` reports **syntax OK** on a file whose `ok` is undeclared at
compile time, and the failure is at RUNTIME:

    Can't call method "like" without a package or object reference

So perl compiles a method call on `$@`, which is not what the program means, and
only discovers the problem when it runs. Our Unknown is arguably the more honest
answer -- but it is not the answer this goal needs, and perl's tree is what the
oracle compares against.

This is the same defect as `01a0ddc5` (`new Foo;` builds `new(Foo())` where perl
builds `'Foo'->new`), from the other side: there we build a wrong tree where
perl builds a method call; here we build none where perl builds a wrong one.

## Why route A rather than B or C

**A. Read the required file.** Chosen.

- The argument is a literal relative path in every one of the 474 occurrences.
  No computed case to guard against.
- `test.pl` declares **78 distinct subs** including all of `ok`, `is`, `like`,
  `plan`, `isnt`, `done_testing`, `skip`, `pass`, `fail`, and parses with only
  **10 Unknown nodes** in 64,807 bytes. It is readable today.
- The machinery exists. `01a0c13f-50ee` taught the parser that a `sub NAME`
  earlier in the SAME FILE is a declaration the call site can see, and measured
  316 of T2's 354 parenless-call sites as that case. Pointing the same
  resolution at a required file's declarations is the same mechanism with a
  different source.
- Call-site volume, counted across `t/`: `is` 3569, `ok` 759, `like` 713,
  `skip` 316, `plan` 175, `cmp_ok` 83, `pass` 64, `isnt` 44.

**B. Declare the ceiling at 226 of 620.** Rejected. It redefines the goal to
exclude 76% of its own population, and the excluded part is the part perl's
suite is actually made of.

**C. Leave it open.** Rejected. The goal is then permanently unreachable and
nothing says why.

## The boundary that keeps A from becoming "execute the program"

Stated so the implementing issue inherits it:

1. A LITERAL string argument only. `require $file` and `require "./$name.pl"`
   stay unresolved. Measured: zero of the 474 occurrences are computed, so this
   costs nothing today and is the whole safety property.
2. Resolve relative to the FILE BEING PARSED, then to the parse root. `t/` files
   say `./test.pl` and run with `t/` as cwd; `require './t/test.pl'` (17 files)
   is the same file reached from the repository root.
3. Take only `sub NAME` declarations. Not variables, not imports, not side
   effects.
4. One level. A required file that itself requires another is not followed --
   measure whether that costs anything before deciding otherwise.
5. Unreadable or unparseable helper: the call sites stay Unknown, as now. A
   missing file must not fail the parse.

## Filed

`01a0de92-e551-7ece-be3c-0f8afc1d1e6b` implements route A with the five
boundaries above.

## What this issue does NOT decide

Whether a genuinely undeclared parenless call should build perl's indirect-object
tree. That is `01a0ddc5`'s question and it is in v0.1. This issue is only about
the callees that ARE declared, in a file we can read.
