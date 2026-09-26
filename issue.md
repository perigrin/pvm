---
title: Five refusal citations point at closed issues, and nothing checks
state: pending
urgency: normal
milestone: v0.1
created: 2026-09-24T00:18:19.052459875Z
updated: 2026-09-26T03:02:15.895792547Z
---

Five corpus cases carry a `refuses:` record citing an issue that is
CLOSED. A reader following the citation lands on a finished issue that
says nothing about the refusal in front of them.

Found by `crochet:assess` on `corpus-complete`. The milestone fixed
exactly this decay for four files (`d95cf4ee`) and never checked whether
any others had it. They did.

## Measured

    $ grep -rh 'refuses:' conformance/mdtest/*.md | grep -o '01a0[a-f0-9-]*' | sort -u

Three of the six cited issues are `done`:

    01a0c35f   done   adjacency-11_oo.md
    01a0c432   done   adjacency-07_subroutines.md, argument-extent.md (x2)
    01a0c730   done   named-operators.md

The other three are live and correct: `01a0ce57` (`x=` unlexed),
`01a0cf64` (`-e` split in two), `01a0d087` (ungated `defer`).

THE CITATIONS SURVIVED THE FORMAT PORT, which is itself the finding.
The corpus moved from 212 `.t` files to 65 markdown topics and the
generator copied every refusal record verbatim -- correctly, since
losing one would turn a known gap into a silent pass. It copied the
stale ones too, because nothing reads a citation.

## Why this keeps happening

A citation is a string in a block. Nothing reads it, so nothing
notices when the issue it names closes -- and issues close routinely,
including for reasons that have nothing to do with the refusal. Three
of the five above were closed because the MILESTONE they belonged to
finished, not because the parser learned anything.

The corpus already has the machinery to catch this shape: `run.go`
reports a file marked refusing that now PASSES, which is the same
failure mode in the other direction. What it does not check is whether
the marker's issue still exists to be read.

## What this needs

A check over every `refuses:` record asking `git zhi` whether the cited
issue is open, failing when it is not. It has to tolerate `git zhi`
being absent -- the suite runs where the chain may not exist -- so a
skip when the binary is missing, and a failure when it is present and
says `done`.

Repointing the five is the other half and is mechanical once the check
names them: each needs an issue describing the gap it actually has, the
way `01a0cf3e`'s four were repointed at `01a0d087` and friends.

## What a closed citation actually costs, measured 2026-09-26

Found by `crochet:chain-review` on the `corpus-covers-t2` chain, and it is
the concrete answer to "why does this matter".

`conformance/mdtest/argument-extent.md` holds five tier-07 cases on
parenless argument extent -- greedy, prototype-cut, and the undeclared
callee as `parses: no`. Both refusing cases cite `01a0c432-fbd5`, state
`done`.

A reader followed that citation, found finished work, and concluded the
corpus had no case for the construct. An issue was then filed to write
coverage that already existed, with a milestone built around it. It was
cancelled once the topic was read (`01a0db76`).

So a closed citation does not merely fail to help -- it reads as tracked
work and actively misleads, where `unfiled` at least reads as an
admission. That asymmetry is the argument for this check.

## The existing check verifies existence, not openness

Worth stating so nobody concludes this check is already written.
`TestRefusalCitationMustResolve/corpus` PASSES today while four citation
sites name `done` issues -- measured at b9f77d9c. It asks whether an id
RESOLVES, not whether it is OPEN, which is exactly the gap
`TestRefusalCitationsAreLive` fills.

Also measured: `unfiled` is NOT part of this defect. It is a deliberate
sentinel (`file.go:275`, `unfiledRefusal`), the topic-format spelling of
`selfRecorded`, meaning the case is its own record;
`TestRefusalCitationMustResolve/unfiled_is_not_looked_up` asserts it and
passes. A chain-review issue proposing to reject `unfiled` was cancelled
on that evidence. This issue's five are the whole of it.

## Acceptance criteria

- [ ] A check names every refusal citing a closed issue (`go test ./internal/conformance/ -run '^TestRefusalCitationsAreLive$' -count=1 -v | grep -q '^--- PASS'`)
- [ ] No corpus case cites a closed issue (`go test ./internal/conformance/ -run '^TestRefusalCitationsAreLive$' -count=1 | grep -q '^ok'`)
- [ ] The check skips rather than fails where git-zhi is absent (`PATH=/usr/bin:/bin go test ./internal/conformance/ -run '^TestRefusalCitationsAreLive$' -count=1 | grep -qE '^(ok|\?)'`)
- [ ] The whole conformance suite is green (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
