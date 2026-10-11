---
title: "sort with a comparator orphans its list: sort $s 4,1,3 canons as sort($s);4 , 1 , 3"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-28T23:15:23.877389868Z
updated: 2026-10-11T15:17:45.716823017Z
sessions:
- start_sha: f196c8a1cfd85c12aac31986c2dfd014375fbf79
  end_sha: 0aaccdcbebea1a3bfae177cf641ba27c12f1b049
  commits: 1
  started_at: 2026-09-29T00:06:37.443647694Z
  ended_at: 2026-09-29T00:08:14.962842045Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T00:06:37.443647694Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T00:08:14.962842045Z
observed_paths:
- internal/parse/call.go
- internal/parse/canon.go
- internal/parse/parse.go
- internal/parse/sortcmp_test.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/testdata/t1.ratchet
- internal/parse/testdata/t1fidelity.ratchet
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

### Whole-suite criteria removed (2026-10-11)

perigrin, on a milestone verify that would have run for 9+ hours: "Why are
we doing 73 full test suite runs?!?" Criteria that only re-ran a whole suite
or package (`go test ./...`, `make test`, the whole conformance or parse
package) restated "make test passed", which gates every commit, while CI
gates every PR -- and `git zhi verify` re-ran each copy for every done issue.
They are removed; this issue's named tests and ratchet checks remain.

## Acceptance criteria

- [ ] sort with a comparator variable or sub name takes the list as its argument and canons to a fixpoint (`go test ./internal/parse/ -run '^TestSortComparator$' -count=1 -v | grep -q '^--- PASS: TestSortComparator'`)
- [ ] NEGATIVE: `sort map ...`, `sort values ...` and `sort keys ...` are still a builtin applied to the list (`go test ./internal/parse/ -run '^TestSortComparator$' -count=1 -v | grep -q '^--- PASS: TestSortComparator'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
