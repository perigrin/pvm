---
title: The corpus suite pins no total, so an empty corpus passes every gate
state: done
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-26T18:45:58.933910487Z
updated: 2026-10-23T15:12:03.217569328Z
sessions:
- start_sha: 054e75e6f0548ae9b596ac0d16dadd06e0d100a9
  end_sha: 62cf900f71db986393d6d645b6814afefaea26e4
  commits: 2
  started_at: 2026-10-23T14:45:39.958488044Z
  ended_at: 2026-10-23T15:12:03.217569328Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-23T14:45:39.958488044Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-23T15:12:03.217569328Z
observed_paths:
- internal/conformance/lint.go
- internal/conformance/ratchet_test.go
---

The corpus suite pins no total. An empty or unreadable `conformance/mdtest/`
reports "0 cases, 0 clean, 0 refusing" and every gate passes by vacuous truth.
T2 has exactly this guard; the corpus does not.

Found by asking the question the B::SoN session named: **has anyone made this
check fail on purpose?**

## The precedent, measured

T2 pins its denominator and I proved it fires. Building a fake corpus root with
`t/base` and `t/comp` present and ZERO `.t` files in them:

    PERL5_CORPUS=<fake> go test ./internal/parse/ -run TestT2CoreParses

    t2_test.go:88: T2 core is 0 files, want 56 (spec 7.3)
    t2_test.go:381: T2 core: 0 of 0 files clean (NaN%)
    --- FAIL: TestT2CoreParses

Note the directory EXISTS -- the `t.Skipf` at `t2_test.go:36` does not catch
this, and the `want 56` pin is what does. That is the distinction: skipping when
the corpus is absent is correct, passing when it is present and empty is not.

## What the corpus has instead

Per-topic guards only:

    croak_test.go:177   if len(cases) == 0
    topic_test.go:63    if len(cases) == 0
    topic_test.go:169   if len(cases) != 1   (a fixture, not the corpus)

Each of those protects ONE topic from being empty. Nothing asserts the corpus
has 216 cases, or any cases at all. `renderRatchet` writes
`# N cases, M clean, K refusing` from `len(keys)`, and
`TestRatchetHeaderAgreesWithItsBody` (landed today) checks the header against
the body -- so 0/0/0 is internally consistent and passes.

So: delete `conformance/mdtest/*.md` and the suite goes green. The corpus that
reached 216 of 216 today has no floor under it.

## Why this is the fifth instrument, not a new class

    corpus ratchet header   could say 999 cases and pass   FIXED today
    t1canon                 46 canon-broken files AS its baseline   FIXED (01a0dde4)
    lintOps                 reads Concise text: op names, not flags   01a0de11
    the 620-file t/ goal    gated by no ratchet at all   01a0deeb
    the corpus TOTAL        no floor; 0 cases passes   THIS

All five share one shape: the check ran, passed, and measured something adjacent
to the question. And the one property that separates a real guard from these is
whether anyone has made it FAIL on purpose. T2's `want 56` was written by
someone who did.

## Scope

1. Pin the corpus total the way T2 pins its file count -- a floor, not an exact
   match. The corpus grew 212 to 216 cases in one afternoon, so an exact pin
   would fail on the next contributed case and teach the reader to update it
   without thinking. A floor fails only on LOSS.
2. Decide the floor. T2 uses an exact `want 56` because its corpus is perl's and
   does not move; ours does, so a floor near the current 216 with a comment
   saying why it is a floor is the honest form.
3. PROVE IT FAILS before landing it. Move the topics aside, run the suite, see
   red, put them back. A guard nobody has seen fail is the thing this issue is
   about.
4. Check the same question for `TestClassCorpusRatchet` (pin 19) and the three
   parse ratchets: is there a state where each passes on nothing? `t1.ratchet`
   pins `986 files` in its header but `01a0ddfb` records that those headers are
   derived on write and skipped on read, so the pin may not be enforced.

## Context

- paths: `internal/conformance/ratchet_test.go` (`TestCorpusRatchet`,
  `TestRatchetHeaderAgreesWithItsBody`), `internal/conformance/topic_test.go:63`,
  `internal/parse/t2_test.go:88` (the model), `internal/conformance/ratchet.go`
  (`renderRatchet`, `corpusState`)
- commands: `go test ./internal/conformance/ -count=1`,
  `go test ./internal/parse/ -run TestT2CoreParses -count=1`
- related: `01a0ddfb` (parse ratchet headers unchecked on read), `01a0deeb`
  (the t/ goal has no ratchet), `01a0de11` (lintOps sees names, not flags)

## Acceptance criteria

- [ ] The corpus suite fails when the corpus is empty (`go test ./internal/conformance/ -run 'TestCorpus' -count=1 | grep -q '^ok'`)
- [ ] The floor is a FLOOR, with a comment saying why not an exact match (`grep -qiE 'floor|at least' internal/conformance/ratchet_test.go`)
- [ ] REGRESSION GUARD, passes today: the corpus stays complete at 216 (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: T2's own denominator pin still fires (`go test ./internal/parse/ -run TestT2CoreParses -count=1 | grep -q '^ok'`)
