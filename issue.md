---
title: "An empty slot inside brackets refuses: `[\"x\",, \"y\"]`"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T15:23:27.484454088Z
updated: 2026-09-29T15:23:27.484454088Z
---

An empty slot inside `[...]` or `{...}` refuses: `["x",, "y"]`. parseParenList drops a separator with no element after it, and parseBracketed does not.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'my $a = ["x",, "y"]; my $h = {a => 1,, b => 2}; my $c = [1, => 2];'
    my $a = ['x', 'y'];
    my $h = {'a', 1, 'b', 2};
    my $c = [1, 2];

perl.git t/porting/bench.t:274-282, `"--bisect=Ir",,`.

## Acceptance criteria

- [ ] An empty slot in brackets is dropped, and canon is a fixpoint (`go test ./internal/parse/ -run '^TestBracketEmptySlot$' -count=1 -v | grep -q '^--- PASS: TestBracketEmptySlot'`)
