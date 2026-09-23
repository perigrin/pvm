---
title: Tier 04's nonassoc measurement lives in prose and no file asserts it
state: done
urgency: normal
milestone: corpus-complete
created: 2026-09-23T02:07:02.765668461Z
updated: 2026-09-23T19:38:10.610122112Z
sessions:
- start_sha: 66e6d10dfed49b493965299ae1c5d6fcc9088c25
  end_sha: 66e6d10dfed49b493965299ae1c5d6fcc9088c25
  commits: 0
  started_at: 2026-09-23T19:38:10.523402954Z
  ended_at: 2026-09-23T19:38:10.610122112Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-23T19:38:10.523402954Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-23T19:38:10.610122112Z
---

`04_operators/README.md` carries a measured passage on which of perly.y's
nonassoc levels actually reject repetition -- it found that 8 of 11 do
NOT. It is the most valuable thing the tier knows and no file asserts any
of it.

That is the header-prose-is-not-coverage trap in the tier's own
documentation, and this milestone's measurement found the same shape
twice more elsewhere. Prose is what a reader trusts; a claim that lives
only in prose is a claim the corpus does not make.

## What is true, measured under 5.42.0

    $ perl -e 'my @x = (1 .. 2 .. 3);'
    syntax error at -e line 1, near "2 .."

    $ perl -e 'print 1 <=> 2 <=> 3'
    syntax error at -e line 1, near "2 <=>"

    $ perl -e 'print 1 < 2 < 3'
    1                                  <- legal, CHAINS

    $ perl -e 'print "[", (9 < 1 < 5), "]\n"'
    []                                 <- chained, not (9<1)<5

The last is the decisive one. Non-chained, `(9<1)` is `""` and `"" < 5`
is true, so a pre-5.32 parser prints `1`. Chained, it is false. Same
source, opposite answer, no diagnostic either way -- and chaining is
UNCONDITIONAL in 5.42.0, no pragma required.

## The hazard within one level

perlop's level 13 is documented `chain/na`, and the two halves differ:

    1 == 1 == 1        legal, chains
    1 <=> 2 <=> 3      syntax error
    "a" cmp "b" cmp "c"  syntax error

A parser that models level 13 as ONE associativity class is wrong
whichever it picks. Nothing in the corpus claims this.

## Files

Two, in `04_operators`:

- The chaining claim, pinned by output: `9 < 1 < 5` is false, which
  separates a chaining parser from a left-associative one.
- The nonassoc refusals, as `--- expect parsent` files: `1 .. 2 .. 3`
  and `1 <=> 2 <=> 3` are syntax errors perl itself rejects, which is
  what a `parsent` file is for. This also moves the README's measurement
  into a place that fails when it stops being true.

The tier's README passage stays, but becomes a pointer to the files
rather than the only record.

## Acceptance criteria

- [ ] A file pins that 9 < 1 < 5 is false, which separates a chaining parser from a left-associative one (`awk '/^--- source/{s=1;next} /^--- expect/{s=0} s' conformance/04_operators/*.t | grep -qE '9 *< *1 *< *5'`)
- [ ] A parsent file records a nonassoc repetition that perl itself rejects (`grep -rlq 'expect parsent' conformance/04_operators/`)
- [ ] Every new file is validated against pinned perl 5.42.0, which for a parsent file means perl refuses it (`go test ./internal/conformance/ -run '^TestTierOperatorsPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsPerlValidated'`)
- [ ] The tier's nonassoc measurement is asserted by a file rather than only described in its README (`go test ./internal/conformance/ -run '^TestTierOperatorsNonassocMeasured$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsNonassocMeasured'`)
- [ ] The whole conformance suite is green with the new files in it (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
