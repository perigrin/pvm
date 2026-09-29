---
title: "Attributes on a list declaration refuse: `my ($a, @b) : shared = ...`"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T17:04:12.929225916Z
updated: 2026-09-29T17:04:12.929225916Z
---

A list declaration with attributes refuses: `my ($cows, @go, %bong) : teapots = qw[ jibber jabber joo ];`. The list target is parsed as one expression, where `: teapots` reads as a ternary's colon with nothing to match. A scalar declaration's attributes are read between target and initialiser; a list's are not.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'package n; sub MODIFY_SCALAR_ATTRIBUTES{} sub MODIFY_ARRAY_ATTRIBUTES{}
          sub MODIFY_HASH_ATTRIBUTES{} my ($c, @g, %b) : teapots = qw[a b c];'
    my($c, @g, %b) :teapots = ('a', 'b', 'c');

perl.git t/op/attrs.t:348 and t/uni/attrs.t:190.

## Acceptance criteria

- [ ] A list declaration takes attributes before its initialiser (`go test ./internal/parse/ -run '^TestListDeclarationAttributes$' -count=1 -v | grep -q '^--- PASS: TestListDeclarationAttributes'`)
