---
title: A sub declared in one package may count as declared in another
state: done
urgency: normal
milestone: v0.1
created: 2026-10-06T10:59:35.644284842Z
updated: 2026-10-06T12:07:52.091156162Z
sessions:
- start_sha: c391da4f920448d9a062cf049a4e591d9ed1d24a
  end_sha: c391da4f920448d9a062cf049a4e591d9ed1d24a
  commits: 0
  started_at: 2026-10-06T12:07:51.931978144Z
  ended_at: 2026-10-06T12:07:52.091156162Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-06T12:07:51.931978144Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-06T12:07:52.091156162Z
---

## Context

- paths: `internal/parse` (`declareSub`, `Import`, `internal/parse/refscalar.go`)
- commands: `make test`; ratchets; perl 5.42

Found while building 01a10dc1-3530-762f-9a03-177445629c6b, unconfirmed. `declareSub` records `sub f` without its package, so `sub f` declared inside `package Foo` makes a bare `f()` in package main count as declared. The `\$`-slot refusal (`ref_scalar_slot`) refuses a declared non-lvalue sub's result, so `sref(f())` in main after `package Foo; sub f {}` may be wrongly refused (perl: main::f is undeclared there, so the call compiles). No ratchet file hits it today.

## Steps

- [ ] Reproduce with a minimal program; measure with `perl -c`
- [ ] If confirmed: failing test first; record the package with each declared sub; ratchets must not regress; `make test`; commit

## Acceptance criteria

- [ ] a sub declared in another package does not count as declared in main (`go test ./internal/parse/ -run '^TestDeclaredSubIsPackageScoped$' -count=1 -v | grep -q '^--- PASS: TestDeclaredSubIsPackageScoped'`)
