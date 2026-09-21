---
title: git-zhi verify reports 0/0 on a milestone whose done issues all have Acceptance sections
state: in-progress
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-21T18:42:18.436831282Z
updated: 2026-09-21T22:29:10.752498598Z
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
