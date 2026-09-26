---
title: The chain tool is four minors behind, and the gate step it breaks is the AC audit
state: pending
urgency: normal
milestone: v0.1
created: 2026-09-26T04:21:41.258041012Z
updated: 2026-09-26T04:21:41.258041012Z
---

The chain tool is four minor versions behind what the plugin expects, and
the gate step that would have caught a malformed acceptance criterion is
unrunnable because of it.

Filed by `crochet:chain-review` round 2 rather than left as an advisory
note, because an advisory with no owner is the unlabelled deferral this
project's Plan Discipline forbids.

## Measured, 2026-09-26

    $ git zhi version | head -1
    git-zhi 0.5.2 (linux/amd64)

    $ grep git_zhi_min_version <plugin>/.claude-plugin/plugin.json
    "git_zhi_min_version": "0.7.2"

Preflight's spec says warn and continue, and that is correct -- a version
check must never block a skill. But it also means the gap accumulates
silently, and this gate hit two consequences of it in one session.

## What it cost, twice

**Step 2.5 of `chain-review` could not run.** The step says:

    git-zhi verify <milestone> --dry-run

and cross-check the listed commands. On this install `--dry-run` reads only
DONE issues, so an all-pending chain prints nothing and exits 1. Both
rounds had to reproduce the extractor by hand over
`milestone show --format json` -- which found three defective criteria in
round 1 and two more in round 2, so the step matters.

**`issue edit` has no `--title`.** Two issues in this milestone were
cancelled and re-filed for no reason but a stale title
(`01a0db77` -> `01a0dba9`, `01a0db79-6a25` -> `01a0db9f`), which leaves
cancelled rows in the chain that read as failures rather than as renames.

Whether either is fixed in 0.7.2 is NOT established -- that is this
issue's first task, not its premise.

## Scope

1. Measure what 0.7.2 changes for the two symptoms above. `--dry-run` on a
   pending chain, and whether `issue edit` grew a title flag. If neither
   moved, this issue closes with that recorded and the version gap stops
   being blamed for them.
2. Upgrade, or record here why not.
3. If `--dry-run` still cannot see pending issues, that is a gate defect
   rather than a version gap, and `chain-review`'s step 2.5 should say so
   instead of prescribing a command that yields nothing.

## Context

- paths: the crochet plugin cache, `.claude/crochet/capabilities.json`
- commands: `git zhi version`, `git zhi verify <milestone> --dry-run`
- entrypoints: `crochet:preflight` check 1 (the version comparison), `crochet:chain-review` step 2.5

## Acceptance criteria

- [ ] The installed version is no longer 0.5.x (`! git zhi version | head -1 | grep -q ' 0\.5\.'`)
- [ ] The findings are written down where the next reader hits them (`grep -q 'dry-run' docs/audits/2026-09-26-chain-tool-version.md`)
- [ ] Whichever way the dry-run measurement went, this issue says so (`grep -qE 'pending chain|DONE issues' docs/audits/2026-09-26-chain-tool-version.md`)
