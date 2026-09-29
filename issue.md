---
title: "A parenthesised comparison joins the chain beside it: ($a == $b) == 0 becomes $a == $b == 0"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T06:29:44.028699987Z
updated: 2026-09-29T06:39:32.261320743Z
sessions:
- start_sha: 1c3d665b4ed12f5ee63d275633b45474baf71c1f
  end_sha: a1440c2f33201ee9ce8a8b103237425c43880b35
  commits: 1
  started_at: 2026-09-29T06:29:53.220417898Z
  ended_at: 2026-09-29T06:39:32.261320743Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T06:29:53.220417898Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T06:39:32.261320743Z
observed_paths:
- internal/parse/canon.go
- internal/parse/chain.go
- internal/parse/expr_test.go
- internal/parse/t2_test.go
- internal/parse/testdata/perlgitt.ratchet
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
