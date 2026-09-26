---
title: Decide what a runtime-required test.pl means for this goal's ceiling
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T16:29:00.320261836Z
updated: 2026-09-26T16:29:00.320261836Z
---

Perl's own suite gets `ok`, `is`, `like` and `plan` from `t/test.pl`, which it
`require`s at RUNTIME. `require` is deliberately unresolved, so every
`like $@, qr/.../` span in this goal's 7,745 Unknown nodes is downstream of that
choice. DECISION ISSUE -- the deliverable is a written answer, and it may be
"no change".

## Why this blocks the goal's ceiling

Measured 2026-09-26, recurring Unknown spans in perl.git t/:

      38  "is $s, join('', map chr,"
      35  "like $@, _create_mismatc"
      24  "like $@, qr/^Type of arg"
      23  "like shift @warnings, qr"
      18  "like $@, qr/^Assigned va"
      17  "is $a, 123;"
      14  "like $@, qr/Modification"
      14  "like $@, qr/^Too many ar"

Every one is a PARENLESS CALL to a test function. `01a0c13f-50ee` closed the
declared case: a `sub NAME` earlier in the same file resolves, and 316 of T2's
354 parenless-call sites were that. These are the other kind -- the callee is
defined in `test.pl` and reached by:

    require './test.pl';

`internal/parse/use.go:118` leaves `require` unresolved on purpose. That is
correct for a parser: `require` is a runtime operation and its argument can be
computed.

## What was already established, so it is not re-litigated

`01a0c13f-50ee` measured that ZERO of the 56 T2 files import `Test::More` -- the
four that mention it do so in comments explaining why they avoid it. Perl's core
suite deliberately does not depend on the module ecosystem. So "resolve the
module" is not the shape of an answer here; the callees live in a FILE, required
by relative path, at runtime.

And perl itself refuses an undeclared parenless call:

    perl -MO=Deparse -e 'ok 8;'    ->  syntax error near "ok 8"

which is why declining these is currently CORRECT rather than a gap. A tree for
`like $@, qr/x/` where perl builds none would be WRONG in the oracle's sense.

## The actual question

Perl compiles these files successfully because `require './test.pl'` has ALREADY
RUN by the time the calls are compiled -- it is at the top of the file, and
`require` executes at runtime in file order, before the later statements are
compiled. So perl knows `ok` is declared and we do not.

Three routes, and this issue is to choose one and write it down:

**A. Read the required file.** Follow `require './test.pl'` with a literal
relative path, parse it, and take its `sub NAME` declarations. Honest, matches
what perl actually does, and bounded -- a literal string argument only, never a
computed one.

**B. Declare this goal's ceiling.** State that files whose test functions come
from a runtime `require` cannot reach zero Unknowns, count them, and make that
the goal's denominator. Cheap and honest, and it means "parse all of t/ cleanly"
becomes "parse all of t/ that does not require at runtime".

**C. Neither, and record why.** If A is a slippery slope toward executing the
program and B redefines the goal, the answer may be that this stays open and the
goal is understood as asymptotic.

MEASURE THE SIZE OF THE QUESTION FIRST: how many of the 435 dirty files contain
`require './test.pl'` or a sibling, and how many of those would go clean if its
`sub NAME` set were known? That number decides whether this is the goal's main
obstacle or a footnote.

## Scope

1. Count the dirty files that `require` a `.pl` helper by literal relative path.
2. For a sample, measure how many Unknowns would clear if that file's `sub NAME`
   declarations were in scope. Use the existing in-file resolution machinery --
   `knowsShape` at `internal/parse/call.go` and whatever `01a0c13f-50ee` added --
   pointed at the helper's declarations.
3. Choose A, B or C, and write the decision into
   `docs/plans/2026-09-26-perl-git-t-parses.md` with the measurement behind it.
4. If A: file it as its own issue with the literal-path-only boundary stated.
   Do not implement it here.

## Context

- paths: `internal/parse/use.go:118` (`require` left unresolved),
  `internal/parse/call.go` (`knowsShape`), `docs/plans/2026-09-26-perl-git-t-parses.md`
- commands: `go test ./internal/parse/ -count=1`
- files: `/home/perigrin/dev/perl5/t/test.pl` is the helper; `t/base/*.t` and
  `t/comp/*.t` are dense users of it
- related: `01a0c13f-50ee` (declared/undeclared boundary), `01a0ad3f` (import
  resolution, measured worth zero on T2)

## Acceptance criteria

- [ ] The count of dirty files depending on a runtime-required helper is recorded (`grep -qiE 'require .*test\.pl.*[0-9]+ (of|files)|[0-9]+ dirty files require' docs/plans/2026-09-26-perl-git-t-parses.md`)
- [ ] A decision of A, B or C is written down with its reasoning (`grep -qiE 'decision: (read the required file|declare the ceiling|no change)' docs/plans/2026-09-26-perl-git-t-parses.md`)
- [ ] REGRESSION GUARD, passes today: the parse suite holds (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the corpus ratchet holds (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
