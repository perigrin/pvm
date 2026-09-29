---
title: "`$$r[0]` is a deref of `$r[0]`: a bare deref sigil swallows the subscript"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T11:34:53.432939093Z
updated: 2026-09-29T11:47:10.448316411Z
sessions:
- start_sha: 58ec17c604eea4eb0da7275b0880828d474bb3f9
  end_sha: 8ef76e349e30a181a72fd26c176383cbd4f31bc0
  commits: 1
  started_at: 2026-09-29T11:34:54.628395795Z
  ended_at: 2026-09-29T11:47:10.448316411Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T11:34:54.628395795Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T11:47:10.448316411Z
observed_paths:
- internal/parse/barederef_test.go
- internal/parse/call.go
- internal/parse/canon.go
- internal/parse/sites.go
- internal/parse/testdata/t1canon.ratchet
- internal/parse/testdata/t1fidelity.ratchet
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
