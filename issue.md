---
title: TestExternalDocumentManagerNew hangs under load instead of failing fast
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-06T03:24:26.294062586Z
updated: 2026-10-06T03:24:26.294062586Z
---

## Context

- paths: `internal/cli/docs` (`TestExternalDocumentManagerNew`)
- commands: `make test`; `go test ./internal/cli/docs/ -run '^TestExternalDocumentManagerNew$' -count=20`

On 2026-10-06, during the type-libraries milestone (issue 01a0fd9b-4a5b's rebase gate), `TestExternalDocumentManagerNew` hung for 10 minutes at load average ~12 while other agents ran `make test` concurrently, failing the suite on the test timeout. Alone the package passed in 3.75s and the full re-run was green. A test that can hang rather than fail fast is waiting on something unbounded (a lock, a network call, a subprocess, a channel). Find what it blocks on under load; the fix must remove the unbounded wait, not raise a timeout.

## Steps

- [ ] Reproduce under parallel load (run alongside `go test ./internal/parse/`), with `-timeout 2m` and a goroutine dump on timeout; record what it blocks on
- [ ] Fix the unbounded wait; re-run under the same load 20 times with no hang; `make test`; commit

## Acceptance criteria

- [ ] the test does not hang under repeated runs (`go test ./internal/cli/docs/ -run '^TestExternalDocumentManagerNew$' -count=20 -timeout 5m`)
