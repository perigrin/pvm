---
title: "`*my_push = \\&CORE::push` leaves my_push shapeless: an alias should take its target's prototype"
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T18:27:47.126800997Z
updated: 2026-09-30T18:35:28.053919094Z
sessions:
- start_sha: 1a15ffd68adb98185d45dc93c6585a6cc8a8b33e
  end_sha: fed0bab5ed153c134139df0f67090f36146d3ffc
  commits: 1
  started_at: 2026-09-30T18:27:48.875514558Z
  ended_at: 2026-09-30T18:35:28.053919094Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-30T18:27:48.875514558Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-30T18:35:28.053919094Z
observed_paths:
- internal/parse/coreprotos.go
- internal/parse/globalias_test.go
- internal/parse/testdata/t1.ratchet
- internal/parse/use.go
---

`BEGIN { *my_push = \&CORE::push; } my_push @arr2, 4, 5, 6;` refuses. A glob assignment in BEGIN records the name with its prototype unknown. An alias takes its target's prototype.

Measured with perl 5.42: `prototype(\&my_push)` is `\@@` after the alias, and `sub f ($$) {} BEGIN { *g = \&f }` gives `g` the prototype `$$`. The parser has no table of builtin prototypes, so one is generated from perl 5.42's `prototype("CORE::name")` over regen/keywords.pl's list: 188 of 266 keywords have one.

PerlOnJava unit/core_subroutine_refs.t:82-84.

## Acceptance criteria

- [ ] A BEGIN glob alias takes its target's prototype (`go test ./internal/parse/ -run '^TestGlobAliasPrototype$' -count=1 -v | grep -q '^--- PASS: TestGlobAliasPrototype'`)
