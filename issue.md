---
title: "`overload::constant ...` refuses: a loaded module's subs are not known by qualified name"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T18:15:10.503591614Z
updated: 2026-09-29T18:15:11.702655708Z
sessions:
- start_sha: 7bc912c2b1f6571957c11606f1ca1b8dfbec6e67
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T18:15:11.702655708Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T18:15:11.702655708Z
---

A loaded module's subs are callable by their qualified names whether or not it exports them, and the parser knew only the exported bare names. After `use overload;`, `overload::constant 'integer' => sub {...}` is lib/overload.pm's `sub constant`, and it refused as an unknown word.

Measured with perl 5.42: `use overload; BEGIN { overload::constant "integer" => sub { return shift } }` compiles, and Deparse writes `overload::constant('integer', sub {...})`.

perl.git t/op/overload_integer.t:20, t/re/overload.t.

## Acceptance criteria

- [ ] A loaded module's declared subs are known by qualified name (`go test ./internal/parse/ -run '^TestQualifiedModuleSub$' -count=1 -v | grep -q '^--- PASS: TestQualifiedModuleSub'`)
