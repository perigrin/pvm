---
title: DESTROY { ... } without sub is read as a call that takes the next statement
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T03:56:54.872473231Z
updated: 2026-09-29T03:56:55.802662661Z
sessions:
- start_sha: 12eb345487d543ca4032f88c401ae4e77d397464
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T03:56:55.802662661Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T03:56:55.802662661Z
---

`DESTROY { ... }` and `AUTOLOAD { ... }` at statement start declare the sub,
exactly as the phasers declare theirs. toke.c:

    case KEY_AUTOLOAD: case KEY_DESTROY: case KEY_BEGIN: ... case KEY_END:
        if (PL_expect == XSTATE)
            return yyl_sub(aTHX_ PL_bufptr, key);

The parser reads the phasers but not these two, so the block became a
`WORD BLOCK` call that took the NEXT statement as its argument -- a wrong tree:

    DESTROY { $m .= 'x' }
    f();                        canon  DESTROY {$m .= 'x';}f();

Measured on 5.42.0:

    $ perl -MO=Deparse -e 'package Thingie; DESTROY { 1 } AUTOLOAD { 2 }'
    package Thingie;
    sub DESTROY { 1; }
    sub AUTOLOAD { 2; }

Reach: comp/hints.t, op/ref.t and op/undef.t first refuse at one.

## Acceptance criteria

- [ ] DESTROY and AUTOLOAD blocks at statement start are `sub NAME BLOCK` declarations and canon to a fixpoint (`go test ./internal/parse/ -run '^TestSpecialSubWithoutSub$' -count=1 -v | grep -q '^--- PASS: TestSpecialSubWithoutSub'`)
- [ ] NEGATIVE: `$obj->DESTROY` is still a method name (`go test ./internal/parse/ -run '^TestSpecialSubWithoutSub$' -count=1 -v | grep -q '^--- PASS: TestSpecialSubWithoutSub'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1`)
