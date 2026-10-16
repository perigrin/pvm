---
title: Every construct CORE.pmt uses works in a library .pmt
state: cancelled
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fa46-a8d1-77c5-97d8-66cf3b4ec55d
- 01a0fa46-be8b-7823-9c6c-5be1406b57b8
- 01a0fa46-ece2-7c78-87ed-84da0b5d65f9
- 01a0fa46-f147-75f6-bd3c-4aa5f5644ee5
created: 2026-10-02T05:00:17.750893693Z
updated: 2026-10-16T18:48:30.707224854Z
transitions:
- state: cancelled
  actor: human:git-zhi
  timestamp: 2026-10-02T17:18:12.872763941Z
---

RFC 0001, "One language for every `.pmt`" (perigrin, 2026-10-02): everything CORE.pmt can say, a library's `.pmt` can say -- typed signatures, `multi`, `:context`, the invocant colon, declared syntax. Nothing in the reader may special-case CORE. A test declares each construct in a module `.pmt` (not CORE) and parses a call through it. Chalk is to read the same files in time, so the syntax stays parseable from the RFC alone.

## Acceptance criteria

- [ ] Each construct works from a library .pmt (`go test ./internal/parse/ -run '^TestLibraryTpmSpeaksCoreLanguage$' -count=1 -v | grep -q '^--- PASS: TestLibraryTpmSpeaksCoreLanguage'`)
- [ ] The RFC section is implemented (`grep -q '^### One language for every .*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
