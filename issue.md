---
title: "`overload::constant ...` refuses: a loaded module's subs are not known by qualified name"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T18:15:10.503591614Z
updated: 2026-09-29T18:30:31.856758263Z
sessions:
- start_sha: 7bc912c2b1f6571957c11606f1ca1b8dfbec6e67
  end_sha: b7dfdf5c7c63e5c4c972825917d269bcc956f422
  commits: 1
  started_at: 2026-09-29T18:15:11.702655708Z
  ended_at: 2026-09-29T18:30:31.856758263Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T18:15:11.702655708Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T18:30:31.856758263Z
observed_paths:
- internal/parse/parse.go
- internal/parse/qualifiedmodsub_test.go
- internal/parse/shape.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/use.go
---

A loaded module's subs are callable by their qualified names whether or not it exports them, and the parser knew only the exported bare names. After `use overload;`, `overload::constant 'integer' => sub {...}` is lib/overload.pm's `sub constant`, and it refused as an unknown word.

Measured with perl 5.42: `use overload; BEGIN { overload::constant "integer" => sub { return shift } }` compiles, and Deparse writes `overload::constant('integer', sub {...})`.

perl.git t/op/overload_integer.t:20, t/re/overload.t.

## Acceptance criteria

- [ ] A loaded module's declared subs are known by qualified name (`go test ./internal/parse/ -run '^TestQualifiedModuleSub$' -count=1 -v | grep -q '^--- PASS: TestQualifiedModuleSub'`)
