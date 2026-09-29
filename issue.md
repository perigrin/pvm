---
title: "A parenthesised comparison joins the chain beside it: ($a == $b) == 0 becomes $a == $b == 0"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T06:29:44.028699987Z
updated: 2026-09-29T06:29:44.028699987Z
---

The comparison parser read a parenthesised comparison as a still-open termrelop or termeqop. So `($a == $b) == 0` extended the chain into `$a == $b == 0`, a different program at Unknown=0, and `($a cmp $b) == 0` refused as a second eqop. Parentheses end the production in perly.y; the grouped comparison is a term.

Measured with perl 5.42:

    $ perl -MO=Deparse,-p -e '($a == $b) == 0; ($a cmp $b) == 0;'
    (($a == $b) == 0);
    (($a cmp $b) == 0);
    $ perl -e 'my ($a,$b)=(1,2); print +(($a == $b) == 0) ? "y":"n", (($a == $b == 0) ? "y":"n")'
    yn

Canon must also keep the grouping. A non-chaining comparison Binary (`<=>`, `cmp`, `isa`, `~~`) cannot repeat at its level on either side, so `($a <=> $b) <=> 0` written bare is a syntax error.

perl.git t/opbasic/cmp.t refuses 4 times on this, at `ok(($raw cmp $cooked) == 0)` and its siblings.

## Acceptance criteria

- [ ] A parenthesised comparison is a term, and canon keeps its parentheses (`go test ./internal/parse/ -run '^TestParenthesisedComparisonIsATerm$' -count=1 -v | grep -q '^--- PASS: TestParenthesisedComparisonIsATerm'`)
- [ ] REGRESSION GUARD: comparisons across levels still parse (`go test ./internal/parse/ -run '^TestComparisonAcrossLevels$' -count=1 -v | grep -q '^--- PASS: TestComparisonAcrossLevels'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
