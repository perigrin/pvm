---
title: "TestParseRatchet resolves no imports: T1 needs a loader"
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T19:30:14.317092503Z
updated: 2026-09-30T19:30:15.430930928Z
sessions:
- start_sha: b813a6a6e535a1e1e296a74dc15a2b453d9e97de
  end_sha: ""
  commits: 0
  started_at: 2026-09-30T19:30:15.430930928Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-30T19:30:15.430930928Z
---

TestParseRatchet parses T1 with parse.Parse, which resolves no imports. So every T1 file that calls an imported sub without parentheses refuses on a name perl knows: Test::More's `is` and `ok` alone account for about 200 of the 236 dirty files.

perigrin decided (2026-09-30) that the ratchet gets a loader: each file goes through ParseFileFrom, rooted at its own directory and at the @INC of the perl the corpus is measured against (conformance.PerlPath, pinned to 5.42.0).

## Acceptance criteria

- [ ] TestParseRatchet resolves imports through perl 5.42's @INC (`sh -c 'grep -q "with perl 5.42" internal/parse/testdata/t1.ratchet'`)
