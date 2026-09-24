---
title: Five refusal citations point at closed issues, and nothing checks
state: pending
urgency: normal
milestone: v0.1
created: 2026-09-24T00:18:19.052459875Z
updated: 2026-09-24T15:27:50.062966854Z
---

Five corpus files carry `STATUS refuses` citing issues that are CLOSED.
A reader following the citation lands on a finished issue that says
nothing about the refusal in front of them.

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

A citation is a string in a comment. Nothing reads it, so nothing
notices when the issue it names closes -- and issues close routinely,
including for reasons that have nothing to do with the refusal. Three
of the five above were closed because the MILESTONE they belonged to
finished, not because the parser learned anything.

The corpus already has the machinery to catch this shape: `run.go`
reports a file marked refusing that now PASSES, which is the same
failure mode in the other direction. What it does not check is whether
the marker's issue still exists to be read.

## What this needs

A check over every `STATUS refuses` line asking `git zhi` whether the
cited issue is open, failing when it is not. It has to tolerate `git
zhi` being absent -- the suite runs where the chain may not exist -- so
a skip when the binary is missing, and a failure when it is present and
says `done`.

Repointing the five is the other half and is mechanical once the check
names them: each needs an issue describing the gap it actually has, the
way `01a0cf3e`'s four were repointed at `01a0d087` and friends.

## Acceptance criteria

- [ ] A check names every refusal citing a closed issue (`go test ./internal/conformance/ -run '^TestRefusalCitationsAreLive$' -count=1 -v | grep -q '^--- PASS'`)
- [ ] No corpus file cites a closed issue (`go test ./internal/conformance/ -run '^TestRefusalCitationsAreLive$' -count=1 | grep -q '^ok'`)
- [ ] The check skips rather than fails where git-zhi is absent (`PATH=/usr/bin:/bin go test ./internal/conformance/ -run '^TestRefusalCitationsAreLive$' -count=1 | grep -qE '^(ok|\?)'`)
- [ ] The whole conformance suite is green (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
