---
title: "A label before a block's closing brace refuses: `sub f { ...; _: }`"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T16:51:34.072457106Z
updated: 2026-09-29T16:51:34.072457106Z
---

A label with no statement after it before a block's `}` refuses: `sub MODIFY_CODE_ATTRIBUTES { ...; _: }`. The parser takes `LABEL: ;` as a labelled empty statement, but after a label it only looks for a `;`; a `}` there went to the expression path and refused, and the block lost its closer.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'sub f { $x = 1; _: } sub g { L: }'
    sub f {
        $x = 1;
        _: ;

perl.git t/op/attrs.t:306 and 324, t/uni/attrs.t, t/comp/parser.t:594 `eval{$1=eval{a:}}`.

## Acceptance criteria

- [ ] A label before a block's `}` labels an empty statement and round-trips (`go test ./internal/parse/ -run '^TestLabelBeforeClose$' -count=1 -v | grep -q '^--- PASS: TestLabelBeforeClose'`)
