---
title: "`&{; $c }` refuses: the braces after `&` never hold statements"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T20:37:38.054345818Z
updated: 2026-09-29T20:43:43.047799421Z
sessions:
- start_sha: ceece6a8325cbfa3675d34f61eef2a7b50f11497
  end_sha: 1dfb3e29036b6b71055c50b1f716583f58b70fdc
  commits: 1
  started_at: 2026-09-29T20:37:38.915025349Z
  ended_at: 2026-09-29T20:43:43.047799421Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T20:37:38.915025349Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T20:43:43.047799421Z
observed_paths:
- internal/parse/ampblock_test.go
- internal/parse/term.go
- internal/parse/testdata/perlgitt.ratchet
---

The braces after `&` are read only as one expression, so `&{; $c }(1)` and `goto &{ 1; $coderef; }` refuse. perly.y gives `&` the same `indirob: ... | block` operand that `$` and `@` have, and derefHoldsStatements already makes that call for them.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'prototype \&{; do { f: sub ($) {} } }; &{; $c }(1);'
    prototype \&{do { f: sub ($) { } };};
    &{$c;}(1);

perl.git t/op/goto.t:674 and t/op/goto-sub.t:383.

## Acceptance criteria

- [ ] `&{; ...}` reads a block and canons to a fixpoint (`go test ./internal/parse/ -run '^TestAmpDerefBlock$' -count=1 -v | grep -q '^--- PASS: TestAmpDerefBlock'`)
