---
title: Num-to-Int truncation is lossy under --warn-lossy-coercion, and a string used as a reference is reported
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-21T20:52:37.129152123Z
updated: 2026-10-21T20:52:37.129152123Z
---

Found by the 01a16016-46b4 review (2026-10-21): Num -> Int truncation is silent under --warn-lossy-coercion on every build (pu, 2db182e6, 692a560c). `"ab" x (20/8)` -- perl 5.42 -w is silent and uses 2; the argument's Num is truncated to the Int the repetition count wants. RFC 0001 "What a parameter type means": an argument of type A fits T when A <: T, otherwise the coercion is lossy and reported under --warn-lossy-coercion. Num is not <: Int.

Also from that review, a gap every build shares: a dualvar or string used as a reference (`$x->{k}`, `@$x`, `$x->()`, `$x->foo` -- perl dies "Can't use string as a HASH ref" / "Can't locate object method") is silent on pu, r2, head and branch.

## Steps
- [ ] Report Num -> Int truncation under --warn-lossy-coercion, silent by default
- [ ] Report a non-reference used as a reference where perl dies

## Acceptance criteria

### Positive Scenarios
- [ ] Num to Int truncation is reported under the flag (`go test ./internal/infer/ -run '^TestNumToIntLossyUnderFlag$' -count=1 -v | grep -q '^--- PASS: TestNumToIntLossyUnderFlag'`)

### Negative Scenarios
- [ ] psc check over perl's lib adds no unlisted line (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
