---
title: "`&{; $c }` refuses: the braces after `&` never hold statements"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T20:37:38.054345818Z
updated: 2026-09-29T20:37:38.915025349Z
sessions:
- start_sha: ceece6a8325cbfa3675d34f61eef2a7b50f11497
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T20:37:38.915025349Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T20:37:38.915025349Z
---

The braces after `&` are read only as one expression, so `&{; $c }(1)` and `goto &{ 1; $coderef; }` refuse. perly.y gives `&` the same `indirob: ... | block` operand that `$` and `@` have, and derefHoldsStatements already makes that call for them.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'prototype \&{; do { f: sub ($) {} } }; &{; $c }(1);'
    prototype \&{do { f: sub ($) { } };};
    &{$c;}(1);

perl.git t/op/goto.t:674 and t/op/goto-sub.t:383.

## Acceptance criteria

- [ ] `&{; ...}` reads a block and canons to a fixpoint (`go test ./internal/parse/ -run '^TestAmpDerefBlock$' -count=1 -v | grep -q '^--- PASS: TestAmpDerefBlock'`)
