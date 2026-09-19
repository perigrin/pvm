---
title: Three ratchets and none of them runs in CI
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-17T20:22:39.30654244Z
updated: 2026-09-19T07:19:59.708068341Z
sessions:
- start_sha: 0db016fd18cba84ae70dcfb32f3453a5771bd27d
  end_sha: ""
  commits: 0
  started_at: 2026-09-19T07:11:37.042871786Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T07:11:37.042871786Z
---

Three ratchets exist and none of them runs in CI. Each skips when its corpus is absent, `ci.yml` sets no corpus variable, so `go test ./...` passes with all three silently skipped.

## The measurement

`.github/workflows/ci.yml:41` runs `go test ./... -count=1` and sets no
corpus environment. What that leaves:

    internal/lexer         TestLexerRatchet        skips: no PERL5_CORPUS
                           TestLexDotTGoldenStream skips: same
    internal/parse         TestParseRatchet        skips: no T1_CORPUS
                           TestParsedFilesRoundTrip skips: same
    internal/parseoracle   TestRatchetCorpus       skips: needs -parseoracle.corpus

Verified by reading the skip sites: `internal/lexer/ratchet_test.go:27,33`,
`internal/parse/ratchet_test.go` (`t1Files`), and
`internal/parseoracle/ratchet_test.go:412`.

The skips are correct behaviour. A contributor without a perl5 checkout and a
PerlOnJava checkout should not see failures, and `t.Skipf` is how Go says so.
The defect is that nothing anywhere supplies those corpora to a machine that
would fail loudly.

## Why this matters more than it looks

The M1 gate (01a0a70f) states the principle itself, about a different metric:
"WRONG = 0 is trivially satisfiable by a parser that answers nothing — every
M0 lesson about guards that cannot fail applies."

A ratchet that skips is a guard that cannot fail. `TestParseRatchet` was
mutation-checked three ways when it landed at 4bc971ec and it catches
movement in either direction — on a machine with the corpus. On CI it reports
success without reading a single file.

This is not hypothetical. This session attempted three parser fixes, each
verifiably correct against toke.c or against running perl:

    subscript shape        shipped      exactly neutral
    dereference leaf       held         408 -> 406 clean, 4944 -> 4986
    right-paren lookahead  reverted     408 -> 406 clean, 4944 -> 4972

Two locally correct changes made the corpus worse. They were caught by
hand-rolled throwaway sweeps because no ratchet existed yet. The ratchet
exists now and would catch them — but only where it runs.

## There is already a working pattern

`.github/workflows/parse-fidelity.yml` fetches a corpus properly and has been
doing it since 2026-09-14. Its shape, at `:200-227`:

- a pin file (`internal/parseoracle/testdata/corpus.pin`) recording the exact
  revision, with its own ABOUTME explaining why: "a measurement taken across
  a skew in any field is an anecdote, not a result"
- `actions/cache` keyed on that revision
- `git clone --filter=blob:none --no-checkout` then `checkout --detach`, a
  partial clone rather than a shallow one, because the pin is a specific
  commit rather than a branch tip
- an explicit verification step that `rev-parse HEAD` equals the pin, failing
  with "a cached corpus went stale" rather than measuring the wrong tree

