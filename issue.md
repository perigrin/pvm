---
title: for my ($k, $v) (LIST) read its variable list as the loop head, and op/for-many.t is nothing else
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-28T22:03:29.384507988Z
updated: 2026-09-28T22:11:49.155686388Z
sessions:
- start_sha: 9b7b13111f748942d278e8fa83872337288084f6
  end_sha: 01e775966608237f2380b27d2433f94283e5146b
  commits: 1
  started_at: 2026-09-28T22:03:40.114129927Z
  ended_at: 2026-09-28T22:11:49.155686388Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-28T22:03:40.114129927Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-28T22:11:49.155686388Z
observed_paths:
- internal/parse/control.go
- internal/parse/formany_test.go
- internal/parse/term.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/testdata/t1.ratchet
---

`for my ($k, $v) (LIST)` -- the multi-variable foreach of perl 5.36 -- read its
variable list as the loop HEAD, leaving the real list with nowhere to go. The
simple spellings scored Unknown=0 with a wrong tree; op/for-many.t carried 81
Unknowns, all resync fallout from the `}` that closed each loop.

Measured on 5.42.0, no feature needed:

    $ perl -e 'for my ($k, $v) (a=>1, b=>2) { print "$k=$v\n" }'
    a=1
    b=2

The same file spells forty empty slots in one variable list (`($a,,,, $b)`),
which perl accepts in ANY parenthesised list after the first element:

    $ perl -e 'my @a = (1,,,2); print scalar(@a)'     -> 2
    $ perl -e 'my @a = (,1);'                         -> syntax error

## Acceptance criteria

- [ ] The multi-variable foreach parses, canons and re-parses to its canon (`go test ./internal/parse/ -run '^TestForLoopManyVariables$' -count=1 -v | grep -q '^--- PASS: TestForLoopManyVariables'`)
- [ ] An empty list slot is dropped after an element and a leading one still refuses (`go test ./internal/parse/ -run '^TestListEmptySlot$' -count=1 -v | grep -q '^--- PASS: TestListEmptySlot'`)
- [ ] op/for-many.t is clean in the perl.git t/ ratchet (`grep -q '^0 op/for-many.t$' internal/parse/testdata/perlgitt.ratchet`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1`)
