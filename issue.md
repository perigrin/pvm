---
title: git-zhi verify reports 0/0 on a milestone whose done issues all have Acceptance sections
state: in-progress
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-21T18:42:18.436831282Z
updated: 2026-09-21T22:49:09.567030647Z
sessions:
- start_sha: b41801021927f68fcbf799c60d6b947393d983de
  end_sha: ""
  commits: 0
  started_at: 2026-09-21T22:29:10.752498598Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T22:29:10.752498598Z
---

`git-zhi verify m3-conformance-corpus` reports `total: 0` -- it finds no
acceptance criteria at all, across four issues in `done` state whose
bodies each carry a `## Acceptance` section with backtick-delimited
commands.

Measured:

    $ git-zhi verify m3-conformance-corpus --format json
    {"milestone":"m3-conformance-corpus","total":0,"passed":0,...,"results":[]}

    $ git-zhi verify m3-conformance-corpus --dry-run
    (no output)

    $ git-zhi issue list --milestone m3-conformance-corpus --state done
      01a0c35d  The corpus file format, decided rather than spiked           done
      01a0c35d  The glossary is the token assertion's dictionary             done
      01a0c35e  The dependency lint: ops-subset                              done
      01a0c35f  Tier READMEs                                                 done

    $ git-zhi issue show 01a0c35d-6e43 | grep -n '^## '
      52:## Acceptance

`verify --help` says it "finds all done issues, extracts backtick-delimited
commands from their Acceptance Criteria sections". The issues are done and
the sections exist, so either the section header must read `## Acceptance
Criteria` rather than `## Acceptance`, or the milestone lookup is not
reaching these issues.

This matters because verify is the gate that is supposed to catch an
acceptance criterion naming a test that does not exist -- which is exactly
the drift found by hand on 01a0c35f-a487, where AC1 named
`TestIntroducesReadFromReadmes` and AC3 named
`TestAspirationalIntroducesReported`, neither of which was ever written.
A verify that reports 0/0 reports success-shaped output for an issue whose
criteria are unrunnable.

All six criteria on 01a0c35f-a487 were confirmed passing by running each
command by hand at 0ce515cb.

## The cause, measured on a throwaway

The header name is the whole bug, and the word `criteria` is the load-bearing
part of it. Three disposable issues were created in a disposable milestone,
identical but for the header line, each carrying one criterion --
`- [ ] a trivial command runs` with `true` in backticks and parens:

    ## Acceptance            -> not found; the issue contributes nothing
    ## Acceptance criteria   -> found; the command runs and passes
    ## Acceptance Criteria   -> found; the command runs and passes

    $ git-zhi verify zzscratch-verify-probe
    01a0c61a  probe B lower criteria
      Positive:
        true OK
    01a0c61a  probe C title criteria
      Positive:
        true OK

    zzscratch-verify-probe: 2/2 acceptance criteria passing

Probe A -- the bare `## Acceptance` spelling -- does not appear in the output
at all. Case does not matter; the word `criteria` does. The probes were
purged afterwards and the scratch milestone left empty.

The same evidence exists without a probe. `git-zhi verify m1-parse-the-core`
reads exactly two issues, 01a0b6a2 and 01a0b721, and both spell the header
`## Acceptance criteria`. Every other done issue in that milestone spells it
`## Acceptance` and is skipped in silence -- as were all 28 done issues here.

So the help text is accurate and the house style was wrong. There is nothing
to fix in git-zhi.

## What was done

The header was renamed from `## Acceptance` to `## Acceptance criteria` on
all 28 done issues in this milestone. Each body was captured to a file
first, exactly one line was changed, and the result written back with
`--body-file`; all 28 were then re-read and confirmed byte-identical to the
intended text, so no prose was lost.

Discovery went from 0 commands to 141.

**House style, going forward: the section header is `## Acceptance criteria`.**
A bare `## Acceptance` is invisible to the gate, which is the worst possible
failure mode -- a milestone whose criteria have never been run reports
success-shaped output.

## What is still open

Twelve issues elsewhere in the chain still carry the bare header, four of
them open issues in this milestone (01a0c547-516b, 01a0c5b9-ac65,
01a0c5e6-9f98, 01a0c605-eff0). They were left alone deliberately: other
agents held them at the time and a body rewrite would have clobbered
concurrent edits. Each will go silently unverified the moment it is marked
done, so the header must be corrected before closing it. The third
criterion below is the check.

## Acceptance criteria

- [ ] verify finds criteria in this milestone rather than reporting none (`test "$(git-zhi verify m3-conformance-corpus --dry-run | wc -l)" -gt 0`)
- [ ] it finds at least one criterion per done issue, so no done issue is silently skipped (`test "$(git-zhi verify m3-conformance-corpus --dry-run | wc -l)" -ge 28`)
- [ ] no done issue in this milestone carries the invisible bare header (`test "$(git-zhi issue list --milestone m3-conformance-corpus --state done --format json | grep -c '## Acceptance\\n')" = 0`)
