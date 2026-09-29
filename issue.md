---
title: "`&m('x')` opens a match: a quote-operator word after & is not read as a name"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T15:46:32.849199889Z
updated: 2026-09-29T15:46:32.849199889Z
---

After the `&` sigil a quote-operator word opens a quote instead of naming a sub: `&m('amper')` lexes as a match whose delimiter is `(`, and likewise for q, qq, qr, qw, qx, s, tr and y.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'sub m {1} sub s {2} sub y {3} my @r = (&m("a"), &s("b"), &y("c"), &qw("d"));'
    my(@r) = (&m('a'), &s('b'), &y('c'), &qw('d'));

perl.git t/comp/opsubs.t:122-179.

## Acceptance criteria

- [ ] A quote-operator word after `&` is a name (`go test ./internal/lexer/ -run '^TestAmpQuoteOpName$' -count=1 -v | grep -q '^--- PASS: TestAmpQuoteOpName'`)
