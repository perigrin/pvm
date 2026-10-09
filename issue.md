---
title: The parse oracle's default subject is internal/parse, and its corpus baseline is regenerated against it
state: in-progress
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0ad3f-e929-77a8-9b00-2a80b06ec42c
blocks:
- 01a0a8cd-8fcc-72a8-9ba1-c916eee9992d
- 01a0bae2-e4fe-74e9-82cb-562c3fa8adf6
- 01a11eb2-21e9-75a7-ae19-ce6f6a358be2
created: 2026-09-19T00:00:00Z
updated: 2026-10-09T03:26:34.183689416Z
---

## Context

- paths: `internal/parseoracle/run.go` (`askSubject`), `internal/parseoracle/ratchet_test.go` (`TestRatchetCorpus`, `TestCorpusBaselineIsIntact`), `internal/parseoracle/goparser_test.go` (`goSubject`), `internal/parse/cmd/subject`, `internal/parseoracle/testdata/ratchet/baseline.txt`, `docs/specs/perl-parser/00-findings.md` §0.11
- commands: `make test` (green before every commit; never `--no-verify`); the ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet|TestT2CoreParses)$' -count=1`, `go test ./internal/conformance/ -run '^TestCorpusRatchet$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; perl.git t/ at `~/dev/perl5/t` (the checkout is not built, so `perl -c` on a file that requires `Config` fails for that reason alone)

Goal (perigrin, 2026-10-09): finish M1 with the parse oracle measuring internal/parse.

## Re-measured, 2026-10-09 at 287e85b8

The premise this issue was filed on has moved. `run.go:381` no longer calls `parser.New()`: e9ab4733 ("the oracle stops parsing Perl") removed the in-process tree-sitter fallback, and `askSubject` now REQUIRES a subject (`RunOptions.Subject`), erroring without one. The tests hand it `goSubject` (`internal/parse/cmd/subject`, internal/parse as a subprocess). But:

    TestRatchetCorpus         skipped unless -parseoracle.corpus is passed
    TestGoParserWrongIsZero   skipped unless PARSEORACLE_SHIM is set
    baseline.txt              last regenerated 20deb6de, 2026-09-14 -- BEFORE e9ab4733,
                              so its verdicts are tree-sitter's
    baseline statuses         405 exact, 174 no-answer, 12 wider, 26 excluded, 2 WRONG,
                              1 error, 1 interpreter, 1 revision, 1 threads
    WRONG rows                comp/our.t (8), op/universal.t (16) -- tree-sitter defects

So nothing in make test has ever measured internal/parse against the corpus, and the per-file ratchet (which already fails in both directions, `exact` included) records the old parser.

## Steps

- [ ] Write failing test `TestDefaultSubjectIsInternalParse`: a `RunOptions` with no `Subject` measures internal/parse, run as a subprocess (`internal/parse/cmd/subject`), so the oracle still parses no Perl in-process (e9ab4733's point) while its default is the M1 parser
- [ ] Write failing test `TestSubjectAnswersAllFiveMarkers` (in `internal/parse/cmd/subject`): every `SubjectFacts` marker is answered from an internal/parse tree, `KnowsCallSites` and `KnowsPrototypes` set honestly (true only where the tree carries them)
- [ ] Measure the full sweep's cost on this machine; make it run where the gates run -- in `make test` if it fits, otherwise a CI job that fails the build, and say which in findings §0.11
- [ ] Regenerate `baseline.txt` against internal/parse in the same commit; report the bucket movement in bytes and counts in the commit body; record the starting exact-match rate (files `exact` of files measured) in findings §0.11
- [ ] `make test`; commit

## Acceptance criteria

- [ ] the oracle's default subject is internal/parse (`go test ./internal/parseoracle/ -run '^TestDefaultSubjectIsInternalParse$' -count=1 -v | grep -q '^--- PASS: TestDefaultSubjectIsInternalParse'`)
- [ ] the subject answers all five markers honestly (`go test ./internal/parse/cmd/subject/ -run '^TestSubjectAnswersAllFiveMarkers$' -count=1 -v | grep -q '^--- PASS: TestSubjectAnswersAllFiveMarkers'`)
- [ ] the corpus ratchet passes against internal/parse (`go test ./internal/parseoracle/ -run '^TestRatchetCorpus$' -count=1 -v -parseoracle.corpus | grep -q '^--- PASS: TestRatchetCorpus'`)
- [ ] findings §0.11 records the parser measured and the starting exact rate (`grep -q 'measures internal/parse' docs/specs/perl-parser/00-findings.md`)
