---
title: "psc check reports nothing for programs perl refuses to compile: my $x = ; and a typed our"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-20T19:35:48.450215751Z
updated: 2026-10-20T19:35:48.450215751Z
---

Found 2026-10-21 while reviewing the core-module-types doc fix. psc check exits 0 with no output on programs perl refuses at compile time, on pu (33ad8d4c) and on feat/core-module-types alike:

- `my $x = ;` -- perl 5.42: "syntax error at ... near "= ;"". psc check: nothing, rc 0.
- `our Hash[Str] %Config;` -- perl 5.42: "No such class Hash ... syntax error". psc check: nothing, rc 0; `psc parse` prints a tree (declaration "our", binary "%", index "["), though TestTypedOurOnlyInDeclarations's refusedAs reports the `:export` form refused.

internal/psc/check_command.go reports a refusal only when its code is in `perlRefuses` (codes perl refused every time in the corpus); everything else is dropped silently. A program perl will not compile, reported clean, is the worst answer a checker can give.

## Steps
- [ ] Measure which refusal codes these two produce and why they are not reported (or why the second is not refused at all)
- [ ] Report them, without adding a false positive over perl 5.42's lib baseline

## Acceptance criteria

### Positive Scenarios
- [ ] psc check reports a bare assignment with no right-hand side (`go test ./internal/psc/ -run '^TestCheckReportsMissingRightHandSide$' -count=1 -v | grep -q '^--- PASS: TestCheckReportsMissingRightHandSide'`)
- [ ] psc check over perl's lib adds no unlisted line (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)

### Negative Scenarios
- [ ] psc check reports a typed our in a program (`go test ./internal/psc/ -run '^TestCheckReportsTypedOurInProgram$' -count=1 -v | grep -q '^--- PASS: TestCheckReportsTypedOurInProgram'`)
