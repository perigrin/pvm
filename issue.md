---
title: TestShebangPassthrough flakes on a 48s fixture against a 60s timeout
state: done
urgency: normal
milestone: parser-fidelity-harness
created: 2026-09-12T20:58:22.461237475Z
updated: 2026-09-20T20:10:57.76472411Z
sessions:
- start_sha: cc4a3d3a4dff59d2e6da66443724d83ccbfb75ec
  end_sha: b6e7c351f8e8cd83c9d92e404cce11020f563dac
  commits: 1
  started_at: 2026-09-20T19:58:14.472261163Z
  ended_at: 2026-09-20T20:10:57.76472411Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-20T19:58:14.472261163Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-20T20:10:57.76472411Z
observed_paths:
- internal/parseoracle/oracle.go
- internal/parseoracle/oracle_test.go
---

TestShebangPassthrough ran op/taint.t against DefaultTimeout's 60 seconds.
Measured what it costs:

    without -T   150ms     perl rejects it and compiles nothing
    with -T      32s idle, 42s under a parallel test run

This issue reported 48s. On a loaded CI runner that clears 60.

## Done, 2026-09-20 (b6e7c351)

    full suite green, 0 FAIL across the repo

- [x] The expensive half gets an explicit timeout rather than the default
- [x] DefaultTimeout's comment says what is true
- [x] The numbers carry a guard that fails when they go stale

## The default's comment was false, and that is why 60s looked generous

    // Compiling a corpus file costs tens of milliseconds; anything
    // approaching this is wedged.

op/taint.t costs 32 seconds through the probe -- three orders of magnitude
off. It now says what is true, with both halves measured:

    perl -Tc op/taint.t          0.6s    the compile alone
    the same file through here   32s     the compile plus the CHECK walk

## Where the time is, and where it is not

The cost is the probe's WALK, not the fixture. Ruled out along the way:

    the fixture itself     0.6s for a bare perl -Tc
    save_BEGINs            0.14s with or without
    the stash walk         118 entries for a bare perl

That leaves the CHECK block's CV walk, which is the oracle's real
measurement work. Not trimmed here: making it faster needs understanding why
this file in particular is expensive, which is a different piece of work
from stopping the flake.

## Why the test's timeout and not the default's

Raising the default would slow every genuinely wedged file in the corpus
sweep to accommodate one fixture that is merely slow. parse-fidelity.yml
already sets 8m for the sweep on the same reasoning. 4m here is ~7x the
loaded measurement, and a wedged child still fails -- four minutes later
instead of one.

## A guard, because the comments carry numbers

Two comments cite 32s and draw a conclusion from it. A comment carrying a
measurement nothing re-runs goes quietly wrong, which is exactly what
happened to the "tens of milliseconds" line.

TestExpensiveFixtureStillFitsItsBudget fails if the cost reaches half the
budget it justified -- the moment the reasoning stops holding. One
direction, loosely, because the absolute number moves with the machine.

## Noted while here

`shimT` reads ORACLE_SHIM, while the rest of the package's corpus tests read
PARSEORACLE_SHIM. It falls back to /tmp/oracletree/t, so these tests run
either way, but the two names are a trap for anyone setting one and
expecting both. Not fixed here -- it is unrelated to the flake and deserves
its own decision about which name wins.
