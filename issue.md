---
title: "A relational comparison under an equality one refuses: $a < $b == $c is valid perl"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T06:14:05.911471897Z
updated: 2026-09-29T06:14:05.911471897Z
---

The comparison parser rejects any comparison whose left operand is already a comparison of a different class. perly.y's yyerror productions reject only a second operator at the *same* level (termrelop followed by a relop, termeqop followed by an eqop). Relop binds tighter than eqop, so `$a < $b == $c` is `($a < $b) == $c`.

Measured with perl 5.42:

    $ perl -MO=Deparse,-p -e '$a < $b == $c; $a == $b < $c;'
    (($a < $b) == $c);
    ($a == ($b < $c));
    $ perl -e '$a <=> $b == $c'
    syntax error at -e line 1, near "$b =="

The first refusal in perl.git t/cmd/mod.t is `(-(-$x) < 0) == ($x < 0)`. The refusal fixture in refusal_test.go pinned the wrong behavior.

## Acceptance criteria

- [ ] A relop under an eqop parses and two eqop-level operators still refuse (`go test ./internal/parse/ -run '^TestComparisonAcrossLevels$' -count=1 -v | grep -q '^--- PASS: TestComparisonAcrossLevels'`)
- [ ] REGRESSION GUARD: every refusal site keeps a code (`go test ./internal/parse/ -run '^TestEveryUnknownSiteHasACode$' -count=1 -v | grep -q '^--- PASS: TestEveryUnknownSiteHasACode'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
