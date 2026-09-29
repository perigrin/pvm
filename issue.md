---
title: A prototype's optional slots are not counted, so ($;$) parses as unary and drops its second argument
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T02:17:43.942577971Z
updated: 2026-09-29T02:17:43.942577971Z
---

ShapeOf counts only the MANDATORY slots of a prototype, so a prototype with
one mandatory slot and optional ones after it -- or two optional ones -- is
read as a named unary, and a parenless call loses every argument past the
first:

    sub f ($;$) {} f $a, $b;       parser: (f($a), $b)     perl: f($a, $b)

perl's line is the TOTAL slot count. Measured on 5.42.0 with
`perl -MO=Deparse,-p -e 'sub f (P) {} my ($a,$b); f $a, $b;'`:

    $       (f($a), $b)       unary
    ;$      (f($a), $b)       unary
    _       (f($a), $b)       unary
    *       (&f($a), $b)      unary
    \@      (&f($a), $b)      unary
    ;\@     (&f($a), $b)      unary
    $;$     f($a, $b)         list
    ;$$     f($a, $b)         list
    \@;$    &f($a, $b)        list

Found adding perl's interpreter-defined subs (01a0eaf2), where
`Internals::SvREADONLY` is `\[$%@];$` and `Internals::SvREADONLY $x, 1` must
take both arguments.

## Acceptance criteria

- [ ] A prototype with exactly one slot, mandatory or optional, is unary; more than one is a list (`go test ./internal/parse/ -run '^TestPrototypeArity$' -count=1 -v | grep -q '^--- PASS: TestPrototypeArity'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1`)
