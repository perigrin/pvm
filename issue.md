---
title: The canon sweep checks stability, not fidelity; quotient the parenless call
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-19T12:24:07.603930506Z
updated: 2026-10-09T17:21:18.379742056Z
sessions:
- start_sha: 657d0caf0f47c00e90171f5eed8589e133a9d877
  end_sha: ef0efc46cb2a0d24b82c6e69ce9d3ec7f357750e
  commits: 1
  started_at: 2026-09-19T12:44:41.147317627Z
  ended_at: 2026-09-19T12:57:06.598396655Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T12:44:41.147317627Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-19T12:57:06.598396655Z
observed_paths:
- .github/workflows/ci.yml
- internal/parse/canon.go
- internal/parse/fidelity.go
- internal/parse/fidelity_test.go
- internal/parse/pin_test.go
- internal/parse/testdata/t1canon.ratchet
- internal/parse/testdata/t1fidelity.ratchet
---

The canon corpus sweep is a STABILITY check, not a fidelity check, and the
difference is the whole point of 01a0a8d8.

`canonMatches` compares `canon(parse(S))` against
`canon(parse(canon(parse(S))))`. The source appears on neither side after the
first emission, so a tree that is not a parse of its source can still be a
fixpoint. Measured, with the wrong tree TestCanonCatchesMisgrouping builds:

    WRONG tree: once="print((1 + 2) * 3);"
                twice="print((1 + 2) * 3);"   fixpoint

## Why the literal criterion was abandoned

`tokens(canon(parse(S))) == tokens(S)` is the formulation that carries
fidelity, and no tree-faithful emitter can satisfy it. Measured over T1: 658
files disagreed and **519 of them differed by one thing** -- a paren around a
parenless call.

    print "x";      parses to   call print -> term "x"
    print("x");     parses to   call print -> term "x"

The same tree, correctly. They mean the same thing and the tree records
meaning, so canon must print one form for both.

## What to build

Quotient by EXACTLY that equivalence and no more. The comparison becomes
fidelity again while remaining satisfiable:

    tokens(canon(parse(S)))  ==  tokens(canon(parse(S)))
                                 where S is first normalised for
                                 parenless-call parens

Two routes, and the choice needs measuring rather than arguing:

1. **Normalise the source side.** Re-emit S through a paren-inserting pass
   before tokenising, so both sides agree on call spelling and nothing else.
2. **Compare tokens modulo a known rewrite.** A token-stream diff that
   accepts an inserted `(`/`)` pair immediately around a call's arguments and
   rejects every other difference.

Route 2 is narrower and does not need a second emitter; route 1 is easier to
explain. Measure both against the 519 before choosing.

## The test this has to pass

Whatever is built must FAIL on the mis-grouped tree above, in the corpus
sweep and not only in a hand-built unit test. That is the acceptance bar,
because it is the thing the current sweep cannot do.

## Acceptance criteria

- [x] The corpus sweep fails on a tree that is not a parse of its source (`go test ./internal/parse/ -run '^TestCanonFidelityCatchesMisgrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonFidelityCatchesMisgrouping'`)
- [x] The parenless-call equivalence is the ONLY one quotiented, named and tested (`go test ./internal/parse/ -run '^TestCanonQuotientIsNarrow$' -count=1 -v | grep -q '^--- PASS: TestCanonQuotientIsNarrow'`)
- [x] The fidelity result is ratcheted per file (`go test ./internal/parse/ -run '^TestCanonFidelityRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonFidelityRatchet'`)

Blocked by nothing; the emitter is in place as of 4bb5af57. Worth doing after
01a0b983 and 01a0b99b, since both move the denominator.

Plan: M3 metrics. Spec: §7.5 (c').


## Done, 2026-09-19 (ef0efc46)

    T1 fidelity   228 -> 345 clean (23.1% -> 35.0%)
    T1 canon      611 -> 649 clean (62.0% -> 65.8%)
    full suite green, 0 FAIL across the repo

All three acceptance criteria verified by running each command as written.
The ratchet was verified to fail in BOTH directions before committing.

## Route 2, as scoped

A token-stream comparison that accepts the known rewrites and rejects
everything else. No second emitter was needed.

## The quotient is two rewrites, not one

The issue scoped one. Measuring found a second that is equally unavoidable:

  1. A parenless call acquires parens -- 519 of 658 files, the measurement
     that forced 01a0a8d8 to abandon the literal criterion.
  2. A block's last statement acquires its semicolon:

         $ perl -e 'sub f { 1 } sub g { 1; } print f(), g(), "\n";'
         11

Both are spellings the tree cannot record BECAUSE they do not change
meaning, which is the test for admitting one. TestCanonQuotientIsNarrow
asserts both halves: each rewrite accepted, and a grouping difference of the
same shape rejected.

## It immediately earned its keep

Four emitter defects the stability sweep could not see, found by running
fidelity over T1:

    (a => 1)            -> (a(), 1)        92 files
    sub f { 1 } f();    -> {1;}f();       177 files
    if (1) {} else {}   -> (else {})
    for my $x (1) {}    -> for (my $x) (1)

The fat comma was the significant one: `Fat` marks the element before `=>`,
and the two commas are not interchangeable -- `(a => 1)` is the string "a"
where `(a, 1)` calls a (§4.5.4). It arrives two ways, as the flag inside a
bracketed form and as a Binary inside a call's arguments, and both autoquote.

`else`/`elsif` are nested Conditionals rather than conditions, and a
foreach's loop variable is a Declaration outside the parens. My Conditional
case had parenthesised every non-Block child.

## The gap this closes

01a0a8d8's corpus sweep measured stability and was described, honestly but
narrowly, as resting on two hand-built tests. The corpus now carries fidelity
too: 35.0% of T1 emits back what its source says, against 65.8% that merely
emits stably. That gap is the parser's remaining defects, and it is now
visible per file rather than inferred.
