---
title: A slurpy in a .pmt takes no bare element type, and its sigils are the caller's view
state: cancelled
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fa46-a8d1-77c5-97d8-66cf3b4ec55d
blocks:
- 01a0fa46-be8b-7823-9c6c-5be1406b57b8
created: 2026-10-02T01:42:13.264416913Z
updated: 2026-10-16T18:48:30.501210147Z
transitions:
- state: cancelled
  actor: human:git-zhi
  timestamp: 2026-10-02T17:18:12.471857469Z
---

RFC 0001, "A slurpy takes no bare element type": `Str @args` is rejected (ambiguous container and assertion); a slurpy is untyped or carries an explicit container type, `List[Str] @args`. In a `.pmt` a sigil is the caller's view, as a prototype's characters are: in `(Array @, List @)` only the final `List` or `Hash` parameter is slurpy, which perl's own signatures reject ("Multiple slurpy parameters not allowed"). Re-mark the RFC section *Implemented*.

## Acceptance criteria

- [ ] Bare typed slurpies are rejected and caller-view sigils parse (`go test ./internal/parse/ -run '^TestTpmSlurpyRule$' -count=1 -v | grep -q '^--- PASS: TestTpmSlurpyRule'`)
- [ ] The RFC section is implemented (`grep -qF '### A slurpy takes no bare element type (*Implemented*' docs/rfcs/0001-type-libraries.md`)
