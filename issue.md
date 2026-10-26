---
title: Twelve acceptance criteria are unverifiable because their command wraps to a second line
state: done
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-21T23:57:09.409395856Z
updated: 2026-10-26T00:44:22.956255165Z
sessions:
- start_sha: a3a0ed0c97ce48a68f5db427979f9d8acf88d3c8
  end_sha: a3a0ed0c97ce48a68f5db427979f9d8acf88d3c8
  commits: 0
  started_at: 2026-09-22T00:44:33.665567052Z
  ended_at: 2026-09-22T01:04:59.586016846Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-22T00:44:33.665567052Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-22T01:04:59.586016846Z
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

## A third rule, found while fixing this one

Reflowing the wrapped criteria closed the gap this issue names, and
discovery still fell short: 196 criteria declared, 179 found. Three issues
contributed ZERO, and every one of their criteria was correctly formatted
under both known rules.

Proven on throwaway probes -- two issues identical but for a trailing
subsection:

    **2026-10-26:** two criteria grepped a memory note that exists only on one machine, so they could not pass anywhere else. They now grep the same strings in docs/contributing/git-zhi-acceptance-criteria.md (PR #494).

## Acceptance criteria
    - [ ] trivially true (`true`)              discovered

    ## Acceptance criteria
    - [ ] trivially true (`true`)
    ### A following subsection                 contributes NOTHING

**A `###` subsection after the criteria block makes the entire block
invisible.** Not the part after it -- all of it.

The three affected issues each lost their criteria to a good-faith note
added during execution: `### Acceptance correction, during execution`,
`### Scope note, added during the lint's execution`, `### The format is
already settled`. 17 criteria, silently unrun.

This is the worst of the three rules. The header rule fails uniformly and
the wrapping rule reports `unverifiable`; this one deletes correct criteria
with no signal. The fix needs no tool change -- order the note before the
criteria and leave the criteria last, which is what was done here.

One more trap, also found here: a criterion must not invoke `git-zhi
verify` on its own milestone. That is verify running inside itself and
reports as unverifiable. `--dry-run` cannot recurse and is the safe form.

## Acceptance criteria

- [ ] verify discovers a command for every criterion, so none is unverifiable (`test "$(git-zhi verify m3-conformance-corpus --dry-run | wc -l)" -ge "$(git-zhi issue list --milestone m3-conformance-corpus --state done --format json | python3 -c 'import json,sys; print(sum(1 for i in json.load(sys.stdin) for l in i.get("body","").split(chr(10)) if l.startswith("- [")))')"`)
- [ ] every criterion in this milestone carries its command on the same line as its checkbox (`test "$(git-zhi issue list --milestone m3-conformance-corpus --state done --format json | python3 -c 'import json,sys; print(sum(1 for i in json.load(sys.stdin) for l in i.get("body","").split(chr(10)) if l.startswith("- [") and not l.rstrip().endswith(chr(96)+")")))')" = 0`)
- [ ] the two silently-enforced rules are written where an author reads them before writing an AC (`grep -q 'SAME LINE' docs/contributing/git-zhi-acceptance-criteria.md`)
- [ ] that record names the header rule too, which is the failure that hid an entire milestone (`grep -q 'Acceptance criteria   found' docs/contributing/git-zhi-acceptance-criteria.md`)
- [ ] no criterion was lost in the reflow: discovery is at least what it was before (`test "$(git-zhi verify m3-conformance-corpus --dry-run | wc -l)" -ge 157`)
