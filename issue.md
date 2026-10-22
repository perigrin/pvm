---
title: "Final-lens residue of core-module-types: goto past an exit, BEGIN past an exit, a sub named guard, rehoist cost"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-22T17:12:18.057513935Z
updated: 2026-10-22T17:12:18.057513935Z
---

From the final quality lens over the narrow round (b76b14ef..d4786743, 2026-10-22; scratchpad/review6-final-core-module-types.md, probes scratchpad/fl6/). perigrin's ruling: one narrow round, then the final lens, then close -- these are follow-ups. None of the 2149 corpus probes regressed; each was found by a probe written for the review.

- N1 (false positive, new in the narrow round): code past an exit reached by `goto LABEL` has its writes undone -- flow.go:366-369 exprStmt calls Unreached, undo in symbols.go:331-338 and ExitScope. fl6/probes-edge/G6.pl `{ goto FALLBACK if ...; die; FALLBACK: $conf = {}; } bless $conf` and G7.pl: perl ok, pu..base silent, head reports. Fix: end the unreached region at the first labelled statement after the exit.
- N2 (false positive, new): a BEGIN block past an exit loses its writes; Hoist covers named subs only (infer.go:341,:379; symbols.go:346-361,395-404). fl6/probes-logic/b1.pl. Fix: wrap compile-time phaser blocks in Hoist.
- N3 (minor, new): markLoopDeref (flow.go:539) recognises a guard scope by name, so a user sub named `guard` counts as one (fl6/probes-sec/sgguard/p.pl). Fix: a guard flag on Scope.
- N4 (performance, new): rehoist (symbols.go:395-404) costs hoisted subs x unreached exits; a crafted 188 KB file goes 1.7s on b76b14ef to 29.5s (fl6/probes-sec/db/p.pl).
- O1 (older): Hoist does not protect a named sub's writes inside an if body or ternary arm that always exits (flow.go:214-223 branch; the ternary undo in ops.go); h1.pl, t2.pl all builds; r5.pl, t1.pl regressed in 7201d0dd.
- O2 (recorded trade-off): under --warn-possible-undef, `print uniq(@r)` with an undef in @r is silent where perl warns (01a163bb-c4ba's exemption, matching 2db182e6 and 692a560c); write it into RFC 0001 beside possible-undef.
- Suggestions: calls.go:569 `listed` and callReturnsList (:243-245) share one isList; calls.go:148-161 copies the Selected loop.

## Steps
- [ ] Fix N1-N3 with a test each, measured against perl 5.42
- [ ] Bound rehoist's cost (N4)
- [ ] Record O2 in RFC 0001; fix or re-file O1

## Acceptance criteria

### Positive Scenarios
- [ ] A label after an exit is reachable (`go test ./internal/infer/ -run '^TestLabelAfterExitIsReached$' -count=1 -v | grep -q '^--- PASS: TestLabelAfterExitIsReached'`)
- [ ] A BEGIN past an exit keeps its writes (`go test ./internal/infer/ -run '^TestBeginAfterExitKeepsWrites$' -count=1 -v | grep -q '^--- PASS: TestBeginAfterExitKeepsWrites'`)

### Negative Scenarios
- [ ] psc check over perl's lib adds no unlisted line (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
