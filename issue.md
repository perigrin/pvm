---
title: A branch join drops an else that narrows nothing, as legacy did
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-13T01:51:16.083606254Z
updated: 2026-10-13T01:51:16.083606254Z
---

## Context

- paths: `internal/infer/lowered_flow.go` (the branch join, `applyBranchJoin`)
- commands: `go test ./internal/infer/ -run '^TestLoweredGoldenRatchet$'`

Found 2026-10-13 by the flow port (01a12a85-3100), which copied it from legacy so the goldens match: when an else branch narrows nothing it contributes nothing to the join, so after `if ($x isa Foo) {} else {}` PSC says `$x` is `Object` (fixtures `flow/isa_else`, `or_else`, `not_and` leave `Ref`). The else branch is exactly where the guard failed, so the join must include the un-narrowed type. perl 5.42: `my $x; if ($x isa Foo) {1} else {2}` leaves `$x` undef. Fix after the cutover (01a0c115-df67), with 01a132ab; each golden edit gets a `Diagnostics-changed:` trailer.

## Steps

- [ ] Failing test first; an else branch joins with the type its guard leaves; edit goldens under trailers; precision must not fall; `make test`; commit

## Acceptance criteria

- [ ] an un-narrowing else keeps the pre-guard type in the join (`go test ./internal/infer/ -run '^TestElseBranchJoinsUnnarrowed$' -count=1 -v | grep -q '^--- PASS: TestElseBranchJoinsUnnarrowed'`)
