---
title: "`*y` opens a transliteration: a quote-op word after a glob star is its name"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T20:27:35.370188078Z
updated: 2026-09-29T20:27:35.370188078Z
---

`*x = *y;` lexes `y; ...` as a transliteration, so everything after the glob star becomes an unterminated quote. After a glob's `*`, toke.c reads the name with scan_ident, and a quote-operator word there is the glob's name. A `*` that multiplies is not a glob star: `$b*s/a/b/` still substitutes.

Measured with perl 5.42:

    $ perl -MO=Deparse -e '*x = *y; *s = *tr; my $r = \*q; $a = $b*s/a/b/;'
    *x = *y;
    *s = *tr;
    my $r = \*q;
    $a = $b * s/a/b/;

perl.git t/op/gv.t:1132.

## Acceptance criteria

- [ ] A quote-operator word after a glob star is a name (`go test ./internal/lexer/ -run '^TestGlobQuoteOpName$' -count=1 -v | grep -q '^--- PASS: TestGlobQuoteOpName'`)
