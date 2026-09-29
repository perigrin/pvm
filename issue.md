---
title: "A block-form declaration takes the next line's for/if as a modifier: sub t {1} then if (...) {...} is one statement"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T01:29:36.846496384Z
updated: 2026-09-29T01:29:36.846496384Z
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
