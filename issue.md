---
title: A module loaded with a computed import list does not record its qualified subs
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T18:34:33.826064748Z
updated: 2026-09-29T18:34:34.817454998Z
sessions:
- start_sha: b7dfdf5c7c63e5c4c972825917d269bcc956f422
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T18:34:34.817454998Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T18:34:34.817454998Z
---

b7dfdf5c records a loaded module's subs under their qualified names, but after the import list is read -- and resolveImports returns early on a computed list. `use overload '""' => sub {...}` loads overload.pm and imports nothing statically readable, so `overload::constant qr => sub {...}` later in the file still refused. Loading is not importing: the qualified names do not depend on the list.

perl.git t/re/overload.t:65 and 247; its only loads of overload have computed lists (lines 18, 42, 80, 92, 104).

## Acceptance criteria

- [ ] A module loaded with a computed import list still has its subs known by qualified name (`go test ./internal/parse/ -run '^TestQualifiedModuleSub$' -count=1 -v | grep -q '^--- PASS: TestQualifiedModuleSub'`)
