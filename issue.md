---
title: Decide how a builtin with no prototype opts out of prototype derivation
state: done
urgency: normal
milestone: type-libraries
blocks:
- 01a0fa46-be8b-7823-9c6c-5be1406b57b8
- 01a0fa46-d9f7-7aab-b4c6-19b4050b7907
created: 2026-10-02T01:42:06.335001949Z
updated: 2026-10-02T17:17:25.86018853Z
sessions:
- start_sha: fa483015ee7f62858d7ecee96d26a43b7c67ff8d
  end_sha: e3277247cfd51a2d6ad7c569b474fadaccc69d74
  commits: 1
  started_at: 2026-10-02T04:51:32.12435992Z
  ended_at: 2026-10-02T04:59:11.15603338Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-02T04:51:32.12435992Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-02T04:59:11.15603338Z
observed_paths:
- docs/rfcs/0001-type-libraries.md
---

RFC 0001 open question 9. Prototypes derive from typed signatures, but about 20 builtins have no prototype because their parse is their own (`print`'s filehandle slot, `sort`'s block). `print`'s typed signature would derive a prototype `print` does not have, and applying it would change how `print` parses. Absent and empty differ for `:prototype` already (`()` is niladic), so the opt-out needs its own spelling.

perigrin decides; the decision lands in RFC 0001 as `### Builtins that keep their own parse (*Decided*)`.

## Acceptance criteria

- [ ] The RFC records the decision, decided or since implemented (`grep -q '^### Builtins that keep their own parse (\*\(Decided\|Implemented\)\*' docs/rfcs/0001-type-libraries.md`)
