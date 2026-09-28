---
title: "The main stash spelled with an empty package name: $::{n} and %::"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T16:41:03.74207397Z
updated: 2026-09-28T22:18:05.479845413Z
sessions:
- start_sha: 9b7b13111f748942d278e8fa83872337288084f6
  end_sha: 01e775966608237f2380b27d2433f94283e5146b
  commits: 1
  started_at: 2026-09-28T21:55:54.99011159Z
  ended_at: 2026-09-28T22:18:05.479845413Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-28T21:55:54.99011159Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-28T22:18:05.479845413Z
observed_paths:
- internal/parse/control.go
- internal/parse/formany_test.go
- internal/parse/term.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/testdata/t1.ratchet
---

The `main` stash spelled with an empty package name -- `$::{name}`, `%::` --
refuses. The same access on a NAMED package parses clean, so the defect is the
empty name before `::` and nothing else.

## Measured 2026-09-26 at `2dc301fd`

Each program is the whole file:

    my $r = $::{foo};      Unknown=1   not_a_term ":"
    my @k = keys %::;      Unknown=1   not_a_term ":;"
    my %h = %::;           Unknown=1   not_a_term ":;"

    my $r = $Foo::{bar};   Unknown=0   <- a NAMED package is fine
    my @k = keys %Foo::;   Unknown=0   <- a NAMED package is fine
    *Foo:: = \%Bar::;      Unknown=0
    my $x = \%Foo::;       Unknown=0
    my $r = $Foo::bar;     Unknown=0

Both refusing forms are valid perl 5.42.0:

    /home/perigrin/.local/bin/perl -e 'my $r = $::{foo}; print "ok\n"'   -> ok
    /home/perigrin/.local/bin/perl -e 'my @k = keys %::; print "ok\n"'   -> ok

## Reach

Appears in 58 of perl.git t/'s 620 files, 56 of them dirty. It is the second
largest of the six causes `not_a_term` bucketing found, and `t/op/gv.t`,
`t/op/lex.t`, `t/op/stash.t` and `t/op/package_aliases.t` are dense with it.

## Not to be confused with the cascade

`not_a_term` is 1,584 nodes but only 41 of them are the first Unknown in their
file; the other 1,543 are resync fallout. This construct is one of the 41-node
cause surface. See the `not_a_term` section of
`docs/plans/2026-09-26-perl-git-t-parses.md`.

## Acceptance criteria

- [ ] `$::{name}` parses with zero Unknown (`go test ./internal/parse/ -count=1`)
- [ ] `%::` and `keys %::` parse with zero Unknown (`go test ./internal/parse/ -count=1`)
- [ ] A named package's stash still parses, no regression (`go test ./internal/parse/ -count=1`)
- [ ] REGRESSION GUARD: the corpus ratchet holds (`go test ./internal/conformance/ -count=1`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1`)