That workflow already proves every hard part. This issue is applying it to a
second corpus (PerlOnJava's `unit/`) and wiring the result into a job that
runs on every push rather than on a schedule.

## Scope

Decide where these run. `parse-fidelity.yml` is a separate workflow for a
reason — the oracle sweep takes 8 minutes — and a T1 parse sweep takes about
1.4 seconds for 986 files, so it belongs in `ci.yml` rather than beside the
oracle.

Not in scope: changing any skip to a failure. A contributor without the
corpora must still get a green local run; the corpus-present path is what
must be made to exist somewhere, not the corpus-absent path made to fail.

Also worth deciding, and cheap: whether `ci.yml` should fail when a ratchet
SKIPS rather than merely when it fails. `go test` reports skips in its
output; a CI step that greps for `--- SKIP` among the ratchet tests would
catch the next time a corpus path silently stops resolving. Without that,
this whole class of defect can recur without anyone noticing.

## Acceptance

- [x] A pin file records the PerlOnJava revision the T1 baseline was measured against (`test -f internal/parse/testdata/t1corpus.pin`)
- [x] CI fetches that corpus at the pinned revision and verifies HEAD matches before measuring (`grep -q 'rev-parse HEAD' .github/workflows/ci.yml`)
- [x] `TestParseRatchet` runs rather than skips in CI (`grep -qE 'T1_CORPUS' .github/workflows/ci.yml`)
- [x] `TestLexerRatchet` runs rather than skips in CI (`grep -qE 'PERL5_CORPUS' .github/workflows/ci.yml`)
- [x] A skipped ratchet fails the build rather than passing quietly (`grep -qE 'SKIP' .github/workflows/ci.yml`)
- [x] The corpus fetch is cached on the pin, so a green run does not re-clone (`grep -q 'actions/cache' .github/workflows/ci.yml`)
- [x] A contributor with no corpus still gets a green `go test ./...` locally (`go test ./internal/parse/ ./internal/lexer/ -count=1 2>&1 | grep -qv FAIL`)
- [x] The whole suite is green (`go test ./internal/... 2>&1 | grep -qv FAIL`)

The last two are the negative scenarios and they pull against each other on
purpose: the corpus must be required where it exists and optional where it
does not. A fix that makes the tests fail without a corpus has traded one
broken state for another.

Found by a paad:alignment pass on 4bc971ec, which is also where
`TestParseRatchet` and its 986-file baseline landed.

Spec: §7.3.5. Plan: M1 metrics.


## Done, 2026-09-19 (c3e65b4f)

    full suite     green, 0 FAIL across the repo
    new job        ratchets (corpus-backed), on every push
    new pin        internal/parse/testdata/t1corpus.pin

All eight acceptance criteria verified by running each command as written,
including the two negative scenarios that pull against each other.

## Reproduced first

    $ T1_CORPUS=/nonexistent PERL5_CORPUS=/nonexistent \
        go test ./internal/parse/ ./internal/lexer/
    ok      tamarou.com/pvm/internal/parse   0.048s
    ok      tamarou.com/pvm/internal/lexer   0.015s

Green, having read no corpus file at all. That is what every CI run had been
reporting.

## The skip check is the point

`go test` reports a skip and exits 0, so the job greps the log for each
corpus-backed ratchet BY NAME and fails on `--- SKIP:` -- or on a test that
neither passed nor skipped, which means it did not run at all. Named rather
than a blanket SKIP grep: other tests in these packages legitimately skip and
a blanket check would fail on those.

Verified in both directions before committing. With the corpora, all four
report PASS; with the corpora absent, all four are detected as SKIP and the
step exits 1.

## One new pin, not two

The lexer's corpus is perl5's t/, and internal/parseoracle/testdata/corpus.pin
ALREADY pins it at the revision this project measures against -- the local
checkout is at 94e5086, exactly what that pin records. A second file holding
the same fact is a second file to forget to update, so only T1 gained a pin.

The T1 pin carries a revision and no interpreter, and says why: the oracle
needs one because it RUNS perl to get an answer to compare against, while
this ratchet reads the .t files as text.

## Why parseoracle-pin.sh was not reused

It requires interpreter, revision and threads, and has ten tests guarding
that strictness. Loosening it to accept a revision-only pin would weaken the
oracle's guard to save four lines of shell. The workflow reads the revision
inline instead.

## Placement

A new job rather than a change to the matrix: the corpora are identical on
every OS, so fetching them three times measures nothing extra. Not in
parse-fidelity.yml, because that sweep runs perl over 620 files and takes
minutes, while a T1 parse sweep is ~1.4s for 986 files and belongs on every
push. Windows and macOS keep the plain suite with the ratchets skipping,
which is correct.

## The guard has its own guard

TestCIRunsTheRatchets asserts the workflow as TEXT -- both corpora supplied,
both pins read, the checkout verified, and every ratchet NAMED in the skip
check. A ratchet added later and left out of that list would skip in CI
unnoticed, which is this issue happening again.

## Worth recording

This session's previous issue (01a0b721) is the case for this one. The lexer
ratchet caught a regression that scored BETTER on node count while swallowing
10,124 bytes -- on my machine, with the corpus present. On CI as it stood, it
would have passed silently.
