---
title: A branch join drops an else that narrows nothing, as legacy did
state: in-progress
urgency: normal
milestone: v0.1
created: 2026-10-13T01:51:16.083606254Z
updated: 2026-10-16T11:47:48.243469873Z
sessions:
- start_sha: 9ff9be88b6f25ac747648420c2ff2433c1b55456
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T11:47:48.243469873Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T11:47:48.243469873Z
---

## Context

- paths: `internal/infer/lowered_flow.go` (the branch join, `applyBranchJoin`)
- commands: `go test ./internal/infer/ -run '^TestLoweredGoldenRatchet$'`

Found 2026-10-13 by the flow port (01a12a85-3100), which copied it from legacy so the goldens match: when an else branch narrows nothing it contributes nothing to the join, so after `if ($x isa Foo) {} else {}` PSC says `$x` is `Object` (fixtures `flow/isa_else`, `or_else`, `not_and` leave `Ref`). The else branch is exactly where the guard failed, so the join must include the un-narrowed type. perl 5.42: `my $x; if ($x isa Foo) {1} else {2}` leaves `$x` undef. Fix after the cutover (01a0c115-df67), with 01a132ab; each golden edit gets a `Diagnostics-changed:` trailer.

Also (2026-10-14, round-2 narrowing fixes): in `if (!defined $x) {...} else {...}` on a defined `$x`, the then-branch is unreachable (None) and the else adds nothing to the join, so `$x` is None after the statement -- hiding any diagnostic on it. The fix here must make the else contribute its own type.

## Steps

- [ ] Failing test first; an else branch joins with the type its guard leaves; edit goldens under trailers; precision must not fall; `make test`; commit

## Acceptance criteria

- [ ] an un-narrowing else keeps the pre-guard type in the join (`go test ./internal/infer/ -run '^TestElseBranchJoinsUnnarrowed$' -count=1 -v | grep -q '^--- PASS: TestElseBranchJoinsUnnarrowed'`)
