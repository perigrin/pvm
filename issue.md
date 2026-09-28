---
title: "sort with a comparator orphans its list: sort $s 4,1,3 canons as sort($s);4 , 1 , 3"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-28T23:15:23.877389868Z
updated: 2026-09-28T23:15:23.877389868Z
---

`sort SUBNAME LIST` and `sort $subref LIST` read the comparator as sort's only
argument and orphan the list, at Unknown=0:

    my @b = sort $s 4,1,3;     canon  my @b = sort($s);4 , 1 , 3;     WRONG
    my @b = sort $s @h;        canon  my @b = sort($s);@h;            WRONG

Measured on 5.42.0:

    $ perl -MO=Deparse -e 'my @b = sort $s 4,1,3; @c = sort foo @h;'
    my(@b) = (sort $s 4, 1, 3);
    @c = (sort foo @h);

The bareword case needs perl's own rule, because `sort map ...`, `sort values
...` and `sort keys ...` are a builtin applied to the list, not a comparator.
Read toke.c's sort handling (the check after KEY_sort) before choosing, and
record which words it excludes.

Reach: grep finds `sort WORD|$VAR TERM` in 9 perl.git t/ files -- an upper
bound, since it includes the builtin forms. op/sort.t carries 35 Unknowns.

## Acceptance criteria

- [ ] sort with a comparator variable or sub name takes the list as its argument and canons to a fixpoint (`go test ./internal/parse/ -run '^TestSortComparator$' -count=1 -v | grep -q '^--- PASS: TestSortComparator'`)
- [ ] NEGATIVE: `sort map ...`, `sort values ...` and `sort keys ...` are still a builtin applied to the list (`go test ./internal/parse/ -run '^TestSortComparator$' -count=1 -v | grep -q '^--- PASS: TestSortComparator'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1`)
