---
title: Every construct CORE.pmt uses works in a library .pmt
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-02T05:00:17.750893693Z
updated: 2026-10-02T05:00:17.750893693Z
---

RFC 0001, "One language for every `.pmt`" (perigrin, 2026-10-02): everything CORE.pmt can say, a library's `.pmt` can say -- typed signatures, `multi`, `:context`, the invocant colon, declared syntax. Nothing in the reader may special-case CORE. A test declares each construct in a module `.pmt` (not CORE) and parses a call through it. Chalk is to read the same files in time, so the syntax stays parseable from the RFC alone.

## Acceptance criteria

- [ ] Each construct works from a library .pmt (`go test ./internal/parse/ -run '^TestLibraryPmtSpeaksCoreLanguage$' -count=1 -v | grep -q '^--- PASS: TestLibraryPmtSpeaksCoreLanguage'`)
- [ ] The RFC section is implemented (`grep -q '^### One language for every .*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
