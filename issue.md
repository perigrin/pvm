---
title: Decide whether psc warns about a redundant multi candidate
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-06T05:27:24.54116464Z
updated: 2026-10-06T05:27:24.54116464Z
---

## Context

- paths: `internal/types/select.go`, `internal/psc`
- commands: `make test`

RFC 0001 "Multi declarations" permits (does not require) psc to warn about a redundant multi candidate -- one no call could ever select because another candidate always wins. Not built by 01a0fd9b-4a5b or 01a0fd9b-4ca3; the alignment gate on 4ca3 found no issue owning it. This records the deferral.

## Steps

- [ ] Decide with perigrin whether psc should warn; if yes, write a failing test for a candidate shadowed by a more specific one in every context, implement the warning, `make test`, commit; if no, record the decision in the RFC and close

## Acceptance criteria

- [ ] the decision is recorded in RFC 0001 (`sh -c 'grep -qi "redundant" docs/rfcs/0001-type-libraries.md'`)
