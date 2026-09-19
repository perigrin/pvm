---
title: The canon sweep checks stability, not fidelity; quotient the parenless call
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-19T12:24:07.603930506Z
updated: 2026-09-19T12:44:41.147317627Z
sessions:
- start_sha: 657d0caf0f47c00e90171f5eed8589e133a9d877
  end_sha: ""
  commits: 0
  started_at: 2026-09-19T12:44:41.147317627Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T12:44:41.147317627Z
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

## Acceptance

- [ ] The corpus sweep fails on a tree that is not a parse of its source (`go test ./internal/parse/ -run '^TestCanonFidelityCatchesMisgrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonFidelityCatchesMisgrouping'`)
- [ ] The parenless-call equivalence is the ONLY one quotiented, named and tested (`go test ./internal/parse/ -run '^TestCanonQuotientIsNarrow$' -count=1 -v | grep -q '^--- PASS: TestCanonQuotientIsNarrow'`)
- [ ] The fidelity result is ratcheted per file (`go test ./internal/parse/ -run '^TestCanonFidelityRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonFidelityRatchet'`)

Blocked by nothing; the emitter is in place as of 4bb5af57. Worth doing after
01a0b983 and 01a0b99b, since both move the denominator.

Plan: M3 metrics. Spec: §7.5 (c').
