---
title: "One pragma, two answers: the same bitwise expression is 10 or 8"
state: pending
urgency: normal
milestone: corpus-complete
blocked_by:
- 01a0cc04-3303-7103-8769-9edca646b56e
created: 2026-09-23T02:07:02.458719721Z
updated: 2026-09-23T02:36:47.973078328Z
---

The same four bytes compute two different answers depending on a line
earlier in the file, with no warning and no error. Measured under perl
5.42.0:

    $ perl -e 'no warnings; print "12" & "10"'
    10

    $ perl -e 'use v5.28; no warnings; print "12" & "10"'
    8

Without the `bitwise` feature -- which `use v5.28` and later enable --
`&` `|` `^` `~` are POLYMORPHIC: string operands get a string bitwise
operation, so `"12" & "10"` ANDs the characters and yields `"10"`. With
the feature they always treat operands as numbers, and `12 & 10` is `8`.

Two valid programs. No diagnostic either way.

## Why this is the sharpest file in the milestone

The corpus has found two constructs whose parse depends on a pragma --
`say` without its feature is a method call, `isa` without its feature is
a filehandle print. Both were worth files because the same bytes mean
different things.

This is the same shape and worse in one respect: those two FAIL, loudly
or at runtime. This one SUCCEEDS with a different answer. A parser can be
wrong here and every test that does not pin the pragma will agree with
it.

## The other half, measured

The dotted spellings are hard-gated rather than silent:

    $ perl -e 'print "AB" &. "ab"'
    syntax error at -e line 1, near "&."

    $ perl -e 'use feature "bitwise"; no warnings; print "AB" &. "ab"'
    AB

So level 14/15 hides BOTH gating kinds at once: the dotted forms are
syntax errors without the feature, and the undotted forms silently change
meaning with it. One issue, because they are the same pragma and the pair
is the claim.

## Files

Two, and they must be a pair to say anything: one file under the pragma
and one without, with the same operands, pinned to different output. The
dotted spelling gets its `parsent` half in the ungated file.

Placement is tier 04, after the bitwise operators exist to be gated.

## Prerequisite

Depends on the bitwise/shift issue: a pragma pair cannot claim operators
the corpus has not introduced.

## Acceptance criteria

- [ ] A corpus file pins the ungated string-bitwise answer and another pins the numeric answer under the pragma (`awk '/^--- source/{s=1;next} /^--- expect/{s=0} s' conformance/04_operators/*.t | grep -q 'use v5.28'`)
- [ ] Both files are validated against pinned perl 5.42.0, which is what makes the pair a measurement (`go test ./internal/conformance/ -run '^TestTierOperatorsPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsPerlValidated'`)
- [ ] The dependency lint passes over the tier (`go test ./internal/conformance/ -run '^TestTierOperatorsLint$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsLint'`)
- [ ] The whole conformance suite is green with the new files in it (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
