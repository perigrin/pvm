---
title: Tell a subject crash from a refusal in the oracle ratchet, and keep harness failures as runner errors
state: pending
urgency: normal
milestone: m3-oracle-at-scale
created: 2026-10-23T15:16:37.781056222Z
updated: 2026-10-23T15:16:37.781056222Z
---

## Context

- paths: internal/parseoracle/ratchet.go (status), internal/parseoracle/run.go (askSubject, defaultSubject), internal/parseoracle/testdata/ratchet/baseline.txt
- entrypoints: internal/parseoracle/ratchet.go:240

Out-of-scope finding (F2) from the independent review of 01a16602-3e74, on 2026-10-23.

After 3e74, a subject failure is `Result.SubjectErr`, graded as a whole-file no-answer. Two consequences:
- `status()` (ratchet.go:240) renders it as plain "no-answer". baseline.txt therefore cannot tell a psc crash or timeout from an honest refusal, and a crash that later turns into a refusal is invisible to the ratchet.
- A harness failure, such as defaultSubject() failing to build internal/parse or "no command configured", is graded as the subject's no-answer, when it should be a runner error.

## Steps

- [ ] Write a failing test: a SubjectErr file appears distinctly in the baseline row (e.g. its status says subject-error), or the ratchet fails on any SubjectErr
- [ ] Route defaultSubject() build errors and a missing subject command to Err, not SubjectErr
- [ ] make test; commit

## Acceptance criteria

### Positive Scenarios
- [ ] A subject crash is distinguishable from a refusal in the ratchet (`go test ./internal/parseoracle/ -run '^TestRatchetDistinguishesSubjectErrorFromRefusal$' -count=1 -v | grep -q '^--- PASS: TestRatchetDistinguishesSubjectErrorFromRefusal'`)
- [ ] A subject build failure is a runner error (`go test ./internal/parseoracle/ -run '^TestSubjectBuildFailureIsRunnerError$' -count=1 -v | grep -q '^--- PASS: TestSubjectBuildFailureIsRunnerError'`)
