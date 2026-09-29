---
title: "`$$r[0]` is a deref of `$r[0]`: a bare deref sigil swallows the subscript"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T11:34:53.432939093Z
updated: 2026-09-29T11:34:54.628395795Z
sessions:
- start_sha: 58ec17c604eea4eb0da7275b0880828d474bb3f9
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T11:34:54.628395795Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T11:34:54.628395795Z
---

A bare deref sigil swallows the subscript after its variable: `$$r[0]` parses as a deref of `$r[0]`, and `@$r{"a","b"}` as a deref of `$r{"a","b"}`. perl applies the subscript to the DEREFERENCE. bpDeref is 300, below the subscripts' 320, contradicting its own comment ("binds tighter than every infix operator ... its operand is the braced expression or the single variable and nothing more"). The tree is wrong at Unknown=0; canon happened to write the same bytes back, which hid it until canon kept a deref's braces.

Measured with perl 5.42:

    $ perl -e 'my $r=[5,6]; print $$r[1], " ", join(",", @$r[0,1]), "\n"'
    6 5,6
    $ perl -MO=Deparse -e 'my $y = $$r[0]; my @x = @$r[1,2]; my $z = $$h{k};'
    my $y = $r->[0];
    my(@x) = @$r[1, 2];
    my $z = $h->{'k'};

sites.go's hash-slice rule was written against the wrong shape (a Unary `@` over a brace Index) and moves with it.

## Acceptance criteria

- [ ] A bare deref takes only its variable; the subscript applies to the deref (`go test ./internal/parse/ -run '^TestBareDerefTakesTheVariable$' -count=1 -v | grep -q '^--- PASS: TestBareDerefTakesTheVariable'`)
- [ ] REGRESSION GUARD: a slice through a reference is still a decided hash site (`go test ./internal/parse/ -run '^TestDecidableHashStillCommits$' -count=1 -v | grep -q '^--- PASS: TestDecidableHashStillCommits'`)
