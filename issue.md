---
title: Count oracle verdicts per marker site, with a whole-file no-answer at every site
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocks:
- 01a16602-5da6-79c8-aea5-e13d8f14c679
created: 2026-10-22T23:46:25.524606434Z
updated: 2026-10-23T01:08:01.985745542Z
---

## Context

- paths: `internal/parseoracle/verdict.go`, `internal/parseoracle/findings.go`, `internal/parseoracle/run.go`, `internal/parseoracle/sites_test.go` (new)
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
- [ ] RED/GREEN: write TestSubjectDeclineCountsEverySiteNoAnswer. A verdict where the SUBJECT (psc) declines the whole file (`subject.Declined`, or `!subject.OK`) counts every one of perl's sites as no-answer. A narrower rule keyed only on perl's "no optree" must fail it.
- [ ] RED/GREEN: write TestPerlDeclinedFileWithSitesCountsNoAnswer. A file perl declines while sites are still counted for it (the op/magic.t shape: 46 sites, perl-declined) counts every site as no-answer. It is never dropped from the denominator.
- [ ] RED: write TestSubjectErrorIsWholeFileNoAnswer. Split the report's `Err` into oracle-side (perl, shim or cache failure) and subject-side (askSubject failed: psc crashed or timed out, run.go:342-345). A subject-side error counts every perl site as no-answer. An oracle-side error stays excluded. Run it and watch it fail.
- [ ] GREEN: implement the split in run.go/verdict.go. Run it and watch it pass.
- [ ] The grader decides what counts as exact, so this change gets a review against perl before merge, like a parser fix. The reviewer checks each fixture's shape against a real corpus file where one exists (op/magic.t, comp/final_line_num.t, a run/switch* readline no-answer), and writes NEW fixtures for shapes the corpus lacks.
- [ ] Run `make test`; the existing per-file ratchet (TestRatchetCorpus's baseline.txt) must not move. If it moves, stop and explain why before re-baselining.
- [ ] Commit.

## Acceptance criteria

### Positive Scenarios

- [ ] A marker perl did not answer yields one no-answer finding per line (`go test ./internal/parseoracle/ -run '^TestNoAnswerFindingsArePerLine$' -count=1 -v | grep -q '^--- PASS: TestNoAnswerFindingsArePerLine'`)
- [ ] A whole-file no-answer counts every marker site in the file as no-answer (`go test ./internal/parseoracle/ -run '^TestWholeFileNoAnswerCountsEverySite$' -count=1 -v | grep -q '^--- PASS: TestWholeFileNoAnswerCountsEverySite'`)
- [ ] A file psc declines whole counts every perl site as no-answer (`go test ./internal/parseoracle/ -run '^TestSubjectDeclineCountsEverySiteNoAnswer$' -count=1 -v | grep -q '^--- PASS: TestSubjectDeclineCountsEverySiteNoAnswer'`)
- [ ] A perl-declined file that still carries sites counts them all as no-answer (`go test ./internal/parseoracle/ -run '^TestPerlDeclinedFileWithSitesCountsNoAnswer$' -count=1 -v | grep -q '^--- PASS: TestPerlDeclinedFileWithSitesCountsNoAnswer'`)
- [ ] A subject-side error counts as whole-file no-answer, and only an oracle-side error is excluded (`go test ./internal/parseoracle/ -run '^TestSubjectErrorIsWholeFileNoAnswer$' -count=1 -v | grep -q '^--- PASS: TestSubjectErrorIsWholeFileNoAnswer'`)
- [ ] Per-marker bucket counts sum to the file's fact sites (`go test ./internal/parseoracle/ -run '^TestSiteBucketsSumToFactSites$' -count=1 -v | grep -q '^--- PASS: TestSiteBucketsSumToFactSites'`)

### Negative Scenarios

- [ ] A whole-file no-answer never counts any of its sites as exact, even when the file has no findings at all (`go test ./internal/parseoracle/ -run '^TestWholeFileNoAnswerNeverCountsExact$' -count=1 -v | grep -q '^--- PASS: TestWholeFileNoAnswerNeverCountsExact'`)
- [ ] A whole-file no-answer with zero marker sites adds zero sites to every bucket, not one phantom site (`go test ./internal/parseoracle/ -run '^TestWholeFileNoAnswerWithNoSitesAddsNothing$' -count=1 -v | grep -q '^--- PASS: TestWholeFileNoAnswerWithNoSitesAddsNothing'`)
- [ ] A no-answer for one marker leaves the file's other, answered markers graded as exact/wider/WRONG, not swept into no-answer (`go test ./internal/parseoracle/ -run '^TestMarkerNoAnswerDoesNotSpreadToOtherMarkers$' -count=1 -v | grep -q '^--- PASS: TestMarkerNoAnswerDoesNotSpreadToOtherMarkers'`)
- [ ] A no-answer marker with no fact sites in the file emits no findings, so it cannot drive exact below zero (`go test ./internal/parseoracle/ -run '^TestNoAnswerWithoutSitesEmitsNoFindings$' -count=1 -v | grep -q '^--- PASS: TestNoAnswerWithoutSitesEmitsNoFindings'`)
- [ ] Findings that outnumber a marker's fact sites are an error from the counting helper, never a negative exact count (`go test ./internal/parseoracle/ -run '^TestFindingsExceedingSitesIsAnError$' -count=1 -v | grep -q '^--- PASS: TestFindingsExceedingSitesIsAnError'`)
- [ ] Two findings on the same line and marker (for example wider and no-answer) count that site once, in one bucket (`go test ./internal/parseoracle/ -run '^TestSiteCountedInExactlyOneBucket$' -count=1 -v | grep -q '^--- PASS: TestSiteCountedInExactlyOneBucket'`)
- [ ] Excluded files and oracle-side error files contribute no sites to any bucket, so they cannot pad the denominator as exact (`go test ./internal/parseoracle/ -run '^TestExcludedAndErrorFilesContributeNoSites$' -count=1 -v | grep -q '^--- PASS: TestExcludedAndErrorFilesContributeNoSites'`)
