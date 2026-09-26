---
title: A dispatched worktree is based on an ancestor, not the branch tip
state: pending
urgency: normal
milestone: v0.1
created: 2026-09-26T05:18:07.183546099Z
updated: 2026-09-26T05:18:07.183546099Z
---

An agent dispatched with `isolation: "worktree"` gets a worktree based on an
ancestor commit, not on the branch tip. Both workers in `corpus-covers-t2`
were handed `53de041e`, which predates `internal/parse/` entirely.

Filed because both noticed independently and both flagged it, in the same
words: "Worth checking whether other workers in this milestone got the same
stale base."

## Measured, 2026-09-26

Two agents dispatched with `isolation: "worktree"` from a session on
`feat/m1-parse-skeleton` at `58076b8a`. Both reported their worktree
starting at `53de041e` ("fix(infer): builtins whose result follows their
argument"), which is a genuine ancestor:

    $ git merge-base --is-ancestor 53de041e HEAD   # exit 0

but old enough that `internal/parse/` -- the directory one issue's whole
scope named -- did not exist in it.

## Why it is worse than it looks

Both workers caught it, checked for unique commits, and fast-forwarded onto
the branch tip before starting. A worker that did NOT notice has three
routes, all bad:

  - report the issue's `paths:` as missing and stop, which reads as a
    malformed issue rather than a stale checkout;
  - write the files from scratch, since they are absent, producing a
    parallel implementation of code that already exists;
  - work against the old tree and produce a diff that applies to nothing.

None of those is distinguishable from ordinary failure in the worker's
report, which is what makes this worth a check rather than a note.

## What is NOT wrong

The isolation itself worked. Two workers committed to their own branches
with private indexes and both merged into the tip with only the conflicts
you would expect from two issues touching one generated baseline. The
defect is the BASE, not the mechanism.

Also measured: nothing was lost. Both fast-forwards were from a clean tree
with no unique commits.

## Scope

1. Establish whether the base is chosen by the harness, by git, or by the
   session's HEAD at dispatch time. Do not guess -- the three have
   different fixes.
2. If the harness picks it, the fix is to branch from the session's current
   HEAD.
3. If it cannot be fixed here, the mitigation is a line in the dispatch
   prompt: verify the worktree's base is the branch tip before reading any
   file, and fast-forward if not. Both workers did this unprompted; the
   next one may not.

## Context

- paths: n/a -- this is harness behaviour, not repository code
- commands: `git log --oneline -1` in a dispatched worktree, compared against the dispatching session's HEAD
- entrypoints: the Agent tool's `isolation: "worktree"` mode

## Acceptance criteria

- [ ] The base-selection behaviour is written down where the next dispatcher finds it (`grep -qiE 'worktree.*base|stale base' docs/audits/2026-09-26-worktree-base.md`)
- [ ] The finding names which of the three causes it is (`grep -qiE 'harness|session HEAD|git default' docs/audits/2026-09-26-worktree-base.md`)
- [ ] Both observing workers are cited, so the measurement is traceable (`grep -q '53de041e' docs/audits/2026-09-26-worktree-base.md`)
