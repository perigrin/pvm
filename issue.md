---
title: "sub :lvalue {} at statement start reads as a label; $#{...} is opaque"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-25T14:38:33.310030738Z
updated: 2026-10-25T14:38:33.310030738Z
---

Found by the 6030 worker (01a16602-6030, exec/6030) on 2026-10-25.

## Context

- paths: `internal/parse/` (statement start: label vs anonymous sub, and `$#` terms)

1. `sub :lvalue {1};` at the start of a statement parses as the label `sub:` followed by `lvalue {1}`. perl reads an anonymous sub with the :lvalue attribute. Measure with perl 5.44 (Deparse,-p).
2. `$#{$r->[0]}` is a single opaque term, so psc cannot see the deref inside it. Measure with perl's Concise, and parse it as the last-index of the dereferenced array.

## Acceptance criteria

### Positive Scenarios

- [ ] Both shapes give perl's tree (`go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 | grep -q '^ok'`)
