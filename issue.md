---
title: "Use-list reading from the fix round: a parenthesised literal list is names, a version before a list stays one statement, named lists install what perl does"
state: done
urgency: normal
milestone: core-module-types
created: 2026-10-21T05:19:06.072902414Z
updated: 2026-10-21T15:13:38.781974019Z
sessions:
- start_sha: 2db182e610c0a9eebf6a623550bd13f550d219a4
  end_sha: 9711a47ad2aaa4e8463f48604119e9b4ed2bf964
  commits: 2
  started_at: 2026-10-21T05:19:33.101444255Z
  ended_at: 2026-10-21T15:13:38.781974019Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-21T05:19:33.101444255Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-21T15:13:38.781974019Z
observed_paths:
- internal/infer/calls.go
- internal/infer/calls_test.go
- internal/infer/ops.go
- internal/infer/warnpossiblemismatch_test.go
- internal/infer/warnpossibleundef_test.go
- internal/parse/declarations.go
- internal/parse/derive.go
- internal/parse/derive_test.go
- internal/parse/issue1959_test.go
- internal/parse/use.go
---

crochet:review round 2 of PR #488 at 2db182e6 (2026-10-21). Quality report: scratchpad/review2-core-module-types-paad.md, probes scratchpad/r2probes/, psc builds of head, cf08c0ac and pu in scratchpad/r2bin/. A finding pu gets right is a regression this milestone introduced: fix it. A gap pu shares is out of scope here and goes to the follow-up named in the issue.

Quality I9 (use.go:378, also :324, :342: `use M ('x')` read as the empty list; `use Test::More ("no_plan")`, Getopt::Long, Mojolicious::Lite, Moose install nothing and `ok $x, "t"` parses as `$x->ok`), S14 (use.go:137-142: `use M VERSION LIST` splits into two statements), S15 (Moose.tpm with use.go:400-424: a named list installs every default; POSIX fcntl_h gap names no issue), S16 (declarations.go:378-407 misses nested and q() export forms).

## Steps
- [ ] Measure each use-line's installs with perl
- [ ] Report every m1 gate (real shim)

## Acceptance criteria

### Positive Scenarios
- [ ] TestParenLiteralListIsNames (`go test ./internal/parse/ -run '^TestParenLiteralListIsNames$' -count=1 -v | grep -q '^--- PASS: TestParenLiteralListIsNames'`)
- [ ] TestUseVersionListIsOneStatement (`go test ./internal/parse/ -run '^TestUseVersionListIsOneStatement$' -count=1 -v | grep -q '^--- PASS: TestUseVersionListIsOneStatement'`)
- [ ] TestNestedAndQExportFormsRefused (`go test ./internal/parse/ -run '^TestNestedAndQExportFormsRefused$' -count=1 -v | grep -q '^--- PASS: TestNestedAndQExportFormsRefused'`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] TestNamedListInstallsWhatPerlDoes (`go test ./internal/parse/ -run '^TestNamedListInstallsWhatPerlDoes$' -count=1 -v | grep -q '^--- PASS: TestNamedListInstallsWhatPerlDoes'`)
