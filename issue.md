---
title: "A glob refuses a deref operand: `*$$foo`, `*{;undef}`"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T20:03:59.176750092Z
updated: 2026-09-29T20:04:00.374772489Z
sessions:
- start_sha: e03ea6a0b167ad69aea3f745c3e2abb409fe021b
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T20:04:00.374772489Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T20:04:00.374772489Z
---

A glob `*` accepts a name, a variable or one braced expression. It refuses the two deref forms that `$` and `@` already accept: a deref'd scalar (`*$$foo`) and braces that hold statements (`*{;undef}`).

Measured with perl 5.42:

    $ perl -MO=Deparse -e '*$$foo = bless [], _121242::;'
    *$$foo = bless([], '_121242');
    $ perl -MO=Deparse -e '*{;undef} = 3'
    *{undef;} = 3;

perl.git t/op/gv.t:984 and :1020.

## Acceptance criteria

- [ ] `*$$foo` and `*{;undef}` parse and canon to a fixpoint (`go test ./internal/parse/ -run '^TestGlobDeref$' -count=1 -v | grep -q '^--- PASS: TestGlobDeref'`)
