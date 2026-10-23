---
title: Count oracle verdicts per marker site, with a whole-file no-answer at every site
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocks:
- 01a16602-5da6-79c8-aea5-e13d8f14c679
created: 2026-10-22T23:46:25.524606434Z
updated: 2026-10-23T07:06:07.787972442Z
---

## Context

- paths: `internal/parseoracle/compare_facts.go` (`decideMarker`), `internal/parseoracle/verdict.go`, `internal/parseoracle/findings.go`, `internal/parseoracle/run.go`, `internal/parseoracle/sites_test.go` (new)
- commands: `go test ./internal/parseoracle/ -run '^(TestNoAnswerFindingsArePerSite|TestWholeFileNoAnswerCountsEverySite|TestSiteBucketsSumToFactSites)$' -count=1 -v`
- entrypoints: `internal/parseoracle/compare_facts.go` (`decideMarker`), `internal/parseoracle/oracle.go:39` (`Sites`: one entry per marker op), `internal/parseoracle/ratchet_test.go:43` (`corpusReport`)

The oracle grades files; M3's metric grades marker sites. The 2026-10-22 per-site figures came from a scratch test, and the site ratchet re-measures them at the 5.44.0 pin. This issue commits the counting rule that test needed. `Facts.Sites` holds one entry per marker op (oracle.go:39; parse_facts.pl:151 pushes one `[line, name, marker]` per op), so two sites on one line are two entries. `decideMarker`'s `!answered` branch (compare_facts.go) emits one `BucketNoAnswer` finding per site entry, not per line. A whole-file no-answer (`Verdict.Marker == MarkerNone`) counts every one of the file's sites as no-answer, whichever side declined: perl produced no optree, perl declined (the op/magic.t shape), or the subject declined. With both in place, a file's sites minus its findings by bucket gives exact counts, and the counting rule is pinned by small fixture tests rather than by the corpus.

## Steps

- [ ] Create `internal/parseoracle/sites_test.go`, starting with two `// ABOUTME:` lines.
- [ ] RED: write TestNoAnswerFindingsArePerSite, a fixture verdict where perl answered nothing for one marker with three sites, two of them on the same line; it expects three `BucketNoAnswer` findings, one per site entry. Run it and watch it fail.
- [ ] GREEN: emit one no-answer finding per site entry from `decideMarker`'s `!answered` branch in compare_facts.go. Run the test and watch it pass.
- [ ] RED: write TestWholeFileNoAnswerCountsEverySite, a fixture with `Verdict.Marker == MarkerNone` and three sites across two markers; every site counts as no-answer. Run it and watch it fail.
- [ ] GREEN: implement it in verdict.go/findings.go. Run the test and watch it pass.
- [ ] RED/GREEN: write TestSiteBucketsSumToFactSites. For a fixture file, exact + wider + no-answer + WRONG per marker equals `len(f.Facts.Sites[m])`. Implement the per-marker x bucket counting helper the site ratchet will call.
- [ ] RED/GREEN: write TestSubjectDeclineCountsEverySiteNoAnswer. A verdict where the SUBJECT (psc) declines the whole file (`subject.Declined`, or `!subject.OK`) counts every one of perl's sites as no-answer. A narrower rule keyed only on perl's "no optree" must fail it.
- [ ] RED/GREEN: write TestPerlDeclinedFileWithSitesCountsNoAnswer. A file perl declines while sites are still counted for it (the op/magic.t shape: 46 sites, perl-declined) counts every site as no-answer. It is never dropped from the denominator.
- [ ] RED: write TestSubjectErrorIsWholeFileNoAnswer. Split the report's `Err` into oracle-side (perl, shim or cache failure) and subject-side (askSubject failed: psc crashed or timed out, run.go:342-345). A subject-side error counts every perl site as no-answer. An oracle-side error stays excluded. Run it and watch it fail.
- [ ] GREEN: implement the split in run.go/verdict.go. Run it and watch it pass.
- [ ] The grader decides what counts as exact, so this change gets a review against perl before merge, like a parser fix. The reviewer checks each fixture's shape against a real corpus file where one exists (op/magic.t, comp/final_line_num.t, a run/switch* readline no-answer), and writes NEW fixtures for shapes the corpus lacks.
- [ ] Run `make test`. It cannot show whether baseline.txt moves, because TestRatchetCorpus skips without `-parseoracle.corpus` (ratchet_test.go:369-370), so this issue makes no claim about it. The subject-error split may move a file from error to no-answer. The site-ratchet issue, which this one blocks, runs the sweep and re-baselines with the move explained.
- [ ] Commit.

## Acceptance criteria

### Positive Scenarios

- [ ] A marker perl did not answer yields one no-answer finding per site, two for two sites on one line (`go test ./internal/parseoracle/ -run '^TestNoAnswerFindingsArePerSite$' -count=1 -v | grep -q '^--- PASS: TestNoAnswerFindingsArePerSite'`)
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
- [ ] Two sites of one marker on one line produce two findings, and each site is counted once, in one bucket; a marker is never both wider and no-answer, since an unanswered marker is no-answer as a whole (`go test ./internal/parseoracle/ -run '^TestSiteCountedInExactlyOneBucket$' -count=1 -v | grep -q '^--- PASS: TestSiteCountedInExactlyOneBucket'`)
- [ ] Excluded files and oracle-side error files contribute no sites to any bucket, so they cannot pad the denominator as exact (`go test ./internal/parseoracle/ -run '^TestExcludedAndErrorFilesContributeNoSites$' -count=1 -v | grep -q '^--- PASS: TestExcludedAndErrorFilesContributeNoSites'`)
