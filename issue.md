---
title: DESTROY { ... } without sub is read as a call that takes the next statement
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T03:56:54.872473231Z
updated: 2026-10-11T15:17:48.779174684Z
sessions:
- start_sha: 12eb345487d543ca4032f88c401ae4e77d397464
  end_sha: 5d2714d168d308b5530a60aaa642e86854743afc
  commits: 1
  started_at: 2026-09-29T03:56:55.802662661Z
  ended_at: 2026-09-29T04:03:33.662666845Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T03:56:55.802662661Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T04:03:33.662666845Z
observed_paths:
- internal/parse/specialsub_test.go
- internal/parse/t2_test.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/use.go
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

### Whole-suite criteria removed (2026-10-11)

perigrin, on a milestone verify that would have run for 9+ hours: "Why are
we doing 73 full test suite runs?!?" Criteria that only re-ran a whole suite
or package (`go test ./...`, `make test`, the whole conformance or parse
package) restated "make test passed", which gates every commit, while CI
gates every PR -- and `git zhi verify` re-ran each copy for every done issue.
They are removed; this issue's named tests and ratchet checks remain.

## Acceptance criteria

- [ ] DESTROY and AUTOLOAD blocks at statement start are `sub NAME BLOCK` declarations and canon to a fixpoint (`go test ./internal/parse/ -run '^TestSpecialSubWithoutSub$' -count=1 -v | grep -q '^--- PASS: TestSpecialSubWithoutSub'`)
- [ ] NEGATIVE: `$obj->DESTROY` is still a method name (`go test ./internal/parse/ -run '^TestSpecialSubWithoutSub$' -count=1 -v | grep -q '^--- PASS: TestSpecialSubWithoutSub'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
