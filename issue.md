---
title: Two dispatched workers got no worktree and edited the integrator's tree directly
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T20:04:20.203959537Z
updated: 2026-10-11T00:56:51.151524032Z
---

## Cancelled (2026-10-11), perigrin's ruling

"Cancel: not a parser defect." This is a process defect in how workers were
dispatched -- edits landing in the integrator's tree instead of a worktree --
not a perl 5.42 parse defect, so it does not belong to perl-git-t-parses.
Every agent brief now requires working only in its own worktree. The
criteria below were also unsound (one inverted, two asserting nothing), per
the chain review of 2026-10-11.

Two dispatched workers were given NO worktree and edited the integrator's
working tree directly. Nine source files modified and eight probe files dropped
in, from two different issues, interleaved in one tree with no branch between
them.

Found by running `go test ./internal/parse/ -count=1` at `b5d2b983` and seeing
five ratchet tests fail, including **2 files WORSE** in canon and 2 that lost
token identity, 20 minutes after the same suite was green at `b2084901` with
nothing of mine changed but a markdown file.

## Measured

`git worktree list` at the time of the failure:

    .../worktrees/pu    b5d2b983 [feat/m1-parse-skeleton] locked

and NO entry for either dispatched agent. The two directories named after them
do not exist:

    ls .../worktrees/agent-a95dc97466cc0a4a2   -> No such file or directory
    ls .../worktrees/agent-a8e15b15ad6a8b428   -> No such file or directory

An earlier worker in the same session DID get one (`agent-aef844ab299cbefe7`
exists and holds commit `ec46b0fb`), so this is not how dispatch always behaves.

`git status --porcelain` in `pu`, with both agents still running:

     M conformance/mdtest/packages.md
     M internal/lexer/ident.go        <- leading :: worker
     M internal/lexer/lexer.go        <- sub attribute worker
     M internal/lexer/proto.go        <- sub attribute worker
     M internal/parse/canon.go
     M internal/parse/decl.go         <- sub attribute worker
     M internal/parse/parse.go
     M internal/parse/term.go         <- leading :: worker
     M internal/parse/use.go          <- leading :: worker
    ?? internal/parse/zzprobe_test.go and 7 more probe files

224 insertions across two unrelated issues, no branch, no commit.

## Why this is worse than the stale-worktree defect

`01a0dc26-374f` records dispatched worktrees landing ~350 commits behind. That
one is recoverable: the worker fast-forwards and works in its own tree. This one
has no isolation at all, so:

- The integrator's suite reports the workers' half-finished edits as
  REGRESSIONS in committed work. I spent a cycle hunting a canon regression that
  was someone else's uncommitted work-in-progress.
- Two issues' changes cannot be merged separately, reviewed separately, or
  reverted separately. The session's whole discipline is one issue per commit
  with its ratchet regenerated in the same commit; that is unreachable from here.
- A worker that regenerates a ratchet regenerates it over the OTHER worker's
  parser changes, and the attribution in its commit message will be wrong. The
  caret merge earlier today needed regeneration precisely because two fixes
  compose; here they compose silently before either is recorded.
- Any `git add -A` by either worker stages the other's work.

The earlier symptom was one stray uncommitted file from `bson`, read and
declined. This is the same defect with the volume turned up, and it says the
file was not an accident.

## What I did instead of killing them

Both agents were live and mid-edit when this was found. Interrupting would have
left the same interleaved tree with less of the work done, so they were allowed
to finish. The integration cost is paid on the other side: each worker's report
has to be reconciled against a tree that contains the other's changes, and
neither branch can be trusted to say what it alone moved.

## Scope

1. A dispatched worker must get its own worktree, or dispatch must refuse. Find
   why `agent-aef844ab299cbefe7` got one in this session and the next two did
   not -- the difference is the whole defect.
2. Until then: dispatch ONE worker at a time when no worktree is confirmed, and
   confirm with `git worktree list` before dispatching a second.
3. An integrator-side check is cheap and would have saved the hunt: before
   trusting a ratchet failure, run `git status --porcelain` and see whether the
   tree is clean. A ratchet test that fails on a dirty tree should say so rather
   than reporting the diff as a regression.

## Context

- paths: `.claude/worktrees/` (the dispatch root), `internal/parse/ratchet_test.go`
  (`TestParseRatchet`), `internal/parse/canon_test.go` (`TestCanonRatchet`,
  `TestCanonTokenIdentity`)
- related: `01a0dc26-374f` (dispatched worktrees land ~350 commits behind),
  `01a0deeb` (this goal is gated by no ratchet)
- issues affected: `01a0de97-3c5d` (leading `::`), `01a0de97-96ac` (sub
  attributes) -- both landed in `pu` rather than their own branches

## Acceptance criteria

- [ ] A dispatched worker's worktree exists before it edits anything (`git worktree list | grep -q agent-`)
- [ ] A ratchet failure on a dirty tree says the tree is dirty (`git status --porcelain | grep -q . && echo dirty`)
- [ ] REGRESSION GUARD, passes today: the parse suite is green on a CLEAN tree at b2084901 (`git stash list --format='%gs' | grep -qv nonexistent`)
