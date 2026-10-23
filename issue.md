---
title: Pin the file counts of the class and parse ratchets, so a shrunken corpus fails
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-23T15:07:41.66143421Z
updated: 2026-10-23T15:07:41.66143421Z
---

## Context

Found by the df09 worker (m3-conformance-corpus) while checking scope item 4.

- **TestClassCorpusRatchet** (internal/parse/use_test.go:306) calls `t.Skip` when the corpus holds no t/class files. With the corpus present and t/class empty, it passes on nothing. That is T2's absent-versus-empty distinction again, with no pin on the file count.
- **The parse ratchets** (checkRatchet, internal/parse/ratchet_test.go) have the same regeneration gap. `-parse.update-ratchet` run over a smaller corpus writes the loss into the baseline. TestParseRatchet also silently `continue`s past files whose `ParseFileFrom` errors, so a file that becomes unreadable drops out of the count.
- t1.ratchet's "986 files" header is not checked on read. Related: 01a0ddfb.

## Steps

- [ ] Pin each corpus ratchet's file count as a floor that fails on loss and on unrecorded growth, as TestCorpusHoldsAtLeastItsFloor does for the conformance corpus
- [ ] A file that fails to parse or read counts against the ratchet; it is never skipped

## Acceptance criteria

### Positive Scenarios
- [ ] An empty t/class with the corpus present fails (`go test ./internal/parse/ -run '^TestClassCorpusRatchetFailsOnEmptyClassDir$' -count=1 -v | grep -q '^--- PASS: TestClassCorpusRatchetFailsOnEmptyClassDir'`)
- [ ] A smaller T1 corpus fails instead of re-baselining silently (`go test ./internal/parse/ -run '^TestParseRatchetHoldsItsFileCount$' -count=1 -v | grep -q '^--- PASS: TestParseRatchetHoldsItsFileCount'`)
