---
title: Count oracle verdicts per marker site, with a whole-file no-answer at every site
state: pending
urgency: normal
milestone: m3-oracle-at-scale
created: 2026-10-22T23:46:25.524606434Z
updated: 2026-10-22T23:46:25.524606434Z
---

## Context

- paths: `internal/parseoracle/verdict.go`, `internal/parseoracle/findings.go`, `internal/parseoracle/sites_test.go` (new)
- commands: `go test ./internal/parseoracle/ -run '^(TestNoAnswerFindingsArePerLine|TestWholeFileNoAnswerCountsEverySite|TestSiteBucketsSumToFactSites)$' -count=1 -v`
- entrypoints: `internal/parseoracle/verdict.go` (`decideMarker`), `internal/parseoracle/ratchet_test.go:43` (`corpusReport`)

The oracle grades files; M3's metric grades marker sites. The 2026-10-22 per-site measurement (5340 sites, 4933 exact, 400 wider, 7 no-answer, 0 WRONG) came from a scratch test that needed one scratch change. This issue makes that change properly. `decideMarker`'s `!answered` branch emits one `BucketNoAnswer` finding per line. A whole-file no-answer (`Verdict.Marker == MarkerNone`, perl "reported success but produced no optree") counts every one of the file's sites as no-answer. With both in place, a file's sites minus its findings by bucket gives exact counts, and the counting rule is pinned by small fixture tests rather than by the corpus.

## Steps

- [ ] Create `internal/parseoracle/sites_test.go`, starting with two `// ABOUTME:` lines.
- [ ] RED: write TestNoAnswerFindingsArePerLine, a fixture verdict where perl answered nothing for one marker on two lines; it expects two `BucketNoAnswer` findings, one per line. Run it and watch it fail.
- [ ] GREEN: emit per-line no-answer findings from `decideMarker`'s `!answered` branch. Run the test and watch it pass.
- [ ] RED: write TestWholeFileNoAnswerCountsEverySite, a fixture with `Verdict.Marker == MarkerNone` and three sites across two markers; every site counts as no-answer. Run it and watch it fail.
- [ ] GREEN: implement it in verdict.go/findings.go. Run the test and watch it pass.
- [ ] RED/GREEN: write TestSiteBucketsSumToFactSites. For a fixture file, exact + wider + no-answer + WRONG per marker equals `len(f.Facts.Sites[m])`. Implement the per-marker x bucket counting helper the site ratchet will call.
- [ ] Run `make test`; the existing per-file ratchet (TestRatchetCorpus's baseline.txt) must not move. If it moves, stop and explain why before re-baselining.
- [ ] Commit.

## Acceptance criteria

### Positive Scenarios

- [ ] A marker perl did not answer yields one no-answer finding per line (`go test ./internal/parseoracle/ -run '^TestNoAnswerFindingsArePerLine$' -count=1 -v | grep -q '^--- PASS: TestNoAnswerFindingsArePerLine'`)
- [ ] A whole-file no-answer counts every marker site in the file as no-answer (`go test ./internal/parseoracle/ -run '^TestWholeFileNoAnswerCountsEverySite$' -count=1 -v | grep -q '^--- PASS: TestWholeFileNoAnswerCountsEverySite'`)
- [ ] Per-marker bucket counts sum to the file's fact sites (`go test ./internal/parseoracle/ -run '^TestSiteBucketsSumToFactSites$' -count=1 -v | grep -q '^--- PASS: TestSiteBucketsSumToFactSites'`)
