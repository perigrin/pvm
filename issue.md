---
title: unfiled is a legal refusal citation and nothing rejects it
state: pending
urgency: normal
milestone: corpus-covers-t2
created: 2026-09-26T02:09:54.934577593Z
updated: 2026-09-26T02:09:54.934577593Z
---

`unfiled` is a legal value of `refuses:` and nothing rejects it. Sixteen
cases reached the corpus with it. This issue makes the corpus refuse to
hold an uncited refusal, so the sixteen cannot come back.

## Why a check and not just the cleanup

The cleanup is issues 01a0db77, 01a0db78 and 01a0db79 in this milestone.
Without a check, the next case written under time pressure spells
`unfiled` again, and the corpus goes back to recording gaps that point
nowhere. The corpus already has this shape of check for the neighbouring
failure -- `staleMarker` fires when a citation stops being TRUE -- and
this is the same guarantee for a citation that was never a citation.

## The measurement that motivates it

At b0f1a9e9, sixteen cases across nine topics carried `refuses: unfiled`.
All sixteen genuinely refuse, so none was hiding a passing case; what they
hid is WHICH work fixes them. Three clusters, none obvious from the
marker: a glossary dispute (4), two real lexer gaps (2), and ten parser
refusals of which three already had an issue nobody had cited.

## What the check must and must not do

MUST reject `refuses: unfiled`, and reject it at the reader so no test has
to remember to look.

MUST NOT require the cited issue to be OPEN -- `01a0d0c7` in `v0.1` owns
that, and duplicating it here would leave one guarantee half-owned by two
issues.

MUST NOT accept an arbitrary non-empty string. `refuses: yes`,
`refuses: TODO` and `refuses: see above` are the same defect wearing
different words. The shape to require is an issue id, which the corpus
already spells consistently: `01a0` then hex, optionally hyphenated.

Verify the check by REMOVING a citation and watching the read fail. A
validator that cannot be made to fire is the defect this milestone is
about.

## Context

- paths: `internal/conformance/topic.go`, `internal/conformance/topic_test.go`
- commands: `go test ./internal/conformance/ -count=1`
- entrypoints: `internal/conformance/topic.go` (`parseBehavior`, which reads the `refuses:` line)

## Acceptance criteria

- [ ] The reader rejects an uncited refusal (`go test ./internal/conformance/ -run '^TestRefusalNamesAnIssue$' -count=1 -v | grep -q '^--- PASS: TestRefusalNamesAnIssue'`)
- [ ] It rejects a placeholder, not only the literal word unfiled (`go test ./internal/conformance/ -run '^TestRefusalRejectsAPlaceholder$' -count=1 -v | grep -q '^--- PASS: TestRefusalRejectsAPlaceholder'`)
- [ ] No case in the corpus records an unfiled refusal (`! grep -rq 'refuses: unfiled' conformance/mdtest/`)
- [ ] The whole corpus still reads, so the new rule broke no existing case (`go test ./internal/conformance/ -run '^TestMdtestTopicsMakeTheSameClaims$' -count=1 -v | grep -q '^--- PASS: TestMdtestTopicsMakeTheSameClaims'`)
- [ ] The format spec states the rule (`grep -qE 'unfiled|names an issue' conformance/mdtest/FORMAT.md`)
- [ ] The suite is green (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
