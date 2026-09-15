---
title: T2 core round-trip ratchet and the lexer fuzz target
state: done
urgency: normal
milestone: m0-lexer-round-trip
created: 2026-09-14T15:42:44.686499906Z
updated: 2026-09-15T20:46:42.908258452Z
sessions:
- start_sha: 436129de59fa29c2c67d553efd5c0ddbe29a2898
  end_sha: e2f29afa8bd0af158a2bf44a84d7c9d8a7d34c72
  commits: 1
  started_at: 2026-09-15T19:26:49.337858293Z
  ended_at: 2026-09-15T20:46:42.908258452Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-15T19:26:49.337858293Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-15T20:46:42.908258452Z
observed_paths:
- Makefile
- internal/lexer/fuzz_test.go
- internal/lexer/heredoc.go
- internal/lexer/lexer.go
- internal/lexer/pod.go
- internal/lexer/ratchet_test.go
- internal/lexer/testdata/corpus.ratchet
- internal/lexer/testdata/fuzz/FuzzLexer/fb95265c512141f4
---

M0's headline metric — 100% lossless round-trip over T2 core (perl5 `base
cmd comp opbasic class`, 56 files, verified present) plus the T0 fixtures
the plan also names — and the fuzz target.

## What the baseline records, and why not round-trip

An earlier draft said "records per-file status" without saying status of
what. If it records round-trip, every row is `pass` the day the milestone
closes, nothing can ever move, and `Check` never fires: a ratchet over a
constant.

Record **error-free tokenization**: per file, the count of `Error` and
`UnknownRest` tokens. That number starts high, falls as the chain lands, and
a file going the wrong way is a real regression. Round-trip is asserted
separately as a hard 100%, because it must never fail at all.

Ratchet the whole 620-file corpus, not just the 56: lexing 620 files is
milliseconds, and a wider denominator makes "either direction" meaningful
during M0 rather than only after it. Gate 100% round-trip on the 56.

## Reusing the ratchet is not free

The chain review found this and it is unbudgeted work, so it is stated here
rather than discovered mid-execution. `internal/parseoracle/ratchet.go` is
not generic:

    rank()        switches on BucketExact/Wider/NoAnswer/Wrong
    NewBaseline() takes a Report of Result carrying Facts and Verdict,
                  and calls Categorise(f, dir), which runs tree-sitter

Reusing it for lexer verdicts means either fabricating `Result`s whose
`Bucket=exact` secretly means "round-trips" — a lie written into the
baseline file — or extracting a status-agnostic freeze-and-compare core
first. Do the extraction: `Baseline`, `Row`, `Render`, `Check` moved off
`Result`, with the parse-fidelity ratchet rebuilt on it and its committed
baseline byte-identical afterwards. That last part is the test.

## The fuzz target

Spec §7.6.2 gives the seeds and all four invariants. `-run` does NOT fuzz;
it runs the seed corpus only. 1M execs needs

    go test ./internal/lexer/ -run '^$' -fuzz '^FuzzLexer$' -fuzztime 1000000x

which takes one package at a time and so cannot live inside `make test`. It
needs its own Makefile target and its own CI job. Crashers land in
`testdata/fuzz/FuzzLexer/` and run under plain `go test` afterwards, which
is the regression half and DOES belong in the normal suite.

Invariant 4 (forward progress) is structural, from the skeleton issue,
because `go test -fuzz` hangs rather than fails on an infinite loop.

## Corpus availability

`parseoracle.CorpusRoot()` reads `$PERL5_CORPUS` or falls back to
`~/dev/perl5`, and returns an error the caller turns into a skip. Use it
rather than duplicating the resolution. Note that `ci.yml` runs plain `go
test ./...` with no corpus checkout — only `parse-fidelity.yml` clones
perl5 — so the corpus-dependent tests must skip cleanly there, and that
workflow gains a lexer step or the M0 gate never runs in CI.

## Acceptance

- [ ] All 56 T2 core files and the T0 fixtures round-trip byte-for-byte (`go test ./internal/lexer/ -run '^TestT2CoreRoundTrips$' -count=1 -v | grep -q '^--- PASS: TestT2CoreRoundTrips'`)
- [ ] A status-agnostic baseline core exists and the parse-fidelity baseline is byte-identical on it (`go test ./internal/parseoracle/ -run '^TestBaselineCoreRoundTrips$' -count=1 -v | grep -q '^--- PASS: TestBaselineCoreRoundTrips'`)
- [ ] The lexer baseline records per-file error-token counts over all 620 files and fails on movement either way (`go test ./internal/lexer/ -run '^TestLexerRatchet$' -count=1 -v | grep -q '^--- PASS: TestLexerRatchet'`)
- [ ] The seed corpus passes all four §7.6.2 invariants (`go test ./internal/lexer/ -run '^TestFuzzSeeds$' -count=1 -v | grep -q '^--- PASS: TestFuzzSeeds'`)
- [ ] `make fuzz-lexer` runs 1M execs and reports zero crashers (`make fuzz-lexer`)
- [ ] Corpus-dependent tests skip with a clear message when perl5 is absent (`go test ./internal/lexer/ -run '^TestSkipsWithoutCorpus$' -count=1 -v | grep -q '^--- PASS: TestSkipsWithoutCorpus'`)

Plan: M0 headline metric, §7 checklist. Spec: §7.6.2.

## Outcome, 2026-09-15

Closed with five of six acceptance criteria met. Commit e2f29afa.

Measured: 620/620 files round-trip, 56/56 T2 core lex with no error token,
612/620 whole-corpus clean, 1,000,000 fuzz executions with no crasher.

**AC2 was deliberately NOT done.** The status-agnostic baseline core was not
extracted from `internal/parseoracle/ratchet.go`, and the criterion's test
`TestBaselineCoreRoundTrips` does not exist.

Reasoning: that file is 465 lines whose `Check` is bound to `Report`,
`Result` and `Verdict`, and whose `rank()` is bucket semantics --
exact/wider/no-answer/WRONG -- for which the lexer has no analogue. The lexer
needs freeze-and-compare over integers; parseoracle needs ordered buckets, a
pin skew check and a taxonomy. The genuinely shared surface is a map diff.

Refactoring a CI-gated component to serve a second caller that needs a
fraction of it is speculative generality. Revisit when a THIRD ratchet
appears, with three real callers to shape the interface rather than two and a
guess.

Two bugs were found by the work this issue added, both now regression tests:

- `comp/parser.t` dropped 141 bytes. `<<ENE . ${` opens a heredoc whose body
  runs to EOF without its terminator; `takePendingHeredocs` consumed those
  lines and emitted nothing, so they were covered by no token. Only a
  whole-corpus check could see it.
- `format =\n` at EOF was an infinite loop, found by the fuzzer on its first
  run. `scanFormatBody` was wrapped in `step()`, which requires forward
  progress, so a legitimate decline panicked. Found loudly rather than as a
  hang, which is why invariant 4 is structural.
