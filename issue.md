---
title: "Decide a type for \\$'s scalar container"
state: done
urgency: normal
milestone: type-libraries
blocks:
- 01a0fa46-be8b-7823-9c6c-5be1406b57b8
created: 2026-10-02T01:42:07.024701879Z
updated: 2026-10-02T05:45:34.709133245Z
sessions:
- start_sha: 32ed4bbd8aac49e99ed796c1496b5b6e4f132973
  end_sha: 4ee57eb8009511283e7a3120ba36301640feb9fe
  commits: 1
  started_at: 2026-10-02T05:40:12.721902732Z
  ended_at: 2026-10-02T05:45:34.709133245Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-02T05:40:12.721902732Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-02T05:45:34.709133245Z
observed_paths:
- docs/rfcs/0001-type-libraries.md
---

RFC 0001 open question 11. The prototype `\$` requires an actual scalar variable -- measured on 5.42, `sref(1)` dies "Type of arg 1 to main::sref must be scalar (not constant item)" -- where `$` takes any expression. Derivation needs a parameter type for the container, distinct from a scalar value, as `Array` is distinct from `List`.

perigrin decides; the decision lands in RFC 0001 as `### The scalar container (*Decided*)`.

## Acceptance criteria

- [ ] The RFC records the decision (`grep -qF '### The scalar container (*Decided*' docs/rfcs/0001-type-libraries.md`)
