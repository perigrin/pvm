---
title: A literal require of a .pl helper contributes its sub NAME declarations
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T16:36:04.048888797Z
updated: 2026-09-26T16:36:04.048888797Z
---

A `require` of a LITERAL relative `.pl` path contributes its `sub NAME`
declarations to the requiring file's scope. This is the decided route A from the
ceiling issue, and it is the single largest item in the perl.git t/ goal.

## Why: 96% of the goal's Unknown nodes are behind it

Measured 2026-09-26 at `2dc301fd` over `/home/perigrin/dev/perl5/t`, 620 files:

    files requiring a .pl helper:   dirty=394   clean=108
    files NOT requiring one:        dirty= 41   clean= 77
    Unknown nodes in require-ers:   7466
    Unknown nodes in the rest:       279

    CEILING WITHOUT THIS:  226 of 620

394 of 435 dirty files hold 7,466 of 7,745 Unknown nodes. The remaining 41 dirty
files hold 279 between them. Without this the goal is arithmetically unreachable.

## Perl builds a tree here, and it is the WRONG one

Not a syntax error, which is what makes our Unknown defensible but insufficient:

    perl -MO=Deparse -e 'like $@, qr/x/, "msg";'
      ->  $@->like, qr/x/, '???';           INDIRECT OBJECT, compiles

    perl -MO=Deparse -e 'sub like {1} like $@, qr/x/, "msg";'
      ->  like $@, qr/x/, 'msg';            a real call

    perl -c t/base/lex.t                    syntax OK

The failure is at RUNTIME -- "Can't call method \"like\" without a package or
object reference". So perl compiles a method call on `$@` and discovers the
problem only when it runs. Our refusal is arguably more honest, and it is still
not the tree the oracle compares against.

## The helper is readable today

`t/test.pl` declares **78 distinct subs** -- `ok`, `is`, `like`, `plan`, `isnt`,
`done_testing`, `skip`, `pass`, `fail` among them -- and parses with **10 Unknown
nodes in 64,807 bytes**.

Call-site volume across `t/`, counted: `is` 3569, `ok` 759, `like` 713,
`skip` 316, `plan` 175, `cmp_ok` 83, `pass` 64, `isnt` 44.

## The machinery exists

`01a0c13f-50ee` taught the parser that a `sub NAME` earlier in the SAME FILE is a
declaration the call site can see, and measured 316 of T2's 354 parenless-call
sites as that case. `knowsShape` at `internal/parse/call.go` and the recording
site in `parseRoot` are what this extends. This is the same resolution with a
different SOURCE of declarations, not a new mechanism.

`01a0ad3f` built import resolution for `use Module`, and it measured worth
EXACTLY ZERO on T2 -- zero of the 56 files import `Test::More`. That is why this
is about a required FILE and not about modules.

## THE BOUNDARY, and it is the whole safety property

1. A LITERAL string argument only. `require $file` and
   `require "./$name.pl"` stay unresolved. Measured: zero of the 474 occurrences
   in `t/` are computed, so this costs nothing today.
2. Resolve relative to the FILE BEING PARSED, then to the parse root. `t/` files
   say `./test.pl` and run with `t/` as cwd; `require './t/test.pl'` (17 files) is
   the same file reached from the repository root.
3. Take only `sub NAME` declarations. Not variables, not imports, not side
   effects, not `BEGIN` blocks.
4. ONE LEVEL. A required file that itself requires another is not followed.
   Measure whether that costs anything before deciding otherwise.
5. An unreadable or unparseable helper leaves the call sites Unknown, as now. A
   missing file must NOT fail the parse -- `t/` files are parsed from many cwds.

Spellings to handle, counted:

    372  require './test.pl'
     78  require "./test.pl"
     22  require './charset_tools.pl'
     17  require './t/test.pl'
     15  require './loc_tools.pl'
     14  require "unicore/UCD.pl"
     11  require './thread_it.pl'
      5  require 'test.pl'

## Measure before and after, per file

The recurring defect in this repo is to read a count off failing files, name a
cause, and never measure the cause separately -- it has overturned five briefs
this milestone. So:

- Before: 185 of 620 clean, 7,745 nodes.
- After: report BOTH files-clean and nodes, and verify every mover actually
  contains a `require` of a literal `.pl`.
- A file that moves WITHOUT one is unattributed; stop and report rather than
  regenerating over it.

## Ratchets, all failing in both directions

- `internal/conformance/testdata/corpus.ratchet` -- 216 of 216, 0 refusing. The
  corpus is COMPLETE; do not regress it.
- `internal/parse/testdata/t1.ratchet` -- 986 files, 554 clean, 1960 nodes.
  T1 is PerlOnJava, NOT perl.git; it has its own helpers and may not move at all.
- `internal/parse/testdata/t1canon.ratchet` -- 721 clean, 265 disagreeing.
- `internal/parse/testdata/t1fidelity.ratchet` -- 369 clean, 617 differing.
- the `shortfall` map in `internal/parse/t2_test.go` -- T2 core IS 56 of this
  goal's 620 files, so it should move a lot.
- `TestClassCorpusRatchet` in `internal/parse/use_test.go`, pin 21 -- not in the
  list above and has bitten three workers.

Canon is the live hazard: four workers this milestone broke
`TestCanonTokenIdentity` or the canon fixpoint. Run
`go test ./internal/parse/ -run '^TestCanon' -count=1` EARLY.

## Context

- paths: `internal/parse/use.go:118` (`require` unresolved), `internal/parse/call.go`
  (`knowsShape`), `internal/parse/parse.go` (`parseRoot`'s declaration recording)
- commands: `go test ./internal/parse/ -count=1`, `go test ./internal/conformance/ -count=1`
- corpus: `/home/perigrin/dev/perl5/t`, 620 `.t` files; `t/test.pl` is the helper
- related: `01a0c13f-50ee` (in-file declarations), `01a0ad3f` (import resolution,
  worth zero on T2), `01a0ddc5` (`new Foo;` builds the wrong tree -- the same
  indirect-object question from the other side)

## Acceptance criteria

- [ ] A literal `require './helper.pl'` brings its `sub NAME` set into scope (`go test ./internal/parse/ -run '^TestRequiredFileDeclarations$' -count=1 -v | grep -q '^--- PASS: TestRequiredFileDeclarations'`)
- [ ] A COMPUTED require stays unresolved (`go test ./internal/parse/ -run '^TestRequiredFileDeclarations$' -count=1 | grep -q '^ok'`)
- [ ] A missing helper does not fail the parse (`go test ./internal/parse/ -run '^TestRequiredFileDeclarations$' -count=1 | grep -q '^ok'`)
- [ ] Canon holds its fixpoint (`go test ./internal/parse/ -run '^TestCanon' -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the corpus stays 216 of 216 (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the parse ratchets hold (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
