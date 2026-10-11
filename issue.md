---
title: "A lowercase bareword filehandle refuses: print foo \"x\" is a handle in perl whatever the case"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T04:11:40.019954771Z
updated: 2026-10-11T15:17:48.980893266Z
sessions:
- start_sha: 5d2714d168d308b5530a60aaa642e86854743afc
  end_sha: f2a1da6dc9413d6e9dc27b8333880632eda6b19d
  commits: 1
  started_at: 2026-09-29T04:11:40.598658473Z
  ended_at: 2026-09-29T04:17:25.594151482Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T04:11:40.598658473Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T04:17:25.594151482Z
observed_paths:
- internal/parse/barehandle_test.go
- internal/parse/call.go
- internal/parse/testdata/perlgitt.ratchet
---

A lowercase bareword in print's handle slot refuses, because the parser took
only ALL-CAPS words as handles:

    print foo "x";            Unknown=1
    print tmp "tvi925\n";     Unknown=1        op/while.t:13

perl's rule is not a naming convention. Measured on 5.42.0 with -MO=Deparse:

    print foo "x";                 print foo 'x';      a handle
    print foo 1;                   print foo 1;        a handle
    sub foo {1} print foo "x";     print foo('x');     a declared sub: a call
    print length "x";              print 1;            a builtin

So a word followed by a term is a handle unless it names a builtin or a sub
already declared or imported -- tables the parser has.

Reach: io/print.t, op/while.t, op/svleak.t (`printf uNopened 42`), and
op/eval.t and run/script.t, which `print try '...'` before `sub try` is
declared.

### Whole-suite criteria removed (2026-10-11)

perigrin, on a milestone verify that would have run for 9+ hours: "Why are
we doing 73 full test suite runs?!?" Criteria that only re-ran a whole suite
or package (`go test ./...`, `make test`, the whole conformance or parse
package) restated "make test passed", which gates every commit, while CI
gates every PR -- and `git zhi verify` re-ran each copy for every done issue.
They are removed; this issue's named tests and ratchet checks remain.

## Acceptance criteria

- [ ] A bareword in the handle slot is a handle unless it is a builtin or a known sub, in any case (`go test ./internal/parse/ -run '^TestBarewordFilehandleCase$' -count=1 -v | grep -q '^--- PASS: TestBarewordFilehandleCase'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
