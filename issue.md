---
title: "`defer { } { }` takes the second block as an argument: defer is not a statement form"
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T17:50:50.996259488Z
updated: 2026-09-30T17:57:44.43078771Z
sessions:
- start_sha: c9401c372fe12638de681ca135ac5a13004384a7
  end_sha: 826a07cd41c07a965caad08affb84b6b9cce9f0f
  commits: 1
  started_at: 2026-09-30T17:50:51.825715248Z
  ended_at: 2026-09-30T17:57:44.43078771Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-30T17:50:51.825715248Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-30T17:57:44.43078771Z
observed_paths:
- internal/parse/control.go
- internal/parse/deferstmt_test.go
- internal/parse/testdata/t1.ratchet
- internal/parse/testdata/t1fidelity.ratchet
- internal/parse/use.go
---

Under `use feature 'defer'`, `defer BLOCK` is read as an unknown word taking a block. So a `{` after the defer's block becomes an anonymous-hash argument: `{ defer { ... } { ... } }` canons as `defer {1;}{2};` or refuses. perl makes defer a statement form ending at its block.

Measured with perl 5.42: `{ defer { $l .= "o" } { defer { $l .= "i," } } $l .= "m"; }` deparses as a defer, a bare block and an assignment, all siblings.

PerlOnJava unit/defer.t:83.

## Acceptance criteria

- [ ] `defer BLOCK` is a statement (`go test ./internal/parse/ -run '^TestDeferIsAStatement$' -count=1 -v | grep -q '^--- PASS: TestDeferIsAStatement'`)
