---
title: A call in a scalar slot is selected in scalar context, and possible-undef forcing does not depend on the slot's list type
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-22T14:56:33.210791633Z
updated: 2026-10-22T14:56:33.210791633Z
---

From the review of 01a163bb-c4ba (2026-10-22), outside the narrow round by perigrin's ruling:

- A list-returning multi passed to a scalar slot is still typed List, so --warn-possible-undef forces it: `lc(shuffle(@r))`, `lc(Dumper(1))`, `lc(uniq(@r))`, `uc(Dumper(1))` (scratchpad/rv6/p/b27, b28, c08, c10) report "expects Str, got List" on 7201d0dd and later; perl, 2db182e6 and 692a560c are clean. Probable root: a call in a scalar slot is not selected in scalar context. Default mode is clean.
- Inconsistent forcing: print's and unlink's `List[Str]` slots force a possible-undef argument; join's, sprintf's and printf's plain `List` slots never do, so `join(",", max(@e))`, `sprintf("%s", max(@e))`, `printf("%s\n", max(@e))` are silent though perl warns (b14, b18, b19, b32). Predates the milestone's last round.
- Recorded trade-off: `floor($o->one, 2)` (f106) -- perl dies on arity; 2db182e6 and 692a560c report it; since 4587 a method call may return a list, so its arity is not checked.

## Steps
- [ ] Select a call in a scalar slot in scalar context
- [ ] Make --warn-possible-undef's forcing independent of whether the slot's declared list type admits Undef

## Acceptance criteria

### Positive Scenarios
- [ ] A list-returning call in a scalar slot is not possible-undef (`go test ./internal/infer/ -run '^TestListCallInScalarSlotNotPossibleUndef$' -count=1 -v | grep -q '^--- PASS: TestListCallInScalarSlotNotPossibleUndef'`)

### Negative Scenarios
- [ ] psc check over perl's lib adds no unlisted line (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
