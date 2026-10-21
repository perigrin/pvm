---
title: "Test harness: undecided_declared needs a positive control, and the hand-rolled declarations swaps become OverlayDeclarations"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-21T21:06:05.230953794Z
updated: 2026-10-21T21:06:05.230953794Z
---

crochet:review round 1 S11 and S13, still untracked at round 3 (coverage F5); destination given by 01a16016-481d.

1. `TestPSCCheckUndecidedCallReportsNothing` (test/e2e/psc_check_test.go): its second half checks that undecided_declared.pl reports nothing under the My::Lib overlay, but the file is also silent without the overlay, so the half passes whether or not the overlay's multi is joined. It needs a positive control: a call the overlay's declaration makes report, under the same overlay.
2. Tests swap the package `declarations` variable by hand (save, assign a layeredFS, restore in Cleanup; 21 assignments at a893c4c7) where `parse.OverlayDeclarations` does the same in one call; e.g. internal/parse/issue1959_test.go TestNestedAndQExportFormsRefused.

## Steps
- [ ] Add the positive control to TestPSCCheckUndecidedCallReportsNothing
- [ ] Replace the hand-rolled swaps with OverlayDeclarations

## Acceptance criteria

### Positive Scenarios
- [ ] The undecided-call test has a control that reports (`go test ./test/e2e/ -run '^TestPSCCheckUndecidedCallReportsNothing$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckUndecidedCallReportsNothing'`)

### Negative Scenarios
- [ ] No hand-rolled declarations swap remains (`! grep -rn 'declarations = ' internal --include='*_test.go' | grep -v '= saved' | grep -q .`)
