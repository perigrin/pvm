---
title: "The runner: perl adjudicates first, and a stale refusal marker fails"
state: pending
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c35d-b59e-77f7-91a3-2a889ae3051a
- 01a0c447-3cbc-77b9-8759-0e0c7571339c
blocks:
- 01a0c35e-d2b6-7993-9059-5675239aea65
- 01a0c35f-1685-786b-abbe-88ed9ed361a3
- 01a0c35f-57cc-7a0f-83ac-53654f46f235
created: 2026-09-21T09:48:39.607140769Z
updated: 2026-09-21T14:03:36.848167414Z
---

The runner is what makes a corpus file a test. The spike has one at
`internal/conformance/run.go`, carrying a banner saying it has not been
through assess/refinement/execution. This issue removes that banner by
earning it.

## The order of adjudication, which the spike got right

perl runs FIRST. If perl's actual output disagrees with `--- expect output`,
the file reports `CORPUS BUG` and the parser is never consulted -- a file
whose expectation perl does not honour is not measuring the parser, it is
wrong. Only then the parser, then the token facts.

That order is the whole reason the corpus can claim to be derived from perl
rather than from our belief about perl.

## The skip discipline, which is the part that rots

- A file whose refusal is KNOWN skips, naming the refusal.
- A file marked as refusing that PASSES fails loudly, because a stale marker
  is how a corpus stops measuring anything.

The second half is implemented and untested; it needs a test that proves it.

## Open: the refusal citation

`refusalIssue()` returns an issue id, `selfRecorded` ("this file"), or "".
`03_leading_decimal.t` cites issue 01a0c13f-97f5. Nothing checks the cited
issue EXISTS, so a citation can outlive its issue -- which nearly happened
during a chain prune that destroyed one ref.

## Acceptance

- [ ] When perl's output disagrees with the file, the result is CORPUS BUG and the parser is never consulted (`go test ./internal/conformance/ -run '^TestPerlAdjudicatesFirst$' -count=1 -v | grep -q '^--- PASS: TestPerlAdjudicatesFirst'`)
- [ ] A file marked as refusing that actually passes fails loudly rather than skipping (`go test ./internal/conformance/ -run '^TestStaleRefusalMarkerFails$' -count=1 -v | grep -q '^--- PASS: TestStaleRefusalMarkerFails'`)
- [ ] A cited refusal issue id that does not resolve fails the file, naming the id (`go test ./internal/conformance/ -run '^TestRefusalCitationMustResolve$' -count=1 -v | grep -q '^--- PASS: TestRefusalCitationMustResolve'`)
- [ ] perl is invoked through pinned 5.42.0 rather than PATH, so MEASURED perl 5.42.0 in the header is true (`go test ./internal/conformance/ -run '^TestRunnerUsesPinnedPerl$' -count=1 -v | grep -q '^--- PASS: TestRunnerUsesPinnedPerl'`)
- [ ] A skip that names no refusal fails, so "skips limited to documented refusals" is observed rather than asserted (`go test ./internal/conformance/ -run '^TestCorpusSkipsAreDocumented$' -count=1 -v | grep -q '^--- PASS: TestCorpusSkipsAreDocumented'`)
- [ ] The spike banner is gone from run.go (`! grep -qi 'spike' internal/conformance/run.go`)
