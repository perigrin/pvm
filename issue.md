---
title: "A leading :: names main, and 84 files spell a call that way"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T16:40:48.477005613Z
updated: 2026-09-26T16:40:48.477005613Z
---

A call written with a leading `::` -- `::ok(1)`, the `main::` shorthand --
refuses. The explicit `main::ok(1)` parses clean, so the defect is the elided
package name.

## Measured 2026-09-26 at `2dc301fd`

Each program is the whole file:

    ::ok(1);            Unknown=1   trailing_tokens, span "::ok(1);"
    ::ok 1;             Unknown=1   trailing_tokens, span "::ok 1;"
    my $r = ::f();      Unknown=1   not_a_term "::"

    main::ok(1);        Unknown=0   <- the explicit form is fine

Valid perl 5.42.0:

    /home/perigrin/.local/bin/perl -e 'sub main::ok { print "ok\n" } ::ok(1);'  -> ok

Note it refuses in BOTH positions, and with two different refusal codes: as a
statement it is `trailing_tokens`, as a term it is `not_a_term`. A fix has to
teach the name lexer that `::` may lead, which covers both sites.

## Reach

Appears in 84 of perl.git t/'s 620 files, 76 of them dirty -- the largest of the
six causes `not_a_term` bucketing found. perl's own suite uses it heavily inside
`sub`s that need to reach the test functions past a lexically-scoped package,
e.g. `t/re/pat_rt_report.t` and `t/op/write.t`.

## Not to be confused with the cascade

`not_a_term` is 1,584 nodes but only 41 are the first Unknown in their file; the
other 1,543 are resync fallout. This construct is one of the 41-node cause
surface. See the `not_a_term` section of
`docs/plans/2026-09-26-perl-git-t-parses.md`.

## Acceptance criteria

- [ ] `::ok(1);` as a statement parses with zero Unknown (`go test ./internal/parse/ -count=1`)
- [ ] `my $r = ::f();` as a term parses with zero Unknown (`go test ./internal/parse/ -count=1`)
- [ ] `main::ok(1)` still parses, no regression (`go test ./internal/parse/ -count=1`)
- [ ] REGRESSION GUARD: the corpus ratchet holds (`go test ./internal/conformance/ -count=1`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1`)
