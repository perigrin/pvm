---
title: The croak extraction runner, split by perl -c because the filename lies
state: pending
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c35e-1b37-7304-ba39-65a6ef57d4e0
created: 2026-09-21T09:49:43.940902431Z
updated: 2026-09-21T10:00:07.165502514Z
---

The negatives -- files that must NOT parse -- are the one category not
authored. `t/lib/croak/`'s cases are extracted by a runner, so they carry
none of the five-artifact per-file cost. That is a structural difference in
the work, not a faster rate.

## The split is measured, and the filename lies

Measured over 340 cases, independently of an earlier assessment whose
numbers did not reproduce:

    201   compile-fail      perl -c rejects. These are the negatives.
     64   runtime-only      perl -c ACCEPTS. croak at run time.
     75   exit 0            neither. Not negatives at all.

NAME count is 251, not the 254 an earlier pass reported.

**Trusting the directory name would assert must-not-parse on 139 cases perl
parses.** A parser that correctly parses a runtime-croak file would be
marked wrong. The `perl -c` split is not a refinement of the extraction, it
is what makes extraction sound at all.

## Acceptance

- [ ] A runtime-croak case (perl -c accepts, dies at run time) is not emitted as a must-not-parse negative (`go test ./internal/conformance/ -run '^TestRuntimeCroakIsNotANegative$' -count=1 -v | grep -q '^--- PASS: TestRuntimeCroakIsNotANegative'`)
- [ ] The three-way split is computed by running perl -c rather than read from a checked-in list that goes stale (`go test ./internal/conformance/ -run '^TestCroakSplitIsMeasured$' -count=1 -v | grep -q '^--- PASS: TestCroakSplitIsMeasured'`)
- [ ] Drift from the measured 201/64/75 is reported rather than silently absorbed (`go test ./internal/conformance/ -run '^TestCroakCountsReported$' -count=1 -v | grep -q '^--- PASS: TestCroakCountsReported'`)
- [ ] Extraction cites source by filename and pinned revision and copies no file into the repo (`go test ./internal/conformance/ -run '^TestCroakExtractionCitesRatherThanCopies$' -count=1 -v | grep -q '^--- PASS: TestCroakExtractionCitesRatherThanCopies'`)
