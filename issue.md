---
title: A user sub that returns one value does not flatten into an enclosing call
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-21T21:54:58.395549694Z
updated: 2026-10-21T21:54:58.395549694Z
---

From the 01a16016-4587 review (2026-10-21): a call to a user sub whose every return yields one value flattens anyway, so the arity check of the enclosing call is skipped. `sub one { return 7 } fmod(one())`, `sub one($) { return $_[0] } fmod(one(7))`, the signatures form, and `fmod(one(), 1, 2)` -- perl dies "Usage: POSIX::fmod(x, y)"; pu, head and fix/4587 are silent; 2db182e6 reported them only because it flattened nothing (which gave it false positives elsewhere). A scalar inferred ReturnType is no proof of one value (`wantarray ? (7,3) : 7`).

## Steps
- [ ] Infer a user sub's return count: one value when every return and the final expression yield exactly one value and no path consults wantarray
- [ ] A call to such a sub does not flatten

## Acceptance criteria

### Positive Scenarios
- [ ] A single-return user sub keeps the enclosing arity check (`go test ./internal/infer/ -run '^TestSingleReturnSubKeepsArity$' -count=1 -v | grep -q '^--- PASS: TestSingleReturnSubKeepsArity'`)

### Negative Scenarios
- [ ] A wantarray sub still flattens (`go test ./internal/infer/ -run '^TestWantarraySubFlattens$' -count=1 -v | grep -q '^--- PASS: TestWantarraySubFlattens'`)
