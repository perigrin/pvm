---
title: given, when and default are statement forms, and default takes no parens
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T16:28:50.389622276Z
updated: 2026-09-26T17:25:39.478660256Z
sessions:
- start_sha: 2dc301fdeeec2303d2bfcdc1cae708b09b190c0b
  end_sha: 9325864f86f42d125118a800b15eb0b62d95d6c3
  commits: 3
  started_at: 2026-09-26T16:30:24.025683777Z
  ended_at: 2026-09-26T17:25:39.478660256Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T16:30:24.025683777Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-26T17:25:39.478660256Z
observed_paths:
- docs/plans/2026-09-26-perl-git-t-parses.md
- internal/parse/control.go
- internal/parse/switch_test.go
- internal/parse/testdata/t1.ratchet
---

`given`/`when`/`default` refuse. Three dirty files in perl.git t/, but
`op/switch.t` is dense with them -- ~45 of the brace-context hits come from that
one file.

## Measured in isolation, 2026-09-26 at `59029a2c`

    use feature 'switch';
    my $x = 1;
    given ($x) {
        when (1) { print "one\n" }
        default  { print "other\n" }
    }

      Unknown=2

Perl accepts it on 5.42.0 at `/home/perigrin/.local/bin/perl`, warnings aside:

    perl -e 'use feature "switch"; no warnings; given (1) { when (1) {...} }'
      ->  one

## Reach is SMALL, and that is the point of doing it

Only 3 of the 620 files contain it. This is not a mass fix -- it is a cheap one
whose cost is knowable in advance, and `op/switch.t` is one of `op/`'s 196 dirty
files.

Do not inflate the estimate: the brace-context probe attributed ~45 hits to
`when`/`default`/`given`, which is a count of OCCURRENCES in a handful of files,
not of files unlocked. The plan records this distinction because presence counts
overturned three of four guesses on this goal already.

## The feature is deprecated, and that is a decision to record

`given`/`when` is deprecated in modern perl and warns under `use warnings`. It
still compiles, and perl's own suite still tests it, so a parser aiming at
perl.git t/ has to read it. Write that reasoning down in a comment rather than
leaving a reader to wonder why deprecated syntax was implemented.

Measure what perl does with each spelling before choosing a shape:

    given (EXPR) BLOCK
    when (EXPR) BLOCK
    when (LIST) BLOCK          smartmatch against a list
    default BLOCK              no parens at all
    my @l = do { given ($_) {...} }    given in expression position, in op/switch.t

`default` taking no parens is the one most likely to need its own rule -- it is
a bare word followed directly by a block, which is the `WORD BLOCK` family
`01a0d087` just landed. Check whether that machinery already covers it.

## Scope

1. `given`, `when` and `default` are statement forms. Measure whether `when`'s
   parenthesised argument is an ordinary expression or needs smartmatch shape.
2. `default BLOCK` has no parens; check whether `01a0d087`'s `WORD BLOCK ARG`
   reading already handles it before adding a rule.
3. `op/switch.t` holds `do { given ($_) {...} }` -- given in EXPRESSION position.
   Measure it; it may be out of scope and should be said so if it is.
4. Canon must re-emit all three. Run `go test ./internal/parse/ -run '^TestCanon'
   -count=1` EARLY -- canon has bitten four workers this milestone.

## Context

- paths: `internal/parse/parse.go` (`statementKeywords`),
  `internal/parse/control.go` (`parseControlFlow`), `internal/parse/canon.go`
- commands: `go test ./internal/parse/ -count=1`,
  `go test ./internal/conformance/ -count=1`
- files: `/home/perigrin/dev/perl5/t/op/switch.t` is the dense one
- related: `01a0d087` landed `WORD BLOCK ARG`, which `default BLOCK` may reuse

## Acceptance criteria

- [ ] `given`/`when`/`default` parse as statements (`go test ./internal/parse/ -run '^TestGivenWhenDefault$' -count=1 -v | grep -q '^--- PASS: TestGivenWhenDefault'`)
- [ ] Canon re-emits them and holds its fixpoint (`go test ./internal/parse/ -run '^TestCanon' -count=1 | grep -q '^ok'`)
- [ ] The decision to implement a deprecated feature is recorded in a comment (`grep -qi 'deprecated' internal/parse/control.go`)
- [ ] REGRESSION GUARD, passes today: the parse ratchets hold (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the corpus ratchet holds (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
