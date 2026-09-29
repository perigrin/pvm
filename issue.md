---
title: "A block-form declaration takes the next line's for/if as a modifier: sub t {1} then if (...) {...} is one statement"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T01:29:36.846496384Z
updated: 2026-09-29T01:54:17.672270229Z
sessions:
- start_sha: 92e674eb3b0bbe192651d845f0206f37ed6f80eb
  end_sha: 56a58656f2b657ee6be44be7c94f3b4dc0dbca2a
  commits: 1
  started_at: 2026-09-29T01:45:28.952809945Z
  ended_at: 2026-09-29T01:54:17.672270229Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T01:45:28.952809945Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T01:54:17.672270229Z
observed_paths:
- internal/parse/control.go
- internal/parse/subthenstmt_test.go
- internal/parse/t2_test.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/testdata/t1.ratchet
---

A declaration that ends in its own block does not end its statement, so the
control-flow statement on the next line is read as a MODIFIER on it:

    sub t { 1 }
    if ($x) { 2 }            one statement, Unknown=0, WRONG

    sub test { ... }
    for my $b (0, 1) { ... } nested loops fall apart, Unknown > 0

perl reads two siblings. Measured on 5.42.0:

    $ perl -MO=Deparse -e 'sub t { 1 } if ($x) { 2 }'
    sub t { 1; }
    if ($x) { '???'; }

applyModifier refuses after a `;` and after a Call ending in a block, but not
after a Declaration ending in one -- a named sub with a body, `package P {...}`,
or a lexical sub. For `if` and `while` the orphaned block then reads as a bare
block, so the tree is wrong at zero Unknowns and only canon sees it.

Found bisecting comp/utf.t for 01a0eab6-fc60, whose fix moved the damage from
one span to two; the cause was this, sitting after `sub test { ... }`.

## Acceptance criteria

- [ ] A sub, package block or lexical sub followed by for/if/while/foreach is two statements (`go test ./internal/parse/ -run '^TestBlockDeclarationTakesNoModifier$' -count=1 -v | grep -q '^--- PASS: TestBlockDeclarationTakesNoModifier'`)
- [ ] REGRESSION GUARD: `$x = sub { 1 } if $y` keeps its modifier (`go test ./internal/parse/ -run '^TestBlockDeclarationTakesNoModifier$' -count=1 -v | grep -q '^--- PASS: TestBlockDeclarationTakesNoModifier'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1`)
