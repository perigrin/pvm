---
title: Decide the type names CORE.pmt needs beyond the lattice
state: done
urgency: normal
milestone: type-libraries
blocks:
- 01a0fa46-a8d1-77c5-97d8-66cf3b4ec55d
- 01a0fa46-fccf-7f53-a476-c02f8f07616d
created: 2026-10-02T01:42:05.536728833Z
updated: 2026-10-02T17:17:25.640388805Z
sessions:
- start_sha: 84ecc3da82ef42d1f02a9b917c70a498b4e103d0
  end_sha: fa483015ee7f62858d7ecee96d26a43b7c67ff8d
  commits: 1
  started_at: 2026-10-02T03:25:40.371418742Z
  ended_at: 2026-10-02T03:30:53.629949543Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-02T03:25:40.371418742Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-02T03:30:53.629949543Z
observed_paths:
- docs/rfcs/0001-type-libraries.md
---

Typed declarations name types the lattice in `internal/types` lacks or has not spelled: `FileHandle` (the lattice has `Glob` and `GlobRef`; the paper has `IO`, pvm does not), and unions (`Str|Undef` is proposed: the lattice is a bitmask, so a union is an OR). An unknown type name in a `.pmt` must fail, so the set has to be settled before typed parsing validates names.

perigrin decides; the decision lands in RFC 0001 as `### Type names (*Decided*)`.

## Acceptance criteria

- [ ] The RFC records the decision, decided or since implemented (`grep -q '^### Type names (\*\(Decided\|Implemented\)\*' docs/rfcs/0001-type-libraries.md`)
