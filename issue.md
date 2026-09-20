---
title: A parenless call consumes its arguments, so its statement stops falling to Unknown
state: pending
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0c13f-50ee-77c8-afe5-9c5e6ebb9638
created: 2026-09-20T23:55:53.237977507Z
updated: 2026-09-20T23:56:34.702946237Z
---

The decision issue establishes WHICH route the `default:` arm of
`internal/parse/call.go` takes for a parenless call to a callee this parser
does not know. This issue takes it, and moves the two ratchets in the same
commit.

## Prerequisites

- [ ] Decide what a parenless call to an unknown callee does with its
      arguments (the blocker). Do not start before it names a route.

## Context

- paths: `internal/parse/call.go`, `internal/parse/shape.go`,
  `internal/parse/t2_test.go`, `internal/parse/testdata/t1.ratchet`
- commands: `go test ./internal/parse/ -count=1`
- entrypoints: `internal/parse/call.go:160-175` (the `default:` arm),
  `internal/parse/shape.go:149` (`knowsShape`)

## What this fixes

Every parenless call to a sub this parser has not seen declared, WITH ANY
ARGUMENT, currently loses its enclosing statement to Unknown -- not because
the Call node is wrong (§4.8.3 is satisfied: `Call{Resolved:false}` is
emitted) but because no arguments are consumed and the leftover tokens sink
the statement.

Measured at b6e7c351: ~330 of T2's 649 Unknown nodes, including all 179 in
`opbasic/arith.t` and the ~151 across `ok`/`is`/`print` sites. This is the
single largest item in the T2 gap.

## Steps

- [ ] Write a failing test asserting `ok 8;` parses with no Unknown node,
      alongside `ok $x`, `ok "s"`, `ok 8,'name'` and `f 13 % 4`, in a file
      that declares or imports the callee the way the chosen route requires
- [ ] Run it and watch it fail for the right reason -- the statement is
      Unknown, not the Call (`go test ./internal/parse/ -run '^TestParenlessCallTakesArguments$' -count=1 -v`)
- [ ] Implement the chosen route, minimally
- [ ] Run the test to green
- [ ] Run the guard the blocker's route-A risk names: a bareword hash key
      (`$h{foo}`), a class name (`Foo->new(1)`) and a fat comma LHS
      (`foo => 1`) each still parse as they do today, with no Call swallowing
      the rest of the statement
- [ ] Measure the T1 ratchet delta in NODES and in BYTES, and the T2 per-file
      movement, before regenerating anything
- [ ] Regenerate `internal/parse/testdata/t1.ratchet` and update the
      `shortfall` map in `internal/parse/t2_test.go` in the SAME commit, with
      the byte measurement in the commit message
- [ ] Run the whole package: `go test ./internal/parse/ -count=1`
- [ ] Commit

## Acceptance Criteria

### Positive Scenarios

- [ ] A parenless call to an unknown-or-declared callee consumes its
      arguments and the statement is not Unknown (`go test ./internal/parse/ -run '^TestParenlessCallTakesArguments$' -count=1 -v | grep -q '^--- PASS: TestParenlessCallTakesArguments'`)
- [ ] A bareword that is NOT a call -- hash key, class name, fat comma LHS --
      still parses unchanged (`go test ./internal/parse/ -run '^TestBarewordIsNotAlwaysACall$' -count=1 -v | grep -q '^--- PASS: TestBarewordIsNotAlwaysACall'`)
- [ ] The T2 shortfall map is regenerated in the same commit and the metric is
      green (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)
- [ ] The T1 ratchet is regenerated in the same commit and is green (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet'`)
- [ ] The whole parse package is green (`go test ./internal/parse/ -count=1`)

Spec: §4.8.3. Plan: M1. Survey: 01a0c10b.
