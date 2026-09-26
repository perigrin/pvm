---
title: The perl.git t/ goal is gated by no ratchet, so it can regress silently
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T18:12:35.095313178Z
updated: 2026-09-26T21:10:58.31586339Z
sessions:
- start_sha: 2257e08ee261e0a0303f9201d0210d81f034b326
  end_sha: ""
  commits: 0
  started_at: 2026-09-26T21:10:58.31586339Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T21:10:58.31586339Z
---

The perl.git t/ goal is gated by no ratchet. `parse.Parse` takes no loader, so
every existing baseline is blind to the resolution the goal depends on, and the
620-file number can regress silently.

Found by the worker on `01a0de92`, which moved 185 to 278 clean while all six
ratchet files stayed BYTE-IDENTICAL.

## Measured

    internal/parse/parse.go:377   func Parse(src []byte) *Node

No path, no loader. `01a0de92` added `ParseFileFrom(path, root)` and
`ParseWithLoader(src, loader)`, and nothing outside its own test calls either.

So when that change took perl.git t/ from 185 clean to 278 and 7,490 Unknown
nodes to 4,223 -- a 43.6% reduction -- these all stayed unchanged, and none of
them COULD have moved:

    internal/conformance/testdata/corpus.ratchet     216 of 216
    internal/parse/testdata/t1.ratchet               986 files, 554 clean
    internal/parse/testdata/t1canon.ratchet          721 clean
    internal/parse/testdata/t1fidelity.ratchet       369 clean
    internal/parse/t2_test.go shortfall              unchanged
    internal/parse/use_test.go TestClassCorpusRatchet pin 21

My brief for that issue asserted T2 "should move a lot", since T2 core IS 56 of
this goal's 620 files. That was unreachable: T2 measures the loader-less parser.

## Why this is a defect and not a fact about scope

The 620-file figure is the goal's only measure of progress, and it lives in
nothing but a throwaway probe. Every number in
`docs/plans/2026-09-26-perl-git-t-parses.md` was produced by a `zz*_test.go`
that its author then deleted, by instruction. So:

- A change that regresses perl.git t/ parsing passes the whole suite.
- The next measurement of the goal starts by rewriting the probe, and two
  probes written a day apart are not guaranteed to count the same way. This
  session already produced one such disagreement: a `grep -c 'parses: no'`
  that counted prose gave 5 where the block parser gave 3.

Three instruments in this repo have now been found reporting a number nobody
checked -- the corpus ratchet header (fixed, `TestRatchetHeaderAgreesWithItsBody`),
`t1canon` recording 46 canon-broken files as its baseline (fixed by `01a0dde4`),
and `lintOps` reading Concise's rendered text (`01a0de11`). This is the fourth,
and it is the one measuring the active goal.

## Scope

1. A ratchet over perl.git t/, keyed per file, holding the Unknown count --
   the same shape as `t1.ratchet`. It must call the LOADER-AWARE entry point,
   or it measures the wrong thing and duplicates T2.
2. It must SKIP cleanly when `/home/perigrin/dev/perl5` is absent, the way
   `t2Files` does via `PERL5_CORPUS` -- a ratchet that fails on a machine
   without the checkout is worse than none. Read `t2_test.go:24` for the
   pattern.
3. Its header must agree with its body, per `01a0ddfb`: derived on write and
   CHECKED on read. The corpus ratchet's header could say 999 cases and pass
   until that was fixed today.
4. Decide whether it belongs in `internal/parse` beside t1/t2 or in its own
   package. T2 already reads perl.git t/ from `internal/parse`, so beside it is
   the obvious answer, but the loader requirement may argue otherwise.
5. Record the baseline as measured, not as remembered. At the time of filing:
   620 files, 278 clean, 4,223 Unknown nodes.

## Context

- paths: `internal/parse/parse.go:377` (`Parse`), `internal/parse/module.go`
  (`ParseFileFrom`, `ParseWithLoader`), `internal/parse/t2_test.go:24`
  (`t2Files`, the skip pattern), `internal/parse/ratchet_test.go` (the t1 shape)
- commands: `go test ./internal/parse/ -count=1`
- corpus: `/home/perigrin/dev/perl5/t`, 620 `.t` files walked recursively
- related: `01a0ddfb` (the three parse ratchet headers are unchecked),
  `01a0de11` (lintOps sees names, not flags), `01a0dde4` (t1canon recorded the
  symptom as its baseline)

## Acceptance criteria

- [ ] A ratchet measures perl.git t/ through the loader-aware entry point (`go test ./internal/parse/ -run 'PerlGitT' -count=1 -v | grep -q '^--- PASS'`)
- [ ] It skips rather than fails when the checkout is absent (`PERL5_CORPUS=/nonexistent go test ./internal/parse/ -run 'PerlGitT' -count=1 | grep -q '^ok'`)
- [ ] Its header is checked against its body (`go test ./internal/parse/ -run 'PerlGitT' -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the parse suite holds (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the corpus stays complete (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
