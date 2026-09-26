---
title: Fourteen of seventeen refusals point at nothing actionable
state: pending
urgency: normal
milestone: corpus-covers-t2
created: 2026-09-26T02:51:11.245880623Z
updated: 2026-09-26T02:53:03.560250648Z
---

**CANCELLED BY CHAIN-REVIEW, 2026-09-26. Its count was wrong and what
remains is already owned by `01a0d0c7`.**

## The count

This issue said fourteen of seventeen refusals point at nothing: ten
`unfiled` plus four citing closed issues.

**The ten are not a defect.** `unfiled` is a deliberate sentinel -- the
topic-format spelling of `selfRecorded` ("this file") -- meaning the case
IS its own record. `internal/conformance/file.go:275` documents it and
`TestRefusalCitationMustResolve/unfiled_is_not_looked_up` asserts it,
passing. Its comment says so directly: treating the sixteen as citations
pointing at nothing is "a claim about the corpus that is not true".

So the real number is FOUR, not fourteen:

    argument-extent.md           2   01a0c432-fbd5   done
    adjacency-07_subroutines.md  1   01a0c432-fbd5   done
    adjacency-11_oo.md           1   01a0c35f-a487   done
    named-operators.md           1   01a0c730-b241   done

(Five cases, four citation sites, three distinct closed ids.)

## Why `01a0d0c7` already covers it

`01a0d0c7` in `v0.1` -- "Five refusal citations point at closed issues,
and nothing checks" -- names exactly these, with the same three ids, and
owns the check (`TestRefusalCitationsAreLive`). Its count of five is
cases; mine of four was sites. Same defect.

It also has the better diagnosis, which this issue would have lost: three
of the five closed because THE MILESTONE they belonged to finished, not
because the parser learned anything. So a citation needs re-measuring
rather than re-pointing.

## What this issue found that 01a0d0c7 should carry forward

One thing, and it is why the chain-review ran at all: `argument-extent.md`
holds five tier-07 cases on parenless argument extent, and because both
refusing cases cite a `done` issue, a reader concluded the corpus had no
such case and filed an issue to write coverage that already existed
(`01a0db76`, now cancelled).

That is the cost of a closed citation stated concretely, and it belongs in
`01a0d0c7` as the reason the check matters.

## Also worth carrying: the existing check verifies the wrong property

`TestRefusalCitationMustResolve/corpus` passes today while four citations
name `done` issues -- measured. It resolves that an id EXISTS, not that it
is OPEN. That is the precise gap `TestRefusalCitationsAreLive` fills, and
naming the difference stops someone concluding the check already exists.
