---
title: "TestParseRatchet re-parses every imported module per file: about 280 seconds"
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T21:55:19.277316665Z
updated: 2026-10-01T05:07:00.171928466Z
sessions:
- start_sha: 555b4ff88e27fe5ed3f19052f9d50091492432ac
  end_sha: ""
  commits: 0
  started_at: 2026-10-01T05:07:00.171928466Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-01T05:07:00.171928466Z
---

TestParseRatchet takes about 280 seconds now that T1 resolves imports through perl 5.42's @INC and perl.git's core sources, up from about 130 before the source roots were added and the CPAN modules were installed. It was a few seconds when T1 parsed without a loader. Each file's ParseFileFrom builds a fresh resolver, so Test::More, Moose, Mojolicious and their dependencies are re-parsed for every file that uses them. A resolver cache shared across one ratchet run's files would parse each module once.

## Acceptance criteria

- [ ] TestParseRatchet parses each module once per run (`sh -c 'go test ./internal/parse/ -run "^TestParseRatchet$" -count=1 2>&1 | grep -qE "^ok.*[0-9]{1,2}\.[0-9]+s$"'`)
