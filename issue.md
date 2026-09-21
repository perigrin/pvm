---
title: Twelve acceptance criteria are unverifiable because their command wraps to a second line
state: pending
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-21T23:57:09.409395856Z
updated: 2026-09-21T23:57:09.409395856Z
---

`git-zhi verify m3-conformance-corpus` now runs the milestone's criteria
(after 01a0c546 fixed the header), and reports **12 unverifiable** out of
157.

Unverifiable means verify FOUND the criterion and could not extract a
command from it. Measured: every one of the twelve is a criterion whose
text WRAPS across lines, with the paren-wrapped command on a line after
the first.

Example, from 01a0c547-516b:

    - [x] `opsOf` sees inside a named sub: a file whose signature compiles
          `argcheck`, `argdefelem` and `argelem` and whose body ends in
          `leavesub` yields all four, none of which the main optree carries.

verify's extraction is line-based, so a command on a continuation line is
invisible to it. The criterion reads well and asserts something real; it
simply cannot be executed by the tool.

## Why this matters

It is the same failure mode as the header bug, one layer in. A criterion
that verify cannot execute is a criterion nobody runs, and the milestone
already found -- by hand -- two acceptance criteria naming tests that never
existed. `unverifiable` at least REPORTS itself, unlike the header bug's
silent `total: 0`, so this is the lesser version. But 12 of 157 is 8% of
the milestone's criteria going unchecked.

## Affected

Concentrated in the four issues whose ACs were written late in the
session, when agents wrote their own criteria:

    01a0c547-516b   3 items   (the sub-body lint)
    01a0c5b9-ac65   ?         (numbering conventions)
    and others -- run `git-zhi verify m3-conformance-corpus --format json`
    and read `unverifiable_items` for the current list.

## The fix

Reflow each affected criterion so the `(command)` sits on the same line as
the `- [ ]`, however long that line becomes. The house style elsewhere in
this milestone already does this -- the 145 verifiable criteria are all
single-line.

Worth checking at the same time whether a CONVENTION should be written
down somewhere an author reads before writing ACs, since this is now the
second formatting rule verify enforces silently (the first being the word
`criteria` in the header).

Found by running verify against a clean tree after 01a0c546 landed.
