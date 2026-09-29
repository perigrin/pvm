---
title: "A lowercase bareword filehandle refuses: print foo \"x\" is a handle in perl whatever the case"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T04:11:40.019954771Z
updated: 2026-09-29T04:11:40.019954771Z
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

## Acceptance criteria

- [ ] A bareword in the handle slot is a handle unless it is a builtin or a known sub, in any case (`go test ./internal/parse/ -run '^TestBarewordFilehandleCase$' -count=1 -v | grep -q '^--- PASS: TestBarewordFilehandleCase'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1`)
